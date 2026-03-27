package ratelimit

import (
	"context"
	"time"
)

type redisLimiter struct {
	client    redisStore
	keyPrefix string
	now       func() time.Time
}

func newRedisLimiter(cfg *Config, client redisStore) (*redisLimiter, error) {
	if client == nil {
		return nil, ErrNilRedisProbe
	}

	return &redisLimiter{
		client:    client,
		keyPrefix: normalizeLimiterKeyPrefix(cfg),
		now:       time.Now,
	}, nil
}

func (l *redisLimiter) Backend() string {
	return BackendRedis
}

func (l *redisLimiter) Allow(ctx context.Context, key string, rule Rule) (*Decision, error) {
	normalizedRule, normalizedKey, now, err := l.prepareRequest(key, rule)
	if err != nil {
		return nil, err
	}

	storageKey := l.storageKey(normalizedRule.Name, normalizedKey)
	ctx = normalizeContext(ctx)

	switch normalizedRule.Mode {
	case ModeCooldown:
		return l.allowCooldown(ctx, storageKey, normalizedRule.Interval, now)
	case ModeWindow:
		return l.allowWindow(ctx, storageKey, normalizedRule, now)
	default:
		return nil, ErrInvalidMode
	}
}

func (l *redisLimiter) allowCooldown(ctx context.Context, storageKey string, interval time.Duration, now time.Time) (*Decision, error) {
	ok, err := l.client.SetNX(ctx, storageKey, "1", interval)
	if err != nil {
		return nil, err
	}
	if ok {
		return &Decision{
			Allowed:   true,
			Remaining: 0,
			ResetAt:   now.Add(interval),
		}, nil
	}

	ttl, err := l.ttlOrInterval(ctx, storageKey, interval)
	if err != nil {
		return nil, err
	}

	return &Decision{
		Allowed:    false,
		Remaining:  0,
		RetryAfter: ttl,
		ResetAt:    now.Add(ttl),
	}, nil
}

func (l *redisLimiter) allowWindow(ctx context.Context, storageKey string, rule Rule, now time.Time) (*Decision, error) {
	count, err := l.client.Incr(ctx, storageKey)
	if err != nil {
		return nil, err
	}
	if count == 1 {
		if err := l.client.Expire(ctx, storageKey, rule.Interval); err != nil {
			return nil, err
		}
	}

	ttl, err := l.ttlOrInterval(ctx, storageKey, rule.Interval)
	if err != nil {
		return nil, err
	}

	if count > int64(rule.Limit) {
		return &Decision{
			Allowed:    false,
			Remaining:  0,
			RetryAfter: ttl,
			ResetAt:    now.Add(ttl),
		}, nil
	}

	return &Decision{
		Allowed:   true,
		Remaining: maxInt(rule.Limit-int(count), 0),
		ResetAt:   now.Add(ttl),
	}, nil
}

func (l *redisLimiter) ttlOrInterval(ctx context.Context, storageKey string, interval time.Duration) (time.Duration, error) {
	ttl, err := l.client.TTL(ctx, storageKey)
	if err != nil {
		return 0, err
	}
	if ttl <= 0 {
		return interval, nil
	}
	return ttl, nil
}

func (l *redisLimiter) prepareRequest(key string, rule Rule) (Rule, string, time.Time, error) {
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

func (l *redisLimiter) storageKey(ruleName string, key string) string {
	return l.keyPrefix + ":" + ruleName + ":" + key
}

var _ Limiter = (*redisLimiter)(nil)
