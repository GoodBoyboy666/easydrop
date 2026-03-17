package captcha

import (
	"context"
	"net/url"
	"strings"
)

func verifyTurnstile(ctx context.Context, cfg VerifyConfig, payload Payload) (Result, error) {
	if strings.TrimSpace(cfg.ProviderConfig.SecretKey) == "" {
		return Result{}, ErrEmptySecretKey
	}
	if strings.TrimSpace(payload.Token) == "" {
		return Result{}, ErrEmptyToken
	}

	verifyURL := strings.TrimSpace(cfg.ProviderConfig.VerifyURL)
	if verifyURL == "" {
		verifyURL = "https://challenges.cloudflare.com/turnstile/v0/siteverify"
	}

	raw, err := verifyByForm(ctx, cfg, verifyURL, url.Values{
		"secret":   []string{cfg.ProviderConfig.SecretKey},
		"response": []string{payload.Token},
		"remoteip": []string{payload.RemoteIP},
	})
	if err != nil {
		return Result{}, err
	}

	return buildGenericResult(ProviderTurnstile, raw)
}
