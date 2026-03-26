package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"easydrop/internal/model"
	cookiepkg "easydrop/internal/pkg/cookie"
	"easydrop/internal/pkg/jwt"
	"easydrop/internal/repo"

	"github.com/gin-gonic/gin"
)

type mockJWTManager struct {
	parseTokenFn func(token string) (*jwt.Claims, error)
}

func (m *mockJWTManager) IssueAccessToken(uint, string, bool) (string, error) {
	return "", nil
}

func (m *mockJWTManager) ParseToken(token string) (*jwt.Claims, error) {
	if m.parseTokenFn == nil {
		return nil, jwt.ErrInvalidToken
	}
	return m.parseTokenFn(token)
}

type mockUserRepo struct {
	getByIDFn func(context.Context, uint) (*model.User, error)
}

func (m *mockUserRepo) Create(context.Context, *model.User) error { return nil }
func (m *mockUserRepo) GetByID(ctx context.Context, id uint) (*model.User, error) {
	if m.getByIDFn == nil {
		return nil, nil
	}
	return m.getByIDFn(ctx, id)
}
func (m *mockUserRepo) GetByUsername(context.Context, string) (*model.User, error) {
	return nil, nil
}
func (m *mockUserRepo) GetByEmail(context.Context, string) (*model.User, error) {
	return nil, nil
}
func (m *mockUserRepo) GetByUsernameOrEmail(context.Context, string) (*model.User, error) {
	return nil, nil
}
func (m *mockUserRepo) UpdateAvatarWithStorageUsedTx(context.Context, uint, *string, int64, int64) (*model.User, error) {
	return nil, nil
}
func (m *mockUserRepo) Update(context.Context, *model.User) error { return nil }
func (m *mockUserRepo) Delete(context.Context, uint) error        { return nil }
func (m *mockUserRepo) List(context.Context, repo.UserFilter, repo.ListOptions) ([]model.User, int64, error) {
	return nil, 0, nil
}

func TestAuthRequireLoginAcceptsBearerToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	auth := NewAuth(&mockJWTManager{
		parseTokenFn: func(token string) (*jwt.Claims, error) {
			if token != "bearer-token" {
				t.Fatalf("expected bearer token, got %q", token)
			}
			return &jwt.Claims{UserID: 7, Admin: true}, nil
		},
	}, &mockUserRepo{
		getByIDFn: func(context.Context, uint) (*model.User, error) {
			return &model.User{ID: 7, Admin: true, Status: 1}, nil
		},
	}, cookiepkg.NewAuthCookie(&cookiepkg.Config{Name: "session"}))

	router := gin.New()
	router.GET("/me", auth.RequireLogin, func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.Header.Set("Authorization", "Bearer bearer-token")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
}

func TestAuthRequireLoginFallsBackToCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)

	auth := NewAuth(&mockJWTManager{
		parseTokenFn: func(token string) (*jwt.Claims, error) {
			if token != "cookie-token" {
				t.Fatalf("expected cookie token, got %q", token)
			}
			return &jwt.Claims{UserID: 11, Admin: false}, nil
		},
	}, &mockUserRepo{
		getByIDFn: func(context.Context, uint) (*model.User, error) {
			return &model.User{ID: 11, Status: 1}, nil
		},
	}, cookiepkg.NewAuthCookie(&cookiepkg.Config{Name: "session"}))

	router := gin.New()
	router.GET("/me", auth.RequireLogin, func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "cookie-token"})
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
}

func TestAuthRequireLoginRejectsMissingCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)

	auth := NewAuth(&mockJWTManager{}, &mockUserRepo{}, nil)

	router := gin.New()
	router.GET("/me", auth.RequireLogin, func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", recorder.Code)
	}
}
