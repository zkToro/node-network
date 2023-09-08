package docker

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"
	"time"

	"github.com/goccy/go-json"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"zktoro/zktoro-core-go/utils/workers"
	"zktoro-node/clients/cooldown"
	"zktoro-node/config"
	log "github.com/sirupsen/logrus"
)

const (
	Labelzktoro                          = "network.zktoro"
	LabelzktoroSupervisor                = "network.zktoro.supervisor"
	LabelzktoroSupervisorStrategyVersion = "network.zktoro.supervisor.strategy-version"
	LabelzktoroIsBot                     = "network.zktoro.is-bot"
	LabelzktoroBotID                     = "network.zktoro.bot-id"

	LabelzktoroSettingsAgentLogsEnable = "network.zktoro.settings.agent-logs.enable"
)

type dockerLabel struct {
	Name  string
	Value string
}

var defaultLabels = []dockerLabel{
	{Name: Labelzktoro, Value: "true"},
}

// Client errors
var (
	ErrContainerNotFound = errors.New("container not found")
)

// Container is a resulting container reference, including the ID and configuration
type Container struct {
	Name      string
	ID        string
	ImageHash string
	Config    ContainerConfig
}

// ContainerConfig is configuration for a particular container
type ContainerConfig struct {
	Name            string
	Image           string
	Env             map[string]string
	LinkNetworkIDs  []string
	NetworkID       string
	Ports           map[string]string
	PublishAllPorts bool // auto-publishing ports EXPOSEd in Dockerfile
	Volumes         map[string]string
	Files           map[string][]byte
	MaxLogSize      string
	MaxLogFiles     int
	CPUQuota        int64
	Memory          int64
	Cmd             []string
	DialHost        bool
	Labels          map[string]string
}

// ContainerList contains the full container data.
type ContainerList []types.Container

// FindByID finds the container by the ID.
func (dcl ContainerList) FindByID(id string) (*types.Container, bool) {
	for _, c := range dcl {
		if c.ID == id {
			return &c, true
		}
	}
	return nil, false
}

// FindByName finds the container by the name.
func (dcl ContainerList) FindByName(name string) (*types.Container, bool) {
	for _, c := range dcl {
		for _, n := range c.Names {
			if n == name || n == fmt.Sprintf("/%s", name) {
				return &c, true
			}
		}
	}
	return nil, false
}

// ContainsAny checks is any of the containers contain this name and returns the first one.
func (dcl ContainerList) ContainsAny(name string) (*types.Container, bool) {
	for _, c := range dcl {
		if strings.Contains(c.Names[0], name) {
			return &c, true
		}
	}
	return nil, false
}

type dockerClient struct {
	cli                   *client.Client
	workers               *workers.Group
	username              string
	password              string
	labels                []dockerLabel
	imageDownloadCooldown cooldown.Cooldown
}

func (cfg ContainerConfig) envVars() []string {
	var results []string
	for k, v := range cfg.Env {
		results = append(results, fmt.Sprintf("%s=%s", k, v))
	}
	return results
}

func registryAuthValue(username, password string) string {
	if username == "" && password == "" {
		return ""
	}
	jsonBytes, _ := json.Marshal(map[string]string{
		"username": username,
		"password": password,
	})
	return base64.StdEncoding.EncodeToString(jsonBytes)
}

// PullImage pulls an image using the given ref.
func (d *dockerClient) PullImage(ctx context.Context, refStr string) error {
	if d.imageDownloadCooldown != nil && d.imageDownloadCooldown.ShouldCoolDown(refStr) {
		return fmt.Errorf("too many pull attempts - cooling down: %s", refStr)
	}

	r, err := d.cli.ImagePull(ctx, refStr, types.ImagePullOptions{
		RegistryAuth: registryAuthValue(d.username, d.password),
	})
	if err != nil {
		return err
	}
	defer r.Close()
	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	respStr := strings.ToLower(string(b))
	if strings.Contains(respStr, "downloaded") || strings.Contains(respStr, "up to date") {
		return nil
	}
	return fmt.Errorf("unexpected image pull response: %s", string(b))
}

func (d *dockerClient) Prune(ctx context.Context) error {
	filter := d.labelFilter()
	res, err := d.cli.NetworksPrune(ctx, filter)
	if err != nil {
		return err
	}
	for _, nw := range res.NetworksDeleted {
		log.Infof("pruned network %s", nw)
	}

	cpRes, err := d.cli.ContainersPrune(ctx, filter)
	if err != nil {
		return err
	}
	for _, cp := range cpRes.ContainersDeleted {
		log.Infof("pruned container %s", cp)
	}

	return nil
}

// RemoveImage removes an image.
func (d *dockerClient) RemoveImage(ctx context.Context, refStr string) error {
	filter := filters.NewArgs()
	filter.Add("ancestor", refStr)
	containers, err := d.cli.ContainerList(context.Background(), types.ContainerListOptions{
		All:     true,
		Filters: filter,
		Limit:   1,
	})
	if err != nil {
		return fmt.Errorf("failed to get the container list: %v", err)
	}

	// avoid removing used images
	if len(containers) > 0 {
		return nil
	}

	_, err = d.cli.ImageRemove(ctx, refStr, types.ImageRemoveOptions{})
	if err != nil && strings.Contains(strings.ToLower(err.Error()), "no such image") {
		return nil
	}
	return err
}

func (d *dockerClient) EnsurePublicNetwork(ctx context.Context, name string) (string, error) {
	return d.createNetwork(ctx, name, false)
}

func (d *dockerClient) EnsureInternalNetwork(ctx context.Context, name string) (string, error) {
	return d.createNetwork(ctx, name, true)
}

func (d *dockerClient) createNetwork(ctx context.Context, name string, internal bool) (string, error) {
	// Reuse if network exists.
	networks, err := d.cli.NetworkList(ctx, types.NetworkListOptions{})
	if err != nil {
		return "", err
	}
	for _, network := range networks {
		if network.Name == name {
			return network.ID, nil
		}
	}

	resp, err := d.cli.NetworkCreate(ctx, name, types.NetworkCreate{
		Labels:   labelsToMap(d.labels),
		Internal: internal,
	})
	if err != nil {
		return "", err
	}
	return resp.ID, nil
}

func (d *dockerClient) RemoveNetworkByName(ctx context.Context, networkName string) error {
	networks, err := d.cli.NetworkList(ctx, types.NetworkListOptions{
		Filters: filters.NewArgs(filters.KeyValuePair{Key: "name", Value: networkName}),
	})
	if err != nil {
		return err
	}
	if len(networks) == 0 {
		return nil
	}
	return d.cli.NetworkRemove(ctx, networks[0].ID)
}

func (d *dockerClient) AttachNetwork(ctx context.Context, containerID string, networkID string) error {
	err := d.cli.NetworkConnect(ctx, networkID, containerID, nil)
	if err == nil {
		return nil
	}
	if strings.Contains(err.Error(), "already exists") {
		return nil
	}
	return err
}

func (d *dockerClient) DetachNetwork(ctx context.Context, containerID string, networkID string) error {
	err := d.cli.NetworkDisconnect(ctx, networkID, containerID, true)
	if err == nil {
		return nil
	}

	if strings.Contains(err.Error(), "is not connected") {
		return nil
	}
	return err
}

func withTcp(port string) string {
	return fmt.Sprintf("%s/tcp", port)
}

// copyFile copies content bytes into container at given file path.
func copyFile(cli *client.Client, ctx context.Context, filePath string, content []byte, containerId string) error {
	if len(filePath) == 0 {
		return errors.New("zero length file path")
	}
	if filePath[0] != '/' {
		filePath = "/" + filePath
	}
	dir, file := path.Split(filePath)

	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	err := tw.WriteHeader(&tar.Header{
		Name: file,
		Mode: 0666,
		Size: int64(len(content)),
	})
	if err != nil {
		return err
	}
	_, err = tw.Write(content)
	if err != nil {
		return err
	}
	err = tw.Close()
	if err != nil {
		return err
	}
	return cli.CopyToContainer(ctx, containerId, dir, &buf, types.CopyToContainerOptions{})
}

// GetContainers returns all of the containers.
func (d *dockerClient) GetContainers(ctx context.Context) (ContainerList, error) {
	return d.cli.ContainerList(ctx, types.ContainerListOptions{
		All:     true,
		Filters: d.labelFilter(),
	})
}

// GetContainersByLabel returns all of the containers that has the label.
func (d *dockerClient) GetContainersByLabel(ctx context.Context, name, value string) (ContainerList, error) {
	return d.cli.ContainerList(ctx, types.ContainerListOptions{
		All: true,
		Filters: makeLabelFilter([]dockerLabel{
			{Name: name, Value: value},
		}),
	})
}

// GetzktoroServiceContainers returns all of the non-agent zktoro containers.
func (d *dockerClient) GetzktoroServiceContainers(ctx context.Context) (zktoroContainers ContainerList, err error) {
	containers, err := d.cli.ContainerList(ctx, types.ContainerListOptions{
		All:     true,
		Filters: d.labelFilter(),
	})
	for _, container := range containers {
		if !strings.Contains(container.Names[0][1:], "zktoro-agent") {
			zktoroContainers = append(zktoroContainers, container)
		}
	}
	return
}

// GetContainerByName gets a container by using a name lookup over all containers.
func (d *dockerClient) GetContainerByName(ctx context.Context, name string) (*types.Container, error) {
	containers, err := d.GetContainers(ctx)
	if err != nil {
		return nil, err
	}
	for _, c := range containers {
		if c.Names[0][1:] == name {
			return &c, nil
		}
	}
	return nil, fmt.Errorf("%w with name '%s'", ErrContainerNotFound, name)
}

// GetContainerByName gets a container by using an ID lookup over all containers.
func (d *dockerClient) GetContainerByID(ctx context.Context, id string) (*types.Container, error) {
	containers, err := d.GetContainers(ctx)
	if err != nil {
		return nil, err
	}
	for _, c := range containers {
		if c.ID == id {
			return &c, nil
		}
	}
	return nil, fmt.Errorf("%w with id '%s'", ErrContainerNotFound, id)
}

// InspectContainer returns container details.
func (d *dockerClient) InspectContainer(ctx context.Context, id string) (*types.ContainerJSON, error) {
	info, err := d.cli.ContainerInspect(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get container details: %v", err)
	}
	return &info, nil
}

// Nuke makes sure that all running zktoro containers are stopped and pruned, quickly enough.
func (d *dockerClient) Nuke(ctx context.Context) error {
	var err error
	for i := 0; i < 4; i++ {
		err = d.nuke(ctx)
		if err == nil {
			return nil
		}
		log.WithError(err).Error("failed to nuke - retrying")
	}
	return fmt.Errorf("all nuke retries failed: %v", err)
}

func (d *dockerClient) nuke(ctx context.Context) error {
	// step 1: put the supervisor to the top of the list so it doesn't do funny restarts
	containers, err := d.GetContainers(ctx)
	if err != nil {
		return fmt.Errorf("failed to get zktoro containers list: %v", err)
	}
	supervisorContainer, err := d.GetContainerByName(ctx, config.DockerSupervisorContainerName)
	if err == nil {
		containers = append([]types.Container{*supervisorContainer}, containers...)
	}
	if err != nil && !errors.Is(err, ErrContainerNotFound) {
		return fmt.Errorf("unexpected error while getting supervisor container: %v", err)
	}

	// step 2: stop all and wait until each exit
	for _, container := range containers {
		if err := d.StopContainer(ctx, container.ID); err != nil {
			return fmt.Errorf("failed to stop: %v", err)
		}
		if err := d.WaitContainerExit(ctx, container.ID); err != nil {
			return err
		}
	}

	// step 3: prune everything
	if err := d.Prune(ctx); err != nil {
		return fmt.Errorf("failed to prune: %v", err)
	}

	// step 4: ensure that the containers are really pruned
	for _, container := range containers {
		if err := d.WaitContainerPrune(ctx, container.ID); err != nil {
			return err
		}
	}

	return nil
}

// GetContainerName returns the container name.
func GetContainerName(container types.Container) string {
	return container.Names[0][1:]
}

// StartContainerWithID starts an existing container.
func (d *dockerClient) StartContainerWithID(ctx context.Context, containerID string) error {
	return d.cli.ContainerStart(ctx, containerID, types.ContainerStartOptions{})
}

// StartContainer kicks off a container as a daemon and returns a summary of the container
func (d *dockerClient) StartContainer(ctx context.Context, config ContainerConfig) (*Container, error) {
	log.WithFields(log.Fields{
		"image": config.Image,
		"name":  config.Name,
	}).Info("StartContainer()")
	containers, err := d.GetContainers(ctx)
	if err != nil {
		return nil, err
	}
	// If we already have the container but it is not running, then just start it.
	var foundContainer *types.Container
	for _, c := range containers {
		if len(c.Names) == 0 {
			continue
		}
		foundName := GetContainerName(c) // remove / in the beginning
		if foundName == config.Name {
			foundContainer = &c
			break
		}
	}
	if foundContainer != nil {
		if err := d.cli.ContainerStart(ctx, foundContainer.ID, types.ContainerStartOptions{}); err != nil {
			return nil, err
		}
		inspection, err := d.cli.ContainerInspect(ctx, foundContainer.ID)
		if err != nil {
			return nil, err
		}
		log.WithFields(log.Fields{
			"id":   foundContainer.ID,
			"name": config.Name,
		}).Info("container is starting")
		return &Container{Name: config.Name, ID: foundContainer.ID, Config: config, ImageHash: inspection.Image}, nil
	}

	bindings := make(map[nat.Port][]nat.PortBinding)
	ps := make(nat.PortSet)
	for hp, cp := range config.Ports {
		hostIP := "0.0.0.0"
		parts := strings.Split(hp, ":")
		if len(parts) == 2 {
			hostIP = parts[0]
			hp = parts[1]
		}
		contPort := nat.Port(withTcp(cp))
		ps[contPort] = struct{}{}
		bindings[contPort] = []nat.PortBinding{{
			HostPort: hp,
			HostIP:   hostIP,
		}}
	}

	var volumes []string
	for hostVol, containerMnt := range config.Volumes {
		volumes = append(volumes, fmt.Sprintf("%s:%s", hostVol, containerMnt))
	}

	maxLogSize := config.MaxLogSize
	if maxLogSize == "" {
		maxLogSize = "10m"
	}

	maxLogFiles := config.MaxLogFiles
	if maxLogFiles == 0 {
		maxLogFiles = 10
	}

	cntCfg := &container.Config{
		Image:  config.Image,
		Env:    config.envVars(),
		Labels: labelsToMap(d.labels),
	}
	// add custom labels
	for k, v := range config.Labels {
		cntCfg.Labels[k] = v
	}

	if len(config.Cmd) > 0 {
		cntCfg.Cmd = config.Cmd
	}

	hostCfg := &container.HostConfig{
		NetworkMode:     container.NetworkMode(config.NetworkID),
		PortBindings:    bindings,
		PublishAllPorts: config.PublishAllPorts,
		Binds:           volumes,
		LogConfig: container.LogConfig{
			Config: map[string]string{
				"max-file": fmt.Sprintf("%d", maxLogFiles),
				"max-size": maxLogSize,
			},
			Type: "json-file",
		},
		Resources: container.Resources{
			CPUQuota: config.CPUQuota,
			Memory:   config.Memory,
		},
	}

	if config.DialHost {
		hostCfg.ExtraHosts = append(hostCfg.ExtraHosts, "host.docker.internal:host-gateway")
	}

	cont, err := d.cli.ContainerCreate(
		ctx, cntCfg, hostCfg, nil, nil, config.Name,
	)

	if err != nil {
		return nil, err
	}

	for fn, b := range config.Files {
		if err := copyFile(d.cli, ctx, fn, b, cont.ID); err != nil {
			return nil, err
		}
	}

	if err := d.cli.ContainerStart(ctx, cont.ID, types.ContainerStartOptions{}); err != nil {
		return nil, err
	}

	for _, nwID := range config.LinkNetworkIDs {
		if err := d.AttachNetwork(ctx, cont.ID, nwID); err != nil {
			log.Error("error attaching network", err)
			return nil, err
		}
	}

	inspection, err := d.cli.ContainerInspect(ctx, cont.ID)
	if err != nil {
		return nil, err
	}

	log.WithFields(log.Fields{
		"id":   cont.ID,
		"name": config.Name,
	}).Info("container is starting")
	return &Container{Name: config.Name, ID: cont.ID, Config: config, ImageHash: inspection.Image}, nil
}

// StopContainer kills a container by ID
func (d *dockerClient) StopContainer(ctx context.Context, id string) error {
	return d.stopContainer(ctx, id, "SIGKILL")
}

// InterruptContainer stops a container by sending an interrupt signal.
func (d *dockerClient) InterruptContainer(ctx context.Context, id string) error {
	return d.stopContainer(ctx, id, "SIGINT")
}

// TerminateContainer stops a container by sending an termination signal.
func (d *dockerClient) TerminateContainer(ctx context.Context, id string) error {
	return d.stopContainer(ctx, id, "SIGTERM")
}

// ShutdownContainer stops a container by sending a termination signal and waits until either container exits or context cancels.
func (d *dockerClient) ShutdownContainer(ctx context.Context, id string, timeout *time.Duration) error {
	return d.cli.ContainerStop(ctx, id, timeout)
}

// TerminateContainer stops a container by sending an termination signal.
func (d *dockerClient) stopContainer(ctx context.Context, containerID, signal string) error {
	log.WithFields(
		log.Fields{
			"id":     containerID,
			"signal": signal,
		},
	).Infof("stopping container")

	err := d.cli.ContainerKill(ctx, containerID, signal)
	if err == nil {
		return nil
	}
	if isNoSuchContainerErr(err) || isNotRunningErr(err) {
		return nil
	}
	return err
}

// RemoveContainer kills and a container by ID.
func (d *dockerClient) RemoveContainer(ctx context.Context, containerID string) error {
	return d.cli.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{
		Force: true,
	})
}

func isNoSuchContainerErr(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "no such container")
}

func isNotRunningErr(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "is not running")
}

// WaitContainerExit waits for container exit by checking periodically.
func (d *dockerClient) WaitContainerExit(ctx context.Context, id string) error {
	ticker := time.NewTicker(500 * time.Millisecond)
	logger := log.WithFields(log.Fields{
		"id": id,
	})

	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	for {
		logger.Info("waiting for container exit")
		c, err := d.GetContainerByID(ctx, id)
		if err != nil && errors.Is(err, ErrContainerNotFound) {
			logger.Info("no need to wait for container exit - not found")
			return nil
		}
		if err != nil {
			logger.WithError(err).Error("failed while waiting for container exit")
			return err
		}
		if c.State == "exited" || c.State == "created" {
			return nil
		}
		logger.WithField("containerState", c.State).Info("still waiting for exit")
		<-ticker.C
	}
}

// WaitContainerStart waits for container start by checking periodically.
func (d *dockerClient) WaitContainerStart(ctx context.Context, id string) error {
	ticker := time.NewTicker(time.Second)
	start := time.Now()
	logger := log.WithFields(log.Fields{
		"id": id,
	})

	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	for t := range ticker.C {
		logger.Info("waiting for container start")
		c, err := d.GetContainerByID(ctx, id)
		if err == nil && c != nil && c.State == "running" {
			logger.Info("container started")
			return nil
		}
		if err != nil {
			return err
		}
		// if the conditions are not met within 30 seconds, it's a failure
		if t.After(start.Add(time.Second * 30)) {
			return errors.New("container did not start")
		}
	}
	return nil
}

// WaitContainerPrune waits for container prune by checking periodically.
func (d *dockerClient) WaitContainerPrune(ctx context.Context, id string) error {
	ticker := time.NewTicker(500 * time.Millisecond)
	logger := log.WithFields(log.Fields{
		"id": id,
	})

	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	for {
		logger.Infof("waiting for container prune")
		c, err := d.GetContainerByID(ctx, id)
		if err != nil && errors.Is(err, ErrContainerNotFound) {
			return nil
		}
		if err != nil {
			logger.WithError(err).Error("error while waiting for prune")
			return err
		}
		logger.WithField("containerState", c.State).Info("container state while waiting for prune")
		if !(c.State == "exited" || c.State == "dead") {
			err = fmt.Errorf("cannot prune container with status '%s' - container needs to stop first", c.State)
			logger.WithError(err).Error("error while waiting for prune")
			return err
		}
		<-ticker.C
	}
}

// HasLocalImage checks if we have an image locally.
func (d *dockerClient) HasLocalImage(ctx context.Context, ref string) (bool, error) {
	_, _, err := d.cli.ImageInspectWithRaw(ctx, ref)
	if client.IsErrNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// EnsureLocalImage ensures that we have the image locally.
func (d *dockerClient) EnsureLocalImage(ctx context.Context, name, ref string) error {
	logger := log.WithFields(log.Fields{
		"image": ref,
		"name":  name,
	})
	logger.Info("ensuring local image")
	imageExists, imgErr := d.HasLocalImage(ctx, ref)
	if imgErr != nil {
		return fmt.Errorf("error checking local: %s", imgErr.Error())
	}
	if imageExists {
		log.Infof("found local image for '%s': %s", name, ref)
		return nil
	}

	startTime := time.Now()
	if err := d.PullImage(ctx, ref); err != nil {
		logger.WithError(err).Error("error pulling image")
		return fmt.Errorf("pull error (duration=%s) %s: %v", time.Since(startTime).String(), ref, err.Error())
	}

	log.Infof("pulled '%s' image: %s", name, ref)
	return nil
}

// ImagePull data about an image to pull.
type ImagePull struct {
	Name string
	Ref  string
}

// EnsureLocalImages pulls an image using the given ref asynchronously.
func (d *dockerClient) EnsureLocalImages(ctx context.Context, timeoutPerPull time.Duration, imagePulls []ImagePull) (errs []error) {
	var outputs []*workers.Output
	for _, imagePull := range imagePulls {
		outputs = append(outputs, d.workers.Execute(func() ([]interface{}, error) {
			ctx, cancel := context.WithTimeout(ctx, timeoutPerPull)
			defer cancel()
			return nil, d.EnsureLocalImage(ctx, imagePull.Name, imagePull.Ref)
		}))
	}
	for _, output := range outputs {
		errs = append(errs, output.Error)
	}
	return
}

// GetContainerLogs gets the container logs.
func (d *dockerClient) GetContainerLogs(ctx context.Context, containerID, tail string, truncate int) (string, error) {
	r, err := d.cli.ContainerLogs(ctx, containerID, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Timestamps: true,
		Tail:       tail,
	})
	if err != nil {
		return "", err
	}
	b, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	if truncate >= 0 && len(b) > truncate {
		b = b[:truncate]
	}
	// remove strange 8-byte prefix in each line
	lines := strings.Split(string(b), "\n")
	for i, line := range lines {
		if len(line) == 0 {
			continue
		}
		prefixEnd := strings.Index(line, "2") // timestamp beginning
		if prefixEnd < 0 || prefixEnd > len(line) {
			continue
		}
		lines[i] = line[prefixEnd:]
	}
	return strings.Join(lines, "\n"), nil
}

func (d *dockerClient) labelFilter() filters.Args {
	return makeLabelFilter(d.labels)
}

func makeLabelFilter(labels []dockerLabel) filters.Args {
	filter := filters.NewArgs()
	for _, label := range labels {
		filter.Add("label", fmt.Sprintf("%s=%s", label.Name, label.Value))
	}
	return filter
}

func (d *dockerClient) GetContainerFromRemoteAddr(ctx context.Context, hostPort string) (*types.Container, error) {
	containers, err := d.GetContainers(ctx)
	if err != nil {
		return nil, err
	}
	ipAddr := strings.Split(hostPort, ":")[0]

	var agentContainer *types.Container
	for _, container := range containers {
		for _, network := range container.NetworkSettings.Networks {
			if network.IPAddress == ipAddr {
				agentContainer = &container
				break
			}
		}
		if agentContainer != nil {
			break
		}
	}

	if agentContainer == nil {
		log.WithField("sourceIp", ipAddr).Warn("not a known bot")
		return nil, fmt.Errorf("could not found agent container from ip address: %s", hostPort)
	}

	return agentContainer, nil
}

func initLabels(name string) []dockerLabel {
	if len(name) == 0 {
		return defaultLabels
	}

	return append(
		defaultLabels, dockerLabel{
			Name:  LabelzktoroSupervisor,
			Value: name,
		},
	)
}

func labelsToMap(labels []dockerLabel) map[string]string {
	m := make(map[string]string)
	for _, label := range labels {
		m[label.Name] = label.Value
	}
	return m
}

// SetImagePullCooldown sets the image pull cooldown.
func (d *dockerClient) SetImagePullCooldown(threshold int, cooldownDuration time.Duration) {
	d.imageDownloadCooldown = cooldown.New(threshold, cooldownDuration)
}

// NewDockerClient creates a new docker client
func NewDockerClient(name string) (*dockerClient, error) {
	cli, err := client.NewClientWithOpts()
	if err != nil {
		return nil, err
	}
	return &dockerClient{
		cli:     cli,
		workers: workers.New(1),
		labels:  initLabels(name),
	}, nil
}

// NewAuthDockerClient creates a new docker client with credentials
func NewAuthDockerClient(name string, username, password string) (*dockerClient, error) {
	if len(username) == 0 && len(password) == 0 {
		return NewDockerClient(name)
	}
	cli, err := client.NewClientWithOpts()
	if err != nil {
		return nil, err
	}
	return &dockerClient{
		cli:      cli,
		workers:  workers.New(1),
		username: username,
		password: password,
		labels:   initLabels(name),
	}, nil
}
