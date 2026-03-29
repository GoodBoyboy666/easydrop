package token

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jellydator/ttlcache/v3"
)

type memoryStore struct {
	mu        sync.Mutex
	byToken   *ttlcache.Cache[string, *Record]
	byUserKey *ttlcache.Cache[string, string]
}

func newMemoryStore() *memoryStore {
	return &memoryStore{
		byToken: ttlcache.New(
			ttlcache.WithDisableTouchOnHit[string, *Record](),
		),
		byUserKey: ttlcache.New(
			ttlcache.WithDisableTouchOnHit[string, string](),
		),
	}
}

func (s *memoryStore) Save(_ context.Context, record *Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	userKey := memoryUserKindKey(record.UserID, record.Kind)
	if oldTokenItem := s.byUserKey.Get(userKey); oldTokenItem != nil {
		oldToken := oldTokenItem.Value()
		if oldToken != record.Token {
			s.byToken.Delete(oldToken)
		}
	}

	cloned, err := cloneRecord(record)
	if err != nil {
		return fmt.Errorf("复制 token 记录失败: %w", err)
	}

	s.byToken.Set(record.Token, cloned, ttlcache.NoTTL)
	s.byUserKey.Set(userKey, record.Token, ttlcache.NoTTL)
	return nil
}

func (s *memoryStore) Consume(_ context.Context, kind, tokenValue string, now time.Time) (*Record, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item := s.byToken.Get(tokenValue)
	if item == nil {
		return nil, ErrTokenNotFound
	}
	record := item.Value()

	if !record.ExpiresAt.After(now) {
		s.deleteLocked(record)
		return nil, ErrTokenExpired
	}

	if record.Kind != kind || record.Token != tokenValue {
		return nil, ErrTokenMismatch
	}

	s.deleteLocked(record)
	return cloneRecord(record)
}

func (s *memoryStore) deleteLocked(record *Record) {
	if record == nil {
		return
	}

	s.byToken.Delete(record.Token)
	s.byUserKey.Delete(memoryUserKindKey(record.UserID, record.Kind))
}

func memoryUserKindKey(userID uint, kind string) string {
	return fmt.Sprintf("%d:%s", userID, kind)
}
