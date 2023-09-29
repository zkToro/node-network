package json_rpc

import (
	"context"

	"zktoro/config"
	"zktoro/healthutils"
	"zktoro/services"
	jrp "zktoro/services/json-rpc"

	"zktoro/zktoro-core-go/clients/health"
	"zktoro/zktoro-core-go/utils"
)

func initJsonRpcProxy(ctx context.Context, cfg config.Config) (*jrp.JsonRpcProxy, error) {
	return jrp.NewJsonRpcProxy(ctx, cfg)
}

func initServices(ctx context.Context, cfg config.Config) ([]services.Service, error) {
	// can't dial localhost - need to dial host gateway from container
	cfg.Scan.JsonRpc.Url = utils.ConvertToDockerHostURL(cfg.Scan.JsonRpc.Url)
	cfg.JsonRpcProxy.JsonRpc.Url = utils.ConvertToDockerHostURL(cfg.JsonRpcProxy.JsonRpc.Url)

	proxy, err := initJsonRpcProxy(ctx, cfg)
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

	apiErr, ok := reports.NameContains("service.json-rpc-proxy.api")
	if ok && len(apiErr.Details) > 0 {
		summary.Addf("last time the api failed with error '%s'.", apiErr.Details)
	}

	return summary.Finish()
}

func Run() {
	services.ContainerMain("json-rpc", initServices)
}
