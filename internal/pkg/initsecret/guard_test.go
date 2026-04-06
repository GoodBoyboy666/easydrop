package initsecret

import "testing"

func TestGuardEnsureAndValidate(t *testing.T) {
	t.Parallel()

	guard := NewGuard()

	secret, err := guard.EnsureSecret()
	if err != nil {
		t.Fatalf("EnsureSecret error: %v", err)
	}
	if secret == "" {
		t.Fatal("expected generated secret")
	}

	secondSecret, err := guard.EnsureSecret()
	if err != nil {
		t.Fatalf("EnsureSecret second call error: %v", err)
	}
	if secondSecret != secret {
		t.Fatalf("expected stable secret, got %q and %q", secret, secondSecret)
	}

	if err := guard.Validate(secret); err != nil {
		t.Fatalf("Validate error: %v", err)
	}
}

func TestGuardValidateRejectsMissingAndWrongSecret(t *testing.T) {
	t.Parallel()

	guard := NewGuard()
	if _, err := guard.EnsureSecret(); err != nil {
		t.Fatalf("EnsureSecret error: %v", err)
	}

	if err := guard.Validate(""); err != ErrRequired {
		t.Fatalf("expected ErrRequired, got %v", err)
	}
	if err := guard.Validate("wrong-secret"); err != ErrInvalid {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
}
