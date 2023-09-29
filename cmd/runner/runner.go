package runner

import (
	"context"
	"fmt"
	"os"

	"zktoro/clients/docker"
	"zktoro/config"
	"zktoro/services"
	"zktoro/services/runner"
	"zktoro/store"

	log "github.com/sirupsen/logrus"
)

func initServices(ctx context.Context, cfg config.Config) ([]services.Service, error) {
	shouldDisableAutoUpdate := cfg.AutoUpdate.Disable
	imgStore, err := store.NewZktoroImageStore(ctx, config.DefaultContainerPort, !shouldDisableAutoUpdate)
	if err != nil {
		return nil, fmt.Errorf("failed to create the image store: %v", err)
	}
	dockerClient, err := docker.NewDockerClient("runner")
	if err != nil {
		return nil, fmt.Errorf("failed to create the docker client: %v", err)
	}
	globalDockerClient, err := docker.NewDockerClient("")
	if err != nil {
		return nil, fmt.Errorf("failed to create the docker client: %v", err)
	}

	if cfg.Development {
		log.Warn("running in development mode")
	}

	return []services.Service{
		runner.NewRunner(ctx, cfg, *imgStore, dockerClient, globalDockerClient),
	}, nil
}

// Run runs the runner.
func Run(cfg config.Config) {
	ctx, cancel := services.InitMainContext()
	defer cancel()

	logger := log.WithField("process", "runner")
	logger.Info("starting")
	defer logger.Info("exiting")

	serviceList, err := initServices(ctx, cfg)
	if err != nil {
		logger.WithError(err).Error("could not initialize services")
		return
	}

	err = services.StartServices(ctx, cancel, log.NewEntry(log.StandardLogger()), serviceList)
	if err == services.ErrExitTriggered {
		logger.Info("exiting successfully after internal trigger")
		os.Exit(0)
	}
	if err != nil {
		logger.WithError(err).Error("error running services")
	}
}
