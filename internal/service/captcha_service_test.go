package service

import (
	"context"
	"testing"

	"easydrop/internal/pkg/captcha"
)

func TestCaptchaConfigServiceGetConfigDisabled(t *testing.T) {
	svc := NewCaptchaConfigService(&captcha.AllCaptchaConfig{
		Enabled:  false,
		Provider: captcha.ProviderTurnstile,
		Turnstile: captcha.ProviderConfig{
			SiteKey: "turnstile-site",
		},
	})

	result := svc.GetConfig(context.Background())
	if result.Enabled {
		t.Fatal("expected enabled=false")
	}
	if result.Provider != "turnstile" {
		t.Fatalf("expected provider=turnstile, got %q", result.Provider)
	}
	if result.SiteKey != "" {
		t.Fatalf("expected empty site_key when disabled, got %q", result.SiteKey)
	}
}

func TestCaptchaConfigServiceGetConfigEnabled(t *testing.T) {
	svc := NewCaptchaConfigService(&captcha.AllCaptchaConfig{
		Enabled:  true,
		Provider: captcha.ProviderRecaptcha,
		Recaptcha: captcha.ProviderConfig{
			SiteKey: " recaptcha-site ",
		},
	})

	result := svc.GetConfig(context.Background())
	if !result.Enabled {
		t.Fatal("expected enabled=true")
	}
	if result.Provider != "recaptcha" {
		t.Fatalf("expected provider=recaptcha, got %q", result.Provider)
	}
	if result.SiteKey != "recaptcha-site" {
		t.Fatalf("expected site_key=recaptcha-site, got %q", result.SiteKey)
	}
}

func TestCaptchaConfigServiceGetConfigUnknownProvider(t *testing.T) {
	svc := NewCaptchaConfigService(&captcha.AllCaptchaConfig{
		Enabled:  true,
		Provider: captcha.Provider("unknown"),
	})

	result := svc.GetConfig(context.Background())
	if result.SiteKey != "" {
		t.Fatalf("expected empty site_key for unknown provider, got %q", result.SiteKey)
	}
}

func TestCaptchaConfigServiceGetConfigNilConfig(t *testing.T) {
	svc := NewCaptchaConfigService(nil)
	result := svc.GetConfig(context.Background())

	if result.Enabled {
		t.Fatal("expected enabled=false")
	}
	if result.Provider != "" || result.SiteKey != "" {
		t.Fatalf("expected empty provider/site_key, got provider=%q site_key=%q", result.Provider, result.SiteKey)
	}
}
