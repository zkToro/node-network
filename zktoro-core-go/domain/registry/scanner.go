package registry

import (
	"math/big"
	"strings"

	"github.com/sirupsen/logrus"
	"zktoro/zktoro-core-go/domain"
	"zktoro/zktoro-core-go/domain/registry/regmsg"

	"zktoro/zktoro-core-go/contracts/merged/contract_scanner_pool_registry"
	"zktoro/zktoro-core-go/contracts/merged/contract_scanner_registry"
	"zktoro/zktoro-core-go/utils"
)

const SaveScanner = "SaveScanner"
const EnableScanner = "EnableScanner"
const DisableScanner = "DisableScanner"
const UpdateScannerPool = "UpdateScannerPool"

const ScannerPermissionAdmin = 0
const ScannerPermissionSelf = 1
const ScannerPermissionOwner = 2
const ScannerPermissionManager = 3

type ScannerMessage struct {
	regmsg.Message
	ScannerID  string `json:"scannerId"`
	Permission int    `json:"permission"`
	Sender     string `json:"sender"`
}

func (sm *ScannerMessage) LogFields() logrus.Fields {
	return logrus.Fields{"scannerId": sm.ScannerID}
}

type ScannerSaveMessage struct {
	ScannerMessage
	ChainID int64  `json:"chainId"`
	PoolID  string `json:"poolId"`
	Enabled bool   `json:"enabled"`
}

type UpdateScannerPoolMessage struct {
	regmsg.Message
	PoolID  string  `json:"poolId"`
	ChainID *int64  `json:"chainId,omitempty"`
	Owner   *string `json:"owner,omitempty"`
}

func (uspm *UpdateScannerPoolMessage) LogFields() logrus.Fields {
	return logrus.Fields{"poolId": uspm.PoolID}
}

func NewScannerMessage(evt *contract_scanner_registry.ScannerRegistryScannerEnabled, blk *domain.Block) *ScannerMessage {
	scannerID := utils.HexAddr(evt.ScannerId)
	evtName := DisableScanner
	if evt.Enabled {
		evtName = EnableScanner
	}
	return &ScannerMessage{
		Message:    regmsg.From(evt.Raw.TxHash.Hex(), blk, evtName),
		ScannerID:  strings.ToLower(scannerID),
		Permission: int(evt.Permission),
	}
}

func NewScannerMessageFromPool(evt *contract_scanner_pool_registry.ScannerPoolRegistryScannerEnabled, blk *domain.Block) *ScannerMessage {
	scannerID := utils.HexAddr(evt.ScannerId)
	evtName := DisableScanner
	if evt.Enabled {
		evtName = EnableScanner
	}
	return &ScannerMessage{
		Message:   regmsg.From(evt.Raw.TxHash.Hex(), blk, evtName),
		ScannerID: strings.ToLower(scannerID),
		Sender:    evt.Sender.Hex(),
	}
}

func NewScannerSaveMessage(evt *contract_scanner_registry.ScannerRegistryScannerUpdated, enabled bool, blk *domain.Block) *ScannerSaveMessage {
	scannerID := utils.HexAddr(evt.ScannerId)
	return &ScannerSaveMessage{
		ScannerMessage: ScannerMessage{
			ScannerID: strings.ToLower(scannerID),
			Message:   regmsg.From(evt.Raw.TxHash.Hex(), blk, SaveScanner),
		},
		ChainID: evt.ChainId.Int64(),
		Enabled: enabled,
	}
}

func NewScannerSaveMessageFromPool(evt *contract_scanner_pool_registry.ScannerPoolRegistryScannerUpdated, enabled bool, blk *domain.Block) *ScannerSaveMessage {
	scannerID := utils.HexAddr(evt.ScannerId)
	return &ScannerSaveMessage{
		ScannerMessage: ScannerMessage{
			ScannerID: strings.ToLower(scannerID),
			Message:   regmsg.From(evt.Raw.TxHash.Hex(), blk, SaveScanner),
		},
		ChainID: evt.ChainId.Int64(),
		PoolID:  utils.PoolIDToString(evt.ScannerPool),
		Enabled: enabled,
	}
}

func NewScannerPoolMessageFromTransfer(evt *contract_scanner_pool_registry.ScannerPoolRegistryTransfer, chainID *big.Int, blk *domain.Block) *UpdateScannerPoolMessage {
	return &UpdateScannerPoolMessage{
		Message: regmsg.From(evt.Raw.TxHash.Hex(), blk, UpdateScannerPool),
		PoolID:  utils.PoolIDToString(evt.TokenId),
		Owner:   utils.StringPtr(evt.To.Hex()),
		ChainID: utils.Int64Ptr(chainID.Int64()),
	}
}

func NewScannerPoolMessageFromRegistration(evt *contract_scanner_pool_registry.ScannerPoolRegistryScannerPoolRegistered, owner string, blk *domain.Block) *UpdateScannerPoolMessage {
	return &UpdateScannerPoolMessage{
		Message: regmsg.From(evt.Raw.TxHash.Hex(), blk, UpdateScannerPool),
		PoolID:  utils.PoolIDToString(evt.ScannerPoolId),
		Owner:   utils.StringPtr(owner),
		ChainID: utils.Int64Ptr(evt.ChainId.Int64()),
	}
}

func NewScannerPoolMessageFromEnablement(evt *contract_scanner_pool_registry.ScannerPoolRegistryEnabledScannersChanged, owner string, chainID *big.Int, blk *domain.Block) *UpdateScannerPoolMessage {
	return &UpdateScannerPoolMessage{
		Message: regmsg.From(evt.Raw.TxHash.Hex(), blk, UpdateScannerPool),
		PoolID:  utils.PoolIDToString(evt.ScannerPoolId),
		Owner:   utils.StringPtr(owner),
		ChainID: utils.Int64Ptr(chainID.Int64()),
	}
}
