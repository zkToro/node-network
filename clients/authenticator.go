package clients

import (
	"context"
	"fmt"
	"sync"

	"zktoro-node/clients/docker"
	"zktoro-node/clients/messaging"
	"zktoro-node/config"
)

// IPAuthenticator makes sure ip is an assigned bot or a managed container
type ipAuthenticator struct {
	ctx          context.Context
	dockerClient DockerClient
	msgClient    MessageClient

	agentConfigs  []config.AgentConfig
	agentConfigMu sync.RWMutex
}

func (p *ipAuthenticator) Authenticate(ctx context.Context, hostPort string) error {
	name, err := p.FindContainerNameFromRemoteAddr(ctx, hostPort)
	if err != nil {
		return err
	}

	return p.AuthenticateByContainerName(name)
}

func (p *ipAuthenticator) FindAgentFromRemoteAddr(hostPort string) (*config.AgentConfig, error) {
	agentContainer, err := p.dockerClient.GetContainerFromRemoteAddr(p.ctx, hostPort)
	if err != nil {
		return nil, err
	}

	containerName := agentContainer.Names[0][1:]

	return p.FindAgentByContainerName(containerName)
}

func (p *ipAuthenticator) FindAgentByContainerName(containerName string) (*config.AgentConfig, error) {
	p.agentConfigMu.RLock()
	defer p.agentConfigMu.RUnlock()

	for _, agentConfig := range p.agentConfigs {
		if agentConfig.ContainerName() == containerName {
			return &agentConfig, nil
		}
	}

	return nil, fmt.Errorf("bot container not found")
}

func (p *ipAuthenticator) AuthenticateByContainerName(containerName string) error {
	// check for zktoro managed containers
	managedContainers := []string{
		config.DockerScannerContainerName, config.DockerSupervisorContainerName, config.DockerInspectorContainerName, config.DockerJSONRPCProxyContainerName, config.DockerJWTProviderContainerName,
	}
	for _, managedContainer := range managedContainers {
		if containerName == managedContainer {
			return nil
		}
	}

	// check for bots
	_, err := p.FindAgentByContainerName(containerName)

	return err
}

func (p *ipAuthenticator) FindContainerNameFromRemoteAddr(ctx context.Context, hostPort string) (string, error) {
	agentContainer, err := p.dockerClient.GetContainerFromRemoteAddr(ctx, hostPort)
	if err != nil {
		return "", err
	}

	containerName := agentContainer.Names[0][1:]

	return containerName, nil
}

func (p *ipAuthenticator) handleAgentStatusRunning(payload messaging.AgentPayload) error {
	p.agentConfigMu.Lock()
	p.agentConfigs = payload
	p.agentConfigMu.Unlock()
	return nil
}

func NewBotAuthenticator(ctx context.Context) (IPAuthenticator, error) {
	globalClient, err := docker.NewDockerClient("")
	if err != nil {
		return nil, fmt.Errorf("failed to create the global docker client: %v", err)
	}
	msgClient := messaging.NewClient("bot-auth", fmt.Sprintf("%s:%s", config.DockerNatsContainerName, config.DefaultNatsPort))

	b := &ipAuthenticator{
		ctx:          ctx,
		dockerClient: globalClient,
		msgClient:    msgClient,
	}

	msgClient.Subscribe(messaging.SubjectAgentsStatusRunning, messaging.AgentsHandler(b.handleAgentStatusRunning))

	return b, nil
}
