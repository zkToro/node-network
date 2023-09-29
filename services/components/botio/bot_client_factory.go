package botio

import (
	"context"

	"zktoro/clients"
	"zktoro/clients/agentgrpc"
	"zktoro/config"
	"zktoro/services/components/botio/botreq"
	"zktoro/services/components/metrics"
)

// BotClientFactory creates new bot clients.
type BotClientFactory interface {
	NewBotClient(ctx context.Context, botConfig config.AgentConfig) BotClient
}

type botClientFactory struct {
	resultChannels   botreq.SendOnlyChannels
	msgClient        clients.MessageClient
	lifecycleMetrics metrics.Lifecycle
	dialer           agentgrpc.BotDialer
}

// NewBotClientFactory creates a new bot client factory by reusing provided dependencies.
func NewBotClientFactory(
	resultChannels botreq.SendOnlyChannels, msgClient clients.MessageClient,
	lifecycleMetrics metrics.Lifecycle, dialer agentgrpc.BotDialer,
) BotClientFactory {
	return &botClientFactory{
		resultChannels:   resultChannels,
		msgClient:        msgClient,
		lifecycleMetrics: lifecycleMetrics,
		dialer:           dialer,
	}
}

func (bcf *botClientFactory) NewBotClient(ctx context.Context, botConfig config.AgentConfig) BotClient {
	return NewBotClient(ctx, botConfig, bcf.msgClient, bcf.lifecycleMetrics, bcf.dialer, bcf.resultChannels)
}
