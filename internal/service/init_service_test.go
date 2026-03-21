package service

import (
	"context"
	"errors"
	"testing"

	"easydrop/internal/dto"
)

type mockInitSettingService struct {
	values  map[string]string
	updates []dto.SettingUpdateInput
}

func (m *mockInitSettingService) GetValue(_ context.Context, key string) (string, bool, error) {
	v, ok := m.values[key]
	return v, ok, nil
}

func (m *mockInitSettingService) ListItems(context.Context, dto.SettingListInput) (*dto.SettingListResult, error) {
	return nil, nil
}

func (m *mockInitSettingService) UpdateItem(_ context.Context, input dto.SettingUpdateInput) error {
	m.updates = append(m.updates, input)
	if input.Value != nil {
		if m.values == nil {
			m.values = make(map[string]string)
		}
		m.values[input.Key] = *input.Value
	}
	return nil
}

func (m *mockInitSettingService) GetPublicItems(context.Context) (*dto.SettingPublicResult, error) {
	return nil, nil
}

type mockInitUserService struct {
	created []dto.UserCreateInput
	err     error
}

func (m *mockInitUserService) Create(_ context.Context, input dto.UserCreateInput) (*dto.UserDTO, error) {
	if m.err != nil {
		return nil, m.err
	}
	m.created = append(m.created, input)
	return &dto.UserDTO{ID: 1, Username: input.Username}, nil
}

func (m *mockInitUserService) Get(context.Context, uint) (*dto.UserDTO, error) {
	return nil, nil
}

func (m *mockInitUserService) UpdateProfile(context.Context, dto.UserProfileUpdateInput) (*dto.UserDTO, error) {
	return nil, nil
}

func (m *mockInitUserService) ChangePassword(context.Context, dto.UserChangePasswordInput) error {
	return nil
}

func (m *mockInitUserService) RequestEmailChange(context.Context, dto.UserChangeEmailInput) error {
	return nil
}

func (m *mockInitUserService) ConfirmEmailChange(context.Context, dto.UserChangeEmailConfirmInput) (*dto.UserDTO, error) {
	return nil, nil
}

func (m *mockInitUserService) Update(context.Context, dto.UserUpdateInput) (*dto.UserDTO, error) {
	return nil, nil
}

func (m *mockInitUserService) UploadAvatar(context.Context, dto.UserAvatarUploadInput) (*dto.UserDTO, error) {
	return nil, nil
}

func (m *mockInitUserService) DeleteAvatar(context.Context, uint) error {
	return nil
}

func (m *mockInitUserService) Delete(context.Context, uint) error {
	return nil
}

func (m *mockInitUserService) List(context.Context, dto.UserListInput) (*dto.UserListResult, error) {
	return nil, nil
}

func TestInitServiceGetStatusNotInitialized(t *testing.T) {
	t.Parallel()

	settingSvc := &mockInitSettingService{values: map[string]string{}}
	svc := NewInitService(&mockInitUserService{}, settingSvc)

	res, err := svc.GetStatus(context.Background())
	if err != nil {
		t.Fatalf("GetStatus error: %v", err)
	}
	if res == nil || res.Initialized {
		t.Fatalf("expected uninitialized, got %+v", res)
	}
}

func TestInitServiceInitializeSuccess(t *testing.T) {
	t.Parallel()

	userSvc := &mockInitUserService{}
	settingSvc := &mockInitSettingService{values: map[string]string{}}
	svc := NewInitService(userSvc, settingSvc)

	allowRegister := false
	err := svc.Initialize(context.Background(), dto.InitInput{
		Username:         "admin",
		Nickname:         "管理员",
		Email:            "admin@example.com",
		Password:         "Pass1234",
		SiteName:         "EasyDrop",
		SiteURL:          "http://localhost:8080",
		SiteAnnouncement: "hello",
		AllowRegister:    &allowRegister,
	})
	if err != nil {
		t.Fatalf("Initialize error: %v", err)
	}

	if len(userSvc.created) != 1 {
		t.Fatalf("expected one created user, got %d", len(userSvc.created))
	}
	if !userSvc.created[0].Admin {
		t.Fatal("expected created user to be admin")
	}
	if settingSvc.values[initSettingKey] != "true" {
		t.Fatalf("expected init flag true, got %q", settingSvc.values[initSettingKey])
	}
	if settingSvc.values["site.allow_register"] != "false" {
		t.Fatalf("expected allow_register false, got %q", settingSvc.values["site.allow_register"])
	}
}

func TestInitServiceInitializeAlreadyInitialized(t *testing.T) {
	t.Parallel()

	userSvc := &mockInitUserService{}
	settingSvc := &mockInitSettingService{values: map[string]string{initSettingKey: "true"}}
	svc := NewInitService(userSvc, settingSvc)

	err := svc.Initialize(context.Background(), dto.InitInput{})
	if !errors.Is(err, ErrAlreadyInitialized) {
		t.Fatalf("expected ErrAlreadyInitialized, got %v", err)
	}
	if len(userSvc.created) != 0 {
		t.Fatalf("expected no user created, got %d", len(userSvc.created))
	}
}

func TestInitServiceInitializeCreateUserError(t *testing.T) {
	t.Parallel()

	userSvc := &mockInitUserService{err: ErrUsernameExists}
	settingSvc := &mockInitSettingService{values: map[string]string{}}
	svc := NewInitService(userSvc, settingSvc)

	err := svc.Initialize(context.Background(), dto.InitInput{Username: "admin"})
	if !errors.Is(err, ErrUsernameExists) {
		t.Fatalf("expected ErrUsernameExists, got %v", err)
	}
}
