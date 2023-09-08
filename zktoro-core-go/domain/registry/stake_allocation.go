package registry

import (
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
	"zktoro/zktoro-core-go/contracts/merged/contract_stake_allocator"
	"zktoro/zktoro-core-go/domain"
	"zktoro/zktoro-core-go/domain/registry/regmsg"
	"zktoro/zktoro-core-go/utils"
)

const (
	ScannerPoolAllocatedStake = "ScannerPoolAllocatedStake"
)

type ScannerPoolAllocationMessage struct {
	regmsg.Message
	PoolID          string `json:"poolId"`
	ChangeAmount    string `json:"changeAmount"`
	TotalAmount     string `json:"totalAmount"`
	Increase        bool   `json:"increase"`
	StakePerScanner string `json:"stakePerScanner"`
}

func (spam *ScannerPoolAllocationMessage) LogFields() logrus.Fields {
	return logrus.Fields{}
}

func NewScannerPoolAllocationMessage(l types.Log, blk *domain.Block, evt *contract_stake_allocator.StakeAllocatorAllocatedStake, stakePerManaged *big.Int) *ScannerPoolAllocationMessage {
	return &ScannerPoolAllocationMessage{
		Message:         regmsg.From(l.TxHash.Hex(), blk, ScannerPoolAllocatedStake),
		PoolID:          utils.PoolIDToString(evt.Subject),
		ChangeAmount:    evt.Amount.String(),
		TotalAmount:     evt.TotalAllocated.String(),
		Increase:        evt.Increase,
		StakePerScanner: stakePerManaged.String(),
	}
}
