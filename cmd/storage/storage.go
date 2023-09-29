package storage

import (
	"context"
	"fmt"

	"zktoro/config"
	"zktoro/healthutils"
	"zktoro/services"
	"zktoro/services/storage"

	"zktoro/zktoro-core-go/clients/health"
)

func initServices(ctx context.Context, cfg config.Config) ([]services.Service, error) {
	service, err := storage.NewStorage(
		ctx, fmt.Sprintf("http://%s:5001", config.DockerIpfsContainerName),
		cfg.StorageConfig.Provide,
	)
	if err != nil {
		return nil, err
	}

	return []services.Service{
		health.NewService(
			ctx, "", healthutils.DefaultHealthServerErrHandler,
			health.CheckerFrom(nil, service),
		),
		service,
	}, nil
}

func Run() {
	services.ContainerMain("storage", initServices)
}
