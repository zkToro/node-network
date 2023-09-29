package publisher

import (
	"context"

	"zktoro/healthutils"

	"zktoro/zktoro-core-go/clients/health"
	"zktoro/zktoro-core-go/utils"

	log "github.com/sirupsen/logrus"

	"zktoro/config"
	"zktoro/services"
	"zktoro/services/publisher"
)

func initServices(ctx context.Context, cfg config.Config) ([]services.Service, error) {
	cfg.Publish.APIURL = utils.ConvertToDockerHostURL(cfg.Publish.APIURL)
	cfg.Publish.IPFS.APIURL = utils.ConvertToDockerHostURL(cfg.Publish.IPFS.APIURL)
	cfg.Publish.IPFS.GatewayURL = utils.ConvertToDockerHostURL(cfg.Publish.IPFS.GatewayURL)
	cfg.LocalModeConfig.WebhookURL = utils.ConvertToDockerHostURL(cfg.LocalModeConfig.WebhookURL)

	p, err := publisher.NewPublisher(ctx, cfg)
	if err != nil {
		log.Errorf("Error while initializing Listener: %s", err.Error())
		return nil, err
	}

	return []services.Service{
		health.NewService(
			ctx, "", healthutils.DefaultHealthServerErrHandler,
			health.CheckerFrom(summarizeReports, p),
		),
		p,
	}, nil
}

func summarizeReports(reports health.Reports) *health.Report {
	summary := health.NewSummary()

	batchPublishErr, ok := reports.NameContains("publisher.event.batch-publish.error")
	if ok && len(batchPublishErr.Details) > 0 {
		summary.Addf("failed to publish the last batch with error '%s'", batchPublishErr.Details)
		summary.Status(health.StatusFailing)
	}

	return summary.Finish()
}

func Run() {
	services.ContainerMain("publisher", initServices)
}
