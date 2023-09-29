package scanner

import (
	"context"
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

// BlockAnalyzerService reads TX info, calls agents, and emits results
type BlockAnalyzerService struct {
	ctx           context.Context
	cfg           BlockAnalyzerServiceConfig
	publisherNode protocol.PublisherNodeClient

	lastInputActivity  health.TimeTracker
	lastOutputActivity health.TimeTracker
}

type BlockAnalyzerServiceConfig struct {
	BlockChannel <-chan *domain.BlockEvent
	AlertSender  clients.AlertSender
	MsgClient    clients.MessageClient
	components.BotProcessing
}

func (t *BlockAnalyzerService) publishMetrics(result *botreq.BlockResult) {
	m := metrics.GetBlockMetrics(result.AgentConfig, result.Response, result.Timestamps)
	t.cfg.MsgClient.PublishProto(messaging.SubjectMetricAgent, &protocol.AgentMetricList{Metrics: m})
}

func (t *BlockAnalyzerService) findingToAlert(result *botreq.BlockResult, ts time.Time, f *protocol.Finding) (
	*protocol.Alert, error,
) {
	alertID := alerthash.ForBlockAlert(
		&alerthash.Inputs{
			BlockEvent: result.Request.Event,
			Finding:    f,
			BotInfo: alerthash.BotInfo{
				BotImage: result.AgentConfig.Image,
				BotID:    result.AgentConfig.ID,
			},
		},
	)

	blockNumber, err := utils.HexToBigInt(result.Request.Event.BlockNumber)
	if err != nil {
		return nil, err
	}
	chainId, err := utils.HexToBigInt(result.Request.Event.Network.ChainId)
	if err != nil {
		return nil, err
	}
	tags := map[string]string{
		"agentImage": result.AgentConfig.Image,
		"agentId":    result.AgentConfig.ID,
		"chainId":    chainId.String(),
	}

	alertType := protocol.AlertType_PRIVATE
	if !f.Private && !result.Response.Private {
		alertType = protocol.AlertType_BLOCK
		tags["blockHash"] = result.Request.Event.BlockHash
		tags["blockNumber"] = blockNumber.String()
	}

	addressBloomFilter, err := t.createBloomFilter(f)
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

func (t *BlockAnalyzerService) createBloomFilter(finding *protocol.Finding) (bloomFilter *protocol.BloomFilter, err error) {
	return utils.CreateBloomFilter(finding.Addresses, utils.AddressBloomFilterFPRate)
}

func (t *BlockAnalyzerService) Start() error {
	// Gear 2: receive result from agent
	go func() {
		for result := range t.cfg.Results.Block {
			ts := time.Now().UTC()

			resStr, err := protojson.Marshal(result.Response)
			if err != nil {
				log.Error("error marshaling response", err)
				continue
			}
			log.Debugf(string(resStr))

			rt := &clients.AgentRoundTrip{
				AgentConfig:       result.AgentConfig,
				EvalBlockRequest:  result.Request,
				EvalBlockResponse: result.Response,
			}

			if len(result.Response.Findings) == 0 {
				if err := t.cfg.AlertSender.NotifyWithoutAlert(
					rt, result.Timestamps,
				); err != nil {
					log.WithError(err).Panic("failed to notify without alert")
				}
			}

			for _, f := range result.Response.Findings {
				alert, err := t.findingToAlert(result, ts, f)
				if err != nil {
					log.WithError(err).Error("failed to transform finding to alert")
					continue
				}
				if err := t.cfg.AlertSender.SignAlertAndNotify(
					rt, alert, result.Request.Event.Network.ChainId, result.Request.Event.BlockNumber, result.Timestamps,
				); err != nil {
					log.WithError(err).Panic("failed sign alert and notify")
				}
			}
			t.publishMetrics(result)

			t.lastOutputActivity.Set()
		}
	}()

	// Gear 1: loops over blocks and distributes to all agents
	go func() {
		// for each block
		for block := range t.cfg.BlockChannel {
			// convert to message
			blockEvt, err := block.ToMessage()
			if err != nil {
				log.WithError(err).Error("error converting block event to message (skipping)")
				continue
			}

			// create a request
			requestId := uuid.Must(uuid.NewUUID())
			request := &protocol.EvaluateBlockRequest{RequestId: requestId.String(), Event: blockEvt}

			// forward to the pool
			t.cfg.RequestSender.SendEvaluateBlockRequest(request)

			t.lastInputActivity.Set()
		}
	}()

	return nil
}

func (t *BlockAnalyzerService) Stop() error {
	return nil
}

func (t *BlockAnalyzerService) Name() string {
	return "block-analyzer"
}

// Health implements the health.Reporter interface.
func (t *BlockAnalyzerService) Health() health.Reports {
	return health.Reports{
		t.lastInputActivity.GetReport("event.input.time"),
		t.lastOutputActivity.GetReport("event.output.time"),
	}
}

func NewBlockAnalyzerService(ctx context.Context, cfg BlockAnalyzerServiceConfig) (*BlockAnalyzerService, error) {
	return &BlockAnalyzerService{
		cfg: cfg,
		ctx: ctx,
	}, nil
}
