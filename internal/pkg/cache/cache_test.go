package cache

import (
	"context"
	"testing"
	"time"

	red "github.com/redis/go-redis/v9"
)

func TestNewCacheWithoutRedisUsesMemory(t *testing.T) {
	c, err := NewCache(nil)
	if err != nil {
		t.Fatalf("NewCache returned error: %v", err)
	}
	if c.Backend() != BackendMemory {
		t.Fatalf("unexpected backend: %s", c.Backend())
	}
}

func TestNewCacheFallsBackToMemoryWhenRedisUnavailable(t *testing.T) {
	client := red.NewClient(&red.Options{
		Addr:         "127.0.0.1:0",
		DialTimeout:  50 * time.Millisecond,
		ReadTimeout:  50 * time.Millisecond,
		WriteTimeout: 50 * time.Millisecond,
	})
	t.Cleanup(func() { _ = client.Close() })

	c, err := NewCache(client)
	if err != nil {
		t.Fatalf("NewCache returned error: %v", err)
	}
	if c.Backend() != BackendMemory {
		t.Fatalf("expected memory backend, got: %s", c.Backend())
	}
}

func TestMemoryCacheTTL(t *testing.T) {
	c, err := NewCache(nil)
	if err != nil {
		t.Fatalf("NewCache returned error: %v", err)
	}

	ctx := context.Background()
	if err := c.Set(ctx, "k1", "v1", 20*time.Millisecond); err != nil {
		t.Fatalf("Set returned error: %v", err)
	}

	value, found, err := c.Get(ctx, "k1")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if !found || value != "v1" {
		t.Fatalf("unexpected get result: found=%v value=%s", found, value)
	}

	time.Sleep(30 * time.Millisecond)
	_, found, err = c.Get(ctx, "k1")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if found {
		t.Fatal("expected key expired")
	}
}

func TestMemoryCachePermanentAndClear(t *testing.T) {
	c, err := NewCache(nil)
	if err != nil {
		t.Fatalf("NewCache returned error: %v", err)
	}

	ctx := context.Background()
	if err := c.Set(ctx, "k1", "v1", 0); err != nil {
		t.Fatalf("Set returned error: %v", err)
	}

	time.Sleep(20 * time.Millisecond)
	value, found, err := c.Get(ctx, "k1")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if !found || value != "v1" {
		t.Fatalf("unexpected get result: found=%v value=%s", found, value)
	}

	if err := c.Delete(ctx, "k1"); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	_, found, err = c.Get(ctx, "k1")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if found {
		t.Fatal("expected deleted key not found")
	}

	if err := c.Set(ctx, "k2", "v2", 0); err != nil {
		t.Fatalf("Set returned error: %v", err)
	}
	if err := c.Clear(ctx); err != nil {
		t.Fatalf("Clear returned error: %v", err)
	}
	_, found, err = c.Get(ctx, "k2")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if found {
		t.Fatal("expected cleared key not found")
	}
}
