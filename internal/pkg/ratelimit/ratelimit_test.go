package ratelimit

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type fakeRedisBackend struct {
	mu          sync.Mutex
	now         func() time.Time
	pingErr     error
	setNXErr    error
	incrErr     error
	expireErr   error
	ttlErr      error
	values      map[string]int64
	expirations map[string]time.Time
}

func newFakeRedisBackend(now func() time.Time) *fakeRedisBackend {
	return &fakeRedisBackend{
		now:         now,
		values:      make(map[string]int64),
		expirations: make(map[string]time.Time),
	}
}

func (f *fakeRedisBackend) Ping(context.Context) error {
	return f.pingErr
}

func (f *fakeRedisBackend) SetNX(_ context.Context, key, value string, ttl time.Duration) (bool, error) {
	if f.setNXErr != nil {
		return false, f.setNXErr
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	f.cleanupLocked(key)
	if _, ok := f.values[key]; ok {
		return false, nil
	}

	f.values[key] = 1
	f.expirations[key] = f.now().UTC().Add(ttl)
	return true, nil
}

func (f *fakeRedisBackend) Incr(_ context.Context, key string) (int64, error) {
	if f.incrErr != nil {
		return 0, f.incrErr
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	f.cleanupLocked(key)
	f.values[key]++
	return f.values[key], nil
}

func (f *fakeRedisBackend) Expire(_ context.Context, key string, ttl time.Duration) error {
	if f.expireErr != nil {
		return f.expireErr
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	if _, ok := f.values[key]; !ok {
		return nil
	}
	f.expirations[key] = f.now().UTC().Add(ttl)
	return nil
}

func (f *fakeRedisBackend) TTL(_ context.Context, key string) (time.Duration, error) {
	if f.ttlErr != nil {
		return 0, f.ttlErr
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	f.cleanupLocked(key)
	expiresAt, ok := f.expirations[key]
	if !ok {
		return -1, nil
	}
	return expiresAt.Sub(f.now().UTC()), nil
}

func (f *fakeRedisBackend) cleanupLocked(key string) {
	expiresAt, ok := f.expirations[key]
	if !ok || expiresAt.After(f.now().UTC()) {
		return
	}

	delete(f.values, key)
	delete(f.expirations, key)
}

func TestNewLimiterFallsBackToMemoryWhenRedisUnavailable(t *testing.T) {
	backend := newFakeRedisBackend(time.Now)
	backend.pingErr = errors.New("redis down")

	limiter, err := newLimiterWithRedis(&Config{KeyPrefix: "test"}, backend, backend)
	if err != nil {
		t.Fatalf("newLimiterWithRedis returned error: %v", err)
	}
	if limiter.Backend() != BackendMemory {
		t.Fatalf("expected memory backend, got %q", limiter.Backend())
	}
}

func TestNewLimiterUsesRedisWhenProbeSucceeds(t *testing.T) {
	backend := newFakeRedisBackend(time.Now)

	limiter, err := newLimiterWithRedis(&Config{KeyPrefix: "test"}, backend, backend)
	if err != nil {
		t.Fatalf("newLimiterWithRedis returned error: %v", err)
	}
	if limiter.Backend() != BackendRedis {
		t.Fatalf("expected redis backend, got %q", limiter.Backend())
	}
}

func TestMemoryLimiterCooldown(t *testing.T) {
	now := time.Unix(10, 0).UTC()
	limiter := newMemoryLimiter(&Config{KeyPrefix: "test"})
	limiter.now = func() time.Time { return now }

	rule := Rule{Name: "cooldown", Mode: ModeCooldown, Interval: time.Second}

	first, err := limiter.Allow(context.Background(), "ip:1.1.1.1", rule)
	if err != nil {
		t.Fatalf("first allow returned error: %v", err)
	}
	if !first.Allowed {
		t.Fatalf("expected first request allowed")
	}

	second, err := limiter.Allow(context.Background(), "ip:1.1.1.1", rule)
	if err != nil {
		t.Fatalf("second allow returned error: %v", err)
	}
	if second.Allowed {
		t.Fatalf("expected second request blocked during cooldown")
	}

	now = now.Add(2 * time.Second)
	third, err := limiter.Allow(context.Background(), "ip:1.1.1.1", rule)
	if err != nil {
		t.Fatalf("third allow returned error: %v", err)
	}
	if !third.Allowed {
		t.Fatalf("expected request allowed after cooldown")
	}
}

func TestMemoryLimiterWindow(t *testing.T) {
	now := time.Unix(20, 0).UTC()
	limiter := newMemoryLimiter(&Config{KeyPrefix: "test"})
	limiter.now = func() time.Time { return now }

	rule := Rule{Name: "window", Mode: ModeWindow, Interval: time.Minute, Limit: 2}

	first, err := limiter.Allow(context.Background(), "user:7", rule)
	if err != nil {
		t.Fatalf("first allow returned error: %v", err)
	}
	if !first.Allowed || first.Remaining != 1 {
		t.Fatalf("unexpected first decision: %+v", first)
	}

	second, err := limiter.Allow(context.Background(), "user:7", rule)
	if err != nil {
		t.Fatalf("second allow returned error: %v", err)
	}
	if !second.Allowed || second.Remaining != 0 {
		t.Fatalf("unexpected second decision: %+v", second)
	}

	third, err := limiter.Allow(context.Background(), "user:7", rule)
	if err != nil {
		t.Fatalf("third allow returned error: %v", err)
	}
	if third.Allowed {
		t.Fatalf("expected third request blocked")
	}

	now = now.Add(61 * time.Second)
	fourth, err := limiter.Allow(context.Background(), "user:7", rule)
	if err != nil {
		t.Fatalf("fourth allow returned error: %v", err)
	}
	if !fourth.Allowed {
		t.Fatalf("expected request allowed after window reset")
	}
}

func TestRedisLimiterWindow(t *testing.T) {
	now := time.Unix(30, 0).UTC()
	backend := newFakeRedisBackend(func() time.Time { return now })
	limiter, err := newRedisLimiter(&Config{KeyPrefix: "test"}, backend)
	if err != nil {
		t.Fatalf("newRedisLimiter returned error: %v", err)
	}
	limiter.now = func() time.Time { return now }

	rule := Rule{Name: "window", Mode: ModeWindow, Interval: time.Minute, Limit: 2}

	for i := 0; i < 2; i++ {
		decision, err := limiter.Allow(context.Background(), "user:9", rule)
		if err != nil {
			t.Fatalf("allow #%d returned error: %v", i+1, err)
		}
		if !decision.Allowed {
			t.Fatalf("expected allow #%d to pass", i+1)
		}
	}

	blocked, err := limiter.Allow(context.Background(), "user:9", rule)
	if err != nil {
		t.Fatalf("blocked allow returned error: %v", err)
	}
	if blocked.Allowed {
		t.Fatalf("expected request blocked after window limit")
	}

	now = now.Add(61 * time.Second)
	afterReset, err := limiter.Allow(context.Background(), "user:9", rule)
	if err != nil {
		t.Fatalf("after reset allow returned error: %v", err)
	}
	if !afterReset.Allowed {
		t.Fatalf("expected request allowed after redis window reset")
	}
}

func TestLimiterWindowConcurrentRequests(t *testing.T) {
	now := time.Unix(40, 0).UTC()
	backend := newFakeRedisBackend(func() time.Time { return now })
	limiter, err := newRedisLimiter(&Config{KeyPrefix: "test"}, backend)
	if err != nil {
		t.Fatalf("newRedisLimiter returned error: %v", err)
	}
	limiter.now = func() time.Time { return now }

	rule := Rule{Name: "window", Mode: ModeWindow, Interval: time.Minute, Limit: 10}

	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		allowed int
		blocked int
	)

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			decision, err := limiter.Allow(context.Background(), "user:15", rule)
			if err != nil {
				t.Errorf("allow returned error: %v", err)
				return
			}

			mu.Lock()
			defer mu.Unlock()
			if decision.Allowed {
				allowed++
			} else {
				blocked++
			}
		}()
	}
	wg.Wait()

	if allowed != 10 {
		t.Fatalf("expected 10 allowed requests, got %d", allowed)
	}
	if blocked != 10 {
		t.Fatalf("expected 10 blocked requests, got %d", blocked)
	}
}
