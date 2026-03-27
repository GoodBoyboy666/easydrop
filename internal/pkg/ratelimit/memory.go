package ratelimit

import (
	"context"
	"sync"
	"time"
)

type memoryEntry struct {
	count     int
	expiresAt time.Time
}

type memoryLimiter struct {
	mu        sync.Mutex
	keyPrefix string
	now       func() time.Time
	items     map[string]memoryEntry
}

func newMemoryLimiter(cfg *Config) *memoryLimiter {
	return &memoryLimiter{
		keyPrefix: normalizeLimiterKeyPrefix(cfg),
		now:       time.Now,
		items:     make(map[string]memoryEntry),
	}
}

func (l *memoryLimiter) Backend() string {
	return BackendMemory
}

func (l *memoryLimiter) Allow(_ context.Context, key string, rule Rule) (*Decision, error) {
	normalizedRule, normalizedKey, now, err := l.prepareRequest(key, rule)
	if err != nil {
		return nil, err
	}

	storageKey := l.storageKey(normalizedRule.Name, normalizedKey)

	l.mu.Lock()
	defer l.mu.Unlock()

	switch normalizedRule.Mode {
	case ModeCooldown:
		return l.allowCooldown(storageKey, normalizedRule.Interval, now), nil
	case ModeWindow:
		return l.allowWindow(storageKey, normalizedRule, now), nil
	default:
		return nil, ErrInvalidMode
	}
}

func (l *memoryLimiter) allowCooldown(storageKey string, interval time.Duration, now time.Time) *Decision {
	entry, found := l.items[storageKey]
	if found && entry.expiresAt.After(now) {
		return &Decision{
			Allowed:    false,
			Remaining:  0,
			RetryAfter: entry.expiresAt.Sub(now),
			ResetAt:    entry.expiresAt,
		}
	}

	resetAt := now.Add(interval)
	l.items[storageKey] = memoryEntry{count: 1, expiresAt: resetAt}
	return &Decision{
		Allowed:   true,
		Remaining: 0,
		ResetAt:   resetAt,
	}
}

func (l *memoryLimiter) allowWindow(storageKey string, rule Rule, now time.Time) *Decision {
	entry, found := l.items[storageKey]
	if !found || !entry.expiresAt.After(now) {
		resetAt := now.Add(rule.Interval)
		l.items[storageKey] = memoryEntry{count: 1, expiresAt: resetAt}
		return &Decision{
			Allowed:   true,
			Remaining: maxInt(rule.Limit-1, 0),
			ResetAt:   resetAt,
		}
	}

	if entry.count >= rule.Limit {
		return &Decision{
			Allowed:    false,
			Remaining:  0,
			RetryAfter: entry.expiresAt.Sub(now),
			ResetAt:    entry.expiresAt,
		}
	}

	entry.count++
	l.items[storageKey] = entry
	return &Decision{
		Allowed:   true,
		Remaining: maxInt(rule.Limit-entry.count, 0),
		ResetAt:   entry.expiresAt,
	}
}

func (l *memoryLimiter) prepareRequest(key string, rule Rule) (Rule, string, time.Time, error) {
	normalizedRule, err := normalizeRule(rule)
	if err != nil {
		return Rule{}, "", time.Time{}, err
	}

	normalizedKey, err := normalizeKey(key)
	if err != nil {
		return Rule{}, "", time.Time{}, err
	}

	nowFn := l.now
	if nowFn == nil {
		nowFn = time.Now
	}

	return normalizedRule, normalizedKey, nowFn().UTC(), nil
}

func (l *memoryLimiter) storageKey(ruleName string, key string) string {
	return l.keyPrefix + ":" + ruleName + ":" + key
}

func normalizeLimiterKeyPrefix(cfg *Config) string {
	if cfg == nil {
		return defaultKeyPrefix
	}
	return normalizeKeyPrefix(cfg.KeyPrefix)
}

func maxInt(value int, fallback int) int {
	if value > fallback {
		return value
	}
	return fallback
}

var _ Limiter = (*memoryLimiter)(nil)
