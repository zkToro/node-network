package domain

import (
	"math/big"
	"testing"

	"github.com/golang/protobuf/jsonpb"
	"github.com/stretchr/testify/assert"
	"zktoro/zktoro-core-go/protocol"
)

func intPtr(val int) *int {
	return &val
}

func TestTransactionEvent_ToMessage(t *testing.T) {
	blockHash := "0x8d2636ff603ef946d97ad797ed13afa31234a3412dacdfecfeb3247230eb1069"
	txHash := "0x99ed5a4e541454219b444250c5c25d0306e73834b185f3aeee3f9627f0cd64c2"
	txHash2 := "0x11ed5a4e541454219b444250c5c25d0306e73834b185f3aeee3f9627f0cd64c2"

	// these are checksum addresses, to confirm that logic lower-cases these
	from := "0xa7d8d9ef8D8Ce8992Df33D8b8CF4Aebabd5bD270"
	to := "0x9C025948e61aeB2EF99503c81d682045f07344c2"

	transferTopic := "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"
	addrTopic := "0x000000000000000000000000a5ca6f2d2d07fc983f954552962b3c21c2db0a9A"

	tt := TrackingTimestampsFromMessage(&protocol.TrackingTimestamps{
		Block:       "2022-01-02T15:04:05Z",
		Feed:        "2022-01-02T15:04:05Z",
		SourceAlert: "2022-01-02T15:04:05Z",
		BotRequest:  "2022-01-02T15:04:05Z",
		BotResponse: "2022-01-02T15:04:05Z",
	})

	evt := &TransactionEvent{
		BlockEvt: &BlockEvent{
			EventType: "block",
			ChainID:   big.NewInt(1),
			Block: &Block{
				BaseFeePerGas:    strPtr("0x1"),
				Difficulty:       strPtr("0x1"),
				ExtraData:        strPtr("0x1"),
				GasLimit:         strPtr("0x1"),
				GasUsed:          strPtr("0x1"),
				Hash:             blockHash,
				LogsBloom:        strPtr("0x1"),
				Miner:            strPtr("0x1"),
				MixHash:          strPtr("0x1"),
				Nonce:            strPtr("0x1"),
				Number:           "0x1",
				ParentHash:       "0xabcdef",
				ReceiptsRoot:     strPtr("0x1"),
				Sha3Uncles:       strPtr("0x1"),
				Size:             strPtr("0x1"),
				StateRoot:        strPtr("0x1"),
				Timestamp:        "0x12345",
				TotalDifficulty:  strPtr("0x1"),
				Transactions:     []Transaction{},
				TransactionsRoot: strPtr("0x1"),
				Uncles:           []*string{strPtr("0x1")},
			},
			Logs: []LogEntry{
				{
					Address:         strPtr(to),
					BlockHash:       &blockHash,
					BlockNumber:     strPtr("0x2"),
					TransactionHash: &txHash,
					Topics: []*string{
						&transferTopic,
						&addrTopic,
					},
				},
				{
					Address:         strPtr(to),
					BlockHash:       &blockHash,
					BlockNumber:     strPtr("0x2"),
					TransactionHash: &txHash2, // should ignore, because doesn't match tx
				},
			},
			Traces: []Trace{
				{
					Action:              TraceAction{To: &to, From: &from},
					BlockHash:           &blockHash,
					BlockNumber:         intPtr(1),
					TransactionHash:     &txHash,
					TransactionPosition: intPtr(5),
					Type:                "transaction",
				},
			},
			Timestamps: tt,
		},
		Transaction: &Transaction{
			BlockHash:            blockHash,
			BlockNumber:          "0x1",
			From:                 from,
			Gas:                  "0x2",
			GasPrice:             "0x3",
			Hash:                 txHash,
			Nonce:                "0x5",
			To:                   &to,
			MaxFeePerGas:         strPtr("0x3"),
			MaxPriorityFeePerGas: strPtr("0x4"),
		},
		Timestamps: tt,
	}
	msg, err := evt.ToMessage()
	assert.NoError(t, err, "error returned from ToMessage")

	js := jsonpb.Marshaler{}
	str, err := js.MarshalToString(msg)
	t.Log(str)

	// I manually checked this json, so this test just ensures this behavior continues
	expected := `{"transaction":{"nonce":"0x5","gasPrice":"0x3","gas":"0x2","to":"0x9c025948e61aeb2ef99503c81d682045f07344c2","hash":"0x99ed5a4e541454219b444250c5c25d0306e73834b185f3aeee3f9627f0cd64c2","from":"0xa7d8d9ef8d8ce8992df33d8b8cf4aebabd5bd270","maxFeePerGas":"0x3","maxPriorityFeePerGas":"0x4"},"receipt":{"status":"0x1","logs":[{"address":"0x9c025948e61aeb2ef99503c81d682045f07344c2","topics":["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef","0x000000000000000000000000a5ca6f2d2d07fc983f954552962b3c21c2db0a9A"],"blockNumber":"0x2","transactionHash":"0x99ed5a4e541454219b444250c5c25d0306e73834b185f3aeee3f9627f0cd64c2","blockHash":"0x8d2636ff603ef946d97ad797ed13afa31234a3412dacdfecfeb3247230eb1069"}],"transactionHash":"0x99ed5a4e541454219b444250c5c25d0306e73834b185f3aeee3f9627f0cd64c2","gasUsed":"0x2","blockHash":"0x8d2636ff603ef946d97ad797ed13afa31234a3412dacdfecfeb3247230eb1069","blockNumber":"0x1"},"network":{"chainId":"0x1"},"traces":[{"action":{"to":"0x9c025948e61aeb2ef99503c81d682045f07344c2","from":"0xa7d8d9ef8d8ce8992df33d8b8cf4aebabd5bd270"},"blockHash":"0x8d2636ff603ef946d97ad797ed13afa31234a3412dacdfecfeb3247230eb1069","blockNumber":"1","transactionHash":"0x99ed5a4e541454219b444250c5c25d0306e73834b185f3aeee3f9627f0cd64c2","transactionPosition":"5","type":"transaction"}],"addresses":{"0x9c025948e61aeb2ef99503c81d682045f07344c2":true,"0xa5ca6f2d2d07fc983f954552962b3c21c2db0a9a":true,"0xa7d8d9ef8d8ce8992df33d8b8cf4aebabd5bd270":true},"block":{"blockHash":"0x8d2636ff603ef946d97ad797ed13afa31234a3412dacdfecfeb3247230eb1069","blockNumber":"0x1","blockTimestamp":"0x12345","baseFeePerGas":"0x1"},"logs":[{"address":"0x9c025948e61aeb2ef99503c81d682045f07344c2","topics":["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef","0x000000000000000000000000a5ca6f2d2d07fc983f954552962b3c21c2db0a9A"],"blockNumber":"0x2","transactionHash":"0x99ed5a4e541454219b444250c5c25d0306e73834b185f3aeee3f9627f0cd64c2","blockHash":"0x8d2636ff603ef946d97ad797ed13afa31234a3412dacdfecfeb3247230eb1069"}],"timestamps":{"block":"2022-01-02T15:04:05Z","feed":"2022-01-02T15:04:05Z","botRequest":"2022-01-02T15:04:05Z","botResponse":"2022-01-02T15:04:05Z","sourceAlert":"2022-01-02T15:04:05Z"},"txAddresses":{"0x9c025948e61aeb2ef99503c81d682045f07344c2":true,"0xa5ca6f2d2d07fc983f954552962b3c21c2db0a9a":true,"0xa7d8d9ef8d8ce8992df33d8b8cf4aebabd5bd270":true}}`
	assert.NoError(t, err, "error returned from json conversion")
	assert.Equal(t, expected, str)
}

func TestTransactionEvent_ToMessage_ContractDeploy(t *testing.T) {
	blockHash := "0x8d2636ff603ef946d97ad797ed13afa31234a3412dacdfecfeb3247230eb1069"
	txHash := "0x99ed5a4e541454219b444250c5c25d0306e73834b185f3aeee3f9627f0cd64c2"

	// these are checksum addresses, to confirm that logic lower-cases these
	from := "0x2f73b85d78b38e90c64830c06a96be318a6e2154"

	tt := TrackingTimestampsFromMessage(&protocol.TrackingTimestamps{
		Block:       "2022-01-02T15:04:05Z",
		Feed:        "2022-01-02T15:04:05Z",
		BotRequest:  "2022-01-02T15:04:05Z",
		BotResponse: "2022-01-02T15:04:05Z",
		SourceAlert: "2022-01-02T15:04:05Z",
	})

	evt := &TransactionEvent{
		BlockEvt: &BlockEvent{
			EventType: "block",
			ChainID:   big.NewInt(1),
			Block: &Block{
				BaseFeePerGas:    strPtr("0x2"),
				Difficulty:       strPtr("0x1"),
				ExtraData:        strPtr("0x1"),
				GasLimit:         strPtr("0x1"),
				GasUsed:          strPtr("0x1"),
				Hash:             blockHash,
				LogsBloom:        strPtr("0x1"),
				Miner:            strPtr("0x1"),
				MixHash:          strPtr("0x1"),
				Nonce:            strPtr("0x1"),
				Number:           "0x8",
				ParentHash:       "0xabcdef",
				ReceiptsRoot:     strPtr("0x1"),
				Sha3Uncles:       strPtr("0x1"),
				Size:             strPtr("0x1"),
				StateRoot:        strPtr("0x1"),
				Timestamp:        "0x12345",
				TotalDifficulty:  strPtr("0x1"),
				Transactions:     []Transaction{},
				TransactionsRoot: strPtr("0x1"),
				Uncles:           []*string{strPtr("0x1")},
			},
			Logs:       []LogEntry{},
			Traces:     []Trace{},
			Timestamps: tt,
		},
		Transaction: &Transaction{
			BlockHash:            blockHash,
			BlockNumber:          "0x1",
			From:                 from,
			Gas:                  "0x2",
			GasPrice:             "0x3",
			Hash:                 txHash,
			Nonce:                "0x8",
			MaxFeePerGas:         strPtr("0x4"),
			MaxPriorityFeePerGas: strPtr("0x5"),
		},
		Timestamps: tt,
	}
	msg, err := evt.ToMessage()
	assert.NoError(t, err, "error returned from ToMessage")

	js := jsonpb.Marshaler{}
	str, err := js.MarshalToString(msg)
	t.Log(str)

	// I manually checked this json, so this test just ensures this behavior continues
	expected := `{"transaction":{"nonce":"0x8","gasPrice":"0x3","gas":"0x2","hash":"0x99ed5a4e541454219b444250c5c25d0306e73834b185f3aeee3f9627f0cd64c2","from":"0x2f73b85d78b38e90c64830c06a96be318a6e2154","maxFeePerGas":"0x4","maxPriorityFeePerGas":"0x5"},"receipt":{"status":"0x1","transactionHash":"0x99ed5a4e541454219b444250c5c25d0306e73834b185f3aeee3f9627f0cd64c2","contractAddress":"0xbf2920129f83d75dec95d97a879942cce3dcd387","gasUsed":"0x2","blockHash":"0x8d2636ff603ef946d97ad797ed13afa31234a3412dacdfecfeb3247230eb1069","blockNumber":"0x8"},"network":{"chainId":"0x1"},"addresses":{"0x2f73b85d78b38e90c64830c06a96be318a6e2154":true,"0xbf2920129f83d75dec95d97a879942cce3dcd387":true},"block":{"blockHash":"0x8d2636ff603ef946d97ad797ed13afa31234a3412dacdfecfeb3247230eb1069","blockNumber":"0x8","blockTimestamp":"0x12345","baseFeePerGas":"0x2"},"isContractDeployment":true,"contractAddress":"0xbf2920129f83d75dec95d97a879942cce3dcd387","timestamps":{"block":"2022-01-02T15:04:05Z","feed":"2022-01-02T15:04:05Z","botRequest":"2022-01-02T15:04:05Z","botResponse":"2022-01-02T15:04:05Z","sourceAlert":"2022-01-02T15:04:05Z"},"txAddresses":{"0x2f73b85d78b38e90c64830c06a96be318a6e2154":true,"0xbf2920129f83d75dec95d97a879942cce3dcd387":true}}`
	assert.NoError(t, err, "error returned from json conversion")
	assert.Equal(t, expected, str)
}
