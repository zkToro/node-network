package scanner

import (
	"context"
	"testing"

	"zktoro/zktoro-core-go/clients/health"
	"zktoro/zktoro-core-go/protocol"
	"zktoro/zktoro-core-go/utils"

	"github.com/stretchr/testify/assert"
)

func TestTxAnalyzerService_createBloomFilter(t *testing.T) {
	type fields struct {
		ctx                context.Context
		cfg                TxAnalyzerServiceConfig
		lastInputActivity  health.TimeTracker
		lastOutputActivity health.TimeTracker
	}
	type args struct {
		finding *protocol.Finding
		event   *protocol.TransactionEvent
	}
	tests := []struct {
		name            string
		fields          fields
		args            args
		wantBloomFilter *protocol.BloomFilter
		wantErr         bool
	}{
		{
			name: "tx finding",
			args: args{
				finding: &protocol.Finding{Addresses: []string{"0xaaa"}},
				event: &protocol.TransactionEvent{
					Addresses: map[string]bool{
						"0xaaa": true,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				txAnalyzer := &TxAnalyzerService{}
				gotBloomFilter, err := txAnalyzer.createBloomFilter(tt.args.finding, tt.args.event)
				assert.Equal(t, tt.wantErr, err != nil)

				bf, err := utils.CreateBloomFilterFromProto(gotBloomFilter)
				assert.NoError(t, err)

				// check for finding addresses
				for _, findingAddr := range tt.args.finding.Addresses {
					assert.True(t, bf.Test([]byte(findingAddr)), findingAddr)
				}

				// check for tx addresses
				for txAddr := range tt.args.event.Addresses {
					assert.True(t, bf.Test([]byte(txAddr)), txAddr)
				}
			},
		)
	}
}
