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
	if cfg.AuthCookie.Name != "easydrop_access_token" {
		t.Fatalf("expected default auth cookie name easydrop_access_token, got %q", cfg.AuthCookie.Name)
	}
	if cfg.AuthCookie.Path != "/" {
		t.Fatalf("expected default auth cookie path /, got %q", cfg.AuthCookie.Path)
	}
	if cfg.AuthCookie.SameSite != "lax" {
		t.Fatalf("expected default auth cookie same_site lax, got %q", cfg.AuthCookie.SameSite)
	}
	if cfg.RateLimit.Enabled {
		t.Fatalf("expected rate_limit.enabled default false, got true")
	}
	if cfg.RateLimit.KeyPrefix != "ratelimit" {
		t.Fatalf("expected default rate limit key prefix ratelimit, got %q", cfg.RateLimit.KeyPrefix)
	}
	if len(cfg.RateLimit.Rules) != 0 {
		t.Fatalf("expected default rate limit rules empty, got %v", cfg.RateLimit.Rules)
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

func TestLoadRateLimitRuleOverrides(t *testing.T) {
	dir := t.TempDir()
	content := []byte(
		"rate_limit:\n" +
			"  enabled: true\n" +
			"  key_prefix: custom-limit\n" +
			"  rules:\n" +
			"    auth_write:\n" +
			"      interval: 5s\n" +
			"    comment_write:\n" +
			"      interval: 30s\n" +
			"      limit: 5\n",
	)
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), content, 0o644); err != nil {
		t.Fatalf("write config file failed: %v", err)
	}

	cfg, err := Load(dir, false)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if !cfg.RateLimit.Enabled {
		t.Fatalf("expected rate limit enabled true, got false")
	}
	if cfg.RateLimit.KeyPrefix != "custom-limit" {
		t.Fatalf("expected rate limit key prefix custom-limit, got %q", cfg.RateLimit.KeyPrefix)
	}
	if cfg.RateLimit.Rules["auth_write"].Interval != 5*time.Second {
		t.Fatalf("expected auth_write interval 5s, got %s", cfg.RateLimit.Rules["auth_write"].Interval)
	}
	if cfg.RateLimit.Rules["comment_write"].Limit != 5 {
		t.Fatalf("expected comment_write limit 5, got %d", cfg.RateLimit.Rules["comment_write"].Limit)
	}
}
