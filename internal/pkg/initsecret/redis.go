package initsecret

import (
	"context"
	"time"

	red "github.com/redis/go-redis/v9"
)

type redisStore struct {
	client *red.Client
}

func (s *redisStore) Get(ctx context.Context, key string) (string, bool, error) {
	value, err := s.client.Get(ctx, key).Result()
	if err == nil {
		return value, true, nil
	}
	if err == red.Nil {
		return "", false, nil
	}
	return "", false, err
}

func (s *redisStore) SetNX(ctx context.Context, key, value string, ttl time.Duration) (bool, error) {
	return s.client.SetNX(ctx, key, value, ttl).Result()
}

var _ store = (*redisStore)(nil)
