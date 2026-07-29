package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	cookiepkg "easydrop/internal/pkg/cookie"

	"github.com/gin-gonic/gin"
)

func TestNewOAuthHandlerUsesCookieSecureConfig(t *testing.T) {
	handler := NewOAuthHandler(nil, nil, &cookiepkg.Config{Secure: true}, nil)

	if !handler.oauthStateCookieSecure {
		t.Fatal("expected OAuth state cookie to use Secure config")
	}
}

func TestOAuthStateCookieSecureOptions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name   string
		secure bool
	}{
		{name: "development", secure: false},
		{name: "production", secure: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)

			setOAuthStateCookie(ctx, "state-value", tt.secure)

			cookie := requireSingleOAuthStateCookie(t, recorder)
			if cookie.Value != "state-value" {
				t.Fatalf("expected state cookie value, got %q", cookie.Value)
			}
			assertOAuthStateCookieOptions(t, cookie, tt.secure)
			if cookie.MaxAge != oauthStateMaxAge {
				t.Fatalf("expected MaxAge %d, got %d", oauthStateMaxAge, cookie.MaxAge)
			}
		})
	}
}

func TestClearOAuthStateCookieSecureOptions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)

	clearOAuthStateCookie(ctx, true)

	cookie := requireSingleOAuthStateCookie(t, recorder)
	assertOAuthStateCookieOptions(t, cookie, true)
	if cookie.Value != "" {
		t.Fatalf("expected empty cookie value, got %q", cookie.Value)
	}
	if cookie.MaxAge >= 0 {
		t.Fatalf("expected negative MaxAge, got %d", cookie.MaxAge)
	}
}

func requireSingleOAuthStateCookie(t *testing.T, recorder *httptest.ResponseRecorder) *http.Cookie {
	t.Helper()

	cookies := recorder.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected one cookie, got %d", len(cookies))
	}
	return cookies[0]
}

func assertOAuthStateCookieOptions(t *testing.T, cookie *http.Cookie, secure bool) {
	t.Helper()

	if cookie.Name != oauthStateCookieName {
		t.Fatalf("expected cookie name %q, got %q", oauthStateCookieName, cookie.Name)
	}
	if cookie.Path != "/api/v1" {
		t.Fatalf("expected cookie path /api/v1, got %q", cookie.Path)
	}
	if !cookie.HttpOnly {
		t.Fatal("expected HttpOnly cookie")
	}
	if cookie.Secure != secure {
		t.Fatalf("expected Secure=%t, got %t", secure, cookie.Secure)
	}
	if cookie.SameSite != http.SameSiteLaxMode {
		t.Fatalf("expected SameSite=Lax, got %v", cookie.SameSite)
	}
}
