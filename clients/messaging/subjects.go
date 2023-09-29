package messaging

import (
	"zktoro/config"
	"zktoro/zktoro-core-go/domain"
	"zktoro/zktoro-core-go/protocol"
)

// Message types
const (
	SubjectAgentsAlertSubscribe   = "agents.alert.subscribe"
	SubjectAgentsAlertUnsubscribe = "agents.alert.unsubscribe"
	SubjectAgentsStatusRunning    = "agents.status.running"
	SubjectAgentsStatusAttached   = "agents.status.attached"
	SubjectAgentsStatusStopping   = "agents.status.stopping"
	SubjectAgentsStatusStopped    = "agents.status.stopped"
	SubjectAgentsStatusRestarted  = "agents.status.restarted"
	SubjectMetricAgent            = "metric.agent"
	SubjectScannerBlock           = "scanner.block"
	SubjectScannerAlert           = "scanner.alert"
	SubjectInspectionDone         = "inspection.done"
)

// AgentPayload is the message payload.
type AgentPayload []config.AgentConfig

// AgentMetricPayload is the message payload for metrics.
type AgentMetricPayload *protocol.AgentMetricList

// SubscriptionPayload is the message payload for combiner bot subscriptions.
type SubscriptionPayload []*domain.CombinerBotSubscription

// ScannerPayload is the message payload for general scanner info.
type ScannerPayload struct {
	LatestBlockInput uint64 `json:"latestBlockInput"`
}
