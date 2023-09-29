package config

// BotResourceLimits contain the agent resource limits data.
type BotResourceLimits struct {
	CPUQuota int64 // in microseconds
	Memory   int64 // in bytes
}

// GetAgentResourceLimits calculates and returns the resource limits by
// taking the configuration into account. Zero values mean no limits.
func GetAgentResourceLimits(resourcesCfg ResourcesConfig) *BotResourceLimits {
	var limits BotResourceLimits

	if resourcesCfg.DisableAgentLimits {
		return &limits
	}

	limits.CPUQuota = getDefaultCPUQuotaPerAgent()
	if resourcesCfg.AgentMaxCPUs > 0 {
		limits.CPUQuota = CPUsToMicroseconds(resourcesCfg.AgentMaxCPUs)
	}

	limits.Memory = getDefaultMemoryPerAgent()
	if resourcesCfg.AgentMaxMemoryMiB > 0 {
		limits.Memory = MiBToBytes(resourcesCfg.AgentMaxMemoryMiB)
	}

	return &limits
}

// CPUsToMicroseconds converts given CPU amount to microseconds.
func CPUsToMicroseconds(cpus float64) int64 {
	return int64(cpus * float64(100000))
}

// MiBToBytes converts given MiB amount to bytes.
func MiBToBytes(mib int) int64 {
	return int64(mib * 1048580)
}

// getDefaultCPUQuotaPerAgent returns the default CFS microseconds value allowed per agent
func getDefaultCPUQuotaPerAgent() int64 {
	return CPUsToMicroseconds(0.2)
}

// getDefaultMemoryPerAgent returns the constant default memory allowed per agent.
func getDefaultMemoryPerAgent() int64 {
	return MiBToBytes(1e3)
}
