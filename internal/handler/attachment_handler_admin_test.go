package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"easydrop/internal/dto"
	"easydrop/internal/service"

	"github.com/gin-gonic/gin"
)

func TestAttachmentAdminHandlerListSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	called := false
	h := NewAttachmentAdminHandler(&mockAttachmentService{
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
			if input.CreatedFrom == nil || *input.CreatedFrom != 1700000000 {
				t.Fatalf("expected created_from 1700000000, got %#v", input.CreatedFrom)
			}
			if input.CreatedTo == nil || *input.CreatedTo != 1800000000 {
				t.Fatalf("expected created_to 1800000000, got %#v", input.CreatedTo)
			}
			if input.Limit != 10 || input.Offset != 20 || input.Order != "created_at_desc" {
				t.Fatalf("unexpected list input: %+v", input)
			}
			return &dto.AttachmentListResult{Items: []dto.AttachmentDTO{}, Total: 0}, nil
		},
	})

	c, w := newTestContext(http.MethodGet, "/api/v1/admin/attachments?id=12&user_id=100&biz_type=2&created_from=1700000000&created_to=1800000000&limit=10&offset=20&order=created_at_desc")

	h.List(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !called {
		t.Fatal("expected ListByUser to be called")
	}
}

func TestAttachmentAdminHandlerListInvalidTimeRange(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewAttachmentAdminHandler(&mockAttachmentService{})
	c, w := newTestContext(http.MethodGet, "/api/v1/admin/attachments?created_from=1800000000&created_to=1700000000")

	h.List(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestAttachmentAdminHandlerListInvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewAttachmentAdminHandler(&mockAttachmentService{})
	c, w := newTestContext(http.MethodGet, "/api/v1/admin/attachments?id=0")

	h.List(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestAttachmentAdminHandlerDeleteSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	deleteCalled := false
	h := NewAttachmentAdminHandler(&mockAttachmentService{
		deleteFn: func(ctx context.Context, id uint) error {
			deleteCalled = true
			if id != 9 {
				t.Fatalf("expected id 9, got %d", id)
			}
			return nil
		},
	})

	c, w := newTestContext(http.MethodDelete, "/api/v1/admin/attachments/9")
	c.Params = gin.Params{{Key: "id", Value: "9"}}

	h.Delete(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !deleteCalled {
		t.Fatal("expected Delete to be called")
	}
}

func TestAttachmentAdminHandlerBatchDeletePartialFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewAttachmentAdminHandler(&mockAttachmentService{
		deleteFn: func(ctx context.Context, id uint) error {
			if id == 2 {
				return service.ErrAttachmentNotFound
			}
			return nil
		},
	})

	c, w := newTestContextWithBody(http.MethodPost, "/api/v1/admin/attachments/batch-delete", `{"ids":[1,2,3]}`)

	h.BatchDelete(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var body dto.AttachmentBatchDeleteResult
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}

	if len(body.SuccessIDs) != 2 {
		t.Fatalf("expected 2 success ids, got %d", len(body.SuccessIDs))
	}
	if len(body.Failed) != 1 || body.Failed[0].ID != 2 {
		t.Fatalf("unexpected failed items: %#v", body.Failed)
	}
}

func newTestContextWithBody(method, target, body string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest(method, target, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	return c, w
}
