package handler

import (
	"context"
	"net/http"
	"testing"

	"easydrop/internal/dto"
	"easydrop/internal/service"

	"github.com/gin-gonic/gin"
)

func TestCommentAdminHandlerListSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	called := false
	h := NewCommentAdminHandler(&mockCommentServiceForHandler{
		listFn: func(_ context.Context, input dto.CommentAdminListInput) (*dto.CommentListResult, error) {
			called = true
			if input.PostID == nil || *input.PostID != 11 {
				t.Fatalf("expected post_id 11, got %#v", input.PostID)
			}
			if input.UserID == nil || *input.UserID != 7 {
				t.Fatalf("expected user_id 7, got %#v", input.UserID)
			}
			if input.Page != 3 || input.Size != 10 || input.Order != "created_at_desc" {
				t.Fatalf("unexpected list input: %+v", input)
			}
			return &dto.CommentListResult{Items: []dto.CommentDTO{}, Total: 0}, nil
		},
	})

	c, w := newTestContext(http.MethodGet, "/api/v1/admin/comments?post_id=11&user_id=7&page=3&size=10&order=created_at_desc")
	h.List(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !called {
		t.Fatal("expected List to be called")
	}
}

func TestCommentAdminHandlerGetInvalidPathID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewCommentAdminHandler(&mockCommentServiceForHandler{})
	c, w := newTestContext(http.MethodGet, "/api/v1/admin/comments/0")
	c.Params = gin.Params{{Key: "id", Value: "0"}}

	h.Get(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestCommentAdminHandlerUpdateBindsPathID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewCommentAdminHandler(&mockCommentServiceForHandler{
		updateFn: func(_ context.Context, input dto.CommentUpdateInput) (*dto.CommentDTO, error) {
			if input.ID != 6 {
				t.Fatalf("expected path id 6, got %d", input.ID)
			}
			if input.Content == nil || *input.Content != "updated" {
				t.Fatalf("unexpected content: %#v", input.Content)
			}
			return &dto.CommentDTO{ID: input.ID, Content: *input.Content}, nil
		},
	})

	c, w := newTestContextWithBody(http.MethodPatch, "/api/v1/admin/comments/6", `{"id":99,"content":"updated"}`)
	c.Params = gin.Params{{Key: "id", Value: "6"}}

	h.Update(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestCommentAdminHandlerDeleteNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewCommentAdminHandler(&mockCommentServiceForHandler{
		deleteFn: func(_ context.Context, _ uint) error {
			return service.ErrCommentNotFound
		},
	})

	c, w := newTestContext(http.MethodDelete, "/api/v1/admin/comments/8")
	c.Params = gin.Params{{Key: "id", Value: "8"}}

	h.Delete(c)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}

func TestCommentAdminHandlerNilService(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewCommentAdminHandler(nil)
	c, w := newTestContext(http.MethodGet, "/api/v1/admin/comments")

	h.List(c)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}
