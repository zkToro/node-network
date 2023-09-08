package feeds

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
	"zktoro/zktoro-core-go/clients/graphql"
	"zktoro/zktoro-core-go/domain"
	"zktoro/zktoro-core-go/protocol"
)

type combinerCache struct {
	cache *cache.Cache
	path  string
}

func newCombinerCache(path string) (*combinerCache, error) {
	var alertCache *cache.Cache
	if path != "" {
		d, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("can not read combiner cache file: %v", err)
		}

		var m map[string]cache.Item

		err = json.Unmarshal(d, &m)
		if err != nil {
			m = make(map[string]cache.Item)

			tErr := os.RemoveAll(path)
			if tErr != nil {
				return nil, fmt.Errorf("can not remove malformed combiner cache, :%v", tErr)
			}

			_, tErr = os.Create(path)
			if tErr != nil {
				return nil, fmt.Errorf("can not create new combiner cache file :%v", tErr)
			}
			log.WithError(err).Warn("removed malformed combiner cache")
		}

		alertCache = cache.NewFrom(graphql.DefaultLastNMinutes*2, time.Minute, m)
	} else {
		alertCache = cache.New(graphql.DefaultLastNMinutes*2, time.Minute)
	}

	return &combinerCache{cache: alertCache, path: path}, nil
}

func (c *combinerCache) Exists(subscription *domain.CombinerBotSubscription, alert *protocol.AlertEvent) bool {
	_, exists := c.cache.Get(encodeAlertCacheKey(subscription.Subscriber.BotID, subscription.Subscriber.BotImage, alert.Alert.Hash))
	return exists
}

func (c *combinerCache) Set(subscription *domain.CombinerBotSubscription, alert *protocol.AlertEvent) {
	c.cache.Set(encodeAlertCacheKey(subscription.Subscriber.BotID, subscription.Subscriber.BotImage, alert.Alert.Hash), struct{}{}, cache.DefaultExpiration)
}

// DumpToFile dumps the current cache into a file in JSON format, so that the cache can be used in a persistent way.
func (c *combinerCache) DumpToFile(filePath string) error {
	// Marshal the cache's items into JSON format
	d, err := json.Marshal(c.cache.Items())
	if err != nil {
		return err
	}

	if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
		_, _ = os.Create(filePath)
	}

	// Write the JSON data to the specified file
	err = os.WriteFile(filePath, d, 0644)
	if err != nil {
		return err
	}

	return nil
}

// encodeAlertCacheKey must encode alerts to prevent missing subscriptions to the same target bot
// from several deployed bots
func encodeAlertCacheKey(subscriberBotID, image, alertHash string) string {
	return fmt.Sprintf("%s|%s|%s", subscriberBotID, image, alertHash)
}
