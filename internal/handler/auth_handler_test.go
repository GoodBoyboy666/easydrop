package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"easydrop/internal/dto"
	cookiepkg "easydrop/internal/pkg/cookie"

	"github.com/gin-gonic/gin"
)

type mockAuthService struct {
	loginFn                func(context.Context, dto.LoginInput) (*dto.AuthResult, error)
	registerFn             func(context.Context, dto.RegisterInput) (*dto.AuthResult, error)
	requestPasswordResetFn func(context.Context, dto.PasswordResetRequestInput) error
	confirmPasswordResetFn func(context.Context, dto.PasswordResetConfirmInput) error
	confirmVerifyEmailFn   func(context.Context, dto.EmailVerifyConfirmInput) error
}

func (m *mockAuthService) Register(ctx context.Context, input dto.RegisterInput) (*dto.AuthResult, error) {
	if m.registerFn == nil {
		return nil, nil
	}
	return m.registerFn(ctx, input)
}

func (m *mockAuthService) Login(ctx context.Context, input dto.LoginInput) (*dto.AuthResult, error) {
	if m.loginFn == nil {
		return nil, nil
	}
	return m.loginFn(ctx, input)
}

func (m *mockAuthService) RequestPasswordReset(ctx context.Context, input dto.PasswordResetRequestInput) error {
	if m.requestPasswordResetFn == nil {
		return nil
	}
	return m.requestPasswordResetFn(ctx, input)
}

func (m *mockAuthService) ConfirmPasswordReset(ctx context.Context, input dto.PasswordResetConfirmInput) error {
	if m.confirmPasswordResetFn == nil {
		return nil
	}
	return m.confirmPasswordResetFn(ctx, input)
}

func (m *mockAuthService) ConfirmVerifyEmail(ctx context.Context, input dto.EmailVerifyConfirmInput) error {
	if m.confirmVerifyEmailFn == nil {
		return nil
	}
	return m.confirmVerifyEmailFn(ctx, input)
}

func TestAuthHandlerLoginSetsCookieAndReturnsToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewAuthHandler(&mockAuthService{
		loginFn: func(context.Context, dto.LoginInput) (*dto.AuthResult, error) {
			return &dto.AuthResult{AccessToken: "jwt-token"}, nil
		},
	}, nil, cookiepkg.NewAuthCookie(&cookiepkg.Config{
		Name:     "session",
		Path:     "/",
		SameSite: "lax",
		MaxAge:   2 * time.Hour,
	}))

	router := gin.New()
	router.POST("/login", handler.Login)

	body := bytes.NewBufferString(`{"account":"alice","password":"secret123"}`)
	req := httptest.NewRequest(http.MethodPost, "/login", body)
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	var payload dto.AuthResult
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if payload.AccessToken != "jwt-token" {
		t.Fatalf("expected access token jwt-token, got %q", payload.AccessToken)
	}

	cookies := recorder.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(cookies))
	}
	cookie := cookies[0]
	if cookie.Name != "session" {
		t.Fatalf("expected cookie name session, got %q", cookie.Name)
	}
	if cookie.Value != "jwt-token" {
		t.Fatalf("expected cookie value jwt-token, got %q", cookie.Value)
	}
	if !cookie.HttpOnly {
		t.Fatal("expected cookie HttpOnly to be true")
	}
	if cookie.Secure {
		t.Fatal("expected development cookie Secure to be false")
	}
	if cookie.SameSite != http.SameSiteLaxMode {
		t.Fatalf("expected SameSite=Lax, got %v", cookie.SameSite)
	}
}

func TestAuthHandlerLoginSetsSecureCookieInProduction(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewAuthHandler(&mockAuthService{
		loginFn: func(context.Context, dto.LoginInput) (*dto.AuthResult, error) {
			return &dto.AuthResult{AccessToken: "jwt-token"}, nil
		},
	}, nil, cookiepkg.NewAuthCookie(&cookiepkg.Config{
		Name:   "session",
		Secure: true,
	}))

	router := gin.New()
	router.POST("/login", handler.Login)

	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(`{"account":"alice","password":"secret123"}`))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	cookies := recorder.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(cookies))
	}
	if !cookies[0].Secure {
		t.Fatal("expected production cookie Secure to be true")
	}
}

func TestAuthHandlerLogoutClearsCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewAuthHandler(nil, nil, cookiepkg.NewAuthCookie(&cookiepkg.Config{
		Name: "session",
		Path: "/",
	}))

	router := gin.New()
	router.POST("/logout", handler.Logout)

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	cookies := recorder.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(cookies))
	}
	cookie := cookies[0]
	if cookie.Name != "session" {
		t.Fatalf("expected cookie name session, got %q", cookie.Name)
	}
	if cookie.Value != "" {
		t.Fatalf("expected cleared cookie value, got %q", cookie.Value)
	}
	if cookie.MaxAge >= 0 {
		t.Fatalf("expected negative MaxAge for cleared cookie, got %d", cookie.MaxAge)
	}
}

func TestAuthHandlerRequestPasswordResetPassesCaptchaRemoteIP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewAuthHandler(&mockAuthService{
		requestPasswordResetFn: func(_ context.Context, input dto.PasswordResetRequestInput) error {
			if input.Email != "alice@example.com" {
				t.Fatalf("unexpected email: %s", input.Email)
			}
			if input.Captcha == nil || input.Captcha.RemoteIP != "192.0.2.1" {
				t.Fatalf("unexpected captcha remote ip: %#v", input.Captcha)
			}
			return nil
		},
	}, nil, nil)

	router := gin.New()
	router.POST("/password-reset/request", handler.RequestPasswordReset)

	req := httptest.NewRequest(http.MethodPost, "/password-reset/request", strings.NewReader(`{"email":"alice@example.com","captcha":{"token":"abc"}}`))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "192.0.2.1:1234"
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
}

func TestAuthHandlerConfirmEmailChangeSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewAuthHandler(nil, &mockUserServiceForHandler{
		confirmEmailChangeFn: func(_ context.Context, input dto.UserChangeEmailConfirmInput) (*dto.UserDTO, error) {
			if input.VerificationToken != "confirm-token" {
				t.Fatalf("unexpected token: %s", input.VerificationToken)
			}
			return &dto.UserDTO{ID: 12, Email: "new@example.com", EmailVerified: true}, nil
		},
	}, nil)

	router := gin.New()
	router.POST("/email-change/confirm", handler.ConfirmEmailChange)

	req := httptest.NewRequest(http.MethodPost, "/email-change/confirm", strings.NewReader(`{"token":"confirm-token"}`))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
}
