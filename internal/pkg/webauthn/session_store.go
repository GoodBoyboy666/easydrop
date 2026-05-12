package webauthn

import (
	"github.com/redis/go-redis/v9"
)

// newSessionStore 创建会话存储实例。
// 若 Redis 客户端可用则使用 Redis 实现（支持多实例共享），
// 否则回退到内存实现（单进程内使用）。
func newSessionStore(redisClient *redis.Client) SessionStore {
	if redisClient != nil {
		return newRedisSessionStore(redisClient)
	}
	return newMemorySessionStore()
}
