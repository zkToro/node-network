package registry

import (
	"time"

	"github.com/sirupsen/logrus"
	"zktoro/zktoro-core-go/contracts/merged/contract_scanner_node_version"
	"zktoro/zktoro-core-go/domain"
	"zktoro/zktoro-core-go/domain/registry/regmsg"
)

const ScannerNodeVersionUpdated = "ScannerNodeVersionUpdated"

type ScannerNodeVersionMessage struct {
	regmsg.Message
	NewVersion string `json:"newVersion"`
	OldVersion string `json:"oldVersion"`
}

func (snvm *ScannerNodeVersionMessage) LogFields() logrus.Fields {
	return logrus.Fields{}
}

func NewScannerNodeVersionUpdated(evt *contract_scanner_node_version.ScannerNodeVersionScannerNodeVersionUpdated, blk *domain.Block) *ScannerNodeVersionMessage {
	return &ScannerNodeVersionMessage{
		Message: regmsg.Message{
			Timestamp: time.Now().UTC(),
			Action:    ScannerNodeVersionUpdated,
			Source:    regmsg.SourceFromBlock(evt.Raw.TxHash.Hex(), blk),
		},

		NewVersion: evt.NewVersion,
		OldVersion: evt.OldVersion,
	}
}
