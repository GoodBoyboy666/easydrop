package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"easydrop/internal/dto"

	"github.com/gin-gonic/gin"
)

type mockCaptchaConfigService struct {
	getConfigFn func(ctx context.Context) *dto.CaptchaConfigResult
}

func (m *mockCaptchaConfigService) GetConfig(ctx context.Context) *dto.CaptchaConfigResult {
	if m.getConfigFn == nil {
		return &dto.CaptchaConfigResult{}
	}
	return m.getConfigFn(ctx)
}

func TestCaptchaHandlerGetConfigSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewCaptchaHandler(&mockCaptchaConfigService{
		getConfigFn: func(context.Context) *dto.CaptchaConfigResult {
			return &dto.CaptchaConfigResult{Enabled: true, Provider: "turnstile", SiteKey: "site-key"}
		},
	})

	c, w := newTestContext(http.MethodGet, "/api/v1/captcha/config")
	h.GetConfig(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var body dto.CaptchaConfigResult
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if !body.Enabled || body.Provider != "turnstile" || body.SiteKey != "site-key" {
		t.Fatalf("unexpected response body: %+v", body)
	}
}

func TestCaptchaHandlerGetConfigNilService(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewCaptchaHandler(nil)
	c, w := newTestContext(http.MethodGet, "/api/v1/captcha/config")
	h.GetConfig(c)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}
