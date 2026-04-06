package initsecret

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeStore struct {
	getFn   func(ctx context.Context, key string) (string, bool, error)
	setNXFn func(ctx context.Context, key, value string, ttl time.Duration) (bool, error)
}

func (f *fakeStore) Get(ctx context.Context, key string) (string, bool, error) {
	if f.getFn == nil {
		return "", false, nil
	}
	return f.getFn(ctx, key)
}

func (f *fakeStore) SetNX(ctx context.Context, key, value string, ttl time.Duration) (bool, error) {
	if f.setNXFn == nil {
		return true, nil
	}
	return f.setNXFn(ctx, key, value, ttl)
}

func TestGuardEnsureAndValidateWithMemoryStore(t *testing.T) {
	t.Parallel()

	guard := NewGuard(nil)

	secret, err := guard.EnsureSecret(context.Background())
	if err != nil {
		t.Fatalf("EnsureSecret error: %v", err)
	}
	if secret == "" {
		t.Fatal("expected generated secret")
	}

	secondSecret, err := guard.EnsureSecret(context.Background())
	if err != nil {
		t.Fatalf("EnsureSecret second call error: %v", err)
	}
	if secondSecret != secret {
		t.Fatalf("expected stable secret, got %q and %q", secret, secondSecret)
	}

	if err := guard.Validate(context.Background(), secret); err != nil {
		t.Fatalf("Validate error: %v", err)
	}
}

func TestGuardEnsureReturnsExistingStoreValueWhenSetNXFails(t *testing.T) {
	t.Parallel()

	callCount := 0
	guard := newGuardWithStore(&fakeStore{
		getFn: func(_ context.Context, key string) (string, bool, error) {
			callCount++
			if key != storageKey {
				t.Fatalf("unexpected key: %q", key)
			}
			if callCount == 1 {
				return "", false, nil
			}
			return "shared-secret", true, nil
		},
		setNXFn: func(_ context.Context, key, value string, ttl time.Duration) (bool, error) {
			if key != storageKey {
				t.Fatalf("unexpected key: %q", key)
			}
			if value == "" {
				t.Fatal("expected generated value")
			}
			if ttl != 0 {
				t.Fatalf("expected ttl 0, got %v", ttl)
			}
			return false, nil
		},
	})

	secret, err := guard.EnsureSecret(context.Background())
	if err != nil {
		t.Fatalf("EnsureSecret error: %v", err)
	}
	if secret != "shared-secret" {
		t.Fatalf("expected shared secret, got %q", secret)
	}
}

func TestGuardValidateRejectsMissingAndWrongSecret(t *testing.T) {
	t.Parallel()

	guard := NewGuard(nil)
	secret, err := guard.EnsureSecret(context.Background())
	if err != nil {
		t.Fatalf("EnsureSecret error: %v", err)
	}

	if err := guard.Validate(context.Background(), ""); err != ErrRequired {
		t.Fatalf("expected ErrRequired, got %v", err)
	}
	if err := guard.Validate(context.Background(), "wrong-secret"); err != ErrInvalid {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
	if err := guard.Validate(context.Background(), secret); err != nil {
		t.Fatalf("expected valid secret, got %v", err)
	}
}

func TestGuardValidatePropagatesStoreError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("redis down")
	guard := newGuardWithStore(&fakeStore{
		getFn: func(context.Context, string) (string, bool, error) {
			return "", false, expectedErr
		},
	})

	err := guard.Validate(context.Background(), "secret")
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected propagated store error, got %v", err)
	}
}
