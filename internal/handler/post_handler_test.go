package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"easydrop/internal/dto"
	"easydrop/internal/middleware"
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
	pinHigh := uint(99)
	pinLow := uint(10)
	h := NewPostHandler(&mockPostServiceForPublicHandler{
		listFn: func(_ context.Context, input dto.PostListInput) (*dto.PostListResult, error) {
			called = true
			if input.UserID == nil || *input.UserID != 7 {
				t.Fatalf("expected user_id=7, got %#v", input.UserID)
			}
			if input.TagID == nil || *input.TagID != 2 {
				t.Fatalf("expected tag_id=2, got %#v", input.TagID)
			}
			if input.Content != "hello" {
				t.Fatalf("expected content=hello, got %q", input.Content)
			}
			if input.Hide == nil || *input.Hide {
				t.Fatalf("expected hide=false, got %#v", input.Hide)
			}
			if input.Limit != 5 || input.Offset != 10 || input.Order != "created_at_desc" {
				t.Fatalf("unexpected list input: %+v", input)
			}
			return &dto.PostListResult{Items: []dto.PostDTO{
				{ID: 1, Content: "pinned high", Pin: &pinHigh},
				{ID: 2, Content: "pinned low", Pin: &pinLow},
				{ID: 3, Content: "normal"},
			}, Total: 3}, nil
		},
	})

	c, w := newTestContext(http.MethodGet, "/api/v1/posts?user_id=7&tag_id=2&content=%20hello%20&limit=5&offset=10&order=created_at_desc&hide=true")
	h.List(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !called {
		t.Fatal("expected List to be called")
	}

	var body dto.PostPublicListResult
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if body.Total != 3 {
		t.Fatalf("expected total 3, got %d", body.Total)
	}
	if len(body.PinnedItems) != 2 {
		t.Fatalf("expected 2 pinned items, got %d", len(body.PinnedItems))
	}
	if len(body.Items) != 1 {
		t.Fatalf("expected 1 normal item, got %d", len(body.Items))
	}
	if body.PinnedItems[0].ID != 1 || body.PinnedItems[1].ID != 2 {
		t.Fatalf("unexpected pinned items order: %#v", body.PinnedItems)
	}
	if body.Items[0].ID != 3 {
		t.Fatalf("unexpected normal item: %#v", body.Items[0])
	}
}

func TestPostHandlerGetSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	called := false
	h := NewPostHandler(&mockPostServiceForPublicHandler{
		getFn: func(_ context.Context, id uint) (*dto.PostDTO, error) {
			called = true
			if id != 9 {
				t.Fatalf("expected id=9, got %d", id)
			}
			return &dto.PostDTO{ID: id, Content: "hello"}, nil
		},
	})

	c, w := newTestContext(http.MethodGet, "/api/v1/posts/9")
	c.Params = gin.Params{{Key: "id", Value: "9"}}
	h.Get(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !called {
		t.Fatal("expected Get to be called")
	}
}

func TestPostHandlerGetHiddenPostForPublicViewer(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewPostHandler(&mockPostServiceForPublicHandler{
		getFn: func(_ context.Context, id uint) (*dto.PostDTO, error) {
			return &dto.PostDTO{ID: id, Content: "hidden", Hide: true}, nil
		},
	})

	c, w := newTestContext(http.MethodGet, "/api/v1/posts/3")
	c.Params = gin.Params{{Key: "id", Value: "3"}}
	h.Get(c)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}

func TestPostHandlerGetHiddenPostForAdminViewer(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewPostHandler(&mockPostServiceForPublicHandler{
		getFn: func(_ context.Context, id uint) (*dto.PostDTO, error) {
			return &dto.PostDTO{ID: id, Content: "hidden", Hide: true}, nil
		},
	})

	c, w := newTestContext(http.MethodGet, "/api/v1/posts/3")
	c.Params = gin.Params{{Key: "id", Value: "3"}}
	c.Set(middleware.ContextUserIDKey, uint(1))
	c.Set(middleware.ContextUserAdminKey, true)
	h.Get(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestPostHandlerGetInvalidPathID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewPostHandler(&mockPostServiceForPublicHandler{})
	c, w := newTestContext(http.MethodGet, "/api/v1/posts/0")
	c.Params = gin.Params{{Key: "id", Value: "0"}}
	h.Get(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
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

func TestPostHandlerListAdminViewerCanSeeHidePosts(t *testing.T) {
	gin.SetMode(gin.TestMode)

	called := false
	h := NewPostHandler(&mockPostServiceForPublicHandler{
		listFn: func(_ context.Context, input dto.PostListInput) (*dto.PostListResult, error) {
			called = true
			if input.Hide != nil {
				t.Fatalf("expected hide filter to be nil for admin viewer, got %#v", input.Hide)
			}
			return &dto.PostListResult{Items: []dto.PostDTO{}, Total: 0}, nil
		},
	})

	c, w := newTestContext(http.MethodGet, "/api/v1/posts")
	c.Set(middleware.ContextUserIDKey, uint(1))
	c.Set(middleware.ContextUserAdminKey, true)
	h.List(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !called {
		t.Fatal("expected List to be called")
	}
}
