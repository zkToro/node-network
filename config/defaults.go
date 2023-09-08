package config

const (
	DefaultKeysDirName           = ".keys"
	DefaultCombinerCacheFileName = ".combiner_cache.json"
	DefaultConfigFileName        = "config.yml"
	DefaultWrappedConfigFileName = "wrapped-config.yml"
	DefaultConfigWrapperKey      = "x-zktoro-config"
	DefaultNatsPort              = "4222"
	DefaultContainerPort         = "8089"
	DefaultHealthPort            = "8090"
	DefaultJWTProviderPort       = "8515"
	DefaultStoragePort           = "8525"
	DefaultPublicAPIProxyPort    = "8535"
	DefaultJSONRPCProxyPort      = "8545"
	DefaultzktoroNodeBinaryPath   = "/zktoro-node" // the path for the common binary in the container image
)
