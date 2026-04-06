package main

import (
	"bytes"
	"context"
	"errors"
	"log"
	"strings"
	"testing"

	"easydrop/internal/di"
	"easydrop/internal/dto"
	"easydrop/internal/pkg/initsecret"
	"easydrop/internal/service"

	"github.com/pterm/pterm"
)

func TestPrintBuildInfoBanner(t *testing.T) {
	originalName := appDisplayName
	originalVersion := appVersion
	originalBuildTime := buildTime
	originalCommit := gitCommit
	t.Cleanup(func() {
		appDisplayName = originalName
		appVersion = originalVersion
		buildTime = originalBuildTime
		gitCommit = originalCommit
	})

	appDisplayName = "EasyDrop"
	appVersion = "v1.2.3"
	buildTime = "2026-04-04T12:34:56Z"
	gitCommit = "abcdef1234567890"

	var buf bytes.Buffer
	printBuildInfoBanner(&buf)

	output := pterm.RemoveColorFromString(buf.String())
	expected := []string{
		"Program    : EasyDrop",
		"Version    : v1.2.3",
		"Build Time : 2026-04-04T12:34:56Z",
		"Commit     : abcdef1234567890",
		"EasyDrop Runtime",
		"███████",
		"██████",
	}

	for _, item := range expected {
		if !strings.Contains(output, item) {
			t.Fatalf("expected banner to contain %q, got %q", item, output)
		}
	}

	if !strings.Contains(output, "┌") || !strings.Contains(output, "┘") {
		t.Fatalf("expected pterm box border in output, got %q", output)
	}
}

type mainTestInitService struct {
	statusFn func(ctx context.Context) (*dto.InitStatusResult, error)
}

func (m *mainTestInitService) GetStatus(ctx context.Context) (*dto.InitStatusResult, error) {
	if m.statusFn == nil {
		return &dto.InitStatusResult{}, nil
	}
	return m.statusFn(ctx)
}

func (m *mainTestInitService) Initialize(context.Context, dto.InitInput) error {
	return nil
}

type mainTestInitSecretGuard struct {
	ensureFn func(ctx context.Context) (string, error)
}

func (m *mainTestInitSecretGuard) EnsureSecret(ctx context.Context) (string, error) {
	if m.ensureFn == nil {
		return "secret-123", nil
	}
	return m.ensureFn(ctx)
}

func (m *mainTestInitSecretGuard) Validate(context.Context, string) error {
	return nil
}

func TestPrepareInitSecretPrintsSecretWhenUninitialized(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)

	app := &di.App{
		InitService: &mainTestInitService{
			statusFn: func(context.Context) (*dto.InitStatusResult, error) {
				return &dto.InitStatusResult{Initialized: false}, nil
			},
		},
		InitSecretGuard: &mainTestInitSecretGuard{
			ensureFn: func(context.Context) (string, error) {
				return "secret-123", nil
			},
		},
	}

	if err := prepareInitSecret(context.Background(), app, logger); err != nil {
		t.Fatalf("prepareInitSecret error: %v", err)
	}
	if !strings.Contains(buf.String(), "secret-123") {
		t.Fatalf("expected log to contain init secret, got %q", buf.String())
	}
}

func TestPrepareInitSecretSkipsSecretWhenInitialized(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)

	app := &di.App{
		InitService: &mainTestInitService{
			statusFn: func(context.Context) (*dto.InitStatusResult, error) {
				return &dto.InitStatusResult{Initialized: true}, nil
			},
		},
		InitSecretGuard: &mainTestInitSecretGuard{
			ensureFn: func(context.Context) (string, error) {
				t.Fatal("expected EnsureSecret not to be called")
				return "", nil
			},
		},
	}

	if err := prepareInitSecret(context.Background(), app, logger); err != nil {
		t.Fatalf("prepareInitSecret error: %v", err)
	}
	if buf.Len() != 0 {
		t.Fatalf("expected no log output, got %q", buf.String())
	}
}

func TestPrepareInitSecretReturnsStatusError(t *testing.T) {
	app := &di.App{
		InitService: &mainTestInitService{
			statusFn: func(context.Context) (*dto.InitStatusResult, error) {
				return nil, errors.New("boom")
			},
		},
		InitSecretGuard: &mainTestInitSecretGuard{},
	}

	err := prepareInitSecret(context.Background(), app, log.New(&bytes.Buffer{}, "", 0))
	if err == nil || !strings.Contains(err.Error(), "读取系统初始化状态失败") {
		t.Fatalf("expected wrapped status error, got %v", err)
	}
}

var _ service.InitService = (*mainTestInitService)(nil)
var _ initsecret.Guard = (*mainTestInitSecretGuard)(nil)
