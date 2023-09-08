package registry

import (
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	log "github.com/sirupsen/logrus"
	"zktoro/zktoro-core-go/domain/registry"
)

func equalsAddress(address common.Address, addr string) bool {
	return strings.EqualFold(address.Hex(), addr)
}

func getTopic(log types.Log) string {
	if len(log.Topics) == 0 {
		return ""
	}
	return strings.ToLower(log.Topics[0].Hex())
}

func getLoggerForLog(le types.Log) *log.Entry {
	return log.WithFields(log.Fields{
		"address":     strings.ToLower(le.Address.Hex()),
		"txHash":      le.TxHash.Hex(),
		"blockNumber": le.BlockNumber,
		"blockHash":   le.BlockHash.Hex(),
		"topic":       getTopic(le),
	})
}

func getAllContractAddrs(regContracts registry.RegistryContracts) []string {
	addrs := []string{
		regContracts.AgentRegistry.Hex(),
		regContracts.ScannerRegistry.Hex(),
		regContracts.Dispatch.Hex(),
		regContracts.ZktoroStaking.Hex(),
		regContracts.ScannerNodeVersion.Hex(),
		regContracts.ScannerPoolRegistry.Hex(),
		regContracts.StakeAllocator.Hex(),
		regContracts.Rewards.Hex(),
	}
	return addrs
}

func isAddrIn(addrs []string, ethAddr common.Address) bool {
	for _, addr := range addrs {
		if equalsAddress(ethAddr, addr) {
			return true
		}
	}
	return false
}
