package handler

import (
	"context"
	"net/http"
	"testing"

	"easydrop/internal/dto"
	"easydrop/internal/service"

	"github.com/gin-gonic/gin"
)

type mockPostServiceForPublicHandler struct {
	createFn func(ctx context.Context, input dto.PostCreateInput) (*dto.PostDTO, error)
	getFn    func(ctx context.Context, id uint) (*dto.PostDTO, error)
	updateFn func(ctx context.Context, input dto.PostUpdateInput) (*dto.PostDTO, error)
	deleteFn func(ctx context.Context, id uint) error
	listFn   func(ctx context.Context, input dto.PostListInput) (*dto.PostListResult, error)
}

func (m *mockPostServiceForPublicHandler) Create(ctx context.Context, input dto.PostCreateInput) (*dto.PostDTO, error) {
	if m.createFn == nil {
		return nil, nil
	}
	return m.createFn(ctx, input)
}

func (m *mockPostServiceForPublicHandler) Get(ctx context.Context, id uint) (*dto.PostDTO, error) {
	if m.getFn == nil {
		return nil, nil
	}
	return m.getFn(ctx, id)
}

func (m *mockPostServiceForPublicHandler) Update(ctx context.Context, input dto.PostUpdateInput) (*dto.PostDTO, error) {
	if m.updateFn == nil {
		return nil, nil
	}
	return m.updateFn(ctx, input)
}

func (m *mockPostServiceForPublicHandler) Delete(ctx context.Context, id uint) error {
	if m.deleteFn == nil {
		return nil
	}
	return m.deleteFn(ctx, id)
}

func (m *mockPostServiceForPublicHandler) List(ctx context.Context, input dto.PostListInput) (*dto.PostListResult, error) {
	if m.listFn == nil {
		return nil, nil
	}
	return m.listFn(ctx, input)
}

func TestPostHandlerListSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	called := false
	h := NewPostHandler(&mockPostServiceForPublicHandler{
		listFn: func(_ context.Context, input dto.PostListInput) (*dto.PostListResult, error) {
			called = true
			if input.UserID == nil || *input.UserID != 7 {
				t.Fatalf("expected user_id=7, got %#v", input.UserID)
			}
			if input.TagID == nil || *input.TagID != 2 {
				t.Fatalf("expected tag_id=2, got %#v", input.TagID)
			}
			if input.Hide == nil || *input.Hide {
				t.Fatalf("expected hide=false, got %#v", input.Hide)
			}
			if input.Limit != 5 || input.Offset != 10 || input.Order != "created_at_desc" {
				t.Fatalf("unexpected list input: %+v", input)
			}
			return &dto.PostListResult{Items: []dto.PostDTO{}, Total: 0}, nil
		},
	})

	c, w := newTestContext(http.MethodGet, "/api/v1/posts?user_id=7&tag_id=2&limit=5&offset=10&order=created_at_desc&hide=true")
	h.List(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !called {
		t.Fatal("expected List to be called")
	}
}

func TestPostHandlerListInvalidQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewPostHandler(&mockPostServiceForPublicHandler{})
	c, w := newTestContext(http.MethodGet, "/api/v1/posts?user_id=0")
	h.List(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestPostHandlerNilService(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewPostHandler(nil)
	c, w := newTestContext(http.MethodGet, "/api/v1/posts")
	h.List(c)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestPostHandlerListServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewPostHandler(&mockPostServiceForPublicHandler{
		listFn: func(_ context.Context, _ dto.PostListInput) (*dto.PostListResult, error) {
			return nil, service.ErrInternal
		},
	})

	c, w := newTestContext(http.MethodGet, "/api/v1/posts")
	h.List(c)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}
