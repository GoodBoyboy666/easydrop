package router

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"easydrop/internal/config"
	"easydrop/internal/di"
	"easydrop/internal/dto"
	"easydrop/internal/handler"
	"easydrop/internal/middleware"
	cookiepkg "easydrop/internal/pkg/cookie"
	"easydrop/internal/pkg/ratelimit"
	"easydrop/internal/pkg/storage"
	"easydrop/internal/service"

	"github.com/gin-gonic/gin"
)

type fakeAuthMiddleware struct{}
type allowAuthMiddleware struct{}
type fakeRateLimitMiddleware struct {
	windowRuleNames   map[string]struct{}
	cooldownRuleNames map[string]struct{}
}

type routerMockSettingService struct {
	getValueFn func(ctx context.Context, key string) (string, bool, error)
}

func (m *routerMockSettingService) GetValue(ctx context.Context, key string) (string, bool, error) {
	if m.getValueFn == nil {
		return "", false, nil
	}
	return m.getValueFn(ctx, key)
}

func (m *routerMockSettingService) ListItems(context.Context, dto.SettingListInput) (*dto.SettingListResult, error) {
	return nil, nil
}

func (m *routerMockSettingService) UpdateItem(context.Context, dto.SettingUpdateInput) error {
	return nil
}

func (m *routerMockSettingService) GetPublicItems(context.Context) (*dto.SettingPublicResult, error) {
	return nil, nil
}

func (fakeAuthMiddleware) OptionalLogin(c *gin.Context) {
	c.Next()
}

func (fakeAuthMiddleware) RequireLogin(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "login required"})
}

func (fakeAuthMiddleware) RequireAdmin(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "admin required"})
}

func (allowAuthMiddleware) OptionalLogin(c *gin.Context) {
	c.Next()
}

func (allowAuthMiddleware) RequireLogin(c *gin.Context) {
	c.Next()
}

func (allowAuthMiddleware) RequireAdmin(c *gin.Context) {
	c.Next()
}

func (m fakeRateLimitMiddleware) Window(name string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, ok := m.windowRuleNames[name]; ok {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"message": "rate limited"})
			return
		}
		c.Next()
	}
}

func (m fakeRateLimitMiddleware) Cooldown(name string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, ok := m.cooldownRuleNames[name]; ok {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"message": "rate limited"})
			return
		}
		c.Next()
	}
}

func TestBuildEngineRegistersAllRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := BuildEngine(newTestApp(fakeAuthMiddleware{}))
	routes := r.Routes()

	expected := map[string]struct{}{
		"POST /api/v1/auth/register":                  {},
		"POST /api/v1/auth/login":                     {},
		"POST /api/v1/auth/logout":                    {},
		"POST /api/v1/auth/password-reset/request":    {},
		"POST /api/v1/auth/password-reset/confirm":    {},
		"POST /api/v1/auth/verify-email/confirm":      {},
		"POST /api/v1/auth/email-change/confirm":      {},
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
	assertSecurityHeaders(t, w)

	apiRecorder := httptest.NewRecorder()
	apiReq := httptest.NewRequest(http.MethodGet, "/api/v1/posts", nil)
	r.ServeHTTP(apiRecorder, apiReq)
	if apiRecorder.Code == http.StatusNotFound {
		t.Fatalf("expected /api/v1 routes to remain reachable, got 404")
	}
	assertSecurityHeaders(t, apiRecorder)
}

func TestBuildEngineLimitsNormalRequestBodyToTwoMegabytes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := BuildEngine(newTestApp(fakeAuthMiddleware{}))
	body := `{"username":"` + strings.Repeat("a", int(middleware.OrdinaryMaxRequestBodyBytes)) + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/init", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413 for oversized normal request, got %d", w.Code)
	}
}

func TestBuildEngineLimitsUploadRequestBodyUsingSetting(t *testing.T) {
	gin.SetMode(gin.TestMode)

	app := newTestApp(allowAuthMiddleware{})
	app.RequestBodyLimit = middleware.NewRequestBodyLimit(&routerMockSettingService{
		getValueFn: func(ctx context.Context, key string) (string, bool, error) {
			if key != service.UploadMaxRequestBodySettingKey {
				return "", false, nil
			}
			return "1024", true, nil
		},
	})

	r := BuildEngine(app)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/attachments", bytes.NewReader(make([]byte, 2048)))
	req.ContentLength = 2048
	req.Header.Set("Content-Type", "multipart/form-data; boundary=limit")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413 for oversized upload request, got %d", w.Code)
	}
}

func TestBuildEngineAppliesRateLimitOnlyToWriteRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	app := newTestApp(allowAuthMiddleware{})
	app.Config.RateLimit = ratelimit.Config{
		Enabled: true,
		Rules: map[string]ratelimit.RuleConfig{
			middleware.RuleNameProfileWrite: {
				Interval: time.Minute,
				Limit:    1,
			},
		},
	}
	app.RateLimit = fakeRateLimitMiddleware{
		windowRuleNames: map[string]struct{}{
			middleware.RuleNameProfileWrite: {},
		},
	}

	r := BuildEngine(app)

	writeReq := httptest.NewRequest(http.MethodPatch, "/api/v1/users/me/profile", strings.NewReader(`{"nickname":"x"}`))
	writeReq.Header.Set("Content-Type", "application/json")
	writeResp := httptest.NewRecorder()
	r.ServeHTTP(writeResp, writeReq)
	if writeResp.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 for write route, got %d", writeResp.Code)
	}

	readReq := httptest.NewRequest(http.MethodGet, "/api/v1/posts", nil)
	readResp := httptest.NewRecorder()
	r.ServeHTTP(readResp, readReq)
	if readResp.Code == http.StatusTooManyRequests {
		t.Fatalf("expected read route not to be rate limited")
	}
}

func TestBuildEngineAppliesBusinessRateLimitRules(t *testing.T) {
	gin.SetMode(gin.TestMode)

	app := newTestApp(allowAuthMiddleware{})
	app.Config.RateLimit = ratelimit.Config{
		Enabled: true,
		Rules: map[string]ratelimit.RuleConfig{
			middleware.RuleNameAuthWrite: {
				Interval: time.Second,
			},
			middleware.RuleNameInitWrite: {
				Interval: 10 * time.Second,
			},
			middleware.RuleNameUserSecurityWrite: {
				Interval: 30 * time.Second,
			},
			middleware.RuleNameCommentWrite: {
				Interval: 30 * time.Second,
				Limit:    5,
			},
			middleware.RuleNameAttachmentWrite: {
				Interval: time.Minute,
				Limit:    20,
			},
		},
	}
	app.RateLimit = fakeRateLimitMiddleware{
		windowRuleNames: map[string]struct{}{
			middleware.RuleNameCommentWrite:    {},
			middleware.RuleNameAttachmentWrite: {},
		},
		cooldownRuleNames: map[string]struct{}{
			middleware.RuleNameAuthWrite:         {},
			middleware.RuleNameInitWrite:         {},
			middleware.RuleNameUserSecurityWrite: {},
		},
	}

	r := BuildEngine(app)

	authReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{}`))
	authReq.Header.Set("Content-Type", "application/json")
	authResp := httptest.NewRecorder()
	r.ServeHTTP(authResp, authReq)
	if authResp.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 for auth write route, got %d", authResp.Code)
	}

	postReq := httptest.NewRequest(http.MethodPost, "/api/v1/init", strings.NewReader(`{}`))
	postReq.Header.Set("Content-Type", "application/json")
	postResp := httptest.NewRecorder()
	r.ServeHTTP(postResp, postReq)
	if postResp.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 for init post, got %d", postResp.Code)
	}

	securityReq := httptest.NewRequest(http.MethodPatch, "/api/v1/users/me/password", strings.NewReader(`{"oldPassword":"a","newPassword":"b"}`))
	securityReq.Header.Set("Content-Type", "application/json")
	securityResp := httptest.NewRecorder()
	r.ServeHTTP(securityResp, securityReq)
	if securityResp.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 for user security route, got %d", securityResp.Code)
	}

	commentReq := httptest.NewRequest(http.MethodPost, "/api/v1/posts/1/comments", strings.NewReader(`{"content":"x"}`))
	commentReq.Header.Set("Content-Type", "application/json")
	commentResp := httptest.NewRecorder()
	r.ServeHTTP(commentResp, commentReq)
	if commentResp.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 for comment write route, got %d", commentResp.Code)
	}

	attachmentReq := httptest.NewRequest(http.MethodPost, "/api/v1/attachments", strings.NewReader("file"))
	attachmentReq.Header.Set("Content-Type", "multipart/form-data; boundary=test")
	attachmentResp := httptest.NewRecorder()
	r.ServeHTTP(attachmentResp, attachmentReq)
	if attachmentResp.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 for attachment write route, got %d", attachmentResp.Code)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/init/status", nil)
	getResp := httptest.NewRecorder()
	r.ServeHTTP(getResp, getReq)
	if getResp.Code == http.StatusTooManyRequests {
		t.Fatalf("expected init status not to be rate limited")
	}

	adminReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/posts", strings.NewReader(`{"title":"x"}`))
	adminReq.Header.Set("Content-Type", "application/json")
	adminResp := httptest.NewRecorder()
	r.ServeHTTP(adminResp, adminReq)
	if adminResp.Code == http.StatusTooManyRequests {
		t.Fatalf("expected admin write route not to be rate limited")
	}
}

func TestBuildEngineAppliesSecurityHeadersToSwagger(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := BuildEngine(newTestApp(fakeAuthMiddleware{}))
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/swagger/index.html", nil)

	r.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200 for swagger UI, got %d", recorder.Code)
	}
	assertSecurityHeaders(t, recorder)
}

var _ middleware.Auth = (*fakeAuthMiddleware)(nil)
var _ middleware.RateLimit = (*fakeRateLimitMiddleware)(nil)

func newTestApp(auth middleware.Auth) *di.App {
	return newTestAppWithMode(auth, config.ServerModeDevelopment)
}

func newTestAppWithMode(auth middleware.Auth, mode string) *di.App {
	return &di.App{
		Config: &config.StaticConfig{
			Server: config.ServerConfig{Mode: mode},
		},
		Middleware:             auth,
		SecurityHeaders:        middleware.NewSecurityHeaders(),
		RateLimit:              nil,
		RequestBodyLimit:       middleware.NewRequestBodyLimit(nil),
		AuthHandler:            handler.NewAuthHandler(nil, nil, cookiepkg.NewAuthCookie(nil)),
		CaptchaHandler:         handler.NewCaptchaHandler(nil),
		InitHandler:            handler.NewInitHandler(nil),
		UserHandler:            handler.NewUserHandler(nil),
		UserAdminHandler:       handler.NewUserAdminHandler(nil),
		AttachmentHandler:      handler.NewAttachmentHandler(nil, nil),
		AttachmentAdminHandler: handler.NewAttachmentAdminHandler(nil),
		CommentHandler:         handler.NewCommentHandler(nil),
		CommentAdminHandler:    handler.NewCommentAdminHandler(nil),
		PostAdminHandler:       handler.NewPostAdminHandler(nil),
		PostHandler:            handler.NewPostHandler(nil),
		SettingAdminHandler:    handler.NewSettingAdminHandler(nil),
		TagHandler:             handler.NewTagHandler(nil),
	}
}

func assertSecurityHeaders(t *testing.T, recorder *httptest.ResponseRecorder) {
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
