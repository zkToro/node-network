package config

const (
	EnvHostzktoroDir = "HOST_zktoro_DIR" // for retrieving zktoro dir path on the host os
	EnvDevelopment  = "zktoro_DEVELOPMENT"
	EnvReleaseInfo  = "zktoro_RELEASE_INFO"

	// Agent env vars
	EnvJsonRpcHost        = "JSON_RPC_HOST"
	EnvJsonRpcPort        = "JSON_RPC_PORT"
	EnvJWTProviderHost    = "zktoro_JWT_PROVIDER_HOST"
	EnvJWTProviderPort    = "zktoro_JWT_PROVIDER_PORT"
	EnvPublicAPIProxyHost = "zktoro_PUBLIC_API_PROXY_HOST"
	EnvPublicAPIProxyPort = "zktoro_PUBLIC_API_PROXY_PORT"
	EnvAgentGrpcPort      = "AGENT_GRPC_PORT"
	EnvzktoroBotID         = "zktoro_BOT_ID"
	EnvzktoroBotOwner      = "zktoro_BOT_OWNER"
	EnvzktoroChainID       = "zktoro_CHAIN_ID"
)

// EnvDefaults contain default values for one env.
type EnvDefaults struct {
	DiscoSubdomain string
}

// GetEnvDefaults returns the default values for an env.
func GetEnvDefaults(development bool) EnvDefaults {
	if development {
		return EnvDefaults{
			DiscoSubdomain: "disco-dev",
		}
	}
	return EnvDefaults{
		DiscoSubdomain: "disco",
	}
}
