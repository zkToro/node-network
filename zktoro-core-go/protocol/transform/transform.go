package transform

import (
	"zktoro/zktoro-core-go/clients/webhook/client/models"
	"zktoro/zktoro-core-go/protocol"
	"zktoro/zktoro-core-go/utils"
)

// ToWebhookAlertBatch transforms an alert batch to a webhook alert batch.
func ToWebhookAlertBatch(batch *protocol.AlertBatch) *models.AlertBatch {
	return &models.AlertBatch{
		Alerts:  ToWebhookAlertList(batch),
		Metrics: ToWebhookBotMetricsList(batch),
	}
}

// ToWebhookBotMetricsList transforms an alert batch to a bot metrics list.
func ToWebhookBotMetricsList(batch *protocol.AlertBatch) models.BotMetricsList {
	var metricsList models.BotMetricsList
	for _, metric := range batch.Metrics {
		webhookMetric := &models.BotMetric{
			BotID:     metric.AgentId,
			Timestamp: metric.Timestamp,
		}
		for _, metricSummary := range metric.Metrics {
			webhookMetric.Metrics = append(
				webhookMetric.Metrics, &models.BotMetricSummary{
					Name:    metricSummary.Name,
					Count:   float64(metricSummary.Count),
					Max:     metricSummary.Max,
					Average: metricSummary.Average,
					Sum:     metricSummary.Sum,
					P95:     metricSummary.P95,
				},
			)
		}
		metricsList = append(metricsList, webhookMetric)
	}
	return metricsList
}

// ToWebhookAlertList transforms an alert batch to a webhook alert list.
func ToWebhookAlertList(batch *protocol.AlertBatch) models.AlertList {
	var alertList models.AlertList
	for _, resultsForBlock := range batch.Results {
		for _, blockResult := range resultsForBlock.Results {
			for _, alert := range blockResult.Alerts {
				alertList = append(
					alertList, ToWebhookAlert(
						alert.Alert,
						batch.ChainId,
						resultsForBlock.Block,
						nil,
						nil,
					),
				)
			}
		}
		for _, resultsForTransaction := range resultsForBlock.Transactions {
			for _, transactionResult := range resultsForTransaction.Results {
				for _, alert := range transactionResult.Alerts {
					alertList = append(
						alertList, ToWebhookAlert(
							alert.Alert,
							batch.ChainId,
							resultsForBlock.Block,
							resultsForTransaction.Transaction,
							nil,
						),
					)
				}
			}
		}
	}

	// handle combiner alerts
	for _, combinationAlertResults := range batch.CombinationAlerts {
		for _, result := range combinationAlertResults.Results {
			for _, alert := range result.Alerts {
				alertList = append(alertList, ToWebhookAlert(alert.Alert, batch.ChainId, nil, nil, combinationAlertResults.AlertEvent))
			}
		}
	}
	return alertList
}

// ToWebhookAlert converts given alert and extra data to webhook alert.
func ToWebhookAlert(
	alert *protocol.Alert, chainID uint64, block *protocol.Block,
	transaction *protocol.TransactionEvent,
	sourceAlertEvent *protocol.AlertEvent,
) *models.Alert {
	webhookAlert := &models.Alert{
		AlertID:       alert.Finding.AlertId,
		CreatedAt:     alert.Timestamp,
		Description:   alert.Finding.Description,
		FindingType:   alert.Finding.Type.String(),
		Hash:          alert.Id,
		Metadata:      alert.Finding.Metadata,
		Name:          alert.Finding.Name,
		Protocol:      alert.Finding.Protocol,
		Severity:      alert.Finding.Severity.String(),
		RelatedAlerts: alert.Finding.RelatedAlerts,
		Source: &models.AlertSource{
			Bot: &models.AlertBot{
				ID:        alert.Agent.Id,
				Image:     alert.Agent.Image,
				Reference: alert.Agent.Manifest,
			},
		},
	}
	if block != nil {
		webhookAlert.Source.Block = &models.AlertBlock{
			ChainID:   chainID,
			Hash:      block.BlockHash,
			Number:    block.BlockNumber,
			Timestamp: block.BlockTimestamp,
		}
	}
	if transaction != nil {
		webhookAlert.Addresses = utils.MapKeys(transaction.Addresses)
		webhookAlert.Source.TransactionHash = transaction.Transaction.Hash
	}

	if sourceAlertEvent != nil {
		webhookAlert.Source.SourceEvent = &models.AlertSourceEvent{
			AlertHash: sourceAlertEvent.Alert.Hash,
			BotID:     sourceAlertEvent.Alert.Source.Bot.Id,
		}
	}

	return webhookAlert
}
