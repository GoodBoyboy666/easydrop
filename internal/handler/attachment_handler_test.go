package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
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

type mockSettingService struct {
	getValueFn func(ctx context.Context, key string) (string, bool, error)
}

func (m *mockSettingService) GetValue(ctx context.Context, key string) (string, bool, error) {
	if m.getValueFn == nil {
		return "", false, nil
	}
	return m.getValueFn(ctx, key)
}

func (m *mockSettingService) ListItems(_ context.Context, _ dto.SettingListInput) (*dto.SettingListResult, error) {
	return nil, nil
}

func (m *mockSettingService) UpdateItem(_ context.Context, _ dto.SettingUpdateInput) error {
	return nil
}

func (m *mockSettingService) GetPublicItems(_ context.Context) (*dto.SettingPublicResult, error) {
	return nil, nil
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

func TestAttachmentHandlerUploadPassesOriginalFilename(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var captured dto.AttachmentCreateInput
	h := NewAttachmentHandler(&mockAttachmentService{
		createFn: func(ctx context.Context, input dto.AttachmentCreateInput) (*dto.AttachmentDTO, error) {
			content, err := io.ReadAll(input.Content)
			if err != nil {
				t.Fatalf("read attachment content failed: %v", err)
			}
			input.Content = bytes.NewReader(content)
			captured = input
			return &dto.AttachmentDTO{ID: 1, UserID: input.UserID}, nil
		},
	}, nil)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "Report.Final.PNG")
	if err != nil {
		t.Fatalf("create form file failed: %v", err)
	}
	if _, err := part.Write([]byte("image-content")); err != nil {
		t.Fatalf("write form file failed: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer failed: %v", err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, err := http.NewRequest(http.MethodPost, "/api/v1/attachments", body)
	if err != nil {
		t.Fatalf("create request failed: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	c.Request = req
	c.Set(middleware.ContextUserIDKey, uint(100))

	h.Upload(c)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", w.Code)
	}
	if captured.UserID != 100 {
		t.Fatalf("expected user id 100, got %d", captured.UserID)
	}
	if captured.OriginalFilename != "Report.Final.PNG" {
		t.Fatalf("expected original filename to be passed through, got %q", captured.OriginalFilename)
	}
	if captured.FileSize != int64(len("image-content")) {
		t.Fatalf("expected file size to be populated, got %d", captured.FileSize)
	}
	if len(captured.ContentSample) == 0 {
		t.Fatal("expected uploaded content sample to be passed through")
	}
	content, err := io.ReadAll(captured.Content)
	if err != nil {
		t.Fatalf("read captured content failed: %v", err)
	}
	if len(content) == 0 {
		t.Fatal("expected uploaded content to be passed through")
	}
	if captured.ContentType == "" {
		t.Fatal("expected content type to be populated")
	}
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
	}, nil)

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

	h := NewAttachmentHandler(&mockAttachmentService{}, nil)
	c, w := newTestContext(http.MethodGet, "/api/v1/attachments/12")
	c.Params = gin.Params{{Key: "id", Value: "12"}}

	h.Get(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestAttachmentHandlerGetInvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewAttachmentHandler(&mockAttachmentService{}, nil)
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
	}, nil)

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
			if input.Page != 3 || input.Size != 10 || input.Order != "created_at asc" {
				t.Fatalf("unexpected list input: %+v", input)
			}
			return &dto.AttachmentListResult{Items: []dto.AttachmentDTO{}, Total: 0}, nil
		},
	}, nil)

	c, w := newTestContext(http.MethodGet, "/api/v1/attachments?id=12&biz_type=2&page=3&size=10&order=created_at%20asc")
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
	}, nil)

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
	}, nil)

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

	h := NewAttachmentHandler(&mockAttachmentService{}, nil)
	c, w := newTestContext(http.MethodGet, "/api/v1/attachments?biz_type=abc")
	c.Set(middleware.ContextUserIDKey, uint(100))

	h.List(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestAttachmentHandlerListInvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewAttachmentHandler(&mockAttachmentService{}, nil)
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
	}, nil)

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

func TestAttachmentHandlerUploadBlockedForNonAdminWhenStorageUploadDisabled(t *testing.T) {
	gin.SetMode(gin.TestMode)

	createCalled := false
	h := NewAttachmentHandler(&mockAttachmentService{
		createFn: func(ctx context.Context, input dto.AttachmentCreateInput) (*dto.AttachmentDTO, error) {
			createCalled = true
			return &dto.AttachmentDTO{ID: 1, UserID: input.UserID}, nil
		},
	}, &mockSettingService{
		getValueFn: func(ctx context.Context, key string) (string, bool, error) {
			if key != "storage.upload" {
				t.Fatalf("unexpected setting key: %s", key)
			}
			return "false", true, nil
		},
	})

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "blocked.txt")
	if err != nil {
		t.Fatalf("create form file failed: %v", err)
	}
	if _, err := part.Write([]byte("blocked")); err != nil {
		t.Fatalf("write form file failed: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer failed: %v", err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, err := http.NewRequest(http.MethodPost, "/api/v1/attachments", body)
	if err != nil {
		t.Fatalf("create request failed: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	c.Request = req
	c.Set(middleware.ContextUserIDKey, uint(100))

	h.Upload(c)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", w.Code)
	}
	if createCalled {
		t.Fatal("create should not be called when upload is disabled for non-admin")
	}
}

func TestAttachmentHandlerUploadAllowedForAdminWhenStorageUploadDisabled(t *testing.T) {
	gin.SetMode(gin.TestMode)

	createCalled := false
	h := NewAttachmentHandler(&mockAttachmentService{
		createFn: func(ctx context.Context, input dto.AttachmentCreateInput) (*dto.AttachmentDTO, error) {
			createCalled = true
			return &dto.AttachmentDTO{ID: 1, UserID: input.UserID}, nil
		},
	}, &mockSettingService{
		getValueFn: func(ctx context.Context, key string) (string, bool, error) {
			if key != "storage.upload" {
				t.Fatalf("unexpected setting key: %s", key)
			}
			return "false", true, nil
		},
	})

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "allowed.txt")
	if err != nil {
		t.Fatalf("create form file failed: %v", err)
	}
	if _, err := part.Write([]byte("allowed")); err != nil {
		t.Fatalf("write form file failed: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer failed: %v", err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, err := http.NewRequest(http.MethodPost, "/api/v1/attachments", body)
	if err != nil {
		t.Fatalf("create request failed: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	c.Request = req
	c.Set(middleware.ContextUserIDKey, uint(100))
	c.Set(middleware.ContextUserAdminKey, true)

	h.Upload(c)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", w.Code)
	}
	if !createCalled {
		t.Fatal("create should be called for admin when upload is disabled")
	}
}

func TestAttachmentHandlerUploadValidationErrorReturnsBadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewAttachmentHandler(&mockAttachmentService{
		createFn: func(ctx context.Context, input dto.AttachmentCreateInput) (*dto.AttachmentDTO, error) {
			return nil, service.ErrAttachmentExtensionNotAllowed
		},
	}, nil)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "blocked.txt")
	if err != nil {
		t.Fatalf("create form file failed: %v", err)
	}
	if _, err := part.Write([]byte("blocked")); err != nil {
		t.Fatalf("write form file failed: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer failed: %v", err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, err := http.NewRequest(http.MethodPost, "/api/v1/attachments", body)
	if err != nil {
		t.Fatalf("create request failed: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	c.Request = req
	c.Set(middleware.ContextUserIDKey, uint(100))

	h.Upload(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}
