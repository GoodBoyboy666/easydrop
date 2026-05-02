package webauthn

import (
	"context"
	"errors"
	"fmt"
	"time"

	wa "github.com/go-webauthn/webauthn/webauthn"
	"github.com/redis/go-redis/v9"
)

// errSessionNotFound 表示会话数据不存在或已过期。
var errSessionNotFound = errors.New("会话不存在")

// redisSessionKeyPrefix 是 Redis 中会话数据的键前缀，用于命名空间隔离。
const redisSessionKeyPrefix = "webauthn:session:"

// redisSessionStore 基于 Redis 的会话存储实现。
// 在 Redis 可用时优先使用，支持多实例共享会话数据。
type redisSessionStore struct {
	client *redis.Client
}

// newRedisSessionStore 创建 Redis 会话存储实例。
func newRedisSessionStore(client *redis.Client) *redisSessionStore {
	return &redisSessionStore{client: client}
}

// Save 将会话数据序列化为 JSON 后存入 Redis，并设置 TTL 过期时间。
func (s *redisSessionStore) Save(ctx context.Context, sessionID string, data *wa.SessionData, ttl time.Duration) error {
	raw, err := serializeSessionData(data)
	if err != nil {
		return err
	}
	key := redisSessionKeyPrefix + sessionID
	return s.client.Set(ctx, key, raw, ttl).Err()
}

// Get 从 Redis 中获取并反序列化会话数据。
// 若键不存在 (redis.Nil) 则返回 errSessionNotFound。
func (s *redisSessionStore) Get(ctx context.Context, sessionID string) (*wa.SessionData, error) {
	key := redisSessionKeyPrefix + sessionID
	raw, err := s.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, errSessionNotFound
		}
		return nil, fmt.Errorf("读取会话失败: %w", err)
	}
	return deserializeSessionData(raw)
}

// Delete 从 Redis 中删除指定会话数据。
func (s *redisSessionStore) Delete(ctx context.Context, sessionID string) error {
	key := redisSessionKeyPrefix + sessionID
	return s.client.Del(ctx, key).Err()
}
