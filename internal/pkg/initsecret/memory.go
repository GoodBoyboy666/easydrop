package initsecret

import (
	"context"
	"sync"
	"time"
)

type memoryStore struct {
	mu     sync.RWMutex
	secret string
}

func newMemoryStore() *memoryStore {
	return &memoryStore{}
}

func (s *memoryStore) Get(_ context.Context, key string) (string, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if key != storageKey || s.secret == "" {
		return "", false, nil
	}
	return s.secret, true, nil
}

func (s *memoryStore) SetNX(_ context.Context, key, value string, _ time.Duration) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if key != storageKey {
		return false, nil
	}
	if s.secret != "" {
		return false, nil
	}
	s.secret = value
	return true, nil
}

var _ store = (*memoryStore)(nil)
