package captcha

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func verifyByForm(ctx context.Context, cfg Config, verifyURL string, form url.Values) (map[string]any, error) {
	if strings.TrimSpace(verifyURL) == "" {
		return nil, ErrEmptyVerifyURL
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, verifyURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRequestFailed, err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := cfg.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: defaultOr(cfg.Timeout, defaultVerifyTimeout)}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRequestFailed, err)
	}
	defer resp.Body.Close()

	var raw map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDecodeResponse, err)
	}
	return raw, nil
}

func buildGenericResult(provider Provider, raw map[string]any) (Result, error) {
	result := Result{Provider: provider, Raw: raw}
	if score, ok := toFloat64(raw["score"]); ok {
		result.Score = score
	}
	if v, ok := raw["error-codes"]; ok {
		result.Message = fmt.Sprint(v)
	}
	if success, ok := raw["success"].(bool); ok && success {
		result.Success = true
		return result, nil
	}
	return result, ErrVerifyFailed
}

