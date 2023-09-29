package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"zktoro/zktoro-core-go/security"

	"github.com/docker/docker/api/types"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	log "github.com/sirupsen/logrus"

	"zktoro/clients"
	"zktoro/clients/docker"
	"zktoro/config"
	sec "zktoro/services/components/security"
)

var ErrCannotFindBotForIP = errors.New("cannot find bot for ip")

type JWTProvider interface {
	CreateJWTFromIP(ctx context.Context, ipAddress string, claims map[string]interface{}) (string, error)
}

type jwtProvider struct {
	cfg            config.Config
	key            *keystore.Key
	dockerClient   clients.DockerClient
	jwtCreatorFunc func(key *keystore.Key, claims map[string]interface{}) (string, error)
}

func NewJWTProvider(cfg config.Config) (JWTProvider, error) {
	dc, err := docker.NewDockerClient("")
	if err != nil {
		return nil, fmt.Errorf("failed to create the global docker client: %v", err)
	}
	key, err := security.LoadKey(config.DefaultContainerKeyDirPath)
	if err != nil {
		return nil, err
	}
	return &jwtProvider{
		cfg:            cfg,
		key:            key,
		dockerClient:   dc,
		jwtCreatorFunc: security.CreateScannerJWT,
	}, nil
}

func (p *jwtProvider) CreateJWTFromIP(ctx context.Context, ipAddress string, claims map[string]interface{}) (string, error) {
	logger := log.WithFields(log.Fields{
		"ip": ipAddress,
	})
	bot, err := p.getBotIDForIPAddress(ctx, ipAddress)
	if err != nil {
		logger.WithError(err).Warn("could not get bot by ip")
		return "", ErrCannotFindBotForIP
	}
	logger = logger.WithFields(log.Fields{
		"agentId": bot,
	})

	res, err := sec.CreateBotJWT(p.key, bot, claims, p.jwtCreatorFunc)
	if err != nil {
		logger.WithError(err).Error("error creating jwt")
		return "", err
	}

	return res, nil
}

// agentIDReverseLookup reverse lookup from ip to agent id.
func (p *jwtProvider) getBotIDForIPAddress(ctx context.Context, ipAddr string) (string, error) {
	container, err := p.findContainerByIP(ctx, ipAddr)
	if err != nil {
		return "", err
	}

	botID, err := p.extractBotIDFromContainer(ctx, container)
	if err != nil {
		return "", err
	}

	return botID, nil
}

const envPrefix = config.EnvzktoroBotID + "="

func (p *jwtProvider) extractBotIDFromContainer(ctx context.Context, container types.Container) (string, error) {
	// container struct doesn't have the "env" information, inspection required.
	c, err := p.dockerClient.InspectContainer(ctx, container.ID)
	if err != nil {
		return "", err
	}

	// find the env variable with bot id
	for _, s := range c.Config.Env {
		if env := strings.SplitAfter(s, envPrefix); len(env) == 2 {
			return env[1], nil
		}
	}

	return "", fmt.Errorf("can't extract bot id from container")
}

func (p *jwtProvider) findContainerByIP(ctx context.Context, ipAddr string) (types.Container, error) {
	containers, err := p.dockerClient.GetContainers(ctx)
	if err != nil {
		return types.Container{}, err
	}

	// find the container that has the same ip
	for _, container := range containers {
		for _, network := range container.NetworkSettings.Networks {
			if network.IPAddress == ipAddr {
				return container, nil
			}
		}
	}
	return types.Container{}, fmt.Errorf("can't find container %s", ipAddr)
}
