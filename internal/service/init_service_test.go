package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"easydrop/internal/consts"
	"easydrop/internal/dto"
	"easydrop/internal/model"
	"easydrop/internal/pkg/initsecret"
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

type mockInitSecretGuard struct {
	ensureFn   func(ctx context.Context) (string, error)
	validateFn func(ctx context.Context, secret string) error
}

func (m *mockInitSecretGuard) EnsureSecret(ctx context.Context) (string, error) {
	if m.ensureFn == nil {
		return "init-secret", nil
	}
	return m.ensureFn(ctx)
}

func (m *mockInitSecretGuard) Validate(ctx context.Context, secret string) error {
	if m.validateFn == nil {
		return nil
	}
	return m.validateFn(ctx, secret)
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
	svc := NewInitService(&mockInitUserRepo{}, &mockInitRepo{}, settingSvc, &mockInitCache{}, &mockInitSecretGuard{})

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
	svc := NewInitService(&mockInitUserRepo{}, initRepo, settingSvc, cache, &mockInitSecretGuard{
		validateFn: func(_ context.Context, secret string) error {
			if secret != "secret-123" {
				t.Fatalf("unexpected secret: %q", secret)
			}
			return nil
		},
	})

	allowRegister := false
	err := svc.Initialize(context.Background(), dto.InitInput{
		Secret:           "secret-123",
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
	if len(initRepo.lastInput.Settings) != 5 {
		t.Fatalf("expected 5 init settings, got %d", len(initRepo.lastInput.Settings))
	}
	if value := findInitSettingValue(initRepo.lastInput.Settings, consts.SiteAllowRegisterSettingKey); value != "false" {
		t.Fatalf("expected allow_register=false, got %q", value)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(initRepo.lastInput.AdminUser.Password), []byte("Pass1234")); err != nil {
		t.Fatalf("expected password to be hashed, compare failed: %v", err)
	}
	if got := cache.values[settingCacheKey(consts.SystemInitializedSettingKey)]; got != "true" {
		t.Fatalf("expected init cache true, got %q", got)
	}
	if got := cache.values[settingCacheKey(consts.SiteAllowRegisterSettingKey)]; got != "false" {
		t.Fatalf("expected allow_register cache false, got %q", got)
	}
}

func TestInitServiceInitializeAlreadyInitialized(t *testing.T) {
	t.Parallel()

	initRepo := &mockInitRepo{}
	settingSvc := &mockInitSettingService{values: map[string]string{consts.SystemInitializedSettingKey: "true"}}
	svc := NewInitService(&mockInitUserRepo{}, initRepo, settingSvc, &mockInitCache{}, &mockInitSecretGuard{})

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
	svc := NewInitService(userRepo, &mockInitRepo{}, &mockInitSettingService{values: map[string]string{}}, &mockInitCache{}, &mockInitSecretGuard{})

	err := svc.Initialize(context.Background(), dto.InitInput{
		Secret:   "secret-123",
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
	svc := NewInitService(&mockInitUserRepo{}, initRepo, &mockInitSettingService{values: map[string]string{}}, &mockInitCache{}, &mockInitSecretGuard{})

	err := svc.Initialize(context.Background(), dto.InitInput{
		Secret:   "secret-123",
		Username: "admin",
		Email:    "admin@example.com",
		Password: "Pass1234",
	})
	if !errors.Is(err, ErrAlreadyInitialized) {
		t.Fatalf("expected ErrAlreadyInitialized, got %v", err)
	}
}

func TestInitServiceInitializeRequiresSecret(t *testing.T) {
	t.Parallel()

	svc := NewInitService(&mockInitUserRepo{}, &mockInitRepo{}, &mockInitSettingService{values: map[string]string{}}, &mockInitCache{}, &mockInitSecretGuard{
		validateFn: func(_ context.Context, secret string) error {
			if secret != "" {
				t.Fatalf("expected empty secret, got %q", secret)
			}
			return initsecret.ErrRequired
		},
	})

	err := svc.Initialize(context.Background(), dto.InitInput{
		Username: "admin",
		Email:    "admin@example.com",
		Password: "Pass1234",
	})
	if !errors.Is(err, initsecret.ErrRequired) {
		t.Fatalf("expected ErrRequired, got %v", err)
	}
}

func TestInitServiceInitializeRejectsInvalidSecret(t *testing.T) {
	t.Parallel()

	initRepo := &mockInitRepo{}
	svc := NewInitService(&mockInitUserRepo{}, initRepo, &mockInitSettingService{values: map[string]string{}}, &mockInitCache{}, &mockInitSecretGuard{
		validateFn: func(_ context.Context, secret string) error {
			if secret != "wrong-secret" {
				t.Fatalf("unexpected secret: %q", secret)
			}
			return initsecret.ErrInvalid
		},
	})

	err := svc.Initialize(context.Background(), dto.InitInput{
		Secret:   "wrong-secret",
		Username: "admin",
		Email:    "admin@example.com",
		Password: "Pass1234",
	})
	if !errors.Is(err, initsecret.ErrInvalid) {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
	if initRepo.called {
		t.Fatal("expected init repo not to be called")
	}
}

func findInitSettingValue(settings []repo.SettingValueInput, key string) string {
	for _, setting := range settings {
		if setting.Key == key {
			return setting.Value
		}
	}
	return ""
}
