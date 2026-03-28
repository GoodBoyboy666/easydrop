package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"testing"

	"easydrop/internal/dto"
	"easydrop/internal/middleware"
	"easydrop/internal/service"
)

type mockUserServiceForHandler struct {
	getFn                func(ctx context.Context, id uint) (*dto.UserDTO, error)
	updateProfileFn      func(ctx context.Context, input dto.UserProfileUpdateInput) (*dto.UserDTO, error)
	changePasswordFn     func(ctx context.Context, input dto.UserChangePasswordInput) error
	requestEmailChangeFn func(ctx context.Context, input dto.UserChangeEmailInput) error
	confirmEmailChangeFn func(ctx context.Context, input dto.UserChangeEmailConfirmInput) (*dto.UserDTO, error)
	uploadAvatarFn       func(ctx context.Context, input dto.UserAvatarUploadInput) (*dto.UserDTO, error)
	deleteAvatarFn       func(ctx context.Context, userID uint) error
}

func (m *mockUserServiceForHandler) Create(context.Context, dto.UserCreateInput) (*dto.UserDTO, error) {
	return nil, nil
}

func (m *mockUserServiceForHandler) Get(ctx context.Context, id uint) (*dto.UserDTO, error) {
	if m.getFn == nil {
		return nil, nil
	}
	return m.getFn(ctx, id)
}

func (m *mockUserServiceForHandler) UpdateProfile(ctx context.Context, input dto.UserProfileUpdateInput) (*dto.UserDTO, error) {
	if m.updateProfileFn == nil {
		return nil, nil
	}
	return m.updateProfileFn(ctx, input)
}

func (m *mockUserServiceForHandler) ChangePassword(ctx context.Context, input dto.UserChangePasswordInput) error {
	if m.changePasswordFn == nil {
		return nil
	}
	return m.changePasswordFn(ctx, input)
}

func (m *mockUserServiceForHandler) RequestEmailChange(ctx context.Context, input dto.UserChangeEmailInput) error {
	if m.requestEmailChangeFn == nil {
		return nil
	}
	return m.requestEmailChangeFn(ctx, input)
}

func (m *mockUserServiceForHandler) ConfirmEmailChange(ctx context.Context, input dto.UserChangeEmailConfirmInput) (*dto.UserDTO, error) {
	if m.confirmEmailChangeFn == nil {
		return nil, nil
	}
	return m.confirmEmailChangeFn(ctx, input)
}

func (m *mockUserServiceForHandler) Update(context.Context, dto.UserUpdateInput) (*dto.UserDTO, error) {
	return nil, nil
}

func (m *mockUserServiceForHandler) UploadAvatar(ctx context.Context, input dto.UserAvatarUploadInput) (*dto.UserDTO, error) {
	if m.uploadAvatarFn == nil {
		return nil, nil
	}
	return m.uploadAvatarFn(ctx, input)
}

func (m *mockUserServiceForHandler) DeleteAvatar(ctx context.Context, userID uint) error {
	if m.deleteAvatarFn == nil {
		return nil
	}
	return m.deleteAvatarFn(ctx, userID)
}

func (m *mockUserServiceForHandler) Delete(context.Context, uint) error {
	return nil
}

func (m *mockUserServiceForHandler) List(context.Context, dto.UserListInput) (*dto.UserListResult, error) {
	return nil, nil
}

func TestUserHandlerGetProfileSuccess(t *testing.T) {
	h := NewUserHandler(&mockUserServiceForHandler{
		getFn: func(_ context.Context, id uint) (*dto.UserDTO, error) {
			if id != 7 {
				t.Fatalf("expected user id 7, got %d", id)
			}
			return &dto.UserDTO{ID: 7, Username: "neo"}, nil
		},
	})

	c, w := newTestContext(http.MethodGet, "/api/v1/users/me")
	c.Set(middleware.ContextUserIDKey, uint(7))

	h.GetProfile(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestUserHandlerGetProfileUnauthorized(t *testing.T) {
	h := NewUserHandler(&mockUserServiceForHandler{})
	c, w := newTestContext(http.MethodGet, "/api/v1/users/me")

	h.GetProfile(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestUserHandlerUpdateProfileSuccess(t *testing.T) {
	h := NewUserHandler(&mockUserServiceForHandler{
		updateProfileFn: func(_ context.Context, input dto.UserProfileUpdateInput) (*dto.UserDTO, error) {
			if input.UserID != 9 {
				t.Fatalf("expected user id 9, got %d", input.UserID)
			}
			if input.Nickname == nil || *input.Nickname != "Neo" {
				t.Fatalf("unexpected nickname: %#v", input.Nickname)
			}
			return &dto.UserDTO{ID: 9, Nickname: "Neo"}, nil
		},
	})

	c, w := newTestContextWithBody(http.MethodPatch, "/api/v1/users/me/profile", `{"nickname":"Neo"}`)
	c.Set(middleware.ContextUserIDKey, uint(9))

	h.UpdateProfile(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestUserHandlerChangePasswordInvalidPassword(t *testing.T) {
	h := NewUserHandler(&mockUserServiceForHandler{
		changePasswordFn: func(_ context.Context, _ dto.UserChangePasswordInput) error {
			return service.ErrInvalidPassword
		},
	})

	c, w := newTestContextWithBody(http.MethodPatch, "/api/v1/users/me/password", `{"old_password":"bad","new_password":"NewPass123"}`)
	c.Set(middleware.ContextUserIDKey, uint(11))

	h.ChangePassword(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestUserHandlerRequestEmailChangeSuccess(t *testing.T) {
	h := NewUserHandler(&mockUserServiceForHandler{
		requestEmailChangeFn: func(_ context.Context, input dto.UserChangeEmailInput) error {
			if input.UserID != 12 {
				t.Fatalf("expected user id 12, got %d", input.UserID)
			}
			if input.CurrentPassword != "OldPass123" {
				t.Fatalf("unexpected current password: %s", input.CurrentPassword)
			}
			if input.NewEmail != "new@example.com" {
				t.Fatalf("unexpected new email: %s", input.NewEmail)
			}
			return nil
		},
	})

	c, w := newTestContextWithBody(http.MethodPost, "/api/v1/users/me/email-change", `{"current_password":"OldPass123","new_email":"new@example.com"}`)
	c.Set(middleware.ContextUserIDKey, uint(12))

	h.RequestEmailChange(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestUserHandlerUploadAvatarSuccess(t *testing.T) {
	h := NewUserHandler(&mockUserServiceForHandler{
		uploadAvatarFn: func(_ context.Context, input dto.UserAvatarUploadInput) (*dto.UserDTO, error) {
			if input.UserID != 18 {
				t.Fatalf("expected user id 18, got %d", input.UserID)
			}
			if input.OriginalFilename != "avatar.png" {
				t.Fatalf("unexpected filename: %s", input.OriginalFilename)
			}
			if input.FileSize != int64(len("avatar-content")) {
				t.Fatalf("unexpected file size: %d", input.FileSize)
			}
			if len(input.ContentSample) == 0 {
				t.Fatal("expected avatar content sample")
			}
			content, err := io.ReadAll(input.Content)
			if err != nil {
				t.Fatalf("read avatar content failed: %v", err)
			}
			if len(content) == 0 {
				t.Fatal("expected avatar content")
			}
			return &dto.UserDTO{ID: 18}, nil
		},
	})

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("avatar", "avatar.png")
	if err != nil {
		t.Fatalf("create form file failed: %v", err)
	}
	if _, err := part.Write([]byte("avatar-content")); err != nil {
		t.Fatalf("write form file failed: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer failed: %v", err)
	}

	c, w := newTestContext(http.MethodPost, "/api/v1/users/me/avatar")
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/users/me/avatar", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	c.Request = req
	c.Set(middleware.ContextUserIDKey, uint(18))

	h.UploadAvatar(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestUserHandlerUploadAvatarMissingFile(t *testing.T) {
	h := NewUserHandler(&mockUserServiceForHandler{})
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer failed: %v", err)
	}

	c, w := newTestContext(http.MethodPost, "/api/v1/users/me/avatar")
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/users/me/avatar", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	c.Request = req
	c.Set(middleware.ContextUserIDKey, uint(18))

	h.UploadAvatar(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestUserHandlerDeleteAvatarSuccess(t *testing.T) {
	h := NewUserHandler(&mockUserServiceForHandler{
		deleteAvatarFn: func(_ context.Context, userID uint) error {
			if userID != 21 {
				t.Fatalf("expected user id 21, got %d", userID)
			}
			return nil
		},
	})

	c, w := newTestContext(http.MethodDelete, "/api/v1/users/me/avatar")
	c.Set(middleware.ContextUserIDKey, uint(21))

	h.DeleteAvatar(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestUserHandlerDeleteAvatarUnauthorized(t *testing.T) {
	h := NewUserHandler(&mockUserServiceForHandler{})
	c, w := newTestContext(http.MethodDelete, "/api/v1/users/me/avatar")

	h.DeleteAvatar(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestMapUserErrorStatus(t *testing.T) {
	if got := mapUserErrorStatus(service.ErrEmailExists); got != http.StatusConflict {
		t.Fatalf("expected 409, got %d", got)
	}
	if got := mapUserErrorStatus(service.ErrStorageQuotaExceeded); got != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", got)
	}
	if got := mapUserErrorStatus(service.ErrInvalidPassword); got != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", got)
	}
}

func TestUserHandlerResponseBodyFormat(t *testing.T) {
	h := NewUserHandler(&mockUserServiceForHandler{
		requestEmailChangeFn: func(_ context.Context, _ dto.UserChangeEmailInput) error {
			return nil
		},
	})
	c, w := newTestContextWithBody(http.MethodPost, "/api/v1/users/me/email-change", `{"current_password":"OldPass123","new_email":"new@example.com"}`)
	c.Set(middleware.ContextUserIDKey, uint(30))

	h.RequestEmailChange(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp dto.ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if resp.Message != "ok" {
		t.Fatalf("expected message ok, got %q", resp.Message)
	}
}

func TestUserHandlerUploadAvatarValidationError(t *testing.T) {
	h := NewUserHandler(&mockUserServiceForHandler{
		uploadAvatarFn: func(_ context.Context, input dto.UserAvatarUploadInput) (*dto.UserDTO, error) {
			return nil, service.ErrAvatarExtensionNotAllowed
		},
	})

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("avatar", "avatar.gif")
	if err != nil {
		t.Fatalf("create form file failed: %v", err)
	}
	if _, err := part.Write([]byte("avatar-content")); err != nil {
		t.Fatalf("write form file failed: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer failed: %v", err)
	}

	c, w := newTestContext(http.MethodPost, "/api/v1/users/me/avatar")
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/users/me/avatar", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	c.Request = req
	c.Set(middleware.ContextUserIDKey, uint(18))

	h.UploadAvatar(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}
