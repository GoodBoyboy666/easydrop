package handler

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"testing"

	"easydrop/internal/dto"
	"easydrop/internal/service"

	"github.com/gin-gonic/gin"
)

type mockUserAdminService struct {
	createFn       func(ctx context.Context, input dto.UserCreateInput) (*dto.UserDTO, error)
	updateFn       func(ctx context.Context, input dto.UserUpdateInput) (*dto.UserDTO, error)
	deleteFn       func(ctx context.Context, id uint) error
	listFn         func(ctx context.Context, input dto.UserListInput) (*dto.UserListResult, error)
	uploadAvatarFn func(ctx context.Context, input dto.UserAvatarUploadInput) (*dto.UserDTO, error)
	deleteAvatarFn func(ctx context.Context, userID uint) error
}

func (m *mockUserAdminService) Create(ctx context.Context, input dto.UserCreateInput) (*dto.UserDTO, error) {
	if m.createFn == nil {
		return nil, nil
	}
	return m.createFn(ctx, input)
}

func (m *mockUserAdminService) Get(context.Context, uint) (*dto.UserDTO, error) {
	return nil, nil
}

func (m *mockUserAdminService) UpdateProfile(context.Context, dto.UserProfileUpdateInput) (*dto.UserDTO, error) {
	return nil, nil
}

func (m *mockUserAdminService) ChangePassword(context.Context, dto.UserChangePasswordInput) error {
	return nil
}

func (m *mockUserAdminService) RequestEmailChange(context.Context, dto.UserChangeEmailInput) error {
	return nil
}

func (m *mockUserAdminService) ConfirmEmailChange(context.Context, dto.UserChangeEmailConfirmInput) (*dto.UserDTO, error) {
	return nil, nil
}

func (m *mockUserAdminService) Update(ctx context.Context, input dto.UserUpdateInput) (*dto.UserDTO, error) {
	if m.updateFn == nil {
		return nil, nil
	}
	return m.updateFn(ctx, input)
}

func (m *mockUserAdminService) UploadAvatar(ctx context.Context, input dto.UserAvatarUploadInput) (*dto.UserDTO, error) {
	if m.uploadAvatarFn == nil {
		return nil, nil
	}
	return m.uploadAvatarFn(ctx, input)
}

func (m *mockUserAdminService) DeleteAvatar(ctx context.Context, userID uint) error {
	if m.deleteAvatarFn == nil {
		return nil
	}
	return m.deleteAvatarFn(ctx, userID)
}

func (m *mockUserAdminService) Delete(ctx context.Context, id uint) error {
	if m.deleteFn == nil {
		return nil
	}
	return m.deleteFn(ctx, id)
}

func (m *mockUserAdminService) List(ctx context.Context, input dto.UserListInput) (*dto.UserListResult, error) {
	if m.listFn == nil {
		return nil, nil
	}
	return m.listFn(ctx, input)
}

func TestUserAdminHandlerListSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	called := false
	h := NewUserAdminHandler(&mockUserAdminService{
		listFn: func(_ context.Context, input dto.UserListInput) (*dto.UserListResult, error) {
			called = true
			if input.Username != "neo" || input.Email != "@example.com" {
				t.Fatalf("unexpected list filter: %+v", input)
			}
			if input.Status == nil || *input.Status != 1 {
				t.Fatalf("unexpected status: %+v", input.Status)
			}
			if input.Limit != 10 || input.Offset != 20 || input.Order != "id desc" {
				t.Fatalf("unexpected list options: %+v", input)
			}
			return &dto.UserListResult{Items: []dto.UserDTO{}, Total: 0}, nil
		},
	})

	c, w := newTestContext(http.MethodGet, "/api/v1/admin/users?username=neo&email=@example.com&status=1&limit=10&offset=20&order=id%20desc")
	h.List(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !called {
		t.Fatal("expected List to be called")
	}
}

func TestUserAdminHandlerCreateSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewUserAdminHandler(&mockUserAdminService{
		createFn: func(_ context.Context, input dto.UserCreateInput) (*dto.UserDTO, error) {
			if input.Username != "admin_created" {
				t.Fatalf("unexpected username: %s", input.Username)
			}
			return &dto.UserDTO{ID: 5, Username: input.Username}, nil
		},
	})

	c, w := newTestContextWithBody(http.MethodPost, "/api/v1/admin/users", `{"username":"admin_created","email":"new@example.com","password":"Pass1234"}`)
	h.Create(c)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", w.Code)
	}
}

func TestUserAdminHandlerUpdateBindsPathID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewUserAdminHandler(&mockUserAdminService{
		updateFn: func(_ context.Context, input dto.UserUpdateInput) (*dto.UserDTO, error) {
			if input.ID != 8 {
				t.Fatalf("expected id from path 8, got %d", input.ID)
			}
			if input.Nickname == nil || *input.Nickname != "Trinity" {
				t.Fatalf("unexpected nickname: %#v", input.Nickname)
			}
			return &dto.UserDTO{ID: 8, Nickname: "Trinity"}, nil
		},
	})

	c, w := newTestContextWithBody(http.MethodPatch, "/api/v1/admin/users/8", `{"id":999,"nickname":"Trinity"}`)
	c.Params = gin.Params{{Key: "id", Value: "8"}}
	h.Update(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestUserAdminHandlerUpdateBindsUseDefaultStorageQuota(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewUserAdminHandler(&mockUserAdminService{
		updateFn: func(_ context.Context, input dto.UserUpdateInput) (*dto.UserDTO, error) {
			if input.ID != 8 {
				t.Fatalf("expected id from path 8, got %d", input.ID)
			}
			if input.UseDefaultStorageQuota == nil || !*input.UseDefaultStorageQuota {
				t.Fatalf("expected use_default_storage_quota=true, got %#v", input.UseDefaultStorageQuota)
			}
			return &dto.UserDTO{ID: 8, Nickname: "Trinity"}, nil
		},
	})

	c, w := newTestContextWithBody(http.MethodPatch, "/api/v1/admin/users/8", `{"use_default_storage_quota":true}`)
	c.Params = gin.Params{{Key: "id", Value: "8"}}
	h.Update(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestUserAdminHandlerDeleteUserNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewUserAdminHandler(&mockUserAdminService{
		deleteFn: func(_ context.Context, _ uint) error {
			return service.ErrUserNotFound
		},
	})

	c, w := newTestContext(http.MethodDelete, "/api/v1/admin/users/2")
	c.Params = gin.Params{{Key: "id", Value: "2"}}
	h.Delete(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestUserAdminHandlerUploadAvatarSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewUserAdminHandler(&mockUserAdminService{
		uploadAvatarFn: func(_ context.Context, input dto.UserAvatarUploadInput) (*dto.UserDTO, error) {
			if input.UserID != 33 {
				t.Fatalf("expected user id 33, got %d", input.UserID)
			}
			if input.OriginalFilename != "avatar.png" {
				t.Fatalf("unexpected filename: %s", input.OriginalFilename)
			}
			if len(input.Content) == 0 {
				t.Fatal("expected avatar content")
			}
			return &dto.UserDTO{ID: 33}, nil
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

	c, w := newTestContext(http.MethodPost, "/api/v1/admin/users/33/avatar")
	c.Params = gin.Params{{Key: "id", Value: "33"}}
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/users/33/avatar", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	c.Request = req

	h.UploadAvatar(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestUserAdminHandlerUploadAvatarMissingFile(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewUserAdminHandler(&mockUserAdminService{})
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer failed: %v", err)
	}

	c, w := newTestContext(http.MethodPost, "/api/v1/admin/users/10/avatar")
	c.Params = gin.Params{{Key: "id", Value: "10"}}
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/admin/users/10/avatar", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	c.Request = req

	h.UploadAvatar(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestUserAdminHandlerDeleteAvatarSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewUserAdminHandler(&mockUserAdminService{
		deleteAvatarFn: func(_ context.Context, userID uint) error {
			if userID != 11 {
				t.Fatalf("expected user id 11, got %d", userID)
			}
			return nil
		},
	})

	c, w := newTestContext(http.MethodDelete, "/api/v1/admin/users/11/avatar")
	c.Params = gin.Params{{Key: "id", Value: "11"}}
	h.DeleteAvatar(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestUserAdminHandlerInvalidPathID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewUserAdminHandler(&mockUserAdminService{})
	c, w := newTestContext(http.MethodPatch, "/api/v1/admin/users/0")
	c.Params = gin.Params{{Key: "id", Value: "0"}}
	c.Request, _ = http.NewRequest(http.MethodPatch, "/api/v1/admin/users/0", bytes.NewBufferString(`{"nickname":"n"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Update(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}
