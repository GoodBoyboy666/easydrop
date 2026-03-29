package handler

import (
	"context"
	"net/http"
	"testing"

	"easydrop/internal/dto"
	"easydrop/internal/middleware"
	"easydrop/internal/service"

	"github.com/gin-gonic/gin"
)

type mockPostAdminService struct {
	createFn func(ctx context.Context, input dto.PostCreateInput) (*dto.PostDTO, error)
	getFn    func(ctx context.Context, id uint) (*dto.PostDTO, error)
	updateFn func(ctx context.Context, input dto.PostUpdateInput) (*dto.PostDTO, error)
	deleteFn func(ctx context.Context, id uint) error
	listFn   func(ctx context.Context, input dto.PostListInput) (*dto.PostListResult, error)
}

func (m *mockPostAdminService) Create(ctx context.Context, input dto.PostCreateInput) (*dto.PostDTO, error) {
	if m.createFn == nil {
		return nil, nil
	}
	return m.createFn(ctx, input)
}

func (m *mockPostAdminService) Get(ctx context.Context, id uint) (*dto.PostDTO, error) {
	if m.getFn == nil {
		return nil, nil
	}
	return m.getFn(ctx, id)
}

func (m *mockPostAdminService) Update(ctx context.Context, input dto.PostUpdateInput) (*dto.PostDTO, error) {
	if m.updateFn == nil {
		return nil, nil
	}
	return m.updateFn(ctx, input)
}

func (m *mockPostAdminService) Delete(ctx context.Context, id uint) error {
	if m.deleteFn == nil {
		return nil
	}
	return m.deleteFn(ctx, id)
}

func (m *mockPostAdminService) List(ctx context.Context, input dto.PostListInput) (*dto.PostListResult, error) {
	if m.listFn == nil {
		return nil, nil
	}
	return m.listFn(ctx, input)
}

func TestPostAdminHandlerListSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	called := false
	h := NewPostAdminHandler(&mockPostAdminService{
		listFn: func(_ context.Context, input dto.PostListInput) (*dto.PostListResult, error) {
			called = true
			if input.UserID == nil || *input.UserID != 12 {
				t.Fatalf("expected user_id=12, got %#v", input.UserID)
			}
			if input.TagID == nil || *input.TagID != 3 {
				t.Fatalf("expected tag_id=3, got %#v", input.TagID)
			}
			if input.Content != "admin key" {
				t.Fatalf("expected content=admin key, got %q", input.Content)
			}
			if input.Page != 3 || input.Size != 10 || input.Order != "created_at desc" {
				t.Fatalf("unexpected list input: %+v", input)
			}
			return &dto.PostListResult{Items: []dto.PostDTO{}, Total: 0}, nil
		},
	})

	c, w := newTestContext(http.MethodGet, "/api/v1/admin/posts?user_id=12&tag_id=3&content=%20admin%20key%20&page=3&size=10&order=created_at%20desc")
	h.List(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !called {
		t.Fatal("expected List to be called")
	}
}

func TestPostAdminHandlerGetSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewPostAdminHandler(&mockPostAdminService{
		getFn: func(_ context.Context, id uint) (*dto.PostDTO, error) {
			if id != 8 {
				t.Fatalf("expected id 8, got %d", id)
			}
			return &dto.PostDTO{ID: 8, Content: "hello", Author: dto.PostAuthorDTO{ID: 1}}, nil
		},
	})

	c, w := newTestContext(http.MethodGet, "/api/v1/admin/posts/8")
	c.Params = gin.Params{{Key: "id", Value: "8"}}
	h.Get(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestPostAdminHandlerGetInvalidPathID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewPostAdminHandler(&mockPostAdminService{})
	c, w := newTestContext(http.MethodGet, "/api/v1/admin/posts/0")
	c.Params = gin.Params{{Key: "id", Value: "0"}}
	h.Get(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestPostAdminHandlerCreateSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewPostAdminHandler(&mockPostAdminService{
		createFn: func(_ context.Context, input dto.PostCreateInput) (*dto.PostDTO, error) {
			if input.UserID != 5 {
				t.Fatalf("expected user_id 5, got %d", input.UserID)
			}
			if input.Content != "post content" {
				t.Fatalf("unexpected content: %s", input.Content)
			}
			if !input.Hide {
				t.Fatal("expected hide=true")
			}
			if input.Pin == nil || *input.Pin != 7 {
				t.Fatalf("expected pin=7, got %#v", input.Pin)
			}
			if !input.DisableComment {
				t.Fatal("expected disable_comment=true")
			}
			return &dto.PostDTO{ID: 11, Author: dto.PostAuthorDTO{ID: input.UserID}, Content: input.Content, Hide: input.Hide, DisableComment: input.DisableComment}, nil
		},
	})

	c, w := newTestContextWithBody(http.MethodPost, "/api/v1/admin/posts", `{"user_id":5,"content":"post content","hide":true,"disable_comment":true,"pin":7}`)
	c.Set(middleware.ContextUserIDKey, uint(5))
	h.Create(c)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", w.Code)
	}
}

func TestPostAdminHandlerCreateInvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewPostAdminHandler(&mockPostAdminService{})
	c, w := newTestContextWithBody(http.MethodPost, "/api/v1/admin/posts", `{`)
	h.Create(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestPostAdminHandlerUpdateBindsPathID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewPostAdminHandler(&mockPostAdminService{
		updateFn: func(_ context.Context, input dto.PostUpdateInput) (*dto.PostDTO, error) {
			if input.ID != 9 {
				t.Fatalf("expected id from path 9, got %d", input.ID)
			}
			if input.Content == nil || *input.Content != "updated" {
				t.Fatalf("unexpected content: %#v", input.Content)
			}
			if input.Hide == nil || !*input.Hide {
				t.Fatalf("unexpected hide: %#v", input.Hide)
			}
			if input.DisableComment == nil || !*input.DisableComment {
				t.Fatalf("unexpected disable_comment: %#v", input.DisableComment)
			}
			if input.Pin == nil || *input.Pin != 123 {
				t.Fatalf("unexpected pin: %#v", input.Pin)
			}
			if input.ClearPin == nil || *input.ClearPin {
				t.Fatalf("unexpected clear_pin: %#v", input.ClearPin)
			}
			return &dto.PostDTO{ID: input.ID, Content: *input.Content, Author: dto.PostAuthorDTO{ID: 2}, Hide: *input.Hide, DisableComment: *input.DisableComment, Pin: input.Pin}, nil
		},
	})

	c, w := newTestContextWithBody(http.MethodPatch, "/api/v1/admin/posts/9", `{"id":999,"content":"updated","hide":true,"disable_comment":true,"pin":123,"clear_pin":false}`)
	c.Params = gin.Params{{Key: "id", Value: "9"}}
	h.Update(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestPostAdminHandlerUpdateSupportsClearPin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewPostAdminHandler(&mockPostAdminService{
		updateFn: func(_ context.Context, input dto.PostUpdateInput) (*dto.PostDTO, error) {
			if input.ID != 12 {
				t.Fatalf("expected id from path 12, got %d", input.ID)
			}
			if input.ClearPin == nil || !*input.ClearPin {
				t.Fatalf("expected clear_pin=true, got %#v", input.ClearPin)
			}
			if input.Pin != nil {
				t.Fatalf("expected pin=nil when clear_pin=true, got %#v", input.Pin)
			}
			return &dto.PostDTO{ID: input.ID, Content: "updated", Author: dto.PostAuthorDTO{ID: 2}}, nil
		},
	})

	c, w := newTestContextWithBody(http.MethodPatch, "/api/v1/admin/posts/12", `{"content":"updated","clear_pin":true}`)
	c.Params = gin.Params{{Key: "id", Value: "12"}}
	h.Update(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestPostAdminHandlerDeleteSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	called := false
	h := NewPostAdminHandler(&mockPostAdminService{
		deleteFn: func(_ context.Context, id uint) error {
			called = true
			if id != 10 {
				t.Fatalf("expected id 10, got %d", id)
			}
			return nil
		},
	})

	c, w := newTestContext(http.MethodDelete, "/api/v1/admin/posts/10")
	c.Params = gin.Params{{Key: "id", Value: "10"}}
	h.Delete(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !called {
		t.Fatal("expected Delete to be called")
	}
}

func TestPostAdminHandlerDeleteNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewPostAdminHandler(&mockPostAdminService{
		deleteFn: func(_ context.Context, _ uint) error {
			return service.ErrPostNotFound
		},
	})

	c, w := newTestContext(http.MethodDelete, "/api/v1/admin/posts/10")
	c.Params = gin.Params{{Key: "id", Value: "10"}}
	h.Delete(c)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}

func TestPostAdminHandlerNilService(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewPostAdminHandler(nil)
	c, w := newTestContext(http.MethodGet, "/api/v1/admin/posts")
	h.List(c)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}
