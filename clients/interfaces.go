package clients

import (
	"context"
	"time"

	"zktoro/zktoro-core-go/domain"

	"github.com/docker/docker/api/types"
	"google.golang.org/protobuf/proto"

	"zktoro/clients/docker"
	"zktoro/config"
)

// DockerClient is a client interface for interacting with docker
type DockerClient interface {
	PullImage(ctx context.Context, refStr string) error
	RemoveImage(ctx context.Context, refStr string) error
	EnsurePublicNetwork(ctx context.Context, name string) (string, error)
	EnsureInternalNetwork(ctx context.Context, name string) (string, error)
	AttachNetwork(ctx context.Context, containerID string, networkID string) error
	DetachNetwork(ctx context.Context, containerID string, networkID string) error
	RemoveNetworkByName(ctx context.Context, networkName string) error
	GetContainers(ctx context.Context) (docker.ContainerList, error)
	GetContainersByLabel(ctx context.Context, name, value string) (docker.ContainerList, error)
	GetzktoroServiceContainers(ctx context.Context) (zktoroContainers docker.ContainerList, err error)
	GetContainerByName(ctx context.Context, name string) (*types.Container, error)
	GetContainerByID(ctx context.Context, id string) (*types.Container, error)
	InspectContainer(ctx context.Context, id string) (*types.ContainerJSON, error)
	StartContainerWithID(ctx context.Context, containerID string) error
	StartContainer(ctx context.Context, config docker.ContainerConfig) (*docker.Container, error)
	StopContainer(ctx context.Context, id string) error
	InterruptContainer(ctx context.Context, id string) error
	TerminateContainer(ctx context.Context, id string) error
	ShutdownContainer(ctx context.Context, id string, timeout *time.Duration) error
	RemoveContainer(ctx context.Context, containerID string) error
	WaitContainerExit(ctx context.Context, id string) error
	WaitContainerStart(ctx context.Context, id string) error
	Prune(ctx context.Context) error
	WaitContainerPrune(ctx context.Context, id string) error
	Nuke(ctx context.Context) error
	HasLocalImage(ctx context.Context, ref string) (bool, error)
	EnsureLocalImage(ctx context.Context, name, ref string) error
	EnsureLocalImages(ctx context.Context, timeoutPerPull time.Duration, imagePulls []docker.ImagePull) []error
	GetContainerLogs(ctx context.Context, containerID, tail string, truncate int) (string, error)
	GetContainerFromRemoteAddr(ctx context.Context, hostPort string) (*types.Container, error)
	SetImagePullCooldown(threshold int, cooldownDuration time.Duration)
}

// MessageClient receives and publishes messages.
type MessageClient interface {
	Subscribe(subject string, handler interface{})
	Publish(subject string, payload interface{})
	PublishProto(subject string, payload proto.Message)
}

// AlertAPIClient calls an http api on the analyzer to store alerts
type AlertAPIClient interface {
	PostBatch(batch *domain.AlertBatchRequest, token string) (*domain.AlertBatchResponse, error)
}

type IPAuthenticator interface {
	Authenticate(ctx context.Context, hostPort string) error
	FindAgentFromRemoteAddr(hostPort string) (*config.AgentConfig, error)
	FindContainerNameFromRemoteAddr(ctx context.Context, hostPort string) (string, error)
	FindAgentByContainerName(containerName string) (*config.AgentConfig, error)
}
