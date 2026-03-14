package captcha

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"strings"
)

func verifyGeetestV4(ctx context.Context, cfg Config, payload Payload) (Result, error) {
	if strings.TrimSpace(cfg.SiteKey) == "" {
		return Result{}, ErrEmptyGeetestCaptchaID
	}
	if strings.TrimSpace(cfg.SecretKey) == "" {
		return Result{}, ErrEmptySecretKey
	}
	if strings.TrimSpace(payload.LotNumber) == "" {
		return Result{}, ErrEmptyLotNumber
	}
	if strings.TrimSpace(payload.CaptchaOutput) == "" {
		return Result{}, ErrEmptyCaptchaOutput
	}
	if strings.TrimSpace(payload.PassToken) == "" {
		return Result{}, ErrEmptyPassToken
	}
	if strings.TrimSpace(payload.GenTime) == "" {
		return Result{}, ErrEmptyGenTime
	}

	verifyURL := strings.TrimSpace(cfg.VerifyURL)
	if verifyURL == "" {
		verifyURL = "https://gcaptcha4.geetest.com/validate"
	}

	raw, err := verifyByForm(ctx, cfg, verifyURL, url.Values{
		"captcha_id":     []string{cfg.SiteKey},
		"lot_number":     []string{payload.LotNumber},
		"captcha_output": []string{payload.CaptchaOutput},
		"pass_token":     []string{payload.PassToken},
		"gen_time":       []string{payload.GenTime},
		"sign_token":     []string{signGeetestToken(payload.LotNumber, cfg.SecretKey)},
	})
	if err != nil {
		return Result{}, err
	}

	result := Result{Provider: ProviderGeetestV4, Raw: raw}
	if v, ok := raw["reason"].(string); ok {
		result.Message = v
	}
	if v, ok := raw["result"].(string); ok && strings.EqualFold(v, "success") {
		result.Success = true
		return result, nil
	}
	return result, ErrVerifyFailed
}

func signGeetestToken(lotNumber, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	_, _ = h.Write([]byte(lotNumber))
	return hex.EncodeToString(h.Sum(nil))
}

