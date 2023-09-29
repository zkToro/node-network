package public_api

import (
	"context"

	"zktoro/config"
	"zktoro/healthutils"
	"zktoro/services"
	public_api "zktoro/services/public-api"

	"zktoro/zktoro-core-go/clients/health"

	"zktoro/zktoro-core-go/utils"
)

func initPublicAPIProxy(ctx context.Context, cfg config.Config) (*public_api.PublicAPIProxy, error) {
	return public_api.NewPublicAPIProxy(ctx, cfg)
}

func initServices(ctx context.Context, cfg config.Config) ([]services.Service, error) {
	// can't dial localhost - need to dial host gateway from container
	cfg.PublicAPIProxy.Url = utils.ConvertToDockerHostURL(cfg.PublicAPIProxy.Url)

	proxy, err := initPublicAPIProxy(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return []services.Service{
		health.NewService(
			ctx, "", healthutils.DefaultHealthServerErrHandler,
			health.CheckerFrom(summarizeReports, proxy),
		),
		proxy,
	}, nil
}

func summarizeReports(reports health.Reports) *health.Report {
	summary := health.NewSummary()

	apiErr, ok := reports.NameContains("service.public-api-proxy.api")
	if ok && len(apiErr.Details) > 0 {
		summary.Addf("last time the api failed with error '%s'.", apiErr.Details)
	}

	return summary.Finish()
}

func Run() {
	services.ContainerMain("public-api", initServices)
}
