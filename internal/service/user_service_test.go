package service

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"easydrop/internal/dto"
	"easydrop/internal/model"
	"easydrop/internal/pkg/storage"
	"easydrop/internal/pkg/token"
	"easydrop/internal/repo"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type mockUserRepo struct {
	users map[uint]*model.User
}

type mockTokenManager struct {
	recordByToken map[string]*token.Record
	nextToken     string
	issueErr      error
	consumeErr    error
}

func (m *mockTokenManager) Backend() string {
	return token.BackendMemory
}

func (m *mockTokenManager) Issue(_ context.Context, userID uint, kind string, ttl time.Duration, payload string) (string, error) {
	if m.issueErr != nil {
		return "", m.issueErr
	}
	if m.recordByToken == nil {
		m.recordByToken = make(map[string]*token.Record)
	}
	value := m.nextToken
	if strings.TrimSpace(value) == "" {
		value = "mock-email-change-token"
	}
	now := time.Now().UTC()
	m.recordByToken[value] = &token.Record{
		UserID:    userID,
		Kind:      kind,
		Token:     value,
		CreatedAt: now,
		ExpiresAt: now.Add(ttl),
		Payload:   payload,
	}
	return value, nil
}

func (m *mockTokenManager) Consume(_ context.Context, userID uint, kind, tokenValue string) (*token.Record, error) {
	if m.consumeErr != nil {
		return nil, m.consumeErr
	}
	record, ok := m.recordByToken[tokenValue]
	if !ok {
		return nil, token.ErrTokenNotFound
	}
	if record.UserID != userID || record.Kind != kind {
		return nil, token.ErrTokenMismatch
	}
	delete(m.recordByToken, tokenValue)
	return record, nil
}

type mockEmailSender struct {
	to      []string
	subject string
	body    string
	sendErr error
}

func (m *mockEmailSender) SendHTML(_ context.Context, to []string, subject, htmlBody string) error {
	if m.sendErr != nil {
		return m.sendErr
	}
	m.to = append([]string{}, to...)
	m.subject = subject
	m.body = htmlBody
	return nil
}

type mockEmailService struct {
	sender *mockEmailSender
	err    error
}

func (m *mockEmailService) SendPasswordResetEmail(ctx context.Context, to, tokenValue string, ttl time.Duration) error {
	if m.err != nil {
		return m.err
	}
	if m.sender == nil {
		return nil
	}
	body := "reset:" + tokenValue
	return m.sender.SendHTML(ctx, []string{to}, "重置密码确认", body)
}

func (m *mockEmailService) SendVerifyEmail(ctx context.Context, to, tokenValue string, ttl time.Duration) error {
	if m.err != nil {
		return m.err
	}
	if m.sender == nil {
		return nil
	}
	body := "verify:" + tokenValue
	return m.sender.SendHTML(ctx, []string{to}, "邮箱验证", body)
}

func (m *mockEmailService) SendChangeEmailEmail(ctx context.Context, to, newEmail, tokenValue string, ttl time.Duration) error {
	if m.err != nil {
		return m.err
	}
	if m.sender == nil {
		return nil
	}
	body := "change:" + newEmail + ":" + tokenValue
	return m.sender.SendHTML(ctx, []string{to}, "邮箱修改确认", body)
}

func (m *mockUserRepo) Create(_ context.Context, user *model.User) error {
	if m.users == nil {
		m.users = make(map[uint]*model.User)
	}
	m.users[user.ID] = cloneUser(user)
	return nil
}

func (m *mockUserRepo) GetByID(_ context.Context, id uint) (*model.User, error) {
	user, ok := m.users[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return cloneUser(user), nil
}

func (m *mockUserRepo) GetByUsername(_ context.Context, username string) (*model.User, error) {
	for _, user := range m.users {
		if user.Username == username {
			return cloneUser(user), nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockUserRepo) GetByEmail(_ context.Context, email string) (*model.User, error) {
	for _, user := range m.users {
		if user.Email == email {
			return cloneUser(user), nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockUserRepo) GetByUsernameOrEmail(_ context.Context, value string) (*model.User, error) {
	for _, user := range m.users {
		if user.Username == value || user.Email == value {
			return cloneUser(user), nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockUserRepo) UpdateAvatarWithStorageUsedTx(_ context.Context, userID uint, avatar *string, sizeDelta int64, defaultQuota int64) (*model.User, error) {
	user, ok := m.users[userID]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}

	newUsed := user.StorageUsed + sizeDelta
	if newUsed < 0 {
		newUsed = 0
	}

	effectiveQuota := defaultQuota
	if user.StorageQuota != nil && *user.StorageQuota > 0 {
		effectiveQuota = *user.StorageQuota
	}
	if sizeDelta > 0 && effectiveQuota > 0 && newUsed > effectiveQuota {
		return nil, repo.ErrUserAvatarQuotaExceeded
	}

	user.Avatar = cloneStringPtr(avatar)
	user.StorageUsed = newUsed
	return cloneUser(user), nil
}

func (m *mockUserRepo) Update(_ context.Context, user *model.User) error {
	if _, ok := m.users[user.ID]; !ok {
		return gorm.ErrRecordNotFound
	}
	m.users[user.ID] = cloneUser(user)
	return nil
}

func (m *mockUserRepo) Delete(_ context.Context, id uint) error {
	delete(m.users, id)
	return nil
}

func (m *mockUserRepo) List(_ context.Context, _ repo.UserFilter, _ repo.ListOptions) ([]model.User, int64, error) {
	items := make([]model.User, 0, len(m.users))
	for _, user := range m.users {
		items = append(items, *cloneUser(user))
	}
	return items, int64(len(items)), nil
}

func TestUserServiceUploadAvatarReplacesManagedAvatar(t *testing.T) {
	storageManager, basePath := newTestStorageManager(t)
	oldKey := filepath.ToSlash(filepath.Join(storage.CategoryAvatar, "1", "existing.png"))
	oldContent := []byte("old-avatar")
	if err := storageManager.Upload(context.Background(), oldKey, oldContent, "image/png"); err != nil {
		t.Fatalf("seed old avatar failed: %v", err)
	}

	oldKeyCopy := oldKey
	repo := &mockUserRepo{
		users: map[uint]*model.User{
			1: {
				ID:          1,
				Username:    "alice",
				Email:       "alice@example.com",
				Status:      1,
				Avatar:      &oldKeyCopy,
				StorageUsed: int64(len(oldContent)),
			},
		},
	}
	service := NewUserService(repo, storageManager, nil, nil, nil)

	result, err := service.UploadAvatar(context.Background(), dto.UserAvatarUploadInput{
		UserID:           1,
		OriginalFilename: "profile.JPG",
		ContentType:      "image/jpeg",
		Content:          []byte("new-avatar-content"),
	})
	if err != nil {
		t.Fatalf("UploadAvatar returned error: %v", err)
	}

	if result.Avatar == nil || *result.Avatar == "" {
		t.Fatal("expected avatar URL to be returned")
	}
	if repo.users[1].Avatar == nil {
		t.Fatal("expected stored avatar key to be updated")
	}
	if *repo.users[1].Avatar == oldKey {
		t.Fatal("expected avatar key to change after upload")
	}
	if filepath.Ext(*repo.users[1].Avatar) != ".jpg" {
		t.Fatalf("expected normalized extension, got %s", *repo.users[1].Avatar)
	}
	if repo.users[1].StorageUsed != int64(len("new-avatar-content")) {
		t.Fatalf("expected storage used to equal new avatar size, got %d", repo.users[1].StorageUsed)
	}
	if _, err := os.Stat(filepath.Join(basePath, filepath.FromSlash(oldKey))); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected old avatar to be deleted, stat err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(basePath, filepath.FromSlash(*repo.users[1].Avatar))); err != nil {
		t.Fatalf("expected new avatar object to exist: %v", err)
	}
}

func TestUserServiceDeleteAvatarClearsManagedAvatar(t *testing.T) {
	storageManager, basePath := newTestStorageManager(t)
	avatarKey := filepath.ToSlash(filepath.Join(storage.CategoryAvatar, "2", "avatar.png"))
	content := []byte("avatar-to-delete")
	if err := storageManager.Upload(context.Background(), avatarKey, content, "image/png"); err != nil {
		t.Fatalf("seed avatar failed: %v", err)
	}

	avatarKeyCopy := avatarKey
	repo := &mockUserRepo{
		users: map[uint]*model.User{
			2: {
				ID:          2,
				Username:    "bob",
				Email:       "bob@example.com",
				Status:      1,
				Avatar:      &avatarKeyCopy,
				StorageUsed: int64(len(content)),
			},
		},
	}
	service := NewUserService(repo, storageManager, nil, nil, nil)

	err := service.DeleteAvatar(context.Background(), 2)
	if err != nil {
		t.Fatalf("DeleteAvatar returned error: %v", err)
	}

	if repo.users[2].Avatar != nil {
		t.Fatal("expected stored avatar to be cleared")
	}
	if repo.users[2].StorageUsed != 0 {
		t.Fatalf("expected storage used to be 0, got %d", repo.users[2].StorageUsed)
	}
	if _, err := os.Stat(filepath.Join(basePath, filepath.FromSlash(avatarKey))); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected avatar object to be deleted, stat err=%v", err)
	}
}

func TestUserServiceUploadAvatarQuotaExceededRollsBackObject(t *testing.T) {
	storageManager, basePath := newTestStorageManager(t)
	quota := int64(4)
	repo := &mockUserRepo{
		users: map[uint]*model.User{
			3: {
				ID:           3,
				Username:     "carol",
				Email:        "carol@example.com",
				Status:       1,
				StorageQuota: &quota,
			},
		},
	}
	service := NewUserService(repo, storageManager, nil, nil, nil)

	_, err := service.UploadAvatar(context.Background(), dto.UserAvatarUploadInput{
		UserID:           3,
		OriginalFilename: "avatar.png",
		ContentType:      "image/png",
		Content:          []byte("too-large"),
	})
	if !errors.Is(err, ErrStorageQuotaExceeded) {
		t.Fatalf("expected ErrStorageQuotaExceeded, got %v", err)
	}

	if repo.users[3].Avatar != nil {
		t.Fatal("expected avatar to remain unchanged after quota failure")
	}
	if repo.users[3].StorageUsed != 0 {
		t.Fatalf("expected storage used to remain 0, got %d", repo.users[3].StorageUsed)
	}
	if count := countFiles(t, basePath); count != 0 {
		t.Fatalf("expected uploaded object rollback, found %d files", count)
	}
}

func TestUserServiceDeleteAvatarWithExternalURL(t *testing.T) {
	external := "https://example.com/avatar.png"
	repo := &mockUserRepo{
		users: map[uint]*model.User{
			4: {
				ID:          4,
				Username:    "dave",
				Email:       "dave@example.com",
				Status:      1,
				Avatar:      &external,
				StorageUsed: 7,
			},
		},
	}
	service := NewUserService(repo, nil, nil, nil, nil)

	err := service.DeleteAvatar(context.Background(), 4)
	if err != nil {
		t.Fatalf("DeleteAvatar returned error: %v", err)
	}

	if repo.users[4].Avatar != nil {
		t.Fatal("expected stored avatar to be nil")
	}
	if repo.users[4].StorageUsed != 7 {
		t.Fatalf("expected storage used to remain unchanged, got %d", repo.users[4].StorageUsed)
	}
}

func newTestStorageManager(t *testing.T) (storage.Manager, string) {
	t.Helper()

	basePath := filepath.Join(t.TempDir(), "uploads")
	manager, err := storage.NewManager(&storage.Config{
		Backend: storage.BackendLocal,
		Local: storage.LocalConfig{
			BasePath:  basePath,
			URLPrefix: "",
		},
	})
	if err != nil {
		t.Fatalf("new test storage manager failed: %v", err)
	}

	return manager, basePath
}

func countFiles(t *testing.T, root string) int {
	t.Helper()

	count := 0
	err := filepath.Walk(root, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			count++
		}
		return nil
	})
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("walk files failed: %v", err)
	}
	return count
}

func cloneUser(user *model.User) *model.User {
	if user == nil {
		return nil
	}

	clone := *user
	clone.Avatar = cloneStringPtr(user.Avatar)
	if user.StorageQuota != nil {
		quota := *user.StorageQuota
		clone.StorageQuota = &quota
	}
	return &clone
}

func cloneStringPtr(value *string) *string {
	if value == nil {
		return nil
	}
	clone := *value
	return &clone
}

func TestUserServiceUpdateProfile(t *testing.T) {
	repo := &mockUserRepo{
		users: map[uint]*model.User{
			7: {
				ID:       7,
				Username: "neo",
				Nickname: "old",
				Email:    "neo@example.com",
				Status:   1,
			},
		},
	}
	svc := NewUserService(repo, nil, nil, nil, nil)

	nickname := "  Neo Matrix  "
	result, err := svc.UpdateProfile(context.Background(), dto.UserProfileUpdateInput{UserID: 7, Nickname: &nickname})
	if err != nil {
		t.Fatalf("UpdateProfile returned error: %v", err)
	}
	if result.Nickname != "Neo Matrix" {
		t.Fatalf("expected trimmed nickname, got %q", result.Nickname)
	}
	if repo.users[7].Nickname != "Neo Matrix" {
		t.Fatalf("expected repo nickname updated, got %q", repo.users[7].Nickname)
	}
}

func TestUserServiceUpdateClearsStorageQuotaWhenUseDefaultStorageQuotaIsTrue(t *testing.T) {
	quota := int64(1024)
	repo := &mockUserRepo{
		users: map[uint]*model.User{
			13: {
				ID:           13,
				Username:     "switch",
				Nickname:     "switch",
				Email:        "switch@example.com",
				Status:       1,
				StorageQuota: &quota,
			},
		},
	}
	svc := NewUserService(repo, nil, nil, nil, nil)
	useDefault := true

	result, err := svc.Update(context.Background(), dto.UserUpdateInput{
		ID:                     13,
		UseDefaultStorageQuota: &useDefault,
	})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if repo.users[13].StorageQuota != nil {
		t.Fatalf("expected repo storage quota cleared, got %#v", repo.users[13].StorageQuota)
	}
	if result.StorageQuota != nil {
		t.Fatalf("expected dto storage quota cleared, got %#v", result.StorageQuota)
	}
}

func TestUserServiceChangePassword(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("OldPass123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("seed hash failed: %v", err)
	}

	repo := &mockUserRepo{
		users: map[uint]*model.User{
			8: {
				ID:       8,
				Username: "trinity",
				Nickname: "trinity",
				Email:    "tri@example.com",
				Password: string(hash),
				Status:   1,
			},
		},
	}
	svc := NewUserService(repo, nil, nil, nil, nil)

	err = svc.ChangePassword(context.Background(), dto.UserChangePasswordInput{
		UserID:      8,
		OldPassword: "OldPass123",
		NewPassword: "NewPass123",
	})
	if err != nil {
		t.Fatalf("ChangePassword returned error: %v", err)
	}
	if compareErr := bcrypt.CompareHashAndPassword([]byte(repo.users[8].Password), []byte("NewPass123")); compareErr != nil {
		t.Fatalf("expected password to be updated: %v", compareErr)
	}
}

func TestUserServiceChangePasswordWithWrongOldPassword(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("OldPass123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("seed hash failed: %v", err)
	}

	repo := &mockUserRepo{
		users: map[uint]*model.User{
			9: {
				ID:       9,
				Username: "smith",
				Email:    "smith@example.com",
				Password: string(hash),
				Status:   1,
			},
		},
	}
	svc := NewUserService(repo, nil, nil, nil, nil)

	err = svc.ChangePassword(context.Background(), dto.UserChangePasswordInput{
		UserID:      9,
		OldPassword: "Wrong1234",
		NewPassword: "NewPass123",
	})
	if !errors.Is(err, ErrInvalidPassword) {
		t.Fatalf("expected ErrInvalidPassword, got %v", err)
	}
}

func TestUserServiceRequestAndConfirmEmailChange(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("OldPass123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("seed hash failed: %v", err)
	}

	repo := &mockUserRepo{
		users: map[uint]*model.User{
			10: {
				ID:            10,
				Username:      "morpheus",
				Email:         "old@example.com",
				Password:      string(hash),
				EmailVerified: false,
				Status:        1,
			},
		},
	}
	tokens := &mockTokenManager{nextToken: "confirm-abc"}
	emailSender := &mockEmailSender{}
	emails := &mockEmailService{sender: emailSender}
	svc := &userService{
		userRepo:       repo,
		tokenManager:   tokens,
		emailService:   emails,
		storageManager: nil,
	}

	err = svc.RequestEmailChange(context.Background(), dto.UserChangeEmailInput{
		UserID:          10,
		CurrentPassword: "OldPass123",
		NewEmail:        "new@example.com",
	})
	if err != nil {
		t.Fatalf("RequestEmailChange returned error: %v", err)
	}
	if len(emailSender.to) != 1 || emailSender.to[0] != "new@example.com" {
		t.Fatalf("expected mail sent to new email, got %#v", emailSender.to)
	}

	result, err := svc.ConfirmEmailChange(context.Background(), dto.UserChangeEmailConfirmInput{
		UserID:            10,
		VerificationToken: "confirm-abc",
	})
	if err != nil {
		t.Fatalf("ConfirmEmailChange returned error: %v", err)
	}
	if repo.users[10].Email != "new@example.com" {
		t.Fatalf("expected email updated, got %s", repo.users[10].Email)
	}
	if !repo.users[10].EmailVerified {
		t.Fatal("expected EmailVerified=true after confirm")
	}
	if result.Email != "new@example.com" || !result.EmailVerified {
		t.Fatalf("unexpected dto result: %#v", result)
	}
}

func TestUserServiceRequestEmailChangeWithWrongPassword(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("OldPass123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("seed hash failed: %v", err)
	}

	repo := &mockUserRepo{
		users: map[uint]*model.User{
			12: {
				ID:       12,
				Username: "cypher",
				Email:    "old@example.com",
				Password: string(hash),
				Status:   1,
			},
		},
	}
	tokens := &mockTokenManager{nextToken: "confirm-bad"}
	emailSender := &mockEmailSender{}
	emails := &mockEmailService{sender: emailSender}
	svc := &userService{
		userRepo:       repo,
		tokenManager:   tokens,
		emailService:   emails,
		storageManager: nil,
	}

	err = svc.RequestEmailChange(context.Background(), dto.UserChangeEmailInput{
		UserID:          12,
		CurrentPassword: "WrongPass123",
		NewEmail:        "new@example.com",
	})
	if !errors.Is(err, ErrInvalidPassword) {
		t.Fatalf("expected ErrInvalidPassword, got %v", err)
	}
	if len(emailSender.to) != 0 {
		t.Fatalf("expected no email sent, got %#v", emailSender.to)
	}
}

func TestUserServiceConfirmEmailChangeOldEmailMismatch(t *testing.T) {
	repo := &mockUserRepo{
		users: map[uint]*model.User{
			11: {
				ID:       11,
				Username: "oracle",
				Email:    "changed@example.com",
				Status:   1,
			},
		},
	}
	payload, err := marshalEmailChangePayload("old@example.com", "new@example.com")
	if err != nil {
		t.Fatalf("marshal payload failed: %v", err)
	}
	tokens := &mockTokenManager{recordByToken: map[string]*token.Record{
		"confirm-old": {
			UserID:  11,
			Kind:    token.KindChangeEmail,
			Token:   "confirm-old",
			Payload: payload,
		},
	}}
	svc := &userService{userRepo: repo, tokenManager: tokens}

	_, err = svc.ConfirmEmailChange(context.Background(), dto.UserChangeEmailConfirmInput{UserID: 11, VerificationToken: "confirm-old"})
	if !errors.Is(err, ErrEmailChanged) {
		t.Fatalf("expected ErrEmailChanged, got %v", err)
	}
}
