package cache

import (
	"context"
	"time"

	red "github.com/redis/go-redis/v9"
)

type redisCache struct {
	client    *red.Client
	keyPrefix string
}

func newRedisCache(client *red.Client, keyPrefix string) *redisCache {
	return &redisCache{client: client, keyPrefix: keyPrefix}
}

func (c *redisCache) Backend() string {
	return BackendRedis
}

func (c *redisCache) Get(ctx context.Context, key string) (string, bool, error) {
	value, err := c.client.Get(ctx, c.key(key)).Result()
	if err == nil {
		return value, true, nil
	}
	if err == red.Nil {
		return "", false, nil
	}
	return "", false, err
}

func (c *redisCache) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	if ttl <= 0 {
		return c.client.Set(ctx, c.key(key), value, 0).Err()
	}
	return c.client.Set(ctx, c.key(key), value, ttl).Err()
}

func (c *redisCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, c.key(key)).Err()
}

func (c *redisCache) Clear(ctx context.Context) error {
	pattern := c.keyPrefix + ":*"
	iter := c.client.Scan(ctx, 0, pattern, 0).Iterator()
	keys := make([]string, 0, 16)
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return err
	}
	if len(keys) == 0 {
		return nil
	}
	return c.client.Del(ctx, keys...).Err()
}

func (c *redisCache) key(key string) string {
	return c.keyPrefix + ":" + key
}

var _ Cache = (*redisCache)(nil)
