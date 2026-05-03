package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"easydrop/internal/consts"
	"easydrop/internal/model"
	"easydrop/internal/pkg/oauth"
	"easydrop/internal/repo"

	"github.com/glebarez/sqlite"
	goog "golang.org/x/oauth2"
	"gorm.io/gorm"
)

func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	db.AutoMigrate(&model.User{}, &model.OAuthBind{})
	return db
}

type mockOAuthManager struct {
	enabledProviders []string
	authCodeURL      string
	authCodeURLErr   error
	exchangeToken    *goog.Token
	exchangeErr      error
	fetchUserInfo    *oauth.ProviderUserInfo
	fetchUserInfoErr error
}

func (m *mockOAuthManager) IsProviderEnabled(provider string) bool {
	for _, p := range m.enabledProviders {
		if p == provider {
			return true
		}
	}
	return false
}

func (m *mockOAuthManager) GetEnabledProviders() []string {
	return m.enabledProviders
}

func (m *mockOAuthManager) AuthCodeURL(provider, state string) (string, error) {
	if m.authCodeURLErr != nil {
		return "", m.authCodeURLErr
	}
	if !m.IsProviderEnabled(provider) {
		return "", oauth.ErrProviderDisabled
	}
	if m.authCodeURL != "" {
		return m.authCodeURL, nil
	}
	return "https://example.com/auth?state=" + state, nil
}

func (m *mockOAuthManager) Exchange(ctx context.Context, provider, code string) (*goog.Token, error) {
	if m.exchangeErr != nil {
		return nil, m.exchangeErr
	}
	if m.exchangeToken != nil {
		return m.exchangeToken, nil
	}
	return &goog.Token{AccessToken: "mock-access-token"}, nil
}

func (m *mockOAuthManager) FetchUserInfo(ctx context.Context, provider string, token *goog.Token) (*oauth.ProviderUserInfo, error) {
	if m.fetchUserInfoErr != nil {
		return nil, m.fetchUserInfoErr
	}
	if m.fetchUserInfo != nil {
		return m.fetchUserInfo, nil
	}
	return &oauth.ProviderUserInfo{
		ProviderUserID: "mock-provider-uid",
		Email:          "mock@example.com",
		Nickname:       "MockUser",
	}, nil
}

type mockOAuthBindRepo struct {
	binds []model.OAuthBind
	err   error
}

func (m *mockOAuthBindRepo) Create(ctx context.Context, bind *model.OAuthBind) error {
	if m.err != nil {
		return m.err
	}
	bind.ID = uint(len(m.binds) + 1)
	m.binds = append(m.binds, *bind)
	return nil
}

func (m *mockOAuthBindRepo) FindByProviderAndUID(ctx context.Context, provider, providerUserID string) (*model.OAuthBind, error) {
	if m.err != nil {
		return nil, m.err
	}
	for i := range m.binds {
		b := &m.binds[i]
		if b.Provider == provider && b.ProviderUserID == providerUserID {
			return b, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockOAuthBindRepo) FindByUserIDAndProvider(ctx context.Context, userID uint, provider string) (*model.OAuthBind, error) {
	if m.err != nil {
		return nil, m.err
	}
	for i := range m.binds {
		b := &m.binds[i]
		if b.UserID == userID && b.Provider == provider {
			return b, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *mockOAuthBindRepo) FindByUserID(ctx context.Context, userID uint) ([]model.OAuthBind, error) {
	if m.err != nil {
		return nil, m.err
	}
	var result []model.OAuthBind
	for _, b := range m.binds {
		if b.UserID == userID {
			result = append(result, b)
		}
	}
	return result, nil
}

func (m *mockOAuthBindRepo) Delete(ctx context.Context, id uint) error {
	if m.err != nil {
		return m.err
	}
	for i, b := range m.binds {
		if b.ID == id {
			m.binds = append(m.binds[:i], m.binds[i+1:]...)
			return nil
		}
	}
	return gorm.ErrRecordNotFound
}

func newMockOAuthService() (*oauthService, *mockUserRepo, *mockOAuthManager, *mockOAuthBindRepo) {
	userRepo := &mockUserRepo{users: map[uint]*model.User{}}
	oauthMgr := &mockOAuthManager{
		enabledProviders: []string{"google"},
	}
	bindRepo := &mockOAuthBindRepo{}
	jwtMgr := &mockJWTManager{token: "jwt-oauth-token"}
	settings := &mockAuthSettingService{valueByKey: map[string]string{consts.SiteAllowRegisterSettingKey: "true"}}
	svc := &oauthService{
		oauthManager:  oauthMgr,
		oauthBindRepo: bindRepo,
		userRepo:      userRepo,
		jwtManager:    jwtMgr,
		settings:      settings,
		db:            nil,
	}
	return svc, userRepo, oauthMgr, bindRepo
}

func newMockOAuthServiceWithDB(t *testing.T) (*oauthService, *mockOAuthManager) {
	t.Helper()
	db := openTestDB(t)
	oauthMgr := &mockOAuthManager{
		enabledProviders: []string{"google"},
	}
	bindRepo := repo.NewOAuthBindRepo(db)
	userRepo := repo.NewUserRepo(db)
	jwtMgr := &mockJWTManager{token: "jwt-oauth-token"}
	settings := &mockAuthSettingService{valueByKey: map[string]string{consts.SiteAllowRegisterSettingKey: "true"}}
	svc := &oauthService{
		oauthManager:  oauthMgr,
		oauthBindRepo: bindRepo,
		userRepo:      userRepo,
		jwtManager:    jwtMgr,
		settings:      settings,
		db:            db,
	}
	return svc, oauthMgr
}

func TestOAuthServiceGetEnabledProviders(t *testing.T) {
	svc, _, _, _ := newMockOAuthService()
	providers := svc.GetEnabledProviders()
	if len(providers) != 1 {
		t.Fatalf("expected 1 provider, got %d", len(providers))
	}
	if providers[0].Provider != "google" {
		t.Fatalf("expected google, got %q", providers[0].Provider)
	}
	if providers[0].AuthURL != "/api/v1/auth/oauth/google" {
		t.Fatalf("unexpected auth url: %q", providers[0].AuthURL)
	}
}

func TestOAuthServiceHandleCallbackAlreadyBound(t *testing.T) {
	svc, userRepo, oauthMgr, bindRepo := newMockOAuthService()

	// 已存在的本地用户
	userRepo.users[1] = &model.User{
		ID:       1,
		Username: "existing",
		Nickname: "Existing",
		Email:    "existing@example.com",
		Status:   1,
	}
	// 已绑定记录
	bindRepo.binds = []model.OAuthBind{
		{
			ID:             1,
			UserID:         1,
			Provider:       "google",
			ProviderUserID: "google-uid-123",
			ProviderEmail:  "existing@example.com",
		},
	}
	oauthMgr.fetchUserInfo = &oauth.ProviderUserInfo{
		ProviderUserID: "google-uid-123",
		Email:          "existing@example.com",
		Nickname:       "Existing",
	}

	result, err := svc.HandleCallback(context.Background(), "google", "code-1", "state-1", "state-1")
	if err != nil {
		t.Fatalf("HandleCallback returned error: %v", err)
	}
	if result.AccessToken != "jwt-oauth-token" {
		t.Fatalf("expected jwt-oauth-token, got %q", result.AccessToken)
	}
	// 不应创建新用户
	if len(userRepo.users) > 1 {
		t.Fatalf("expected no new user, got %d users", len(userRepo.users))
	}
}

func TestOAuthServiceHandleCallbackNewUserSilentRegister(t *testing.T) {
	svc, oauthMgr := newMockOAuthServiceWithDB(t)

	oauthMgr.fetchUserInfo = &oauth.ProviderUserInfo{
		ProviderUserID: "google-uid-new",
		Email:          "newuser@example.com",
		Nickname:       "NewUser",
	}

	result, err := svc.HandleCallback(context.Background(), "google", "code-1", "state-1", "state-1")
	if err != nil {
		t.Fatalf("HandleCallback returned error: %v", err)
	}
	if result.AccessToken != "jwt-oauth-token" {
		t.Fatalf("expected jwt-oauth-token, got %q", result.AccessToken)
	}

	// 验证 DB 中确实创建了用户和绑定
	user, err := svc.userRepo.GetByEmail(context.Background(), "newuser@example.com")
	if err != nil {
		t.Fatalf("get user by email: %v", err)
	}
	if !user.EmailVerified {
		t.Fatal("expected new OAuth user to be email verified")
	}
	if !strings.HasPrefix(user.Username, "google_") {
		t.Fatalf("expected username to start with google_, got %q", user.Username)
	}

	bind, err := svc.oauthBindRepo.FindByProviderAndUID(context.Background(), "google", "google-uid-new")
	if err != nil {
		t.Fatalf("get bind: %v", err)
	}
	if bind.UserID != user.ID {
		t.Fatalf("expected bind.UserID %d, got %d", user.ID, bind.UserID)
	}
}

func TestOAuthServiceHandleCallbackNewUserNicknameFallback(t *testing.T) {
	svc, oauthMgr := newMockOAuthServiceWithDB(t)

	oauthMgr.fetchUserInfo = &oauth.ProviderUserInfo{
		ProviderUserID: "google-uid-no-nick",
		Email:          "nonick@example.com",
		Nickname:       "",
	}

	_, err := svc.HandleCallback(context.Background(), "google", "code-1", "state-1", "state-1")
	if err != nil {
		t.Fatalf("HandleCallback returned error: %v", err)
	}
	user, err := svc.userRepo.GetByEmail(context.Background(), "nonick@example.com")
	if err != nil {
		t.Fatalf("get user by email: %v", err)
	}
	if user.Nickname == "" {
		t.Fatal("expected nickname to fall back to username, got empty")
	}
}

func TestOAuthServiceHandleCallbackEmailExistsUnbound(t *testing.T) {
	svc, userRepo, oauthMgr, _ := newMockOAuthService()

	// 已存在同邮箱用户但未绑定
	userRepo.users[1] = &model.User{
		ID:       1,
		Username: "existing",
		Nickname: "Existing",
		Email:    "existing@example.com",
		Status:   1,
	}
	oauthMgr.fetchUserInfo = &oauth.ProviderUserInfo{
		ProviderUserID: "google-uid-another",
		Email:          "existing@example.com",
		Nickname:       "Existing",
	}

	_, err := svc.HandleCallback(context.Background(), "google", "code-1", "state-1", "state-1")
	if !errors.Is(err, ErrOAuthEmailExistsUnbound) {
		t.Fatalf("expected ErrOAuthEmailExistsUnbound, got %v", err)
	}
}

func TestOAuthServiceHandleCallbackStateMismatch(t *testing.T) {
	svc, _, _, _ := newMockOAuthService()

	_, err := svc.HandleCallback(context.Background(), "google", "code-1", "query-state", "cookie-state")
	if !errors.Is(err, ErrOAuthStateMismatch) {
		t.Fatalf("expected ErrOAuthStateMismatch, got %v", err)
	}

	_, err = svc.HandleCallback(context.Background(), "google", "code-1", "", "cookie-state")
	if !errors.Is(err, ErrOAuthStateMismatch) {
		t.Fatalf("expected ErrOAuthStateMismatch for empty state, got %v", err)
	}
}

func TestOAuthServiceHandleCallbackProviderNotConfigured(t *testing.T) {
	svc, _, _, _ := newMockOAuthService()

	_, err := svc.HandleCallback(context.Background(), "github", "code-1", "state-1", "state-1")
	if !errors.Is(err, ErrOAuthNotConfigured) {
		t.Fatalf("expected ErrOAuthNotConfigured, got %v", err)
	}
}

func TestOAuthServiceHandleCallbackUserDisabled(t *testing.T) {
	svc, userRepo, oauthMgr, bindRepo := newMockOAuthService()

	userRepo.users[1] = &model.User{
		ID:       1,
		Username: "banned",
		Nickname: "Banned",
		Email:    "banned@example.com",
		Status:   2, // 封禁
	}
	bindRepo.binds = []model.OAuthBind{
		{
			ID:             1,
			UserID:         1,
			Provider:       "google",
			ProviderUserID: "google-uid-banned",
			ProviderEmail:  "banned@example.com",
		},
	}
	oauthMgr.fetchUserInfo = &oauth.ProviderUserInfo{
		ProviderUserID: "google-uid-banned",
		Email:          "banned@example.com",
		Nickname:       "Banned",
	}

	_, err := svc.HandleCallback(context.Background(), "google", "code-1", "state-1", "state-1")
	if !errors.Is(err, ErrUserDisabled) {
		t.Fatalf("expected ErrUserDisabled, got %v", err)
	}
}

func TestOAuthServiceHandleCallbackEmailMissing(t *testing.T) {
	svc, _, oauthMgr, _ := newMockOAuthService()

	oauthMgr.fetchUserInfo = &oauth.ProviderUserInfo{
		ProviderUserID: "google-uid-noemail",
		Email:          "",
		Nickname:       "NoEmail",
	}

	_, err := svc.HandleCallback(context.Background(), "google", "code-1", "state-1", "state-1")
	if !errors.Is(err, oauth.ErrEmailNotReturned) {
		t.Fatalf("expected ErrEmailNotReturned, got %v", err)
	}
}

func TestOAuthServiceHandleCallbackExchangeFailed(t *testing.T) {
	svc, _, oauthMgr, _ := newMockOAuthService()

	oauthMgr.exchangeErr = errors.New("exchange failed")

	_, err := svc.HandleCallback(context.Background(), "google", "code-1", "state-1", "state-1")
	if err == nil {
		t.Fatal("expected error from exchange failure")
	}
}

func TestOAuthServiceGetUserBindings(t *testing.T) {
	svc, _, _, bindRepo := newMockOAuthService()

	bindRepo.binds = []model.OAuthBind{
		{ID: 1, UserID: 1, Provider: "google", ProviderUserID: "g-1", ProviderEmail: "g@example.com"},
		{ID: 2, UserID: 1, Provider: "github", ProviderUserID: "gh-1", ProviderEmail: "gh@example.com"},
		{ID: 3, UserID: 2, Provider: "google", ProviderUserID: "g-2", ProviderEmail: "other@example.com"},
	}

	binds, err := svc.GetUserBindings(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetUserBindings returned error: %v", err)
	}
	if len(binds) != 2 {
		t.Fatalf("expected 2 binds for user 1, got %d", len(binds))
	}
}

func TestOAuthServiceUnbind(t *testing.T) {
	svc, _, _, bindRepo := newMockOAuthService()

	bindRepo.binds = []model.OAuthBind{
		{ID: 1, UserID: 1, Provider: "google", ProviderUserID: "g-1"},
		{ID: 2, UserID: 1, Provider: "github", ProviderUserID: "gh-1"},
	}

	err := svc.Unbind(context.Background(), 1, 1)
	if err != nil {
		t.Fatalf("Unbind returned error: %v", err)
	}
	if len(bindRepo.binds) != 1 {
		t.Fatalf("expected 1 bind remaining, got %d", len(bindRepo.binds))
	}
}

func TestOAuthServiceUnbindNotFound(t *testing.T) {
	svc, _, _, bindRepo := newMockOAuthService()

	bindRepo.binds = []model.OAuthBind{
		{ID: 1, UserID: 1, Provider: "google", ProviderUserID: "g-1"},
	}

	err := svc.Unbind(context.Background(), 1, 999)
	if !errors.Is(err, ErrOAuthBindNotFound) {
		t.Fatalf("expected ErrOAuthBindNotFound, got %v", err)
	}
}

func TestOAuthServiceBindManually(t *testing.T) {
	svc, _, oauthMgr, bindRepo := newMockOAuthService()

	oauthMgr.fetchUserInfo = &oauth.ProviderUserInfo{
		ProviderUserID: "google-uid-bind",
		Email:          "bind@example.com",
		Nickname:       "BindUser",
	}

	err := svc.BindManually(context.Background(), 1, "google", "code-1", "state-1", "state-1")
	if err != nil {
		t.Fatalf("BindManually returned error: %v", err)
	}
	if len(bindRepo.binds) != 1 {
		t.Fatalf("expected 1 bind, got %d", len(bindRepo.binds))
	}
	if bindRepo.binds[0].UserID != 1 {
		t.Fatalf("expected userID 1, got %d", bindRepo.binds[0].UserID)
	}
}

func TestOAuthServiceBindManuallyStateMismatch(t *testing.T) {
	svc, _, _, _ := newMockOAuthService()

	err := svc.BindManually(context.Background(), 1, "google", "code-1", "query-state", "cookie-state")
	if !errors.Is(err, ErrOAuthStateMismatch) {
		t.Fatalf("expected ErrOAuthStateMismatch, got %v", err)
	}
}

func TestOAuthServiceBindManuallyAlreadyBoundToOtherUser(t *testing.T) {
	svc, _, oauthMgr, bindRepo := newMockOAuthService()

	bindRepo.binds = []model.OAuthBind{
		{ID: 1, UserID: 2, Provider: "google", ProviderUserID: "google-uid-taken"},
	}
	oauthMgr.fetchUserInfo = &oauth.ProviderUserInfo{
		ProviderUserID: "google-uid-taken",
		Email:          "taken@example.com",
		Nickname:       "Taken",
	}

	err := svc.BindManually(context.Background(), 1, "google", "code-1", "state-1", "state-1")
	if !errors.Is(err, ErrOAuthBindAlreadyExists) {
		t.Fatalf("expected ErrOAuthBindAlreadyExists, got %v", err)
	}
}

func TestOAuthServiceBindManuallyAlreadyBoundToSelf(t *testing.T) {
	svc, _, oauthMgr, bindRepo := newMockOAuthService()

	bindRepo.binds = []model.OAuthBind{
		{ID: 1, UserID: 1, Provider: "google", ProviderUserID: "google-uid-self"},
	}
	oauthMgr.fetchUserInfo = &oauth.ProviderUserInfo{
		ProviderUserID: "google-uid-self",
		Email:          "self@example.com",
		Nickname:       "Self",
	}

	err := svc.BindManually(context.Background(), 1, "google", "code-1", "state-1", "state-1")
	if err == nil {
		t.Fatal("expected error for already bound provider")
	}
}

func TestOAuthServiceBindManuallyProviderNotConfigured(t *testing.T) {
	svc, _, _, _ := newMockOAuthService()

	err := svc.BindManually(context.Background(), 1, "github", "code-1", "state-1", "state-1")
	if !errors.Is(err, ErrOAuthNotConfigured) {
		t.Fatalf("expected ErrOAuthNotConfigured, got %v", err)
	}
}

func TestOAuthServiceGetAuthURL(t *testing.T) {
	svc, _, _, _ := newMockOAuthService()

	url, err := svc.GetAuthURL(context.Background(), "google", "state-123")
	if err != nil {
		t.Fatalf("GetAuthURL returned error: %v", err)
	}
	if url == "" {
		t.Fatal("expected non-empty auth URL")
	}
}

func TestOAuthServiceGetAuthURLDisabledProvider(t *testing.T) {
	svc, _, _, _ := newMockOAuthService()

	_, err := svc.GetAuthURL(context.Background(), "github", "state-123")
	if err == nil {
		t.Fatal("expected error for disabled provider")
	}
}

func TestOAuthServiceHandleCallbackRegisterClosed(t *testing.T) {
	db := openTestDB(t)
	oauthMgr := &mockOAuthManager{
		enabledProviders: []string{"google"},
		fetchUserInfo: &oauth.ProviderUserInfo{
			ProviderUserID: "google-uid-new",
			Email:          "newuser@example.com",
			Nickname:       "NewUser",
		},
	}
	bindRepo := repo.NewOAuthBindRepo(db)
	userRepo := repo.NewUserRepo(db)
	jwtMgr := &mockJWTManager{token: "jwt-oauth-token"}
	settings := &mockAuthSettingService{valueByKey: map[string]string{consts.SiteAllowRegisterSettingKey: "false"}}
	svc := &oauthService{
		oauthManager:  oauthMgr,
		oauthBindRepo: bindRepo,
		userRepo:      userRepo,
		jwtManager:    jwtMgr,
		settings:      settings,
		db:            db,
	}

	_, err := svc.HandleCallback(context.Background(), "google", "code-1", "state-1", "state-1")
	if !errors.Is(err, ErrRegisterClosed) {
		t.Fatalf("expected ErrRegisterClosed, got %v", err)
	}
}

// 确保 mock 实现了接口
var _ repo.OAuthBindRepo = (*mockOAuthBindRepo)(nil)
var _ oauth.Manager = (*mockOAuthManager)(nil)
