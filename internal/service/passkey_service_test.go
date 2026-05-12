package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"easydrop/internal/model"
	"easydrop/internal/repo"

	wa "github.com/go-webauthn/webauthn/webauthn"
	"github.com/go-webauthn/webauthn/protocol"
	"gorm.io/gorm"
)

type mockPasskeyRepo struct {
	passkeys      map[uint]*model.PasskeyCredential
	byCredential  map[string]*model.PasskeyCredential
	nextID        uint
	createWithErr error
	updateNameErr error
	updateCredErr error
	findByIDErr   error
	deleteErr     error
}

func newMockPasskeyRepo() *mockPasskeyRepo {
	return &mockPasskeyRepo{
		passkeys:     make(map[uint]*model.PasskeyCredential),
		byCredential: make(map[string]*model.PasskeyCredential),
		nextID:       1,
	}
}

func (m *mockPasskeyRepo) Create(_ context.Context, p *model.PasskeyCredential) error {
	if p.ID == 0 {
		p.ID = m.nextID
		m.nextID++
	}
	clone := *p
	m.passkeys[p.ID] = &clone
	m.byCredential[p.CredentialID] = &clone
	return nil
}

func (m *mockPasskeyRepo) CreateWithLimit(_ context.Context, p *model.PasskeyCredential, max int) (int64, error) {
	if m.createWithErr != nil {
		return 0, m.createWithErr
	}

	var count int64
	for _, pk := range m.passkeys {
		if pk.UserID == p.UserID {
			count++
		}
	}
	if count >= int64(max) {
		return 0, repo.ErrPasskeyLimitExceeded
	}

	if p.ID == 0 {
		p.ID = m.nextID
		m.nextID++
	}
	clone := *p
	m.passkeys[p.ID] = &clone
	m.byCredential[p.CredentialID] = &clone
	return count, nil
}

func (m *mockPasskeyRepo) FindByID(_ context.Context, id uint) (*model.PasskeyCredential, error) {
	if m.findByIDErr != nil {
		return nil, m.findByIDErr
	}
	p, ok := m.passkeys[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	clone := *p
	return &clone, nil
}

func (m *mockPasskeyRepo) FindByCredentialID(_ context.Context, credentialID string) (*model.PasskeyCredential, error) {
	p, ok := m.byCredential[credentialID]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	clone := *p
	return &clone, nil
}

func (m *mockPasskeyRepo) FindByUserID(_ context.Context, userID uint) ([]model.PasskeyCredential, error) {
	var result []model.PasskeyCredential
	for _, p := range m.passkeys {
		if p.UserID == userID {
			result = append(result, *p)
		}
	}
	return result, nil
}

func (m *mockPasskeyRepo) CountByUserID(_ context.Context, userID uint) (int64, error) {
	var count int64
	for _, p := range m.passkeys {
		if p.UserID == userID {
			count++
		}
	}
	return count, nil
}

func (m *mockPasskeyRepo) UpdateName(_ context.Context, id uint, name string) error {
	if m.updateNameErr != nil {
		return m.updateNameErr
	}
	p, ok := m.passkeys[id]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	p.Name = name
	return nil
}

func (m *mockPasskeyRepo) UpdateCredential(_ context.Context, id uint, credential *wa.Credential) error {
	if m.updateCredErr != nil {
		return m.updateCredErr
	}
	p, ok := m.passkeys[id]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	data, _ := json.Marshal(credential)
	p.CredentialJSON = data
	return nil
}

func (m *mockPasskeyRepo) Delete(_ context.Context, id uint) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	p, ok := m.passkeys[id]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	delete(m.byCredential, p.CredentialID)
	delete(m.passkeys, id)
	return nil
}

type mockWebAuthnManager struct {
	creation           *protocol.CredentialCreation
	assertion          *protocol.CredentialAssertion
	credential         *wa.Credential
	waUser             wa.User
	sessionID          string
	beginRegErr        error
	finishRegErr       error
	beginLoginErr      error
	finishLoginErr     error
}

func (m *mockWebAuthnManager) BeginRegistration(_ wa.User) (*protocol.CredentialCreation, string, error) {
	if m.beginRegErr != nil {
		return nil, "", m.beginRegErr
	}
	return m.creation, m.sessionID, nil
}

func (m *mockWebAuthnManager) FinishRegistration(_ wa.User, _ string, _ []byte) (*wa.Credential, error) {
	if m.finishRegErr != nil {
		return nil, m.finishRegErr
	}
	return m.credential, nil
}

func (m *mockWebAuthnManager) BeginDiscoverableLogin() (*protocol.CredentialAssertion, string, error) {
	if m.beginLoginErr != nil {
		return nil, "", m.beginLoginErr
	}
	return m.assertion, m.sessionID, nil
}

func (m *mockWebAuthnManager) FinishDiscoverableLogin(_ wa.DiscoverableUserHandler, _ string, _ []byte) (wa.User, *wa.Credential, error) {
	if m.finishLoginErr != nil {
		return nil, nil, m.finishLoginErr
	}
	return m.waUser, m.credential, nil
}

func testCredential(id string) *wa.Credential {
	rawID, _ := base64.RawURLEncoding.DecodeString(id)
	return &wa.Credential{
		ID:              rawID,
		PublicKey:       []byte("test-public-key"),
		AttestationType: "none",
	}
}

func testCredentialJSON(id string) []byte {
	data, _ := json.Marshal(testCredential(id))
	return data
}

func testUser(id uint) *model.User {
	return &model.User{
		ID:        id,
		Username:  "testuser",
		Nickname:  "测试用户",
		Email:     "test@example.com",
		Password:  "hashed",
		Admin:     false,
		Status:    1,
	}
}

func setupPasskeyService() (*passkeyService, *mockPasskeyRepo, *mockUserRepo, *mockWebAuthnManager) {
	passkeyRepo := newMockPasskeyRepo()
	userRepo := &mockUserRepo{users: make(map[uint]*model.User)}
	waMgr := &mockWebAuthnManager{
		sessionID:  "test-session-id",
		creation:   &protocol.CredentialCreation{},
		assertion:  &protocol.CredentialAssertion{},
		credential: testCredential("test-credential-id"),
	}
	svc := &passkeyService{
		passkeyRepo: passkeyRepo,
		userRepo:    userRepo,
		webauthnMgr: waMgr,
	}
	return svc, passkeyRepo, userRepo, waMgr
}

func TestBeginRegistrationSuccess(t *testing.T) {
	svc, _, userRepo, _ := setupPasskeyService()
	u := testUser(1)
	userRepo.Create(context.Background(), u)

	_, err := svc.BeginRegistration(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
}

func TestBeginRegistrationUserNotFound(t *testing.T) {
	svc, _, _, _ := setupPasskeyService()

	_, err := svc.BeginRegistration(context.Background(), 999)
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestBeginRegistrationLimitReached(t *testing.T) {
	svc, _, userRepo, _ := setupPasskeyService()
	u := testUser(1)
	userRepo.Create(context.Background(), u)

	for i := 0; i < maxPasskeysPerUser; i++ {
		svc.passkeyRepo.Create(context.Background(), &model.PasskeyCredential{
			UserID:       1,
			CredentialID: "cred-" + string(rune('0'+i)),
			Name:         "Passkey",
		})
	}

	_, err := svc.BeginRegistration(context.Background(), 1)
	if !errors.Is(err, ErrPasskeyLimitReached) {
		t.Fatalf("expected ErrPasskeyLimitReached, got %v", err)
	}
}

func TestBeginRegistrationBeginError(t *testing.T) {
	svc, _, userRepo, waMgr := setupPasskeyService()
	u := testUser(1)
	userRepo.Create(context.Background(), u)
	waMgr.beginRegErr = errors.New("begin failed")

	_, err := svc.BeginRegistration(context.Background(), 1)
	if !errors.Is(err, ErrInternal) {
		t.Fatalf("expected ErrInternal, got %v", err)
	}
}

func TestFinishRegistrationSuccess(t *testing.T) {
	svc, passkeyRepo, userRepo, _ := setupPasskeyService()
	u := testUser(1)
	userRepo.Create(context.Background(), u)

	err := svc.FinishRegistration(context.Background(), 1, "sid", []byte("{}"))
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	passkeys, _ := passkeyRepo.FindByUserID(context.Background(), 1)
	if len(passkeys) != 1 {
		t.Fatalf("expected 1 passkey, got %d", len(passkeys))
	}
	if passkeys[0].Name != "通行密钥 1" {
		t.Fatalf("expected auto name '通行密钥 1', got '%s'", passkeys[0].Name)
	}
}

func TestFinishRegistrationUserNotFound(t *testing.T) {
	svc, _, _, _ := setupPasskeyService()

	err := svc.FinishRegistration(context.Background(), 999, "sid", []byte("{}"))
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestFinishRegistrationFinishError(t *testing.T) {
	svc, _, userRepo, waMgr := setupPasskeyService()
	u := testUser(1)
	userRepo.Create(context.Background(), u)
	waMgr.finishRegErr = errors.New("finish failed")

	err := svc.FinishRegistration(context.Background(), 1, "sid", []byte("{}"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestFinishRegistrationLimitReachedInTransaction(t *testing.T) {
	svc, passkeyRepo, userRepo, _ := setupPasskeyService()
	u := testUser(1)
	userRepo.Create(context.Background(), u)

	for i := 0; i < maxPasskeysPerUser; i++ {
		passkeyRepo.Create(context.Background(), &model.PasskeyCredential{
			UserID:         1,
			CredentialID:   "cred-" + string(rune('0'+i)),
			Name:           "Passkey",
			CredentialJSON: testCredentialJSON("cred-" + string(rune('0'+i))),
		})
	}

	err := svc.FinishRegistration(context.Background(), 1, "sid", []byte("{}"))
	if !errors.Is(err, ErrPasskeyLimitReached) {
		t.Fatalf("expected ErrPasskeyLimitReached, got %v", err)
	}
}

func TestBeginLoginSuccess(t *testing.T) {
	svc, _, _, _ := setupPasskeyService()

	result, err := svc.BeginLogin(context.Background())
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if result.SessionID != "test-session-id" {
		t.Fatalf("expected session_id 'test-session-id', got '%s'", result.SessionID)
	}
}

func TestBeginLoginError(t *testing.T) {
	svc, _, _, waMgr := setupPasskeyService()
	waMgr.beginLoginErr = errors.New("begin login failed")

	_, err := svc.BeginLogin(context.Background())
	if !errors.Is(err, ErrInternal) {
		t.Fatalf("expected ErrInternal, got %v", err)
	}
}

func TestFinishLoginSuccess(t *testing.T) {
	svc, passkeyRepo, userRepo, waMgr := setupPasskeyService()
	u := testUser(1)
	userRepo.Create(context.Background(), u)

	cred := testCredential("login-cred-id")
	credJSON, _ := json.Marshal(cred)
	passkeyRepo.Create(context.Background(), &model.PasskeyCredential{
		UserID:         1,
		CredentialID:   base64.RawURLEncoding.EncodeToString(cred.ID),
		CredentialJSON: credJSON,
		Name:           "我的密钥",
	})

	waMgr.waUser = &passkeyUser{user: u, credentials: []wa.Credential{*cred}}
	waMgr.credential = cred

	result, err := svc.FinishLogin(context.Background(), "sid", []byte("{}"))
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if result.ID != 1 {
		t.Fatalf("expected user ID 1, got %d", result.ID)
	}
}

func TestFinishLoginFinishError(t *testing.T) {
	svc, _, _, waMgr := setupPasskeyService()
	waMgr.finishLoginErr = errors.New("finish login failed")

	_, err := svc.FinishLogin(context.Background(), "sid", []byte("{}"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestListSuccess(t *testing.T) {
	svc, passkeyRepo, _, _ := setupPasskeyService()
	now := time.Now()

	passkeyRepo.Create(context.Background(), &model.PasskeyCredential{
		ID:         1,
		UserID:     1,
		Name:       "密钥 A",
		CreatedAt:  now,
		CredentialID: "cred-a",
	})
	passkeyRepo.Create(context.Background(), &model.PasskeyCredential{
		ID:         2,
		UserID:     1,
		Name:       "密钥 B",
		CreatedAt:  now,
		CredentialID: "cred-b",
	})

	items, err := svc.List(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].Name != "密钥 A" {
		t.Fatalf("expected name '密钥 A', got '%s'", items[0].Name)
	}
}

func TestListEmpty(t *testing.T) {
	svc, _, _, _ := setupPasskeyService()

	items, err := svc.List(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected 0 items, got %d", len(items))
	}
}

func TestRenameSuccess(t *testing.T) {
	svc, passkeyRepo, _, _ := setupPasskeyService()
	passkeyRepo.Create(context.Background(), &model.PasskeyCredential{
		ID:           1,
		UserID:       1,
		Name:         "旧名称",
		CredentialID: "cred-rename",
	})

	err := svc.Rename(context.Background(), 1, 1, "新名称")
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	p, _ := passkeyRepo.FindByID(context.Background(), 1)
	if p.Name != "新名称" {
		t.Fatalf("expected name '新名称', got '%s'", p.Name)
	}
}

func TestRenameEmptyName(t *testing.T) {
	svc, _, _, _ := setupPasskeyService()

	err := svc.Rename(context.Background(), 1, 1, "   ")
	if !errors.Is(err, ErrPasskeyNameEmpty) {
		t.Fatalf("expected ErrPasskeyNameEmpty, got %v", err)
	}
}

func TestRenameNotFound(t *testing.T) {
	svc, _, _, _ := setupPasskeyService()

	err := svc.Rename(context.Background(), 1, 999, "新名称")
	if !errors.Is(err, ErrPasskeyNotFound) {
		t.Fatalf("expected ErrPasskeyNotFound, got %v", err)
	}
}

func TestRenameNotOwner(t *testing.T) {
	svc, passkeyRepo, _, _ := setupPasskeyService()
	passkeyRepo.Create(context.Background(), &model.PasskeyCredential{
		ID:           1,
		UserID:       2,
		Name:         "密钥",
		CredentialID: "cred-other",
	})

	err := svc.Rename(context.Background(), 1, 1, "新名称")
	if !errors.Is(err, ErrPasskeyNotFound) {
		t.Fatalf("expected ErrPasskeyNotFound (not owner), got %v", err)
	}
}

func TestDeleteSuccess(t *testing.T) {
	svc, passkeyRepo, _, _ := setupPasskeyService()
	passkeyRepo.Create(context.Background(), &model.PasskeyCredential{
		ID:           1,
		UserID:       1,
		Name:         "密钥",
		CredentialID: "cred-del",
	})

	err := svc.Delete(context.Background(), 1, 1)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	_, err = passkeyRepo.FindByID(context.Background(), 1)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected passkey to be deleted, but still exists")
	}
}

func TestDeleteNotFound(t *testing.T) {
	svc, _, _, _ := setupPasskeyService()

	err := svc.Delete(context.Background(), 1, 999)
	if !errors.Is(err, ErrPasskeyNotFound) {
		t.Fatalf("expected ErrPasskeyNotFound, got %v", err)
	}
}

func TestDeleteNotOwner(t *testing.T) {
	svc, passkeyRepo, _, _ := setupPasskeyService()
	passkeyRepo.Create(context.Background(), &model.PasskeyCredential{
		ID:           1,
		UserID:       2,
		Name:         "密钥",
		CredentialID: "cred-other-del",
	})

	err := svc.Delete(context.Background(), 1, 1)
	if !errors.Is(err, ErrPasskeyNotFound) {
		t.Fatalf("expected ErrPasskeyNotFound (not owner), got %v", err)
	}
}
