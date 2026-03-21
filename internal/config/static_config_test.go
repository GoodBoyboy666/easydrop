package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadDefaultsWithoutConfigFile(t *testing.T) {
	cfg, err := Load("", false)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Server.Mode != ServerModeDevelopment {
		t.Fatalf("expected default mode %q, got %q", ServerModeDevelopment, cfg.Server.Mode)
	}
	if cfg.Server.Addr != ":8080" {
		t.Fatalf("expected default addr :8080, got %q", cfg.Server.Addr)
	}
	if len(cfg.Server.TrustedProxies) != 0 {
		t.Fatalf("expected default trusted proxies empty, got %v", cfg.Server.TrustedProxies)
	}
	if len(cfg.Server.RemoteIPHeaders) != 2 || cfg.Server.RemoteIPHeaders[0] != "X-Forwarded-For" || cfg.Server.RemoteIPHeaders[1] != "X-Real-IP" {
		t.Fatalf("expected default remote ip headers [X-Forwarded-For X-Real-IP], got %v", cfg.Server.RemoteIPHeaders)
	}
	if cfg.Server.ReadTimeout != 10*time.Second {
		t.Fatalf("expected default read timeout 10s, got %s", cfg.Server.ReadTimeout)
	}
	if cfg.Server.WriteTimeout != 15*time.Second {
		t.Fatalf("expected default write timeout 15s, got %s", cfg.Server.WriteTimeout)
	}
	if cfg.Server.ShutdownTimeout != 5*time.Second {
		t.Fatalf("expected default shutdown timeout 5s, got %s", cfg.Server.ShutdownTimeout)
	}
	if cfg.Email.Enable {
		t.Fatalf("expected email.enable default false, got true")
	}
}

func TestLoadModeFromEnv(t *testing.T) {
	t.Setenv("EASYDROP_SERVER_MODE", "production")

	cfg, err := Load("", false)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Server.Mode != ServerModeProduction {
		t.Fatalf("expected mode %q, got %q", ServerModeProduction, cfg.Server.Mode)
	}
}

func TestLoadProductionModeWithLowConfigUsesDefaults(t *testing.T) {
	dir := t.TempDir()
	content := []byte("server:\n  mode: production\n")
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), content, 0o644); err != nil {
		t.Fatalf("write config file failed: %v", err)
	}

	cfg, err := Load(dir, false)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Server.Mode != ServerModeProduction {
		t.Fatalf("expected mode %q, got %q", ServerModeProduction, cfg.Server.Mode)
	}
	if cfg.Server.Addr != ":8080" {
		t.Fatalf("expected default addr :8080, got %q", cfg.Server.Addr)
	}
	if cfg.Server.ReadTimeout != 10*time.Second {
		t.Fatalf("expected read timeout 10s, got %s", cfg.Server.ReadTimeout)
	}
	if cfg.Server.WriteTimeout != 15*time.Second {
		t.Fatalf("expected write timeout 15s, got %s", cfg.Server.WriteTimeout)
	}
	if cfg.Server.ShutdownTimeout != 5*time.Second {
		t.Fatalf("expected shutdown timeout 5s, got %s", cfg.Server.ShutdownTimeout)
	}
}
