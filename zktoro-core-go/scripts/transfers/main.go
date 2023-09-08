package main

import (
	"context"
	"math/big"
	"os"

	log "github.com/sirupsen/logrus"
	"zktoro/zktoro-core-go/contracts/generated/contract_zktoro_staking_0_1_1"
	rd "zktoro/zktoro-core-go/domain/registry"
	"zktoro/zktoro-core-go/domain/registry/regmsg"
	"zktoro/zktoro-core-go/registry"
)

// this script prints any stake transfers between two block ranges+
func main() {
	ctx := context.Background()
	l, err := registry.NewListener(ctx, registry.ListenerConfig{
		JsonRpcURL: os.Getenv("POLYGON_JSON_RPC"),
		Handlers: registry.Handlers{
			TransferSharesHandlers: regmsg.Handlers(
				func(ctx context.Context, logger *log.Entry, msg *rd.TransferSharesMessage) error {
					log.WithFields(log.Fields{
						"to":     msg.To,
						"from":   msg.From,
						"amount": msg.Amount,
						"type":   msg.StakeType,
						"burn":   msg.IsBurn(),
						"mint":   msg.IsMint(),
					}).Info("event")
					return nil
				},
			),
		},
		ContractFilter: &registry.ContractFilter{
			ZktoroStaking: true,
		},
		Topics: []string{contract_zktoro_staking_0_1_1.TransferSingleTopic, contract_zktoro_staking_0_1_1.TransferBatchTopic},
	})
	if err != nil {
		panic(err)
	}

	start := big.NewInt(30172379)
	end := big.NewInt(30189948)

	if err := l.ProcessBlockRange(start, end); err != nil {
		panic(err)
	}
}
