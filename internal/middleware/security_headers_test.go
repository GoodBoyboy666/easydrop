package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"easydrop/internal/pkg/captcha"

	"github.com/gin-gonic/gin"
)

func TestSecurityHeadersApplySetsSecurityHeadersAndDefaultCSP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	securityHeaders := NewSecurityHeaders(nil)
	router := gin.New()
	router.GET("/health", securityHeaders.Apply, func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/health", nil))

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", recorder.Code)
	}
	assertBaseSecurityHeaders(t, recorder)

	expectedCSP := "default-src 'self'; base-uri 'self'; object-src 'none'; frame-ancestors 'none'; script-src 'self'; frame-src 'self'; connect-src 'self'"
	if got := recorder.Header().Get(headerContentSecurityPolicy); got != expectedCSP {
		t.Fatalf("expected Content-Security-Policy %q, got %q", expectedCSP, got)
	}
}

func TestSecurityHeadersApplyAddsCaptchaProviderSourcesToCSP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testCases := []struct {
		name            string
		provider        captcha.Provider
		expectedSources []string
	}{
		{
			name:     "turnstile",
			provider: captcha.ProviderTurnstile,
			expectedSources: []string{
				"https://challenges.cloudflare.com",
			},
		},
		{
			name:     "recaptcha",
			provider: captcha.ProviderRecaptcha,
			expectedSources: []string{
				"https://www.google.com",
				"https://www.gstatic.com",
			},
		},
		{
			name:     "hcaptcha",
			provider: captcha.ProviderHCaptcha,
			expectedSources: []string{
				"https://js.hcaptcha.com",
				"https://hcaptcha.com",
				"https://*.hcaptcha.com",
			},
		},
		{
			name:     "geetest_v4",
			provider: captcha.ProviderGeetestV4,
			expectedSources: []string{
				"https://static.geetest.com",
				"https://gcaptcha4.geetest.com",
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			securityHeaders := NewSecurityHeaders(&captcha.AllCaptchaConfig{
				Enabled:  true,
				Provider: tc.provider,
			})
			router := gin.New()
			router.GET("/health", securityHeaders.Apply, func(c *gin.Context) {
				c.Status(http.StatusNoContent)
			})

			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/health", nil))

			if recorder.Code != http.StatusNoContent {
				t.Fatalf("expected 204, got %d", recorder.Code)
			}

			csp := recorder.Header().Get(headerContentSecurityPolicy)
			if csp == "" {
				t.Fatalf("expected Content-Security-Policy to be set")
			}
			for _, source := range tc.expectedSources {
				if !strings.Contains(csp, source) {
					t.Fatalf("expected Content-Security-Policy to contain %q, got %q", source, csp)
				}
			}
		})
	}
}

func TestSecurityHeadersApplyIgnoresUnknownCaptchaProviderInCSP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	securityHeaders := NewSecurityHeaders(&captcha.AllCaptchaConfig{
		Enabled:  true,
		Provider: captcha.Provider("custom_provider"),
	})
	router := gin.New()
	router.GET("/health", securityHeaders.Apply, func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/health", nil))

	expectedCSP := "default-src 'self'; base-uri 'self'; object-src 'none'; frame-ancestors 'none'; script-src 'self'; frame-src 'self'; connect-src 'self'"
	if got := recorder.Header().Get(headerContentSecurityPolicy); got != expectedCSP {
		t.Fatalf("expected Content-Security-Policy %q, got %q", expectedCSP, got)
	}
}

func TestSecurityHeadersApplySkipsCSPForSwaggerRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)

	securityHeaders := NewSecurityHeaders(&captcha.AllCaptchaConfig{
		Enabled:  true,
		Provider: captcha.ProviderTurnstile,
	})
	router := gin.New()
	router.GET("/api/swagger/index.html", securityHeaders.Apply, func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/swagger/index.html", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	assertBaseSecurityHeaders(t, recorder)
	if got := recorder.Header().Get(headerContentSecurityPolicy); got != "" {
		t.Fatalf("expected Content-Security-Policy to be empty on swagger route, got %q", got)
	}
}

func TestSecurityHeadersApplyPreservesResponseBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	securityHeaders := NewSecurityHeaders(nil)
	router := gin.New()
	router.GET("/payload", securityHeaders.Apply, func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/payload", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if recorder.Body.String() != "ok" {
		t.Fatalf("expected body ok, got %q", recorder.Body.String())
	}
	if got := recorder.Header().Get(headerContentSecurityPolicy); got == "" {
		t.Fatalf("expected Content-Security-Policy to be set")
	}
}

func assertBaseSecurityHeaders(t *testing.T, recorder *httptest.ResponseRecorder) {
	t.Helper()

	if got := recorder.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("expected X-Content-Type-Options nosniff, got %q", got)
	}
	if got := recorder.Header().Get("X-Frame-Options"); got != "DENY" {
		t.Fatalf("expected X-Frame-Options DENY, got %q", got)
	}
	if got := recorder.Header().Get("Referrer-Policy"); got != "strict-origin-when-cross-origin" {
		t.Fatalf("expected Referrer-Policy strict-origin-when-cross-origin, got %q", got)
	}
	if got := recorder.Header().Get("Permissions-Policy"); got != "geolocation=(), microphone=(), camera=()" {
		t.Fatalf("expected Permissions-Policy to be set, got %q", got)
	}
	if got := recorder.Header().Get("Strict-Transport-Security"); got != "" {
		t.Fatalf("expected Strict-Transport-Security to be empty, got %q", got)
	}
}
