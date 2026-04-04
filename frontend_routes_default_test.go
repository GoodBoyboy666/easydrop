//go:build !embed_frontend

package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"easydrop/internal/config"

	"github.com/gin-gonic/gin"
)

func TestDefaultBuildDoesNotServeFrontend(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := newFrontendTestEngine(config.ServerModeDevelopment)
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for / without embed_frontend, got %d", recorder.Code)
	}
}
