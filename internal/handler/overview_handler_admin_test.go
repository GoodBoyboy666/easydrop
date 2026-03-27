package handler

import (
	"context"
	"net/http"
	"testing"

	"easydrop/internal/dto"
	"easydrop/internal/service"

	"github.com/gin-gonic/gin"
)

type mockOverviewAdminService struct {
	getFn func(ctx context.Context) (*dto.AdminOverviewResult, error)
}

func (m *mockOverviewAdminService) Get(ctx context.Context) (*dto.AdminOverviewResult, error) {
	if m.getFn == nil {
		return nil, nil
	}
	return m.getFn(ctx)
}

func TestOverviewAdminHandlerGetSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	called := false
	h := NewOverviewAdminHandler(&mockOverviewAdminService{
		getFn: func(_ context.Context) (*dto.AdminOverviewResult, error) {
			called = true
			return &dto.AdminOverviewResult{
				Totals: dto.AdminOverviewTotals{
					Users:       1,
					Posts:       2,
					Comments:    3,
					Attachments: 4,
				},
				RecentActivity: []dto.AdminOverviewTrendItem{
					{Date: "2026-03-27", Posts: 2, Comments: 3},
				},
			}, nil
		},
	})

	c, w := newTestContext(http.MethodGet, "/api/v1/admin/overview")
	h.Get(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !called {
		t.Fatal("expected Get to be called")
	}
}

func TestOverviewAdminHandlerNilService(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewOverviewAdminHandler(nil)
	c, w := newTestContext(http.MethodGet, "/api/v1/admin/overview")
	h.Get(c)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestOverviewAdminHandlerServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewOverviewAdminHandler(&mockOverviewAdminService{
		getFn: func(_ context.Context) (*dto.AdminOverviewResult, error) {
			return nil, service.ErrInternal
		},
	})

	c, w := newTestContext(http.MethodGet, "/api/v1/admin/overview")
	h.Get(c)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}
