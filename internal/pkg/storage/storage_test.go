package storage

import (
	"strings"
	"testing"
	"time"
)

func TestBuildObjectKeyForFile(t *testing.T) {
	now := time.Date(2026, 3, 17, 10, 0, 0, 0, time.UTC)
	key, err := buildObjectKey(CategoryFile, 0, "Photo.PNG", now)
	if err != nil {
		t.Fatalf("buildObjectKey returned error: %v", err)
	}

	parts := strings.Split(key, "/")
	if len(parts) != 5 {
		t.Fatalf("unexpected file key format: %s", key)
	}
	if parts[0] != CategoryFile || parts[1] != "2026" || parts[2] != "03" || parts[3] != "17" {
		t.Fatalf("unexpected file key path: %s", key)
	}
	if !strings.HasSuffix(parts[4], ".png") {
		t.Fatalf("expected lower-case extension in file key: %s", key)
	}
	if strings.TrimSuffix(parts[4], ".png") == "" {
		t.Fatalf("uuid segment should not be empty")
	}
}

func TestBuildObjectKeyForAvatar(t *testing.T) {
	now := time.Date(2026, 3, 17, 10, 0, 0, 0, time.UTC)
	key, err := buildObjectKey(CategoryAvatar, 42, "avatar.jpg", now)
	if err != nil {
		t.Fatalf("buildObjectKey returned error: %v", err)
	}

	parts := strings.Split(key, "/")
	if len(parts) != 3 {
		t.Fatalf("unexpected avatar key format: %s", key)
	}
	if parts[0] != CategoryAvatar || parts[1] != "42" {
		t.Fatalf("unexpected avatar key path: %s", key)
	}
	if !strings.HasSuffix(parts[2], ".jpg") {
		t.Fatalf("expected avatar extension in key: %s", key)
	}
	if strings.TrimSuffix(parts[2], ".jpg") == "" {
		t.Fatalf("uuid segment should not be empty")
	}
}

func TestBuildObjectKeyForAvatarRequiresUserID(t *testing.T) {
	now := time.Date(2026, 3, 17, 10, 0, 0, 0, time.UTC)
	if _, err := buildObjectKey(CategoryAvatar, 0, "avatar.jpg", now); err == nil {
		t.Fatalf("expected error when user id is zero")
	}
}

func TestBuildObjectKeyWithoutExtension(t *testing.T) {
	now := time.Date(2026, 3, 17, 10, 0, 0, 0, time.UTC)
	key, err := buildObjectKey(CategoryFile, 0, "README", now)
	if err != nil {
		t.Fatalf("buildObjectKey returned error: %v", err)
	}

	parts := strings.Split(key, "/")
	if len(parts) != 5 {
		t.Fatalf("unexpected file key format: %s", key)
	}
	if strings.Contains(parts[4], ".") {
		t.Fatalf("expected filename without extension: %s", key)
	}
}

func TestBuildObjectKeyUsesLastExtension(t *testing.T) {
	now := time.Date(2026, 3, 17, 10, 0, 0, 0, time.UTC)
	key, err := buildObjectKey(CategoryFile, 0, "archive.tar.GZ", now)
	if err != nil {
		t.Fatalf("buildObjectKey returned error: %v", err)
	}

	if !strings.HasSuffix(key, ".gz") {
		t.Fatalf("expected only last extension to be kept: %s", key)
	}
}
