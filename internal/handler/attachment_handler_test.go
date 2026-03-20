package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"easydrop/internal/dto"
	"easydrop/internal/middleware"
	"easydrop/internal/service"

	"github.com/gin-gonic/gin"
)

type mockAttachmentService struct {
	createFn     func(ctx context.Context, input dto.AttachmentCreateInput) (*dto.AttachmentDTO, error)
	getFn        func(ctx context.Context, id uint) (*dto.AttachmentDTO, error)
	updateFn     func(ctx context.Context, input dto.AttachmentUpdateInput) (*dto.AttachmentDTO, error)
	deleteFn     func(ctx context.Context, id uint) error
	listByUserFn func(ctx context.Context, input dto.AttachmentListInput) (*dto.AttachmentListResult, error)
}

func (m *mockAttachmentService) Create(ctx context.Context, input dto.AttachmentCreateInput) (*dto.AttachmentDTO, error) {
	if m.createFn == nil {
		return nil, nil
	}
	return m.createFn(ctx, input)
}

func (m *mockAttachmentService) Get(ctx context.Context, id uint) (*dto.AttachmentDTO, error) {
	if m.getFn == nil {
		return nil, nil
	}
	return m.getFn(ctx, id)
}

func (m *mockAttachmentService) Update(ctx context.Context, input dto.AttachmentUpdateInput) (*dto.AttachmentDTO, error) {
	if m.updateFn == nil {
		return nil, nil
	}
	return m.updateFn(ctx, input)
}

func (m *mockAttachmentService) Delete(ctx context.Context, id uint) error {
	if m.deleteFn == nil {
		return nil
	}
	return m.deleteFn(ctx, id)
}

func (m *mockAttachmentService) ListByUser(ctx context.Context, input dto.AttachmentListInput) (*dto.AttachmentListResult, error) {
	if m.listByUserFn == nil {
		return nil, nil
	}
	return m.listByUserFn(ctx, input)
}

func TestAttachmentHandlerGetSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewAttachmentHandler(&mockAttachmentService{
		getFn: func(ctx context.Context, id uint) (*dto.AttachmentDTO, error) {
			if id != 12 {
				t.Fatalf("expected id 12, got %d", id)
			}
			return &dto.AttachmentDTO{ID: 12, UserID: 100, CreatedAt: time.Now()}, nil
		},
	})

	c, w := newTestContext(http.MethodGet, "/api/v1/attachments/12")
	c.Params = gin.Params{{Key: "id", Value: "12"}}
	c.Set(middleware.ContextUserIDKey, uint(100))

	h.Get(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var body dto.AttachmentDTO
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if body.ID != 12 {
		t.Fatalf("expected id 12, got %d", body.ID)
	}
}

func TestAttachmentHandlerGetUnauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewAttachmentHandler(&mockAttachmentService{})
	c, w := newTestContext(http.MethodGet, "/api/v1/attachments/12")
	c.Params = gin.Params{{Key: "id", Value: "12"}}

	h.Get(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestAttachmentHandlerGetInvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewAttachmentHandler(&mockAttachmentService{})
	c, w := newTestContext(http.MethodGet, "/api/v1/attachments/invalid")
	c.Params = gin.Params{{Key: "id", Value: "invalid"}}
	c.Set(middleware.ContextUserIDKey, uint(100))

	h.Get(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestAttachmentHandlerGetForbiddenByOwnerCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewAttachmentHandler(&mockAttachmentService{
		getFn: func(ctx context.Context, id uint) (*dto.AttachmentDTO, error) {
			return &dto.AttachmentDTO{ID: id, UserID: 200}, nil
		},
	})

	c, w := newTestContext(http.MethodGet, "/api/v1/attachments/12")
	c.Params = gin.Params{{Key: "id", Value: "12"}}
	c.Set(middleware.ContextUserIDKey, uint(100))

	h.Get(c)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}

func TestAttachmentHandlerListSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	called := false
	h := NewAttachmentHandler(&mockAttachmentService{
		listByUserFn: func(ctx context.Context, input dto.AttachmentListInput) (*dto.AttachmentListResult, error) {
			called = true
			if input.ID == nil || *input.ID != 12 {
				t.Fatalf("expected attachment id 12, got %#v", input.ID)
			}
			if input.UserID == nil || *input.UserID != 100 {
				t.Fatalf("expected user id 100, got %#v", input.UserID)
			}
			if input.BizType == nil || *input.BizType != 2 {
				t.Fatalf("expected biz type 2, got %#v", input.BizType)
			}
			if input.Limit != 10 || input.Offset != 20 || input.Order != "created_at asc" {
				t.Fatalf("unexpected list input: %+v", input)
			}
			return &dto.AttachmentListResult{Items: []dto.AttachmentDTO{}, Total: 0}, nil
		},
	})

	c, w := newTestContext(http.MethodGet, "/api/v1/attachments?id=12&biz_type=2&limit=10&offset=20&order=created_at%20asc")
	c.Set(middleware.ContextUserIDKey, uint(100))

	h.List(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !called {
		t.Fatal("expected ListByUser to be called")
	}
}

func TestAttachmentHandlerDeleteSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	deleteCalled := false
	h := NewAttachmentHandler(&mockAttachmentService{
		getFn: func(ctx context.Context, id uint) (*dto.AttachmentDTO, error) {
			return &dto.AttachmentDTO{ID: id, UserID: 100}, nil
		},
		deleteFn: func(ctx context.Context, id uint) error {
			deleteCalled = true
			if id != 9 {
				t.Fatalf("expected id 9, got %d", id)
			}
			return nil
		},
	})

	c, w := newTestContext(http.MethodDelete, "/api/v1/attachments/9")
	c.Params = gin.Params{{Key: "id", Value: "9"}}
	c.Set(middleware.ContextUserIDKey, uint(100))

	h.Delete(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !deleteCalled {
		t.Fatal("expected Delete to be called")
	}
}

func TestAttachmentHandlerDeleteNotOwner(t *testing.T) {
	gin.SetMode(gin.TestMode)

	deleteCalled := false
	h := NewAttachmentHandler(&mockAttachmentService{
		getFn: func(ctx context.Context, id uint) (*dto.AttachmentDTO, error) {
			return &dto.AttachmentDTO{ID: id, UserID: 999}, nil
		},
		deleteFn: func(ctx context.Context, id uint) error {
			deleteCalled = true
			return nil
		},
	})

	c, w := newTestContext(http.MethodDelete, "/api/v1/attachments/9")
	c.Params = gin.Params{{Key: "id", Value: "9"}}
	c.Set(middleware.ContextUserIDKey, uint(100))

	h.Delete(c)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
	if deleteCalled {
		t.Fatal("delete should not be called for non-owner")
	}
}

func TestAttachmentHandlerListInvalidBizType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewAttachmentHandler(&mockAttachmentService{})
	c, w := newTestContext(http.MethodGet, "/api/v1/attachments?biz_type=abc")
	c.Set(middleware.ContextUserIDKey, uint(100))

	h.List(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestAttachmentHandlerListInvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewAttachmentHandler(&mockAttachmentService{})
	c, w := newTestContext(http.MethodGet, "/api/v1/attachments?id=0")
	c.Set(middleware.ContextUserIDKey, uint(100))

	h.List(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestAttachmentHandlerDeleteServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewAttachmentHandler(&mockAttachmentService{
		getFn: func(ctx context.Context, id uint) (*dto.AttachmentDTO, error) {
			return &dto.AttachmentDTO{ID: id, UserID: 100}, nil
		},
		deleteFn: func(ctx context.Context, id uint) error {
			return service.ErrInternal
		},
	})

	c, w := newTestContext(http.MethodDelete, "/api/v1/attachments/9")
	c.Params = gin.Params{{Key: "id", Value: "9"}}
	c.Set(middleware.ContextUserIDKey, uint(100))

	h.Delete(c)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func newTestContext(method, target string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest(method, target, nil)
	c.Request = req
	return c, w
}
