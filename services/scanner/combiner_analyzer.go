package scanner

import (
	"context"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"zktoro/clients/messaging"
	"zktoro/services/components"
	"zktoro/services/components/botio/botreq"
	"zktoro/services/components/metrics"

	"zktoro/zktoro-core-go/clients/health"
	"zktoro/zktoro-core-go/domain"
	"zktoro/zktoro-core-go/protocol/alerthash"
	"zktoro/zktoro-core-go/utils"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/encoding/protojson"

	"zktoro/clients"

	"zktoro/zktoro-core-go/protocol"
)

// CombinerAlertAnalyzerService reads alert info, calls agents, and emits results
type CombinerAlertAnalyzerService struct {
	ctx           context.Context
	cfg           CombinerAlertAnalyzerServiceConfig
	publisherNode protocol.PublisherNodeClient

	lastInputActivity  health.TimeTracker
	lastOutputActivity health.TimeTracker
}

type CombinerAlertAnalyzerServiceConfig struct {
	AlertChannel <-chan *domain.AlertEvent
	AlertSender  clients.AlertSender
	MsgClient    clients.MessageClient
	ChainID      string
	components.BotProcessing
}

func (aas *CombinerAlertAnalyzerService) publishMetrics(result *botreq.CombinationAlertResult) {
	m := metrics.GetCombinerMetrics(result.AgentConfig, result.Response, result.Timestamps)
	aas.cfg.MsgClient.PublishProto(messaging.SubjectMetricAgent, &protocol.AgentMetricList{Metrics: m})
}

func (aas *CombinerAlertAnalyzerService) findingToAlert(result *botreq.CombinationAlertResult, ts time.Time, f *protocol.Finding) (*protocol.Alert, error) {
	alertID := alerthash.ForCombinationAlert(
		&alerthash.Inputs{
			AlertEvent: result.Request.Event,
			Finding:    f,
			BotInfo: alerthash.BotInfo{
				BotImage: result.AgentConfig.Image,
				BotID:    result.AgentConfig.ID,
			},
		},
	)

	chainId := big.NewInt(int64(result.Request.Event.Alert.Source.Block.ChainId))
	tags := map[string]string{
		"agentImage": result.AgentConfig.Image,
		"agentId":    result.AgentConfig.ID,
		"chainId":    chainId.String(),
	}

	alertType := protocol.AlertType_PRIVATE
	if !f.Private && !result.Response.Private {
		alertType = protocol.AlertType_COMBINATION
	}

	addressBloomFilter, err := aas.createBloomFilter(f)
	if err != nil {
		return nil, err
	}

	truncated := truncateFinding(f)

	return &protocol.Alert{
		Id:                 alertID,
		Finding:            f,
		Timestamp:          ts.Format(utils.AlertTimeFormat),
		Type:               alertType,
		Agent:              result.AgentConfig.ToAgentInfo(),
		Tags:               tags,
		Timestamps:         result.Timestamps.ToMessage(),
		Truncated:          truncated,
		AddressBloomFilter: addressBloomFilter,
	}, nil
}

func (aas *CombinerAlertAnalyzerService) createBloomFilter(finding *protocol.Finding) (bloomFilter *protocol.BloomFilter, err error) {
	return utils.CreateBloomFilter(finding.Addresses, utils.AddressBloomFilterFPRate)
}

func (aas *CombinerAlertAnalyzerService) Start() error {
	// Gear 2: receive result from agent
	go func() {
		for result := range aas.cfg.Results.CombinationAlert {
			ts := time.Now().UTC()

			resStr, err := protojson.Marshal(result.Response)
			if err != nil {
				log.Error("error marshaling response", err)
				continue
			}
			log.Debugf(string(resStr))

			rt := &clients.AgentRoundTrip{
				AgentConfig:       result.AgentConfig,
				EvalAlertRequest:  result.Request,
				EvalAlertResponse: result.Response,
			}

			if len(result.Response.Findings) == 0 {
				if err := aas.cfg.AlertSender.NotifyWithoutAlert(
					rt, result.Timestamps,
				); err != nil {
					log.WithError(err).Panic("failed to notify without alert")
				}
			}

			chainIDInt, _ := strconv.Atoi(aas.cfg.ChainID)
			chainID := fmt.Sprintf("0x%x", chainIDInt)
			for _, f := range result.Response.Findings {
				alert, err := aas.findingToAlert(result, ts, f)
				if err != nil {
					log.WithError(err).Error("failed to transform finding to alert")
					continue
				}
				if err := aas.cfg.AlertSender.SignAlertAndNotify(
					rt, alert, chainID, "", result.Timestamps,
				); err != nil {
					log.WithError(err).Panic("failed sign alert and notify")
				}
			}
			aas.publishMetrics(result)

			aas.lastOutputActivity.Set()
		}
	}()

	// Gear 1: loops over alerts and distributes to all agents
	go func() {
		// for each alert
		for alertEvt := range aas.cfg.AlertChannel {
			logger := log.WithFields(
				log.Fields{
					"component": "combinerAnalyzer",
					"target":    alertEvt.Subscriber.BotID,
					"source":    alertEvt.Event.Alert.Source.Bot.Id,
				},
			)

			logger.Debug("received alert")

			// convert to message
			alertEvtMsg, err := alertEvt.ToMessage()
			if err != nil {
				logger.WithError(err).Error("error converting alert event to message (skipping)")
				continue
			}

			// create a request
			requestId := uuid.Must(uuid.NewUUID())
			request := &protocol.EvaluateAlertRequest{RequestId: requestId.String(), Event: alertEvtMsg, TargetBotId: alertEvt.Subscriber.BotID}

			// forward to the pool
			aas.cfg.RequestSender.SendEvaluateAlertRequest(request)

			aas.lastInputActivity.Set()
		}
	}()

	return nil
}

func (aas *CombinerAlertAnalyzerService) Stop() error {
	return nil
}

func (aas *CombinerAlertAnalyzerService) Name() string {
	return "combiner-alert-analyzer"
}

// Health implements the health.Reporter interface.
func (aas *CombinerAlertAnalyzerService) Health() health.Reports {
	return health.Reports{
		aas.lastInputActivity.GetReport("event.input.time"),
		aas.lastOutputActivity.GetReport("event.output.time"),
	}
}

func NewCombinerAlertAnalyzerService(ctx context.Context, cfg CombinerAlertAnalyzerServiceConfig) (*CombinerAlertAnalyzerService, error) {
	return &CombinerAlertAnalyzerService{
		cfg: cfg,
		ctx: ctx,
	}, nil
}
