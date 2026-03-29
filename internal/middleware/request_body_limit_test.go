package middleware

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

type errReadCloser struct {
	err error
}

func (r *errReadCloser) Read(_ []byte) (int, error) {
	if r.err == nil {
		return 0, io.EOF
	}
	return 0, r.err
}

func (r *errReadCloser) Close() error {
	return nil
}

func TestRequestBodyLimitOrdinaryRejectsUnknownLengthOversizedBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(NewRequestBodyLimit(nil).Ordinary)
	router.POST("/limited", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	body := strings.Repeat("a", int(OrdinaryMaxRequestBodyBytes)+1)
	req := httptest.NewRequest(http.MethodPost, "/limited", strings.NewReader(body))
	req.ContentLength = -1
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d", resp.Code)
	}
	if resp.Body.String() != "{\"message\":\"请求体过大\"}" {
		t.Fatalf("unexpected body: %s", resp.Body.String())
	}
}

func TestRequestBodyLimitOrdinaryReturnsBadRequestOnReadError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(NewRequestBodyLimit(nil).Ordinary)
	router.POST("/limited", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodPost, "/limited", http.NoBody)
	req.Body = &errReadCloser{err: errors.New("boom")}
	req.ContentLength = -1
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.Code)
	}
	if resp.Body.String() != "{\"message\":\"请求体读取失败\"}" {
		t.Fatalf("unexpected body: %s", resp.Body.String())
	}
}

func TestRequestBodyLimitOrdinaryKnownLengthDoesNotRebufferBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(NewRequestBodyLimit(nil).Ordinary)
	router.POST("/limited", func(c *gin.Context) {
		if c.Request.ContentLength != 10 {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "unexpected content-length mutation"})
			return
		}

		payload, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "unexpected read error"})
			return
		}
		if string(payload) != "ok" {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "unexpected payload"})
			return
		}

		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodPost, "/limited", strings.NewReader("ok"))
	req.ContentLength = 10
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d, body=%s", resp.Code, resp.Body.String())
	}
}
