package handler

import (
	"context"
	"net/http"
	"testing"

	"easydrop/internal/dto"
	"easydrop/internal/service"

	"github.com/gin-gonic/gin"
)

type mockTagServiceForHandler struct {
	listFn func(ctx context.Context, input dto.TagListInput) (*dto.TagListResult, error)
}

func (m *mockTagServiceForHandler) List(ctx context.Context, input dto.TagListInput) (*dto.TagListResult, error) {
	if m.listFn == nil {
		return &dto.TagListResult{}, nil
	}
	return m.listFn(ctx, input)
}

func TestTagHandlerListSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	called := false
	h := NewTagHandler(&mockTagServiceForHandler{
		listFn: func(_ context.Context, input dto.TagListInput) (*dto.TagListResult, error) {
			called = true
			if input.Keyword != "go" {
				t.Fatalf("expected keyword go, got %q", input.Keyword)
			}
			if input.Page != 3 || input.Size != 10 || input.Order != "hot_desc" {
				t.Fatalf("unexpected list input: %+v", input)
			}
			return &dto.TagListResult{Items: []dto.TagDTO{}, Total: 0}, nil
		},
	})

	c, w := newTestContext(http.MethodGet, "/api/v1/tags?keyword=go&page=3&size=10&order=hot_desc")
	h.List(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !called {
		t.Fatal("expected List to be called")
	}
}

func TestTagHandlerListInvalidQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewTagHandler(&mockTagServiceForHandler{})
	c, w := newTestContext(http.MethodGet, "/api/v1/tags?size=abc")
	h.List(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestTagHandlerNilService(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewTagHandler(nil)
	c, w := newTestContext(http.MethodGet, "/api/v1/tags")
	h.List(c)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestTagHandlerListServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewTagHandler(&mockTagServiceForHandler{
		listFn: func(_ context.Context, _ dto.TagListInput) (*dto.TagListResult, error) {
			return nil, service.ErrInternal
		},
	})

	c, w := newTestContext(http.MethodGet, "/api/v1/tags")
	h.List(c)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}
