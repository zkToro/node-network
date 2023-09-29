package supervisor

import (
	"context"
	"fmt"
	"strconv"

	"zktoro/config"
	"zktoro/healthutils"
	"zktoro/services"
	"zktoro/services/components"
	"zktoro/services/components/registry"
	"zktoro/services/supervisor"

	"zktoro/zktoro-core-go/clients/health"
	"zktoro/zktoro-core-go/security"
	"zktoro/zktoro-core-go/utils"
)

func initServices(ctx context.Context, cfg config.Config) ([]services.Service, error) {
	cfg.Registry.JsonRpc.Url = utils.ConvertToDockerHostURL(cfg.Registry.JsonRpc.Url)
	cfg.Registry.IPFS.APIURL = utils.ConvertToDockerHostURL(cfg.Registry.IPFS.APIURL)
	cfg.Registry.IPFS.GatewayURL = utils.ConvertToDockerHostURL(cfg.Registry.IPFS.GatewayURL)
	cfg.AgentLogsConfig.URL = utils.ConvertToDockerHostURL(cfg.AgentLogsConfig.URL)

	passphrase, err := security.ReadPassphrase()
	if err != nil {
		return nil, err
	}
	key, err := security.LoadKey(config.DefaultContainerKeyDirPath)
	if err != nil {
		return nil, err
	}
	botRegistry, err := registry.New(cfg, key.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to create the bot registry: %v", err)
	}
	botLifecycleConfig := components.BotLifecycleConfig{
		Config:         cfg,
		ScannerAddress: key.Address,
		BotRegistry:    botRegistry,
	}
	svc, err := supervisor.NewSupervisorService(ctx, supervisor.SupervisorServiceConfig{
		Config:             cfg,
		Passphrase:         passphrase,
		Key:                key,
		BotLifecycleConfig: botLifecycleConfig,
	})
	if err != nil {
		return nil, err
	}
	return []services.Service{
		health.NewService(
			ctx, "", healthutils.DefaultHealthServerErrHandler,
			health.CheckerFrom(summarizeReports, svc, botRegistry),
		),
		svc,
	}, nil
}

func summarizeReports(reports health.Reports) *health.Report {
	summary := health.NewSummary()

	containersManager, ok := reports.NameContains("containers.managed")
	if ok {
		count, _ := strconv.Atoi(containersManager.Details)
		if count < config.DockerSupervisorManagedContainers {
			summary.Addf("missing %d containers.", config.DockerSupervisorManagedContainers-count)
			summary.Status(health.StatusFailing)
		} else {
			summary.Addf("all %d service containers are running.", config.DockerSupervisorManagedContainers)
		}
	}

	telemetryErr, ok := reports.NameContains("telemetry-sync.error")
	if ok && len(telemetryErr.Details) > 0 {
		summary.Addf("telemetry sync is failing with error '%s' (non-critical).", telemetryErr.Details)
		// do not change status - non critical
	}

	return summary.Finish()
}

func Run() {
	services.ContainerMain("supervisor", initServices)
}
