package cache

import (
	"context"
	"time"

	red "github.com/redis/go-redis/v9"
)

const (
	BackendMemory = "memory"
	BackendRedis  = "redis"

	defaultKeyPrefix  = "cache"
	redisProbeTimeout = 500 * time.Millisecond
)

// Cache 定义 KV 缓存能力。
type Cache interface {
	Backend() string
	Get(ctx context.Context, key string) (string, bool, error)
	Set(ctx context.Context, key, value string, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Clear(ctx context.Context) error
}

// NewCache 创建缓存实例。
// 当 Redis 客户端可用且启动时探活成功时使用 Redis，否则回退到内存实现。
func NewCache(redisClient *red.Client) (Cache, error) {
	if redisClient == nil {
		return newTTLCache(), nil
	}

	probeCtx, cancel := context.WithTimeout(context.Background(), redisProbeTimeout)
	defer cancel()
	if err := redisClient.Ping(probeCtx).Err(); err != nil {
		return newTTLCache(), nil
	}

	return newRedisCache(redisClient, defaultKeyPrefix), nil
}
