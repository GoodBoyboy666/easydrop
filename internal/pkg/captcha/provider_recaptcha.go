package captcha

import (
	"context"
	"net/url"
	"strings"
)

func verifyRecaptcha(ctx context.Context, cfg Config, payload Payload) (Result, error) {
	if strings.TrimSpace(cfg.SecretKey) == "" {
		return Result{}, ErrEmptySecretKey
	}
	if strings.TrimSpace(payload.Token) == "" {
		return Result{}, ErrEmptyToken
	}

	verifyURL := strings.TrimSpace(cfg.VerifyURL)
	if verifyURL == "" {
		verifyURL = "https://www.google.com/recaptcha/api/siteverify"
	}

	raw, err := verifyByForm(ctx, cfg, verifyURL, url.Values{
		"secret":   []string{cfg.SecretKey},
		"response": []string{payload.Token},
		"remoteip": []string{cfg.RemoteIP},
	})
	if err != nil {
		return Result{}, err
	}

	return buildGenericResult(ProviderRecaptcha, raw)
}

