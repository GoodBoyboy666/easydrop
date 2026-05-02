package webauthn

import (
	"context"
	"time"

	wa "github.com/go-webauthn/webauthn/webauthn"
	"github.com/jellydator/ttlcache/v3"
)

// memorySessionStore 基于内存 (ttlcache) 的会话存储实现。
// 在 Redis 不可用时作为后备方案，数据仅存在于当前进程内存中。
type memorySessionStore struct {
	cache *ttlcache.Cache[string, []byte]
}

// newMemorySessionStore 创建内存会话存储实例。
// 内部使用 ttlcache 实现带 TTL 的键值存储，自动过期清理。
func newMemorySessionStore() *memorySessionStore {
	c := ttlcache.New[string, []byte](
		ttlcache.WithTTL[string, []byte](5 * time.Minute),
	)
	go c.Start()
	return &memorySessionStore{cache: c}
}

// Save 将会话数据序列化为 JSON 后存入内存缓存。
func (s *memorySessionStore) Save(ctx context.Context, sessionID string, data *wa.SessionData, ttl time.Duration) error {
	raw, err := serializeSessionData(data)
	if err != nil {
		return err
	}
	s.cache.Set(sessionID, raw, ttl)
	return nil
}

// Get 从内存缓存中获取并反序列化会话数据。
func (s *memorySessionStore) Get(ctx context.Context, sessionID string) (*wa.SessionData, error) {
	item := s.cache.Get(sessionID)
	if item == nil {
		return nil, errSessionNotFound
	}
	return deserializeSessionData(item.Value())
}

// Delete 从内存缓存中删除指定会话。
func (s *memorySessionStore) Delete(ctx context.Context, sessionID string) error {
	s.cache.Delete(sessionID)
	return nil
}
