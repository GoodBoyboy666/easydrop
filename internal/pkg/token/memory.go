package token

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type memoryStore struct {
	mu        sync.Mutex
	byToken   map[string]*Record
	byUserKey map[string]string
}

func newMemoryStore() *memoryStore {
	return &memoryStore{
		byToken:   make(map[string]*Record),
		byUserKey: make(map[string]string),
	}
}

func (s *memoryStore) Save(_ context.Context, record *Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	userKey := memoryUserKindKey(record.UserID, record.Kind)
	if oldToken, ok := s.byUserKey[userKey]; ok && oldToken != record.Token {
		delete(s.byToken, oldToken)
	}

	cloned, err := cloneRecord(record)
	if err != nil {
		return fmt.Errorf("复制 token 记录失败: %w", err)
	}

	s.byToken[record.Token] = cloned
	s.byUserKey[userKey] = record.Token
	return nil
}

func (s *memoryStore) Consume(_ context.Context, kind, tokenValue string, now time.Time) (*Record, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	record, ok := s.byToken[tokenValue]
	if !ok {
		return nil, ErrTokenNotFound
	}

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

	delete(s.byToken, record.Token)
	delete(s.byUserKey, memoryUserKindKey(record.UserID, record.Kind))
}

func memoryUserKindKey(userID uint, kind string) string {
	return fmt.Sprintf("%d:%s", userID, kind)
}
