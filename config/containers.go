package config

import (
	"fmt"
	"path"
)

const ContainerNamePrefix = "zktoro"

// Docker container names
var (
	DockerSupervisorImage = "zktoro-network/zktoro-node:latest"
	DockerUpdaterImage    = "zktoro-network/zktoro-node:latest"
	UseDockerImages       = "local"

	DockerSupervisorManagedContainers = 6
	DockerUpdaterContainerName        = fmt.Sprintf("%s-updater", ContainerNamePrefix)
	DockerSupervisorContainerName     = fmt.Sprintf("%s-supervisor", ContainerNamePrefix)
	DockerNatsContainerName           = fmt.Sprintf("%s-nats", ContainerNamePrefix)
	DockerIpfsContainerName           = fmt.Sprintf("%s-ipfs", ContainerNamePrefix)
	DockerScannerContainerName        = fmt.Sprintf("%s-scanner", ContainerNamePrefix)
	DockerInspectorContainerName      = fmt.Sprintf("%s-inspector", ContainerNamePrefix)
	DockerJSONRPCProxyContainerName   = fmt.Sprintf("%s-json-rpc", ContainerNamePrefix)
	DockerPublicAPIProxyContainerName = fmt.Sprintf("%s-public-api", ContainerNamePrefix)
	DockerJWTProviderContainerName    = fmt.Sprintf("%s-jwt-provider", ContainerNamePrefix)
	DockerStorageContainerName        = fmt.Sprintf("%s-storage", ContainerNamePrefix)

	DockerNetworkName = DockerScannerContainerName

	DefaultContainerzktoroDirPath      = "/.zktoro"
	DefaultContainerConfigPath        = path.Join(DefaultContainerzktoroDirPath, DefaultConfigFileName)
	DefaultContainerWrappedConfigPath = path.Join(DefaultContainerzktoroDirPath, DefaultWrappedConfigFileName)
	DefaultContainerKeyDirPath        = path.Join(DefaultContainerzktoroDirPath, DefaultKeysDirName)
)
