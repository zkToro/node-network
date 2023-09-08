package registry

import (
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"zktoro/zktoro-core-go/domain"
	"zktoro/zktoro-core-go/domain/registry/regmsg"

	"zktoro/zktoro-core-go/contracts/merged/contract_dispatch"
	"zktoro/zktoro-core-go/utils"
)

var Link = "Link"
var Unlink = "Unlink"

type DispatchMessage struct {
	regmsg.Message
	ScannerID string `json:"scannerId"`
	AgentID   string `json:"agentId"`
}

func (dm *DispatchMessage) LogFields() logrus.Fields {
	return logrus.Fields{
		"scannerId": dm.ScannerID,
		"agentId":   dm.AgentID,
	}
}

func NewDispatchMessage(evt *contract_dispatch.DispatchLink, blk *domain.Block) *DispatchMessage {
	scannerID := utils.HexAddr(evt.ScannerId)
	agentID := utils.Hex(evt.AgentId)
	evtName := Unlink
	if evt.Enable {
		evtName = Link
	}
	return &DispatchMessage{
		Message: regmsg.Message{
			Action:    evtName,
			Timestamp: time.Now().UTC(),
			Source:    regmsg.SourceFromBlock(evt.Raw.TxHash.Hex(), blk),
		},
		ScannerID: strings.ToLower(scannerID),
		AgentID:   agentID,
	}
}

func NewAlreadyLinkedDispatchMessage(evt *contract_dispatch.DispatchAlreadyLinked, blk *domain.Block) *DispatchMessage {
	scannerID := utils.HexAddr(evt.ScannerId)
	agentID := utils.Hex(evt.AgentId)
	evtName := Unlink
	if evt.Enable {
		evtName = Link
	}
	return &DispatchMessage{
		Message: regmsg.Message{
			Action:    evtName,
			Timestamp: time.Now().UTC(),
			Source:    regmsg.SourceFromBlock(evt.Raw.TxHash.Hex(), blk),
		},
		ScannerID: strings.ToLower(scannerID),
		AgentID:   agentID,
	}
}
