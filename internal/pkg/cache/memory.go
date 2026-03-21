package cache

import (
	"context"
	"sync"
	"time"
)

type memoryEntry struct {
	value     string
	expiresAt time.Time
	permanent bool
}

type memoryCache struct {
	mu    sync.RWMutex
	items map[string]memoryEntry
}

func newMemoryCache() *memoryCache {
	return &memoryCache{items: make(map[string]memoryEntry)}
}

func (c *memoryCache) Backend() string {
	return BackendMemory
}

func (c *memoryCache) Get(_ context.Context, key string) (string, bool, error) {
	now := time.Now().UTC()

	c.mu.RLock()
	entry, ok := c.items[key]
	c.mu.RUnlock()
	if !ok {
		return "", false, nil
	}

	if !entry.permanent && !entry.expiresAt.After(now) {
		c.mu.Lock()
		delete(c.items, key)
		c.mu.Unlock()
		return "", false, nil
	}

	return entry.value, true, nil
}

func (c *memoryCache) Set(_ context.Context, key, value string, ttl time.Duration) error {
	entry := memoryEntry{value: value, permanent: ttl <= 0}
	if ttl > 0 {
		entry.expiresAt = time.Now().UTC().Add(ttl)
	}

	c.mu.Lock()
	c.items[key] = entry
	c.mu.Unlock()
	return nil
}

func (c *memoryCache) Delete(_ context.Context, key string) error {
	c.mu.Lock()
	delete(c.items, key)
	c.mu.Unlock()
	return nil
}

func (c *memoryCache) Clear(_ context.Context) error {
	c.mu.Lock()
	c.items = make(map[string]memoryEntry)
	c.mu.Unlock()
	return nil
}

var _ Cache = (*memoryCache)(nil)
