package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"go.yaml.in/yaml/v3"
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
	if !cfg.Server.CSP.Enabled {
		t.Fatalf("expected server.csp.enabled default true, got false")
	}
	if len(cfg.Server.CSP.AllowedSources) != 0 {
		t.Fatalf("expected default server.csp.allowed_sources empty, got %v", cfg.Server.CSP.AllowedSources)
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
	if cfg.Avatar.GravatarBaseURL != "https://www.gravatar.com/avatar/" {
		t.Fatalf("expected default gravatar base url https://www.gravatar.com/avatar/, got %q", cfg.Avatar.GravatarBaseURL)
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

func TestLoadAvatarGravatarBaseURLOverride(t *testing.T) {
	dir := t.TempDir()
	content := []byte(
		"avatar:\n" +
			"  gravatar_base_url: https://gravatar.example.com/avatar\n",
	)
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), content, 0o644); err != nil {
		t.Fatalf("write config file failed: %v", err)
	}

	cfg, err := Load(dir, false)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Avatar.GravatarBaseURL != "https://gravatar.example.com/avatar" {
		t.Fatalf("expected gravatar base url override, got %q", cfg.Avatar.GravatarBaseURL)
	}
}

func TestLoadServerCSPOverride(t *testing.T) {
	dir := t.TempDir()
	content := []byte(
		"server:\n" +
			"  csp:\n" +
			"    enabled: false\n" +
			"    allowed_sources:\n" +
			"      - https://cdn.example.com\n" +
			"      - https://*.example.com\n",
	)
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), content, 0o644); err != nil {
		t.Fatalf("write config file failed: %v", err)
	}

	cfg, err := Load(dir, false)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Server.CSP.Enabled {
		t.Fatalf("expected server.csp.enabled false, got true")
	}
	if len(cfg.Server.CSP.AllowedSources) != 2 {
		t.Fatalf("expected 2 allowed_sources, got %v", cfg.Server.CSP.AllowedSources)
	}
	if cfg.Server.CSP.AllowedSources[0] != "https://cdn.example.com" {
		t.Fatalf("expected first allowed source https://cdn.example.com, got %q", cfg.Server.CSP.AllowedSources[0])
	}
	if cfg.Server.CSP.AllowedSources[1] != "https://*.example.com" {
		t.Fatalf("expected second allowed source https://*.example.com, got %q", cfg.Server.CSP.AllowedSources[1])
	}
}

func TestWriteDefaultConfigFileCreatesConfig(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "runtime")

	if err := WriteDefaultConfigFile(dir); err != nil {
		t.Fatalf("WriteDefaultConfigFile returned error: %v", err)
	}

	configPath := filepath.Join(dir, "config.yaml")
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read generated config failed: %v", err)
	}
	text := string(content)
	if !strings.Contains(text, "server:") {
		t.Fatalf("expected generated config to contain server section, got %q", text)
	}
	if !strings.Contains(text, "csp:") {
		t.Fatalf("expected generated config to contain csp section, got %q", text)
	}
	if !strings.Contains(text, "enabled: true") {
		t.Fatalf("expected generated config to contain csp enabled default, got %q", text)
	}
	if !strings.Contains(text, "read_timeout: 10s") {
		t.Fatalf("expected generated config to contain read_timeout default, got %q", text)
	}

	cfg, err := Load(dir, false)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.Server.Mode != ServerModeDevelopment {
		t.Fatalf("expected mode %q, got %q", ServerModeDevelopment, cfg.Server.Mode)
	}
	if cfg.JWT.Expire != 24*time.Hour {
		t.Fatalf("expected jwt.expire 24h, got %s", cfg.JWT.Expire)
	}
}

func TestWriteDefaultConfigFileReturnsErrExistWhenAlreadyExists(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("server:\n  addr: \":9090\"\n"), 0o644); err != nil {
		t.Fatalf("write existing config failed: %v", err)
	}

	err := WriteDefaultConfigFile(dir)
	if !errors.Is(err, os.ErrExist) {
		t.Fatalf("expected os.ErrExist, got %v", err)
	}
}

func TestWriteDefaultConfigFileIgnoresEnvOverride(t *testing.T) {
	t.Setenv("EASYDROP_SERVER_ADDR", ":9090")

	dir := t.TempDir()
	if err := WriteDefaultConfigFile(dir); err != nil {
		t.Fatalf("WriteDefaultConfigFile returned error: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(dir, "config.yaml"))
	if err != nil {
		t.Fatalf("read generated config failed: %v", err)
	}

	parsed := map[string]any{}
	if err := yaml.Unmarshal(content, &parsed); err != nil {
		t.Fatalf("unmarshal generated yaml failed: %v", err)
	}

	serverRaw, ok := parsed["server"]
	if !ok {
		t.Fatalf("expected server section, got %v", parsed)
	}
	serverMap, ok := serverRaw.(map[string]any)
	if !ok {
		t.Fatalf("expected server section to be map, got %T", serverRaw)
	}

	addrRaw, ok := serverMap["addr"]
	if !ok {
		t.Fatalf("expected server.addr, got %v", serverMap)
	}
	addr, ok := addrRaw.(string)
	if !ok {
		t.Fatalf("expected server.addr to be string, got %T", addrRaw)
	}
	if addr != ":8080" {
		t.Fatalf("expected generated default server.addr :8080, got %q", addr)
	}
}
