package router

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"easydrop/internal/config"
	"easydrop/internal/di"
	"easydrop/internal/handler"
	"easydrop/internal/middleware"
	"easydrop/internal/pkg/storage"

	"github.com/gin-gonic/gin"
)

type fakeAuthMiddleware struct{}

func (fakeAuthMiddleware) OptionalLogin(c *gin.Context) {
	c.Next()
}

func (fakeAuthMiddleware) RequireLogin(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "login required"})
}

func (fakeAuthMiddleware) RequireAdmin(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "admin required"})
}

func TestBuildEngineRegistersAllRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := BuildEngine(newTestApp(fakeAuthMiddleware{}))
	routes := r.Routes()

	expected := map[string]struct{}{
		"POST /api/v1/auth/register":                  {},
		"POST /api/v1/auth/login":                     {},
		"GET /api/v1/captcha/config":                  {},
		"GET /api/v1/init/status":                     {},
		"POST /api/v1/init":                           {},
		"GET /api/v1/settings/public":                 {},
		"GET /api/v1/posts":                           {},
		"GET /api/v1/posts/:id":                       {},
		"GET /api/v1/posts/:id/comments":              {},
		"GET /api/v1/tags":                            {},
		"GET /api/v1/users/me":                        {},
		"PATCH /api/v1/users/me/profile":              {},
		"PATCH /api/v1/users/me/password":             {},
		"POST /api/v1/users/me/email-change":          {},
		"POST /api/v1/users/me/avatar":                {},
		"DELETE /api/v1/users/me/avatar":              {},
		"POST /api/v1/attachments":                    {},
		"GET /api/v1/attachments":                     {},
		"GET /api/v1/attachments/:id":                 {},
		"DELETE /api/v1/attachments/:id":              {},
		"GET /api/v1/comments":                        {},
		"POST /api/v1/posts/:id/comments":             {},
		"GET /api/v1/users/me/comments":               {},
		"GET /api/v1/users/me/comments/:id":           {},
		"PATCH /api/v1/users/me/comments/:id":         {},
		"DELETE /api/v1/users/me/comments/:id":        {},
		"GET /api/v1/admin/users":                     {},
		"POST /api/v1/admin/users":                    {},
		"PATCH /api/v1/admin/users/:id":               {},
		"DELETE /api/v1/admin/users/:id":              {},
		"POST /api/v1/admin/users/:id/avatar":         {},
		"DELETE /api/v1/admin/users/:id/avatar":       {},
		"GET /api/v1/admin/posts":                     {},
		"GET /api/v1/admin/posts/:id":                 {},
		"POST /api/v1/admin/posts":                    {},
		"PATCH /api/v1/admin/posts/:id":               {},
		"DELETE /api/v1/admin/posts/:id":              {},
		"GET /api/v1/admin/attachments":               {},
		"DELETE /api/v1/admin/attachments/:id":        {},
		"POST /api/v1/admin/attachments/batch-delete": {},
		"GET /api/v1/admin/comments":                  {},
		"GET /api/v1/admin/comments/:id":              {},
		"PATCH /api/v1/admin/comments/:id":            {},
		"DELETE /api/v1/admin/comments/:id":           {},
		"GET /api/v1/admin/settings":                  {},
		"PATCH /api/v1/admin/settings/:key":           {},
	}

	seen := make(map[string]struct{}, len(routes))
	for _, rt := range routes {
		seen[rt.Method+" "+rt.Path] = struct{}{}
	}

	for key := range expected {
		if _, ok := seen[key]; !ok {
			t.Fatalf("missing route: %s", key)
		}
	}
}

func TestBuildEngineAppliesMiddlewareGroups(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := BuildEngine(newTestApp(fakeAuthMiddleware{}))

	{
		req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 for login route, got %d", w.Code)
		}
	}

	{
		req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me/comments", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 for me comments route, got %d", w.Code)
		}
	}

	{
		req := httptest.NewRequest(http.MethodPost, "/api/v1/posts/1/comments", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 for create post comment route, got %d", w.Code)
		}
	}

	{
		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusForbidden {
			t.Fatalf("expected 403 for admin route, got %d", w.Code)
		}
	}

	{
		req := httptest.NewRequest(http.MethodGet, "/api/v1/comments", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code == http.StatusUnauthorized || w.Code == http.StatusForbidden {
			t.Fatalf("public comments route should not be blocked by auth middleware, got %d", w.Code)
		}
	}

	{
		req := httptest.NewRequest(http.MethodGet, "/api/v1/settings/public", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code == http.StatusUnauthorized || w.Code == http.StatusForbidden {
			t.Fatalf("public settings route should not be blocked by auth middleware, got %d", w.Code)
		}
	}

	{
		req := httptest.NewRequest(http.MethodGet, "/api/v1/posts/1", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code == http.StatusUnauthorized || w.Code == http.StatusForbidden {
			t.Fatalf("public post detail route should not be blocked by auth middleware, got %d", w.Code)
		}
	}

	{
		req := httptest.NewRequest(http.MethodGet, "/api/v1/posts/1/comments", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code == http.StatusUnauthorized || w.Code == http.StatusForbidden {
			t.Fatalf("public post comments route should not be blocked by auth middleware, got %d", w.Code)
		}
	}

	{
		req := httptest.NewRequest(http.MethodGet, "/api/v1/tags", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code == http.StatusUnauthorized || w.Code == http.StatusForbidden {
			t.Fatalf("public tags route should not be blocked by auth middleware, got %d", w.Code)
		}
	}

	{
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code == http.StatusUnauthorized || w.Code == http.StatusForbidden {
			t.Fatalf("public auth route should not be blocked by auth middleware, got %d", w.Code)
		}
	}

	{
		req := httptest.NewRequest(http.MethodGet, "/api/v1/captcha/config", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code == http.StatusUnauthorized || w.Code == http.StatusForbidden {
			t.Fatalf("public captcha route should not be blocked by auth middleware, got %d", w.Code)
		}
	}

	{
		req := httptest.NewRequest(http.MethodGet, "/api/v1/init/status", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code == http.StatusUnauthorized || w.Code == http.StatusForbidden {
			t.Fatalf("public init status route should not be blocked by auth middleware, got %d", w.Code)
		}
	}

	{
		req := httptest.NewRequest(http.MethodPost, "/api/v1/init", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code == http.StatusUnauthorized || w.Code == http.StatusForbidden {
			t.Fatalf("public init route should not be blocked by auth middleware, got %d", w.Code)
		}
	}
}

func TestBuildEngineNilApp(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := BuildEngine(nil)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/login", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for nil app router, got %d", w.Code)
	}
}

func TestBuildEngineAppliesGinModeFromConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)

	_ = BuildEngine(newTestAppWithMode(fakeAuthMiddleware{}, config.ServerModeProduction))
	if got := gin.Mode(); got != gin.ReleaseMode {
		t.Fatalf("expected gin mode %s, got %s", gin.ReleaseMode, got)
	}

	_ = BuildEngine(newTestAppWithMode(fakeAuthMiddleware{}, config.ServerModeDevelopment))
	if got := gin.Mode(); got != gin.DebugMode {
		t.Fatalf("expected gin mode %s, got %s", gin.DebugMode, got)
	}
}

func TestBuildEngineServesLocalStorageFilesWithAPIURLPrefix(t *testing.T) {
	gin.SetMode(gin.TestMode)

	basePath := t.TempDir()
	filePath := filepath.Join(basePath, storage.CategoryFile, "2026", "03", "hello.txt")
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := os.WriteFile(filePath, []byte("hello attachment"), 0o644); err != nil {
		t.Fatalf("write file failed: %v", err)
	}

	app := newTestApp(fakeAuthMiddleware{})
	app.Config.Storage = storage.Config{
		Backend: storage.BackendLocal,
		Local: storage.LocalConfig{
			BasePath:  basePath,
			URLPrefix: "",
		},
	}

	r := BuildEngine(app)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/file/2026/03/hello.txt", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for local storage file, got %d", w.Code)
	}
	if body := w.Body.String(); body != "hello attachment" {
		t.Fatalf("expected file content, got %q", body)
	}

	apiRecorder := httptest.NewRecorder()
	apiReq := httptest.NewRequest(http.MethodGet, "/api/v1/posts", nil)
	r.ServeHTTP(apiRecorder, apiReq)
	if apiRecorder.Code == http.StatusNotFound {
		t.Fatalf("expected /api/v1 routes to remain reachable, got 404")
	}
}

var _ middleware.Auth = (*fakeAuthMiddleware)(nil)

func newTestApp(auth middleware.Auth) *di.App {
	return newTestAppWithMode(auth, config.ServerModeDevelopment)
}

func newTestAppWithMode(auth middleware.Auth, mode string) *di.App {
	return &di.App{
		Config: &config.StaticConfig{
			Server: config.ServerConfig{Mode: mode},
		},
		Middleware:             auth,
		AuthHandler:            handler.NewAuthHandler(nil),
		CaptchaHandler:         handler.NewCaptchaHandler(nil),
		InitHandler:            handler.NewInitHandler(nil),
		UserHandler:            handler.NewUserHandler(nil),
		UserAdminHandler:       handler.NewUserAdminHandler(nil),
		AttachmentHandler:      handler.NewAttachmentHandler(nil),
		AttachmentAdminHandler: handler.NewAttachmentAdminHandler(nil),
		CommentHandler:         handler.NewCommentHandler(nil),
		CommentAdminHandler:    handler.NewCommentAdminHandler(nil),
		PostAdminHandler:       handler.NewPostAdminHandler(nil),
		PostHandler:            handler.NewPostHandler(nil),
		SettingAdminHandler:    handler.NewSettingAdminHandler(nil),
		TagHandler:             handler.NewTagHandler(nil),
	}
}
