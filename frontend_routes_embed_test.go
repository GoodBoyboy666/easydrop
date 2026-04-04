//go:build embed_frontend

package main

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"easydrop/internal/config"
	"easydrop/internal/pkg/storage"
	"easydrop/internal/router"

	"github.com/gin-gonic/gin"
)

func TestEmbedFrontendServesRootAndPageRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := newFrontendTestEngine(config.ServerModeDevelopment)

	for _, requestPath := range []string{"/", "/login", "/admin/posts"} {
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, requestPath, nil)

		engine.ServeHTTP(recorder, req)

		if recorder.Code != http.StatusOK {
			t.Fatalf("expected 200 for %s, got %d", requestPath, recorder.Code)
		}
		if body := strings.ToLower(recorder.Body.String()); !strings.Contains(body, "<!doctype html>") {
			t.Fatalf("expected %s to return frontend index.html", requestPath)
		}
	}
}

func TestEmbedFrontendServesStaticFiles(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := newFrontendTestEngine(config.ServerModeDevelopment)

	robotsRecorder := httptest.NewRecorder()
	robotsReq := httptest.NewRequest(http.MethodGet, "/robots.txt", nil)
	engine.ServeHTTP(robotsRecorder, robotsReq)

	if robotsRecorder.Code != http.StatusOK {
		t.Fatalf("expected 200 for /robots.txt, got %d", robotsRecorder.Code)
	}
	if body := robotsRecorder.Body.String(); !strings.Contains(body, "User-agent") {
		t.Fatalf("expected robots.txt content, got %q", body)
	}

	assetPath := firstEmbeddedAssetPath(t)
	assetRecorder := httptest.NewRecorder()
	assetReq := httptest.NewRequest(http.MethodGet, assetPath, nil)
	engine.ServeHTTP(assetRecorder, assetReq)

	if assetRecorder.Code != http.StatusOK {
		t.Fatalf("expected 200 for %s, got %d", assetPath, assetRecorder.Code)
	}
	if body := strings.ToLower(assetRecorder.Body.String()); strings.Contains(body, "<!doctype html>") {
		t.Fatalf("expected %s to return static asset instead of index.html", assetPath)
	}
}

func TestEmbedFrontendDoesNotInterceptAPIOrLocalStorageRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	basePath := t.TempDir()
	filePath := filepath.Join(basePath, storage.CategoryFile, "2026", "04", "hello.txt")
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := os.WriteFile(filePath, []byte("hello attachment"), 0o644); err != nil {
		t.Fatalf("write file failed: %v", err)
	}

	app := newFrontendTestApp(config.ServerModeDevelopment)
	app.Config.Storage = storage.Config{
		Backend: storage.BackendLocal,
		Local: storage.LocalConfig{
			BasePath:  basePath,
			URLPrefix: "",
		},
	}

	engine := router.BuildEngine(app)
	registerFrontendRoutes(engine)

	apiRecorder := httptest.NewRecorder()
	apiReq := httptest.NewRequest(http.MethodGet, "/api/v1/posts", nil)
	engine.ServeHTTP(apiRecorder, apiReq)
	if apiRecorder.Code == http.StatusNotFound {
		t.Fatalf("expected /api/v1/posts not to be handled by frontend fallback")
	}
	if body := strings.ToLower(apiRecorder.Body.String()); strings.Contains(body, "<!doctype html>") {
		t.Fatalf("expected /api/v1/posts not to return frontend index.html")
	}

	fileRecorder := httptest.NewRecorder()
	fileReq := httptest.NewRequest(http.MethodGet, "/api/file/2026/04/hello.txt", nil)
	engine.ServeHTTP(fileRecorder, fileReq)
	if fileRecorder.Code != http.StatusOK {
		t.Fatalf("expected 200 for local storage route, got %d", fileRecorder.Code)
	}
	if body := fileRecorder.Body.String(); body != "hello attachment" {
		t.Fatalf("expected file content, got %q", body)
	}

	missingAPIRecorder := httptest.NewRecorder()
	missingAPIReq := httptest.NewRequest(http.MethodGet, "/api/missing", nil)
	engine.ServeHTTP(missingAPIRecorder, missingAPIReq)
	if missingAPIRecorder.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for missing /api route, got %d", missingAPIRecorder.Code)
	}
}

func firstEmbeddedAssetPath(t *testing.T) string {
	t.Helper()

	entries, err := fs.ReadDir(frontendDistFS, "assets")
	if err != nil {
		t.Fatalf("read embedded assets failed: %v", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			return "/assets/" + entry.Name()
		}
	}

	t.Fatal("no embedded asset file found")
	return ""
}
