package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"easydrop/internal/dto"
	"easydrop/internal/model"
	"easydrop/internal/repo"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type mockInitSettingService struct {
	values map[string]string
}

func (m *mockInitSettingService) GetValue(_ context.Context, key string) (string, bool, error) {
	v, ok := m.values[key]
	return v, ok, nil
}

func (m *mockInitSettingService) ListItems(context.Context, dto.SettingListInput) (*dto.SettingListResult, error) {
	return nil, nil
}

func (m *mockInitSettingService) UpdateItem(context.Context, dto.SettingUpdateInput) error {
	return nil
}

func (m *mockInitSettingService) GetPublicItems(context.Context) (*dto.SettingPublicResult, error) {
	return nil, nil
}

type mockInitRepo struct {
	lastInput repo.SystemInitInput
	err       error
	called    bool
}

func (m *mockInitRepo) Initialize(_ context.Context, input repo.SystemInitInput) error {
	m.called = true
	m.lastInput = input
	return m.err
}

type mockInitUserRepo struct {
	usersByUsername map[string]*model.User
	usersByEmail    map[string]*model.User
}

func (m *mockInitUserRepo) Create(context.Context, *model.User) error {
	return nil
}

func (m *mockInitUserRepo) GetByID(context.Context, uint) (*model.User, error) {
	return nil, gorm.ErrRecordNotFound
}

func (m *mockInitUserRepo) GetByUsername(_ context.Context, username string) (*model.User, error) {
	if user, ok := m.usersByUsername[username]; ok {
		clone := *user
		return &clone, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockInitUserRepo) GetByEmail(_ context.Context, email string) (*model.User, error) {
	if user, ok := m.usersByEmail[email]; ok {
		clone := *user
		return &clone, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockInitUserRepo) GetByUsernameOrEmail(context.Context, string) (*model.User, error) {
	return nil, gorm.ErrRecordNotFound
}

func (m *mockInitUserRepo) UpdateAvatarWithStorageUsedTx(context.Context, uint, *string, int64, int64) (*model.User, error) {
	return nil, nil
}

func (m *mockInitUserRepo) Update(context.Context, *model.User) error {
	return nil
}

func (m *mockInitUserRepo) Delete(context.Context, uint) error {
	return nil
}

func (m *mockInitUserRepo) List(context.Context, repo.UserFilter, repo.ListOptions) ([]model.User, int64, error) {
	return nil, 0, nil
}

type mockInitCache struct {
	values map[string]string
}

func (m *mockInitCache) Backend() string {
	return "memory"
}

func (m *mockInitCache) Get(_ context.Context, key string) (string, bool, error) {
	v, ok := m.values[key]
	return v, ok, nil
}

func (m *mockInitCache) Set(_ context.Context, key, value string, _ time.Duration) error {
	if m.values == nil {
		m.values = make(map[string]string)
	}
	m.values[key] = value
	return nil
}

func (m *mockInitCache) Delete(_ context.Context, key string) error {
	delete(m.values, key)
	return nil
}

func (m *mockInitCache) Clear(context.Context) error {
	m.values = make(map[string]string)
	return nil
}

func TestInitServiceGetStatusNotInitialized(t *testing.T) {
	t.Parallel()

	settingSvc := &mockInitSettingService{values: map[string]string{}}
	svc := NewInitService(&mockInitUserRepo{}, &mockInitRepo{}, settingSvc, &mockInitCache{})

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

	initRepo := &mockInitRepo{}
	cache := &mockInitCache{}
	settingSvc := &mockInitSettingService{values: map[string]string{}}
	svc := NewInitService(&mockInitUserRepo{}, initRepo, settingSvc, cache)

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

	if !initRepo.called {
		t.Fatal("expected init repo to be called")
	}
	if !initRepo.lastInput.AdminUser.Admin {
		t.Fatal("expected created admin user")
	}
	if !initRepo.lastInput.AdminUser.EmailVerified {
		t.Fatal("expected initialized admin email to be verified")
	}
	if initRepo.lastInput.AdminUser.Status != 1 {
		t.Fatalf("expected status=1, got %d", initRepo.lastInput.AdminUser.Status)
	}
	if initRepo.lastInput.AllowRegister != "false" {
		t.Fatalf("expected allow_register=false, got %q", initRepo.lastInput.AllowRegister)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(initRepo.lastInput.AdminUser.Password), []byte("Pass1234")); err != nil {
		t.Fatalf("expected password to be hashed, compare failed: %v", err)
	}
	if got := cache.values[settingCacheKey(initSettingKey)]; got != "true" {
		t.Fatalf("expected init cache true, got %q", got)
	}
	if got := cache.values[settingCacheKey("site.allow_register")]; got != "false" {
		t.Fatalf("expected allow_register cache false, got %q", got)
	}
}

func TestInitServiceInitializeAlreadyInitialized(t *testing.T) {
	t.Parallel()

	initRepo := &mockInitRepo{}
	settingSvc := &mockInitSettingService{values: map[string]string{initSettingKey: "true"}}
	svc := NewInitService(&mockInitUserRepo{}, initRepo, settingSvc, &mockInitCache{})

	err := svc.Initialize(context.Background(), dto.InitInput{})
	if !errors.Is(err, ErrAlreadyInitialized) {
		t.Fatalf("expected ErrAlreadyInitialized, got %v", err)
	}
	if initRepo.called {
		t.Fatal("expected init repo not to be called")
	}
}

func TestInitServiceInitializeUsernameExists(t *testing.T) {
	t.Parallel()

	userRepo := &mockInitUserRepo{
		usersByUsername: map[string]*model.User{
			"admin": {ID: 1, Username: "admin"},
		},
	}
	svc := NewInitService(userRepo, &mockInitRepo{}, &mockInitSettingService{values: map[string]string{}}, &mockInitCache{})

	err := svc.Initialize(context.Background(), dto.InitInput{
		Username: "admin",
		Email:    "admin@example.com",
		Password: "Pass1234",
	})
	if !errors.Is(err, ErrUsernameExists) {
		t.Fatalf("expected ErrUsernameExists, got %v", err)
	}
}

func TestInitServiceInitializeInitRepoAlreadyInitialized(t *testing.T) {
	t.Parallel()

	initRepo := &mockInitRepo{err: repo.ErrInitAlreadyInitialized}
	svc := NewInitService(&mockInitUserRepo{}, initRepo, &mockInitSettingService{values: map[string]string{}}, &mockInitCache{})

	err := svc.Initialize(context.Background(), dto.InitInput{
		Username: "admin",
		Email:    "admin@example.com",
		Password: "Pass1234",
	})
	if !errors.Is(err, ErrAlreadyInitialized) {
		t.Fatalf("expected ErrAlreadyInitialized, got %v", err)
	}
}
