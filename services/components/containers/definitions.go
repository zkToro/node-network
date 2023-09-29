package containers

import (
	"fmt"

	"zktoro/clients/docker"
	"zktoro/config"
)

// Label values
const (
	// LabelzktoroSupervisor ensures that our docker client in the supervisor only touches
	// the containers managed by the supervisor service.
	LabelzktoroSupervisor = "supervisor"
	LabelValuezktoroIsBot = "true"
	// LabelValueStrategyVersion is for versioning the critical changes in container management strategy.
	// It's effective in deciding if a bot container should be re-created or not.
	LabelValueStrategyVersion = "2023-06-16T15:00:00Z"
)

// Limits define container limits.
type Limits struct {
	config.LogConfig
	config.BotResourceLimits
}

// NewBotContainerConfig creates a new bot container config.
func NewBotContainerConfig(
	networkID string, botConfig config.AgentConfig,
	logConfig config.LogConfig, resourcesConfig config.ResourcesConfig,
) docker.ContainerConfig {
	limits := config.GetAgentResourceLimits(resourcesConfig)

	return docker.ContainerConfig{
		Name:           botConfig.ContainerName(),
		Image:          botConfig.Image,
		NetworkID:      networkID,
		LinkNetworkIDs: []string{},
		Env: map[string]string{
			config.EnvJsonRpcHost:        config.DockerJSONRPCProxyContainerName,
			config.EnvJsonRpcPort:        config.DefaultJSONRPCProxyPort,
			config.EnvJWTProviderHost:    config.DockerJWTProviderContainerName,
			config.EnvJWTProviderPort:    config.DefaultJWTProviderPort,
			config.EnvPublicAPIProxyHost: config.DockerPublicAPIProxyContainerName,
			config.EnvPublicAPIProxyPort: config.DefaultPublicAPIProxyPort,
			config.EnvAgentGrpcPort:      botConfig.GrpcPort(),
			config.EnvzktoroBotID:        botConfig.ID,
			config.EnvzktoroBotOwner:     botConfig.Owner,
			config.EnvzktoroChainID:      fmt.Sprintf("%d", botConfig.ChainID),
		},
		MaxLogFiles: logConfig.MaxLogFiles,
		MaxLogSize:  logConfig.MaxLogSize,
		CPUQuota:    limits.CPUQuota,
		Memory:      limits.Memory,
		Labels: map[string]string{
			docker.LabelzktoroIsBot:                     LabelValuezktoroIsBot,
			docker.LabelzktoroSupervisorStrategyVersion: LabelValueStrategyVersion,
			docker.LabelzktoroBotID:                     botConfig.ID,
		},
	}
}
