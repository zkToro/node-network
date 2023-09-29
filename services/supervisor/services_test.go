package supervisor

import (
	"context"
	"fmt"
	"os"
	"testing"

	"zktoro/clients/messaging"

	"zktoro/zktoro-core-go/clients/agentlogs"
	"zktoro/zktoro-core-go/release"
	"zktoro/zktoro-core-go/security"
	"zktoro/zktoro-core-go/utils"

	"github.com/ethereum/go-ethereum/accounts/keystore"

	"github.com/docker/docker/api/types"

	mrelease "zktoro/zktoro-core-go/release/mocks"

	"zktoro/clients/docker"
	mock_clients "zktoro/clients/mocks"
	"zktoro/config"
	"zktoro/services/components/containers"
	mock_containers "zktoro/services/components/containers/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	testImageRef               = "some.docker.registry.io/foobar@sha256:cdd4ddccf5e9c740eb4144bcc68e3ea3a056789ec7453e94a6416dcfc80937a4"
	testNodeNetworkID          = "node-network-id"
	testNatsNetworkID          = "nats-network-id"
	testPublicAPINetworkID     = "public-api-network-id"
	testGenericContainerID     = "test-generic-container-id"
	testInspectorContainerID   = "test-inspector-container-id"
	testScannerContainerID     = "test-scanner-container-id"
	testProxyContainerID       = "test-proxy-container-id"
	testPublicAPIContainerID   = "test-public-api-container-id"
	testSupervisorContainerID  = "test-supervisor-container-id"
	testJWTProviderContainerID = "test-jwt-provider-container-id"
)

// TestSuite runs the test suite.
func TestSuite(t *testing.T) {
	suite.Run(t, &Suite{})
}

// Suite is a test suite to test the tx node runner implementation.
type Suite struct {
	r *require.Assertions

	dockerClient  *mock_clients.MockDockerClient
	globalClient  *mock_clients.MockDockerClient
	releaseClient *mrelease.MockClient

	msgClient *mock_clients.MockMessageClient
	botClient *mock_containers.MockBotClient

	supervisor *SupervisorService

	suite.Suite
}

// configMatcher is a wrapper to implement the Matcher interface.
type configMatcher docker.ContainerConfig

// Matches implements the gomock.Matcher interface.
func (m configMatcher) Matches(x interface{}) bool {
	c1, ok := x.(docker.ContainerConfig)
	if !ok {
		return false
	}
	c2 := m

	if c2.Env != nil && c1.Env == nil {
		return false
	}

	for k2, v2 := range c2.Env {
		if v1, ok := c1.Env[k2]; !ok {
			return false
		} else {
			if v1 != v2 {
				return false
			}
		}

	}

	return c1.Name == c2.Name
}

// String implements the gomock.Matcher interface.
func (m configMatcher) String() string {
	return fmt.Sprintf("%+v", (docker.ContainerConfig)(m))
}

// SetupTest sets up the test.
func (s *Suite) SetupTest() {
	s.r = require.New(s.T())
	os.Setenv(config.EnvHostzktoroDir, "/tmp/zktoro")
	ctrl := gomock.NewController(s.T())
	s.dockerClient = mock_clients.NewMockDockerClient(ctrl)
	s.globalClient = mock_clients.NewMockDockerClient(ctrl)
	s.releaseClient = mrelease.NewMockClient(ctrl)
	s.botClient = mock_containers.NewMockBotClient(ctrl)

	s.msgClient = mock_clients.NewMockMessageClient(ctrl)

	dir := s.T().TempDir()
	ks := keystore.NewKeyStore(dir, keystore.StandardScryptN, keystore.StandardScryptP)

	_, err := ks.NewAccount("zktoro123")
	s.r.NoError(err)

	key, err := security.LoadKeyWithPassphrase(dir, "zktoro123")
	s.r.NoError(err)

	supervisor := &SupervisorService{
		ctx:           context.Background(),
		client:        s.dockerClient,
		globalClient:  s.globalClient,
		msgClient:     s.msgClient,
		releaseClient: s.releaseClient,
	}
	supervisor.config.Key = key
	supervisor.config.Config.TelemetryConfig.Disable = true
	supervisor.config.Config.Log.Level = "debug"
	supervisor.config.Config.ChainID = 1
	supervisor.config.Config.AdvancedConfig.IPFSExperiment = true
	supervisor.config.Config.InspectionConfig.InspectAtStartup = utils.BoolPtr(false)
	supervisor.config.Config.AgentLogsConfig.SendIntervalSeconds = 1
	supervisor.botLifecycleConfig.Config = supervisor.config.Config
	supervisor.botLifecycle.BotClient = s.botClient
	s.supervisor = supervisor
}

func (s *Suite) initialContainerCheck() {
	for _, containerName := range knownServiceContainerNames {
		s.dockerClient.EXPECT().GetContainerByName(s.supervisor.ctx, containerName).Return(&types.Container{ID: testGenericContainerID}, nil)
	}

	s.dockerClient.EXPECT().GetContainers(s.supervisor.ctx).Return(
		[]types.Container{
			{
				Names: []string{"/zktoro-agent-name"},
				ID:    testGenericContainerID,
				Labels: map[string]string{
					docker.LabelzktoroSupervisorStrategyVersion: containers.LabelValueStrategyVersion,
				},
			},
			{
				Names: []string{"/zktoro-agent-name"},
				ID:    testGenericContainerID,
				Labels: map[string]string{
					docker.LabelzktoroSupervisorStrategyVersion: "old",
				},
			},
		}, nil,
	)

	// supervisor-managed containers
	for i := 0; i < len(knownServiceContainerNames)+1; i++ {
		s.dockerClient.EXPECT().RemoveContainer(s.supervisor.ctx, testGenericContainerID).Return(nil)
		s.dockerClient.EXPECT().WaitContainerPrune(s.supervisor.ctx, testGenericContainerID).Return(nil)
	}
	for i := 0; i < len(knownServiceContainerNames)+1; i++ {
		s.dockerClient.EXPECT().RemoveNetworkByName(s.supervisor.ctx, gomock.Any()).Return(nil)
	}
}

func (s *Suite) TestStartServices() {
	s.msgClient.EXPECT().Subscribe(messaging.SubjectMetricAgent, gomock.Any())

	s.releaseClient.EXPECT().GetReleaseManifest(gomock.Any()).Return(&release.ReleaseManifest{}, nil).AnyTimes()

	s.initialContainerCheck()
	s.dockerClient.EXPECT().EnsureLocalImage(s.supervisor.ctx, gomock.Any(), gomock.Any()).Times(2) // needs to get nats and ipfs
	s.dockerClient.EXPECT().EnsurePublicNetwork(s.supervisor.ctx, gomock.Any()).Return(testNodeNetworkID, nil)
	s.dockerClient.EXPECT().EnsureInternalNetwork(s.supervisor.ctx, gomock.Any()).Return(testNatsNetworkID, nil) // for nats
	s.dockerClient.EXPECT().StartContainer(
		s.supervisor.ctx, (configMatcher)(
			docker.ContainerConfig{
				Name: config.DockerIpfsContainerName,
			},
		),
	).Return(&docker.Container{}, nil)
	s.dockerClient.EXPECT().StartContainer(
		s.supervisor.ctx, (configMatcher)(
			docker.ContainerConfig{
				Name: config.DockerStorageContainerName,
			},
		),
	).Return(&docker.Container{}, nil)
	s.dockerClient.EXPECT().StartContainer(
		s.supervisor.ctx, (configMatcher)(
			docker.ContainerConfig{
				Name: config.DockerNatsContainerName,
			},
		),
	).Return(&docker.Container{}, nil)
	s.dockerClient.EXPECT().StartContainer(
		s.supervisor.ctx, (configMatcher)(
			docker.ContainerConfig{
				Name: config.DockerJSONRPCProxyContainerName,
			},
		),
	).Return(&docker.Container{ID: testProxyContainerID}, nil)
	s.dockerClient.EXPECT().StartContainer(
		s.supervisor.ctx, (configMatcher)(
			docker.ContainerConfig{
				Name: config.DockerPublicAPIProxyContainerName,
			},
		),
	).Return(&docker.Container{ID: testPublicAPIContainerID}, nil)
	s.dockerClient.EXPECT().StartContainer(
		s.supervisor.ctx, (configMatcher)(
			docker.ContainerConfig{
				Name: config.DockerScannerContainerName,
			},
		),
	).Return(&docker.Container{ID: testScannerContainerID}, nil)
	s.dockerClient.EXPECT().StartContainer(
		s.supervisor.ctx, (configMatcher)(
			docker.ContainerConfig{
				Name: config.DockerJWTProviderContainerName,
			},
		),
	).Return(&docker.Container{ID: testJWTProviderContainerID}, nil)
	s.dockerClient.EXPECT().StartContainer(
		s.supervisor.ctx, (configMatcher)(
			docker.ContainerConfig{
				Name: config.DockerInspectorContainerName,
			},
		),
	).Return(&docker.Container{ID: testProxyContainerID}, nil)
	s.globalClient.EXPECT().GetContainerByName(s.supervisor.ctx, config.DockerSupervisorContainerName).Return(&types.Container{ID: testSupervisorContainerID}, nil).AnyTimes()
	s.dockerClient.EXPECT().AttachNetwork(s.supervisor.ctx, testSupervisorContainerID, testNodeNetworkID)
	s.dockerClient.EXPECT().AttachNetwork(s.supervisor.ctx, testSupervisorContainerID, testNatsNetworkID)
	s.dockerClient.EXPECT().GetContainerByName(s.supervisor.ctx, config.DockerJSONRPCProxyContainerName).Return(&types.Container{ID: testProxyContainerID}, nil).AnyTimes()
	s.dockerClient.EXPECT().GetContainerByName(s.supervisor.ctx, config.DockerInspectorContainerName).Return(&types.Container{ID: testInspectorContainerID}, nil).AnyTimes()
	s.dockerClient.EXPECT().GetContainerByName(s.supervisor.ctx, config.DockerScannerContainerName).Return(&types.Container{ID: testScannerContainerID}, nil).AnyTimes()
	s.dockerClient.EXPECT().GetContainerByName(s.supervisor.ctx, config.DockerScannerContainerName).Return(&types.Container{ID: testScannerContainerID}, nil).AnyTimes()
	s.dockerClient.EXPECT().GetContainerByName(
		s.supervisor.ctx,
		config.DockerJWTProviderContainerName,
	).Return(&types.Container{ID: testJWTProviderContainerID}, nil).AnyTimes()
	s.dockerClient.EXPECT().WaitContainerStart(s.supervisor.ctx, gomock.Any()).Return(nil).AnyTimes()

	s.r.NoError(s.supervisor.start())
}

func (s *Suite) TestDoSyncAgentLogs() {
	s.botClient.EXPECT().LoadBotContainers(gomock.Any()).Return([]types.Container{{
		Labels: map[string]string{
			docker.LabelzktoroSettingsAgentLogsEnable: "true",
		},
	}}, nil)
	s.dockerClient.EXPECT().GetContainerLogs(gomock.Any(), gomock.Any(), "50", 10000)
	s.supervisor.sendAgentLogs = func(agents agentlogs.Agents, authToken string) error {
		return nil
	}
	s.r.NoError(s.supervisor.doSyncAgentLogs())
}

func (s *Suite) TestDoHealthCheck() {
	s.r.NoError(s.supervisor.doHealthCheck())
}
