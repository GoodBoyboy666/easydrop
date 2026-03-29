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

type mockCommentServiceForHandler struct {
	createFn     func(ctx context.Context, input dto.CommentCreateInput) (*dto.CommentDTO, error)
	getFn        func(ctx context.Context, id uint) (*dto.CommentDTO, error)
	updateFn     func(ctx context.Context, input dto.CommentUpdateInput) (*dto.CommentDTO, error)
	deleteFn     func(ctx context.Context, id uint) error
	listByPostFn func(ctx context.Context, input dto.CommentListInput) (*dto.CommentListResult, error)
	listByUserFn func(ctx context.Context, input dto.CommentUserListInput) (*dto.CommentListResult, error)
	listFn       func(ctx context.Context, input dto.CommentAdminListInput) (*dto.CommentListResult, error)
	listPublicFn func(ctx context.Context, input dto.CommentPublicListInput) (*dto.CommentListResult, error)
}

func (m *mockCommentServiceForHandler) Create(ctx context.Context, input dto.CommentCreateInput) (*dto.CommentDTO, error) {
	if m.createFn == nil {
		return nil, nil
	}
	return m.createFn(ctx, input)
}

func (m *mockCommentServiceForHandler) Get(ctx context.Context, id uint) (*dto.CommentDTO, error) {
	if m.getFn == nil {
		return nil, nil
	}
	return m.getFn(ctx, id)
}

func (m *mockCommentServiceForHandler) Update(ctx context.Context, input dto.CommentUpdateInput) (*dto.CommentDTO, error) {
	if m.updateFn == nil {
		return nil, nil
	}
	return m.updateFn(ctx, input)
}

func (m *mockCommentServiceForHandler) Delete(ctx context.Context, id uint) error {
	if m.deleteFn == nil {
		return nil
	}
	return m.deleteFn(ctx, id)
}

func (m *mockCommentServiceForHandler) ListByPost(ctx context.Context, input dto.CommentListInput) (*dto.CommentListResult, error) {
	if m.listByPostFn == nil {
		return nil, nil
	}
	return m.listByPostFn(ctx, input)
}

func (m *mockCommentServiceForHandler) ListByUser(ctx context.Context, input dto.CommentUserListInput) (*dto.CommentListResult, error) {
	if m.listByUserFn == nil {
		return nil, nil
	}
	return m.listByUserFn(ctx, input)
}

func (m *mockCommentServiceForHandler) List(ctx context.Context, input dto.CommentAdminListInput) (*dto.CommentListResult, error) {
	if m.listFn == nil {
		return nil, nil
	}
	return m.listFn(ctx, input)
}

func (m *mockCommentServiceForHandler) ListPublic(ctx context.Context, input dto.CommentPublicListInput) (*dto.CommentListResult, error) {
	if m.listPublicFn == nil {
		return nil, nil
	}
	return m.listPublicFn(ctx, input)
}

func TestCommentHandlerCreateSetsCurrentUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewCommentHandler(&mockCommentServiceForHandler{
		createFn: func(_ context.Context, input dto.CommentCreateInput) (*dto.CommentDTO, error) {
			if input.UserID != 101 {
				t.Fatalf("expected user id 101, got %d", input.UserID)
			}
			if input.PostID != 9 {
				t.Fatalf("expected post id 9, got %d", input.PostID)
			}
			if input.CanViewHidden {
				t.Fatal("expected normal user create request to disallow hidden posts")
			}
			return &dto.CommentDTO{ID: 1, Author: dto.CommentAuthorDTO{ID: input.UserID}, PostID: input.PostID}, nil
		},
	})

	c, w := newTestContextWithBody(http.MethodPost, "/api/v1/posts/9/comments", `{"post_id":7,"user_id":999,"content":"hello"}`)
	c.Params = gin.Params{{Key: "id", Value: "9"}}
	c.Set(middleware.ContextUserIDKey, uint(101))

	h.Create(c)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", w.Code)
	}
}

func TestCommentHandlerCreateAdminCanViewHiddenPost(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewCommentHandler(&mockCommentServiceForHandler{
		createFn: func(_ context.Context, input dto.CommentCreateInput) (*dto.CommentDTO, error) {
			if !input.CanViewHidden {
				t.Fatal("expected admin create request to allow hidden posts")
			}
			return &dto.CommentDTO{ID: 1, Author: dto.CommentAuthorDTO{ID: input.UserID}, PostID: input.PostID}, nil
		},
	})

	c, w := newTestContextWithBody(http.MethodPost, "/api/v1/posts/9/comments", `{"content":"hello"}`)
	c.Params = gin.Params{{Key: "id", Value: "9"}}
	c.Set(middleware.ContextUserIDKey, uint(101))
	c.Set(middleware.ContextUserAdminKey, true)

	h.Create(c)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", w.Code)
	}
}

func TestCommentHandlerCreateInvalidPathID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	createCalled := false
	h := NewCommentHandler(&mockCommentServiceForHandler{
		createFn: func(_ context.Context, _ dto.CommentCreateInput) (*dto.CommentDTO, error) {
			createCalled = true
			return &dto.CommentDTO{}, nil
		},
	})

	c, w := newTestContextWithBody(http.MethodPost, "/api/v1/posts/0/comments", `{"content":"hello"}`)
	c.Params = gin.Params{{Key: "id", Value: "0"}}
	c.Set(middleware.ContextUserIDKey, uint(101))

	h.Create(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
	if createCalled {
		t.Fatal("create should not be called when path id is invalid")
	}
}

func TestCommentHandlerCreateWhenPostCommentDisabled(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewCommentHandler(&mockCommentServiceForHandler{
		createFn: func(_ context.Context, _ dto.CommentCreateInput) (*dto.CommentDTO, error) {
			return nil, service.ErrPostCommentDisabled
		},
	})

	c, w := newTestContextWithBody(http.MethodPost, "/api/v1/posts/9/comments", `{"content":"hello"}`)
	c.Params = gin.Params{{Key: "id", Value: "9"}}
	c.Set(middleware.ContextUserIDKey, uint(101))

	h.Create(c)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", w.Code)
	}
}

func TestCommentHandlerGetNotOwner(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewCommentHandler(&mockCommentServiceForHandler{
		getFn: func(_ context.Context, id uint) (*dto.CommentDTO, error) {
			return &dto.CommentDTO{ID: id, Author: dto.CommentAuthorDTO{ID: 202}}, nil
		},
	})

	c, w := newTestContext(http.MethodGet, "/api/v1/comments/7")
	c.Params = gin.Params{{Key: "id", Value: "7"}}
	c.Set(middleware.ContextUserIDKey, uint(101))

	h.Get(c)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}

func TestCommentHandlerUpdateOnlyOwnComment(t *testing.T) {
	gin.SetMode(gin.TestMode)

	updateCalled := false
	h := NewCommentHandler(&mockCommentServiceForHandler{
		getFn: func(_ context.Context, id uint) (*dto.CommentDTO, error) {
			return &dto.CommentDTO{ID: id, Author: dto.CommentAuthorDTO{ID: 999}}, nil
		},
		updateFn: func(_ context.Context, _ dto.CommentUpdateInput) (*dto.CommentDTO, error) {
			updateCalled = true
			return &dto.CommentDTO{}, nil
		},
	})

	c, w := newTestContextWithBody(http.MethodPatch, "/api/v1/comments/8", `{"content":"updated"}`)
	c.Params = gin.Params{{Key: "id", Value: "8"}}
	c.Set(middleware.ContextUserIDKey, uint(101))

	h.Update(c)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
	if updateCalled {
		t.Fatal("update should not be called when current user is not owner")
	}
}

func TestCommentHandlerDeleteSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	deleteCalled := false
	h := NewCommentHandler(&mockCommentServiceForHandler{
		getFn: func(_ context.Context, id uint) (*dto.CommentDTO, error) {
			return &dto.CommentDTO{ID: id, Author: dto.CommentAuthorDTO{ID: 101}}, nil
		},
		deleteFn: func(_ context.Context, id uint) error {
			deleteCalled = true
			if id != 8 {
				t.Fatalf("expected id 8, got %d", id)
			}
			return nil
		},
	})

	c, w := newTestContext(http.MethodDelete, "/api/v1/comments/8")
	c.Params = gin.Params{{Key: "id", Value: "8"}}
	c.Set(middleware.ContextUserIDKey, uint(101))

	h.Delete(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !deleteCalled {
		t.Fatal("expected delete to be called")
	}
}

func TestCommentHandlerListSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	called := false
	h := NewCommentHandler(&mockCommentServiceForHandler{
		listByUserFn: func(_ context.Context, input dto.CommentUserListInput) (*dto.CommentListResult, error) {
			called = true
			if input.UserID != 101 {
				t.Fatalf("expected user id 101, got %d", input.UserID)
			}
			if input.PostID == nil || *input.PostID != 9 {
				t.Fatalf("expected post id filter 9, got %#v", input.PostID)
			}
			if input.CanViewHidden {
				t.Fatal("expected normal user list request to disallow hidden posts")
			}
			if input.Page != 3 || input.Size != 10 || input.Order != "created_at_asc" {
				t.Fatalf("unexpected list input: %+v", input)
			}
			return &dto.CommentListResult{Items: []dto.CommentDTO{}, Total: 0}, nil
		},
	})

	c, w := newTestContext(http.MethodGet, "/api/v1/users/me/comments?post_id=9&page=3&size=10&order=created_at_asc")
	c.Set(middleware.ContextUserIDKey, uint(101))

	h.List(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !called {
		t.Fatal("expected ListByUser to be called")
	}
}

func TestCommentHandlerListAdminCanViewHiddenPost(t *testing.T) {
	gin.SetMode(gin.TestMode)

	called := false
	h := NewCommentHandler(&mockCommentServiceForHandler{
		listByUserFn: func(_ context.Context, input dto.CommentUserListInput) (*dto.CommentListResult, error) {
			called = true
			if !input.CanViewHidden {
				t.Fatal("expected admin list request to allow hidden posts")
			}
			return &dto.CommentListResult{Items: []dto.CommentDTO{}, Total: 0}, nil
		},
	})

	c, w := newTestContext(http.MethodGet, "/api/v1/users/me/comments?post_id=9")
	c.Set(middleware.ContextUserIDKey, uint(101))
	c.Set(middleware.ContextUserAdminKey, true)

	h.List(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !called {
		t.Fatal("expected ListByUser to be called")
	}
}

func TestCommentHandlerListPublicSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	called := false
	h := NewCommentHandler(&mockCommentServiceForHandler{
		listPublicFn: func(_ context.Context, input dto.CommentPublicListInput) (*dto.CommentListResult, error) {
			called = true
			if input.CanViewHidden {
				t.Fatal("expected anonymous public list to disallow hidden posts")
			}
			if input.Page != 3 || input.Size != 10 || input.Order != "created_at_desc" {
				t.Fatalf("unexpected public list input: %+v", input)
			}
			return &dto.CommentListResult{Items: []dto.CommentDTO{}, Total: 0}, nil
		},
	})

	c, w := newTestContext(http.MethodGet, "/api/v1/comments?page=3&size=10&order=created_at_desc")

	h.ListPublic(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !called {
		t.Fatal("expected ListPublic to be called")
	}
}

func TestCommentHandlerListPublicRejectsUnknownQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)

	called := false
	h := NewCommentHandler(&mockCommentServiceForHandler{
		listPublicFn: func(_ context.Context, input dto.CommentPublicListInput) (*dto.CommentListResult, error) {
			called = true
			if input.CanViewHidden {
				t.Fatal("expected anonymous public list to keep hidden posts filtered")
			}
			if input.Order != "" || input.Page != 0 || input.Size != 0 {
				t.Fatalf("unexpected list defaults: %+v", input)
			}
			return &dto.CommentListResult{Items: []dto.CommentDTO{}, Total: 0}, nil
		},
	})

	c, w := newTestContext(http.MethodGet, "/api/v1/comments?post_id=9&user_id=2")

	h.ListPublic(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !called {
		t.Fatal("expected ListPublic to be called")
	}
}

func TestCommentHandlerListPublicAdminCanViewHiddenPosts(t *testing.T) {
	gin.SetMode(gin.TestMode)

	called := false
	h := NewCommentHandler(&mockCommentServiceForHandler{
		listPublicFn: func(_ context.Context, input dto.CommentPublicListInput) (*dto.CommentListResult, error) {
			called = true
			if !input.CanViewHidden {
				t.Fatal("expected admin viewer to allow hidden posts in public comment list")
			}
			return &dto.CommentListResult{Items: []dto.CommentDTO{}, Total: 0}, nil
		},
	})

	c, w := newTestContext(http.MethodGet, "/api/v1/comments")
	c.Set(middleware.ContextUserIDKey, uint(101))
	c.Set(middleware.ContextUserAdminKey, true)

	h.ListPublic(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !called {
		t.Fatal("expected ListPublic to be called")
	}
}

func TestCommentHandlerListByPostSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	called := false
	h := NewCommentHandler(&mockCommentServiceForHandler{
		listByPostFn: func(_ context.Context, input dto.CommentListInput) (*dto.CommentListResult, error) {
			called = true
			if input.PostID != 9 {
				t.Fatalf("expected post id 9, got %d", input.PostID)
			}
			if input.CanViewHidden {
				t.Fatal("expected normal public viewer to disallow hidden posts")
			}
			if input.Page != 3 || input.Size != 10 || input.Order != "created_at_desc" {
				t.Fatalf("unexpected list input: %+v", input)
			}
			return &dto.CommentListResult{Items: []dto.CommentDTO{}, Total: 0}, nil
		},
	})

	c, w := newTestContext(http.MethodGet, "/api/v1/posts/9/comments?page=3&size=10&order=created_at_desc")
	c.Params = gin.Params{{Key: "id", Value: "9"}}

	h.ListByPost(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !called {
		t.Fatal("expected ListByPost to be called")
	}
}

func TestCommentHandlerListByPostAdminCanViewHiddenPost(t *testing.T) {
	gin.SetMode(gin.TestMode)

	called := false
	h := NewCommentHandler(&mockCommentServiceForHandler{
		listByPostFn: func(_ context.Context, input dto.CommentListInput) (*dto.CommentListResult, error) {
			called = true
			if !input.CanViewHidden {
				t.Fatal("expected admin viewer to allow hidden posts")
			}
			return &dto.CommentListResult{Items: []dto.CommentDTO{}, Total: 0}, nil
		},
	})

	c, w := newTestContext(http.MethodGet, "/api/v1/posts/9/comments")
	c.Params = gin.Params{{Key: "id", Value: "9"}}
	c.Set(middleware.ContextUserIDKey, uint(101))
	c.Set(middleware.ContextUserAdminKey, true)

	h.ListByPost(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !called {
		t.Fatal("expected ListByPost to be called")
	}
}

func TestCommentHandlerListByPostInvalidPathID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewCommentHandler(&mockCommentServiceForHandler{})
	c, w := newTestContext(http.MethodGet, "/api/v1/posts/0/comments")
	c.Params = gin.Params{{Key: "id", Value: "0"}}

	h.ListByPost(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestCommentHandlerListByPostServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewCommentHandler(&mockCommentServiceForHandler{
		listByPostFn: func(_ context.Context, _ dto.CommentListInput) (*dto.CommentListResult, error) {
			return nil, service.ErrPostNotFound
		},
	})

	c, w := newTestContext(http.MethodGet, "/api/v1/posts/9/comments")
	c.Params = gin.Params{{Key: "id", Value: "9"}}

	h.ListByPost(c)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}

func TestCommentHandlerUnauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewCommentHandler(&mockCommentServiceForHandler{})
	c, w := newTestContext(http.MethodGet, "/api/v1/users/me/comments/1")
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	h.Get(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestMapCommentErrorStatus(t *testing.T) {
	if got := mapCommentErrorStatus(service.ErrInvalidCommentPost); got != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", got)
	}
	if got := mapCommentErrorStatus(service.ErrCommentNotFound); got != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", got)
	}
	if got := mapCommentErrorStatus(service.ErrPostCommentDisabled); got != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", got)
	}
	if got := mapCommentErrorStatus(service.ErrInternal); got != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", got)
	}
}
