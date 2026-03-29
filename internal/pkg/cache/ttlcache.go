package cache

import (
	"context"
	"time"

	"github.com/jellydator/ttlcache/v3"
)

type ttlCache struct {
	store *ttlcache.Cache[string, string]
}

func newTTLCache() *ttlCache {
	return &ttlCache{store: ttlcache.New(
		ttlcache.WithDisableTouchOnHit[string, string](),
	)}
}

func (c *ttlCache) Backend() string {
	return BackendMemory
}

func (c *ttlCache) Get(_ context.Context, key string) (string, bool, error) {
	item := c.store.Get(key)
	if item == nil {
		return "", false, nil
	}

	return item.Value(), true, nil
}

func (c *ttlCache) Set(_ context.Context, key, value string, ttl time.Duration) error {
	if ttl <= 0 {
		c.store.Set(key, value, ttlcache.NoTTL)
		return nil
	}

	c.store.Set(key, value, ttl)
	return nil
}

func (c *ttlCache) Delete(_ context.Context, key string) error {
	c.store.Delete(key)
	return nil
}

func (c *ttlCache) Clear(_ context.Context) error {
	c.store.DeleteAll()
	return nil
}

var _ Cache = (*ttlCache)(nil)
