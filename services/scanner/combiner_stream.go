package scanner

import (
	"context"

	"zktoro/clients"
	"zktoro/clients/messaging"

	"zktoro/zktoro-core-go/clients/health"
	"zktoro/zktoro-core-go/domain"
	"zktoro/zktoro-core-go/feeds"

	log "github.com/sirupsen/logrus"
)

// CombinerAlertStreamService pulls alert info from providers and emits to channel
type CombinerAlertStreamService struct {
	cfg         CombinerAlertStreamServiceConfig
	ctx         context.Context
	alertOutput chan *domain.AlertEvent
	alertFeed   feeds.AlertFeed
	msgClient   clients.MessageClient

	subscribeChan     chan string
	unSubscribeChan   chan string
	lastAlertActivity health.TimeTracker
}

type CombinerAlertStreamServiceConfig struct {
	Start uint64
	End   uint64
}

func (t *CombinerAlertStreamService) registerMessageHandlers() {
	t.msgClient.Subscribe(messaging.SubjectAgentsAlertSubscribe, messaging.SubscriptionHandler(t.handleMessageSubscribe))
	t.msgClient.Subscribe(messaging.SubjectAgentsAlertUnsubscribe, messaging.SubscriptionHandler(t.handleMessageUnsubscribe))
}

func (t *CombinerAlertStreamService) ReadOnlyAlertStream() <-chan *domain.AlertEvent {
	return t.alertOutput
}

func (t *CombinerAlertStreamService) handleAlert(evt *domain.AlertEvent) error {
	select {
	case <-t.ctx.Done():
		return nil
	default:
	}

	log.WithFields(
		log.Fields{
			"subscribee":     evt.Event.Alert.Source.Bot.Id,
			"alert":          evt.Event.Alert.Hash,
			"sourceBot":      evt.Subscriber.BotID,
			"sourceOwner":    evt.Subscriber.BotOwner,
			"sourceBotImage": evt.Subscriber.BotImage,
		},
	).Debug("streaming new alert event")

	t.alertOutput <- evt
	t.lastAlertActivity.Set()
	return nil
}

func (t *CombinerAlertStreamService) Start() error {
	t.registerMessageHandlers()
	go func() {
		t.alertFeed.RegisterHandler(t.handleAlert)
		t.alertFeed.Start()
	}()
	return nil
}

func (t *CombinerAlertStreamService) Stop() error {
	if t.alertOutput != nil {
		// drain and close block channel
		func(c chan *domain.AlertEvent) {
			for {
				select {
				case a := <-c:
					log.WithFields(log.Fields{"subscription": a.Event.Alert.Source.Bot.Id}).Info("gracefully draining combination")
				default:
					close(c)
					return
				}
			}
		}(t.alertOutput)
	}
	return nil
}

func (t *CombinerAlertStreamService) Name() string {
	return "combiner-alert-stream"
}

// Health implements health.Reporter interface.
func (t *CombinerAlertStreamService) Health() health.Reports {
	return health.Reports{
		t.lastAlertActivity.GetReport("event.alert.time"),
	}
}

func (t *CombinerAlertStreamService) handleMessageSubscribe(payload messaging.SubscriptionPayload) error {
	for _, subscription := range payload {
		if subscription == nil || subscription.Subscription == nil || subscription.Subscriber == nil {
			continue
		}

		err := t.alertFeed.AddSubscription(subscription)
		if err != nil {
			log.WithFields(
				log.Fields{
					"subscriberBot": subscription.Subscriber.BotID,
					"subscribesTo":  subscription.Subscription.BotId,
				},
			).WithError(err).Warn("can not add subscription")
			continue
		}
	}

	return nil
}

func (t *CombinerAlertStreamService) handleMessageUnsubscribe(payload messaging.SubscriptionPayload) error {
	for _, subscription := range payload {
		if subscription == nil {
			continue
		}

		t.alertFeed.RemoveSubscription(subscription)
	}

	return nil
}

func NewCombinerAlertStreamService(ctx context.Context, alertFeed feeds.AlertFeed, msgClient clients.MessageClient, cfg CombinerAlertStreamServiceConfig) (*CombinerAlertStreamService, error) {
	alertOutput := make(chan *domain.AlertEvent)

	return &CombinerAlertStreamService{
		cfg:         cfg,
		ctx:         ctx,
		msgClient:   msgClient,
		alertOutput: alertOutput,
		alertFeed:   alertFeed,
	}, nil
}
