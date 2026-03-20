package storage

import (
	"context"
	"errors"
	"os"
	"path/filepath"
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

func TestManagerGetSizeForLocalStorage(t *testing.T) {
	basePath := filepath.Join(t.TempDir(), "uploads")
	manager, err := NewManager(&Config{
		Backend: BackendLocal,
		Local: LocalConfig{
			BasePath: basePath,
		},
	})
	if err != nil {
		t.Fatalf("NewManager returned error: %v", err)
	}

	key := "avatar/42/profile.png"
	content := []byte("avatar-bytes")
	if err := manager.Upload(context.Background(), key, content, "image/png"); err != nil {
		t.Fatalf("Upload returned error: %v", err)
	}

	size, err := manager.GetSize(context.Background(), key)
	if err != nil {
		t.Fatalf("GetSize returned error: %v", err)
	}
	if size != int64(len(content)) {
		t.Fatalf("expected size %d, got %d", len(content), size)
	}
}

func TestManagerGetSizeRejectsEmptyObjectKey(t *testing.T) {
	manager, err := NewManager(&Config{
		Backend: BackendLocal,
		Local: LocalConfig{
			BasePath: t.TempDir(),
		},
	})
	if err != nil {
		t.Fatalf("NewManager returned error: %v", err)
	}

	_, err = manager.GetSize(context.Background(), "  ")
	if !errors.Is(err, ErrEmptyObjectKey) {
		t.Fatalf("expected ErrEmptyObjectKey, got %v", err)
	}
}

func TestLocalStorageGetSizeMissingObject(t *testing.T) {
	backend, err := NewLocalStorage(LocalConfig{BasePath: t.TempDir()})
	if err != nil {
		t.Fatalf("NewLocalStorage returned error: %v", err)
	}

	_, err = backend.GetSize(context.Background(), "missing/file.png")
	if err == nil {
		t.Fatal("expected GetSize to fail for missing object")
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected not-exist error, got %v", err)
	}
}
