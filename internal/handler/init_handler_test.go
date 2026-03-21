package handler

import (
	"context"
	"net/http"
	"testing"

	"easydrop/internal/dto"
	"easydrop/internal/service"

	"github.com/gin-gonic/gin"
)

type mockInitHandlerService struct {
	statusFn func(ctx context.Context) (*dto.InitStatusResult, error)
	initFn   func(ctx context.Context, input dto.InitInput) error
}

func (m *mockInitHandlerService) GetStatus(ctx context.Context) (*dto.InitStatusResult, error) {
	if m.statusFn == nil {
		return &dto.InitStatusResult{}, nil
	}
	return m.statusFn(ctx)
}

func (m *mockInitHandlerService) Initialize(ctx context.Context, input dto.InitInput) error {
	if m.initFn == nil {
		return nil
	}
	return m.initFn(ctx, input)
}

func TestInitHandlerStatusSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewInitHandler(&mockInitHandlerService{
		statusFn: func(context.Context) (*dto.InitStatusResult, error) {
			return &dto.InitStatusResult{Initialized: true}, nil
		},
	})

	c, w := newTestContext(http.MethodGet, "/api/v1/init/status")
	h.Status(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestInitHandlerInitializeSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	called := false
	h := NewInitHandler(&mockInitHandlerService{
		initFn: func(_ context.Context, input dto.InitInput) error {
			called = true
			if input.Username != "admin" || input.SiteName != "EasyDrop" {
				t.Fatalf("unexpected input: %+v", input)
			}
			return nil
		},
	})

	c, w := newTestContextWithBody(http.MethodPost, "/api/v1/init", `{"username":"admin","email":"admin@example.com","password":"Pass1234","site_name":"EasyDrop","site_url":"http://localhost:8080"}`)
	h.Initialize(c)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", w.Code)
	}
	if !called {
		t.Fatal("expected Initialize to be called")
	}
}

func TestInitHandlerInitializeAlreadyInitialized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewInitHandler(&mockInitHandlerService{
		initFn: func(context.Context, dto.InitInput) error {
			return service.ErrAlreadyInitialized
		},
	})

	c, w := newTestContextWithBody(http.MethodPost, "/api/v1/init", `{"username":"admin"}`)
	h.Initialize(c)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected status 409, got %d", w.Code)
	}
}

func TestInitHandlerNilService(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewInitHandler(nil)
	c, w := newTestContext(http.MethodGet, "/api/v1/init/status")
	h.Status(c)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}
