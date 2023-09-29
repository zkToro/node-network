package jwt_provider

import (
	"context"

	"zktoro/config"
	"zktoro/healthutils"
	"zktoro/services"
	jwt_provider "zktoro/services/jwt-provider"

	"zktoro/zktoro-core-go/clients/health"
)

func initJWTProvider(cfg config.Config) (*jwt_provider.JWTAPI, error) {
	return jwt_provider.NewJWTAPI(cfg)
}

func initServices(ctx context.Context, cfg config.Config) ([]services.Service, error) {
	jwtProvider, err := initJWTProvider(cfg)
	if err != nil {
		return nil, err
	}

	return []services.Service{
		health.NewService(
			ctx, "", healthutils.DefaultHealthServerErrHandler,
			health.CheckerFrom(nil, jwtProvider),
		),
		jwtProvider,
	}, nil
}

func Run() {
	services.ContainerMain("jwt-provider", initServices)
}
