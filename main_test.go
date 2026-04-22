package main

import (
	"bytes"
	"context"
	"errors"
	"log"
	"os"
	"path/filepath"
	"strconv"
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

func TestEnsureJWTKeysOnStartupDisabled(t *testing.T) {
	configDir := t.TempDir()
	privatePath := filepath.Join(configDir, "jwt", "private.pem")
	publicPath := filepath.Join(configDir, "jwt", "public.pem")
	writeMainTestJWTConfig(t, configDir, privatePath, publicPath)

	if err := ensureJWTKeysOnStartup(configDir, false, log.New(&bytes.Buffer{}, "", 0)); err != nil {
		t.Fatalf("ensureJWTKeysOnStartup error: %v", err)
	}

	if _, err := os.Stat(privatePath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected private key not to be generated, err=%v", err)
	}
	if _, err := os.Stat(publicPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected public key not to be generated, err=%v", err)
	}
}

func TestEnsureJWTKeysOnStartupGeneratesWhenBothMissing(t *testing.T) {
	configDir := t.TempDir()
	privatePath := filepath.Join(configDir, "custom", "keys", "server-private.key")
	publicPath := filepath.Join(configDir, "custom", "keys", "server-public.key")
	writeMainTestJWTConfig(t, configDir, privatePath, publicPath)

	var buf bytes.Buffer
	if err := ensureJWTKeysOnStartup(configDir, true, log.New(&buf, "", 0)); err != nil {
		t.Fatalf("ensureJWTKeysOnStartup error: %v", err)
	}

	if _, err := os.Stat(privatePath); err != nil {
		t.Fatalf("expected generated private key, err=%v", err)
	}
	if _, err := os.Stat(publicPath); err != nil {
		t.Fatalf("expected generated public key, err=%v", err)
	}
	if !strings.Contains(buf.String(), "已自动生成") {
		t.Fatalf("expected log to contain auto-generate message, got %q", buf.String())
	}
}

func TestEnsureJWTKeysOnStartupSkipsWhenBothExist(t *testing.T) {
	configDir := t.TempDir()
	privatePath := filepath.Join(configDir, "jwt", "private.pem")
	publicPath := filepath.Join(configDir, "jwt", "public.pem")
	writeMainTestJWTConfig(t, configDir, privatePath, publicPath)

	if err := generateJWTTokenPair(privatePath, publicPath, false); err != nil {
		t.Fatalf("generateJWTTokenPair error: %v", err)
	}

	privateBefore, err := os.ReadFile(privatePath)
	if err != nil {
		t.Fatalf("read private key before ensure failed: %v", err)
	}
	publicBefore, err := os.ReadFile(publicPath)
	if err != nil {
		t.Fatalf("read public key before ensure failed: %v", err)
	}

	var buf bytes.Buffer
	if err := ensureJWTKeysOnStartup(configDir, true, log.New(&buf, "", 0)); err != nil {
		t.Fatalf("ensureJWTKeysOnStartup error: %v", err)
	}

	privateAfter, err := os.ReadFile(privatePath)
	if err != nil {
		t.Fatalf("read private key after ensure failed: %v", err)
	}
	publicAfter, err := os.ReadFile(publicPath)
	if err != nil {
		t.Fatalf("read public key after ensure failed: %v", err)
	}

	if !bytes.Equal(privateBefore, privateAfter) {
		t.Fatal("expected private key to stay unchanged when both files already exist")
	}
	if !bytes.Equal(publicBefore, publicAfter) {
		t.Fatal("expected public key to stay unchanged when both files already exist")
	}
	if !strings.Contains(buf.String(), "跳过自动生成") {
		t.Fatalf("expected log to contain skip message, got %q", buf.String())
	}
}

func TestEnsureJWTKeysOnStartupReturnsErrorWhenPartiallyMissing(t *testing.T) {
	configDir := t.TempDir()
	privatePath := filepath.Join(configDir, "jwt", "private.pem")
	publicPath := filepath.Join(configDir, "jwt", "public.pem")
	writeMainTestJWTConfig(t, configDir, privatePath, publicPath)

	if err := os.MkdirAll(filepath.Dir(privatePath), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	privateContent := []byte("partial-private-key")
	if err := os.WriteFile(privatePath, privateContent, 0o600); err != nil {
		t.Fatalf("write private key failed: %v", err)
	}

	err := ensureJWTKeysOnStartup(configDir, true, log.New(&bytes.Buffer{}, "", 0))
	if err == nil || !strings.Contains(err.Error(), "必须同时存在或同时不存在") {
		t.Fatalf("expected partial key error, got %v", err)
	}

	privateAfter, readErr := os.ReadFile(privatePath)
	if readErr != nil {
		t.Fatalf("read private key after ensure failed: %v", readErr)
	}
	if !bytes.Equal(privateAfter, privateContent) {
		t.Fatal("expected existing private key to remain unchanged when startup check fails")
	}
	if _, statErr := os.Stat(publicPath); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("expected public key to remain missing, err=%v", statErr)
	}
}

func writeMainTestJWTConfig(t *testing.T, configDir string, privatePath string, publicPath string) {
	t.Helper()

	content := strings.Join([]string{
		"jwt:",
		"  private_key_path: " + strconv.Quote(filepath.ToSlash(privatePath)),
		"  public_key_path: " + strconv.Quote(filepath.ToSlash(publicPath)),
	}, "\n") + "\n"

	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(content), 0o644); err != nil {
		t.Fatalf("write config file failed: %v", err)
	}
}

var _ service.InitService = (*mainTestInitService)(nil)
var _ initsecret.Guard = (*mainTestInitSecretGuard)(nil)
