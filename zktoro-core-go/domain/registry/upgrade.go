package registry

import (
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
	"zktoro/zktoro-core-go/contracts/merged/contract_dispatch"
	"zktoro/zktoro-core-go/domain"
	"zktoro/zktoro-core-go/domain/registry/regmsg"
)

const Upgrade = "Upgrade"
const ConfigurationChange = "ConfigurationChange"

type UpgradeMessage struct {
	regmsg.Message
	Proxy             string `json:"proxy"`
	NewImplementation string `json:"newImplementation"`
}

func (um *UpgradeMessage) LogFields() logrus.Fields {
	return logrus.Fields{"proxy": um.Proxy}
}

func NewUpgradeMessageFromUpgrade(evt *contract_dispatch.DispatchUpgraded, l types.Log, blk *domain.Block) *UpgradeMessage {
	return &UpgradeMessage{
		Message:           regmsg.From(evt.Raw.TxHash.Hex(), blk, Upgrade),
		Proxy:             strings.ToLower(l.Address.Hex()),
		NewImplementation: strings.ToLower(evt.Implementation.Hex()),
	}
}

func NewUpgradeMessageFromConfigurationChange(txHash common.Hash, l types.Log, blk *domain.Block) *UpgradeMessage {
	return &UpgradeMessage{
		Message: regmsg.From(txHash.Hex(), blk, ConfigurationChange),
		Proxy:   strings.ToLower(l.Address.Hex()),
	}
}
