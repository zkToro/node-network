package feeds

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"zktoro/zktoro-core-go/domain"
	clients "zktoro/zktoro-core-go/ethereum/mocks"
	"zktoro/zktoro-core-go/testutils"
	"zktoro/zktoro-core-go/utils"
)

func getTestTransactionFeed(t *testing.T, blockFeed BlockFeed) (*transactionFeed, *clients.MockClient) {
	blocks := make(chan *domain.BlockEvent, 1)
	txs := make(chan *domain.TransactionEvent, 1)
	ctrl := gomock.NewController(t)
	client := clients.NewMockClient(ctrl)
	cache := utils.NewCache(10000)
	return &transactionFeed{
		ctx:       context.Background(),
		blockFeed: blockFeed,
		cache:     cache,
		txCh:      txs,
		blockCh:   blocks,
		client:    client,
		workers:   1,
	}, client
}

func TestTransactionFeed_ForEachTransaction(t *testing.T) {
	bf := NewMockBlockFeed([]*domain.BlockEvent{
		{
			EventType:  domain.EventTypeBlock,
			Block:      testutils.TestBlock(1, 2, 3),
			Timestamps: &domain.TrackingTimestamps{Block: time.Now().UTC()},
		},
		{
			EventType:  domain.EventTypeBlock,
			Block:      testutils.TestBlock(4, 5, 6, 6), // with duplicate
			Timestamps: &domain.TrackingTimestamps{Block: time.Now().UTC()},
		},
		{
			EventType:  domain.EventTypeBlock,
			Block:      testutils.TestBlock(), // empty
			Timestamps: &domain.TrackingTimestamps{Block: time.Now().UTC()},
		},
		{
			EventType:  domain.EventTypeBlock,
			Block:      testutils.TestBlock(7, 8, 9),
			Timestamps: &domain.TrackingTimestamps{Block: time.Now().UTC()},
		},
	})

	txFeed, _ := getTestTransactionFeed(t, bf)

	var evts []*domain.TransactionEvent
	err := txFeed.ForEachTransaction(func(evt *domain.BlockEvent) error { return nil }, func(evt *domain.TransactionEvent) error {
		evts = append(evts, evt)
		return nil
	})

	assert.Equal(t, endOfBlocks, err)
	assert.Len(t, evts, 9)
}

func TestTransactionFeed_ToMessage(t *testing.T) {
	var blockEvents []*domain.BlockEvent
	for i := 0; i < 1000; i++ {
		blockEvents = append(
			blockEvents, &domain.BlockEvent{
				EventType:  domain.EventTypeBlock,
				Block:      testutils.TestBlock(i),
				Timestamps: &domain.TrackingTimestamps{Block: time.Now().UTC()},
			},
		)
	}
	bf := NewMockBlockFeed(blockEvents)

	txFeed, _ := getTestTransactionFeed(t, bf)

	var result *domain.TransactionEvent
	err := txFeed.ForEachTransaction(
		func(evt *domain.BlockEvent) error { return nil }, func(evt *domain.TransactionEvent) error {
			result = evt
			msg, err := result.ToMessage()
			assert.NoError(t, err)
			assert.Equal(t, result.Transaction.Hash, msg.Transaction.Hash)
			return nil
		},
	)
	assert.Error(t, err, endOfBlocks)
	assert.Equal(t, endOfBlocks, err)
}
