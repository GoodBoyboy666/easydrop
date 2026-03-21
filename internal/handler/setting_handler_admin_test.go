package handler

import (
	"context"
	"net/http"
	"testing"

	"easydrop/internal/dto"
	"easydrop/internal/service"

	"github.com/gin-gonic/gin"
)

type mockSettingAdminService struct {
	listItemsFn  func(ctx context.Context, input dto.SettingListInput) (*dto.SettingListResult, error)
	updateItemFn func(ctx context.Context, input dto.SettingUpdateInput) error
	publicFn     func(ctx context.Context) (*dto.SettingPublicResult, error)
}

func (m *mockSettingAdminService) GetValue(_ context.Context, _ string) (string, bool, error) {
	return "", false, nil
}

func (m *mockSettingAdminService) ListItems(ctx context.Context, input dto.SettingListInput) (*dto.SettingListResult, error) {
	if m.listItemsFn == nil {
		return &dto.SettingListResult{}, nil
	}
	return m.listItemsFn(ctx, input)
}

func (m *mockSettingAdminService) UpdateItem(ctx context.Context, input dto.SettingUpdateInput) error {
	if m.updateItemFn == nil {
		return nil
	}
	return m.updateItemFn(ctx, input)
}

func (m *mockSettingAdminService) GetPublicItems(ctx context.Context) (*dto.SettingPublicResult, error) {
	if m.publicFn == nil {
		return &dto.SettingPublicResult{}, nil
	}
	return m.publicFn(ctx)
}

func TestSettingAdminHandlerListSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	called := false
	h := NewSettingAdminHandler(&mockSettingAdminService{
		listItemsFn: func(_ context.Context, input dto.SettingListInput) (*dto.SettingListResult, error) {
			called = true
			if input.Category != "site" || input.Key != "site." {
				t.Fatalf("unexpected filter: %+v", input)
			}
			if input.Limit != 10 || input.Offset != 20 || input.Order != "key_desc" {
				t.Fatalf("unexpected paging: %+v", input)
			}
			return &dto.SettingListResult{Total: 1, Items: []dto.SettingItem{{Key: "site.url", Value: "https://example.com"}}}, nil
		},
	})

	c, w := newTestContext(http.MethodGet, "/api/v1/admin/settings?category=site&key=site.&limit=10&offset=20&order=key_desc")
	h.List(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !called {
		t.Fatal("expected ListItems to be called")
	}
}

func TestSettingAdminHandlerUpdateSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	called := false
	h := NewSettingAdminHandler(&mockSettingAdminService{
		updateItemFn: func(_ context.Context, input dto.SettingUpdateInput) error {
			called = true
			if input.Key != "site.url" || input.Value == nil || *input.Value != "https://example.com" {
				t.Fatalf("unexpected value input: %+v", input)
			}
			return nil
		},
	})

	c, w := newTestContextWithBody(http.MethodPatch, "/api/v1/admin/settings/site.url", `{"value":"https://example.com"}`)
	c.Params = gin.Params{{Key: "key", Value: "site.url"}}
	h.Update(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !called {
		t.Fatal("expected UpdateItem to be called")
	}
}

func TestSettingAdminHandlerNilService(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewSettingAdminHandler(nil)
	c, w := newTestContext(http.MethodGet, "/api/v1/admin/settings")
	h.List(c)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestSettingAdminHandlerMapBadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewSettingAdminHandler(&mockSettingAdminService{
		updateItemFn: func(_ context.Context, _ dto.SettingUpdateInput) error {
			return service.ErrSettingKeyRequired
		},
	})

	c, w := newTestContextWithBody(http.MethodPatch, "/api/v1/admin/settings/site.url", `{"value":"x"}`)
	c.Params = gin.Params{{Key: "key", Value: "site.url"}}
	h.Update(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestSettingAdminHandlerPublicSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	called := false
	h := NewSettingAdminHandler(&mockSettingAdminService{
		publicFn: func(_ context.Context) (*dto.SettingPublicResult, error) {
			called = true
			return &dto.SettingPublicResult{Items: []dto.SettingPublicItem{{Key: "site.name", Value: "EasyDrop"}}}, nil
		},
	})

	c, w := newTestContext(http.MethodGet, "/api/v1/settings/public")
	h.Public(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !called {
		t.Fatal("expected GetPublicItems to be called")
	}
}
