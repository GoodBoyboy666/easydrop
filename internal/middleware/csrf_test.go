package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestDoubleSubmitCSRFRejectsCookieWriteWithoutHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.POST("/write", NewDoubleSubmitCSRF(CSRFOptions{AuthCookieName: "session", CookieName: "csrf"}), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/write", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "session-token"})
	req.AddCookie(&http.Cookie{Name: "csrf", Value: "csrf-token"})
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusForbidden {
		t.Fatalf("expected 403 when csrf header is missing, got %d", resp.Code)
	}
}

func TestDoubleSubmitCSRFAllowsCookieWriteWithMatchingHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.PATCH("/write", NewDoubleSubmitCSRF(CSRFOptions{AuthCookieName: "session", CookieName: "csrf"}), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPatch, "/write", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "session-token"})
	req.AddCookie(&http.Cookie{Name: "csrf", Value: "csrf-token"})
	req.Header.Set(CSRFHeaderName, "csrf-token")
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200 when csrf header matches cookie, got %d", resp.Code)
	}
}

func TestDoubleSubmitCSRFSkipsBearerWriteWithoutCSRF(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.DELETE("/write", NewDoubleSubmitCSRF(CSRFOptions{AuthCookieName: "session", CookieName: "csrf"}), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodDelete, "/write", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "session-token"})
	req.Header.Set("Authorization", "Bearer any-token")
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200 for bearer request without csrf header, got %d", resp.Code)
	}
}

func TestDoubleSubmitCSRFIssuesCookieOnSafeRequestForSession(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/me", NewDoubleSubmitCSRF(CSRFOptions{AuthCookieName: "session", CookieName: "csrf"}), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "session-token"})
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200 for safe request, got %d", resp.Code)
	}

	if _, ok := responseCookie(resp, "csrf"); !ok {
		t.Fatalf("expected csrf cookie to be issued on safe request")
	}
}

func TestIssueCSRFCookieOnSuccessSetsCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.POST("/login", IssueCSRFCookieOnSuccess(CSRFOptions{CookieName: "csrf"}), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/login", nil)
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	if _, ok := responseCookie(resp, "csrf"); !ok {
		t.Fatalf("expected csrf cookie to be issued after successful request")
	}
}

func TestIssueCSRFCookieOnSuccessSkipsFailedResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.POST("/login", IssueCSRFCookieOnSuccess(CSRFOptions{CookieName: "csrf"}), func(c *gin.Context) {
		c.AbortWithStatus(http.StatusUnauthorized)
	})

	req := httptest.NewRequest(http.MethodPost, "/login", nil)
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	if _, ok := responseCookie(resp, "csrf"); ok {
		t.Fatalf("expected csrf cookie not to be issued for failed request")
	}
}

func responseCookie(resp *httptest.ResponseRecorder, name string) (*http.Cookie, bool) {
	if resp == nil {
		return nil, false
	}

	for _, cookie := range resp.Result().Cookies() {
		if cookie.Name == name {
			return cookie, true
		}
	}

	return nil, false
}
