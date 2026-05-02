package service

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	"easydrop/internal/dto"
	"easydrop/internal/model"
	pk "easydrop/internal/pkg/webauthn"
	"easydrop/internal/repo"

	wa "github.com/go-webauthn/webauthn/webauthn"
	"gorm.io/gorm"
)

var (
	// ErrPasskeyLimitReached 表示用户已达到通行密钥数量上限。
	ErrPasskeyLimitReached = errors.New("每个用户最多创建10个通行密钥")
	// ErrPasskeyNotFound 表示指定的通行密钥不存在或不属于当前用户。
	ErrPasskeyNotFound = errors.New("通行密钥不存在")
	// ErrPasskeyNameTooLong 表示通行密钥名称不符合长度要求。
	ErrPasskeyNameTooLong = errors.New("通行密钥名称不能超过15个字符")
)

// maxPasskeysPerUser 是每个用户最多可创建的通行密钥数量。
const maxPasskeysPerUser = 10

// PasskeyService 定义了通行密钥 (WebAuthn) 的业务逻辑接口。
type PasskeyService interface {
	// BeginRegistration 发起通行密钥注册流程，返回客户端所需的创建选项和会话 ID。
	BeginRegistration(ctx context.Context, userID uint) (*dto.PasskeyRegisterBeginResponse, error)
	// FinishRegistration 完成通行密钥注册流程，验证客户端响应并保存凭证。
	FinishRegistration(ctx context.Context, userID uint, sessionID string, body []byte) error
	// BeginLogin 发起通行密钥登录流程，返回客户端所需的断言选项和会话 ID。
	BeginLogin(ctx context.Context) (*dto.PasskeyLoginBeginResponse, error)
	// FinishLogin 完成通行密钥登录流程，验证断言并返回对应的用户。
	FinishLogin(ctx context.Context, sessionID string, body []byte) (*model.User, error)
	// List 列出指定用户的所有通行密钥。
	List(ctx context.Context, userID uint) ([]dto.PasskeyItem, error)
	// Rename 重命名指定通行密钥。
	Rename(ctx context.Context, userID uint, passkeyID uint, name string) error
	// Delete 删除指定通行密钥。
	Delete(ctx context.Context, userID uint, passkeyID uint) error
}

// passkeyService 是 PasskeyService 的具体实现。
type passkeyService struct {
	passkeyRepo repo.PasskeyRepo
	userRepo    repo.UserRepo
	webauthnMgr pk.Manager
}

// NewPasskeyService 创建通行密钥服务实例。
func NewPasskeyService(passkeyRepo repo.PasskeyRepo, userRepo repo.UserRepo, webauthnMgr pk.Manager) PasskeyService {
	return &passkeyService{
		passkeyRepo: passkeyRepo,
		userRepo:    userRepo,
		webauthnMgr: webauthnMgr,
	}
}

// passkeyUser 实现 go-webauthn 的 User 接口，将业务模型适配到 WebAuthn 库。
type passkeyUser struct {
	user        *model.User
	credentials []wa.Credential
}

// WebAuthnID 返回用户的 WebAuthn 唯一标识，以用户 ID 的大端序字节表示。
func (u *passkeyUser) WebAuthnID() []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(u.user.ID))
	return buf
}

// WebAuthnName 返回用户的 WebAuthn 名称，使用邮箱作为唯一可读标识。
func (u *passkeyUser) WebAuthnName() string {
	return u.user.Email
}

// WebAuthnDisplayName 返回用户的展示名称，使用昵称。
func (u *passkeyUser) WebAuthnDisplayName() string {
	return u.user.Nickname
}

// WebAuthnCredentials 返回用户已注册的所有 WebAuthn 凭证。
func (u *passkeyUser) WebAuthnCredentials() []wa.Credential {
	return u.credentials
}

// BeginRegistration 发起通行密钥注册。
// 检查用户是否存在、数量是否达到上限（10 个），加载已有凭证用于排除重复注册。
func (s *passkeyService) BeginRegistration(ctx context.Context, userID uint) (*dto.PasskeyRegisterBeginResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		log.Printf("获取用户失败: %v", err)
		return nil, ErrInternal
	}

	count, err := s.passkeyRepo.CountByUserID(ctx, userID)
	if err != nil {
		log.Printf("查询通行密钥数量失败: %v", err)
		return nil, ErrInternal
	}
	if count >= maxPasskeysPerUser {
		return nil, ErrPasskeyLimitReached
	}

	passkeys, err := s.passkeyRepo.FindByUserID(ctx, userID)
	if err != nil {
		log.Printf("查询通行密钥列表失败: %v", err)
		return nil, ErrInternal
	}

	// 反序列化已有凭证，供 go-webauthn 在注册时排除。
	credentials := make([]wa.Credential, 0, len(passkeys))
	for _, pk := range passkeys {
		var cred wa.Credential
		if err := json.Unmarshal(pk.CredentialJSON, &cred); err != nil {
			log.Printf("解析通行密钥凭证失败: %v", err)
			continue
		}
		credentials = append(credentials, cred)
	}

	waUser := &passkeyUser{
		user:        user,
		credentials: credentials,
	}

	creation, sessionID, err := s.webauthnMgr.BeginRegistration(waUser)
	if err != nil {
		log.Printf("开始注册通行密钥失败: %v", err)
		return nil, ErrInternal
	}

	return &dto.PasskeyRegisterBeginResponse{
		Options:   creationToMap(creation),
		SessionID: sessionID,
	}, nil
}

// FinishRegistration 完成通行密钥注册。
// 验证客户端响应后，自动命名为 "通行密钥 N"（N 为当前数量 + 1）并保存到数据库。
func (s *passkeyService) FinishRegistration(ctx context.Context, userID uint, sessionID string, body []byte) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		log.Printf("获取用户失败: %v", err)
		return ErrInternal
	}

	count, err := s.passkeyRepo.CountByUserID(ctx, userID)
	if err != nil {
		log.Printf("查询通行密钥数量失败: %v", err)
		return ErrInternal
	}
	if count >= maxPasskeysPerUser {
		return ErrPasskeyLimitReached
	}

	passkeys, err := s.passkeyRepo.FindByUserID(ctx, userID)
	if err != nil {
		log.Printf("查询通行密钥列表失败: %v", err)
		return ErrInternal
	}

	credentials := make([]wa.Credential, 0, len(passkeys))
	for _, pk := range passkeys {
		var cred wa.Credential
		if err := json.Unmarshal(pk.CredentialJSON, &cred); err != nil {
			log.Printf("解析通行密钥凭证失败: %v", err)
			continue
		}
		credentials = append(credentials, cred)
	}

	waUser := &passkeyUser{
		user:        user,
		credentials: credentials,
	}

	credential, err := s.webauthnMgr.FinishRegistration(waUser, sessionID, body)
	if err != nil {
		log.Printf("完成注册通行密钥失败: %v", err)
		return errors.New("通行密钥注册失败，请重试")
	}

	credJSON, err := json.Marshal(credential)
	if err != nil {
		log.Printf("序列化凭证失败: %v", err)
		return ErrInternal
	}

	// 自动命名规则: "通行密钥 1", "通行密钥 2", ...
	autoName := fmt.Sprintf("通行密钥 %d", count+1)
	credentialID := base64.RawURLEncoding.EncodeToString(credential.ID)

	passkeyRecord := &model.PasskeyCredential{
		Name:           autoName,
		UserID:         userID,
		CredentialID:   credentialID,
		CredentialJSON: credJSON,
	}

	if err := s.passkeyRepo.Create(ctx, passkeyRecord); err != nil {
		log.Printf("保存通行密钥失败: %v", err)
		return ErrInternal
	}

	return nil
}

// BeginLogin 发起无用户名通行密钥登录（discoverable credential）。
// 无需事先知道用户身份，用户将在浏览器中选择要使用的通行密钥。
func (s *passkeyService) BeginLogin(ctx context.Context) (*dto.PasskeyLoginBeginResponse, error) {
	assertion, sessionID, err := s.webauthnMgr.BeginDiscoverableLogin()
	if err != nil {
		log.Printf("开始通行密钥登录失败: %v", err)
		return nil, ErrInternal
	}

	return &dto.PasskeyLoginBeginResponse{
		Options:   assertionToMap(assertion),
		SessionID: sessionID,
	}, nil
}

// FinishLogin 完成通行密钥登录。
// 通过凭证 ID 反查用户，验证断言合法性后，更新凭证的签名计数以防重放攻击。
func (s *passkeyService) FinishLogin(ctx context.Context, sessionID string, body []byte) (*model.User, error) {
	// discoverableUserHandler 在 go-webauthn 库验证过程中被回调，
	// 通过凭证 rawID 查找对应的用户并加载其所有凭证用于校验。
	handler := func(rawID, userHandle []byte) (wa.User, error) {
		credentialID := base64.RawURLEncoding.EncodeToString(rawID)
		passkey, err := s.passkeyRepo.FindByCredentialID(ctx, credentialID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, fmt.Errorf("凭证未找到")
			}
			log.Printf("按凭证ID查询通行密钥失败: %v", err)
			return nil, ErrInternal
		}

		user, err := s.userRepo.GetByID(ctx, passkey.UserID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, fmt.Errorf("用户不存在")
			}
			log.Printf("获取用户失败: %v", err)
			return nil, ErrInternal
		}

		if user.Status != 1 {
			return nil, ErrUserDisabled
		}

		passkeys, err := s.passkeyRepo.FindByUserID(ctx, user.ID)
		if err != nil {
			log.Printf("查询通行密钥列表失败: %v", err)
			return nil, ErrInternal
		}

		credentials := make([]wa.Credential, 0, len(passkeys))
		for _, pk := range passkeys {
			var cred wa.Credential
			if err := json.Unmarshal(pk.CredentialJSON, &cred); err != nil {
				log.Printf("解析通行密钥凭证失败: %v", err)
				continue
			}
			credentials = append(credentials, cred)
		}

		return &passkeyUser{
			user:        user,
			credentials: credentials,
		}, nil
	}

	waUser, credential, err := s.webauthnMgr.FinishDiscoverableLogin(handler, sessionID, body)
	if err != nil {
		log.Printf("完成通行密钥登录失败: %v", err)
		return nil, errors.New("通行密钥验证失败，请重试")
	}

	pkUser, ok := waUser.(*passkeyUser)
	if !ok {
		log.Printf("用户类型转换失败")
		return nil, ErrInternal
	}

	// 更新凭证的签名计数，防止克隆认证器重放攻击。
	credentialID := base64.RawURLEncoding.EncodeToString(credential.ID)
	passkey, err := s.passkeyRepo.FindByCredentialID(ctx, credentialID)
	if err == nil {
		if err := s.passkeyRepo.UpdateCredential(ctx, passkey.ID, credential); err != nil {
			log.Printf("更新通行密钥签名计数失败: %v", err)
		}
	}

	return pkUser.user, nil
}

// List 列出指定用户的所有通行密钥（仅元数据，不含密钥材料）。
func (s *passkeyService) List(ctx context.Context, userID uint) ([]dto.PasskeyItem, error) {
	passkeys, err := s.passkeyRepo.FindByUserID(ctx, userID)
	if err != nil {
		log.Printf("查询通行密钥列表失败: %v", err)
		return nil, ErrInternal
	}

	items := make([]dto.PasskeyItem, 0, len(passkeys))
	for _, pk := range passkeys {
		items = append(items, dto.PasskeyItem{
			ID:        pk.ID,
			Name:      pk.Name,
			CreatedAt: pk.CreatedAt,
		})
	}
	return items, nil
}

// Rename 重命名通行密钥，需校验名称长度（1-15 字符）和所有权。
func (s *passkeyService) Rename(ctx context.Context, userID uint, passkeyID uint, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return ErrPasskeyNameTooLong
	}

	passkey, err := s.passkeyRepo.FindByID(ctx, passkeyID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPasskeyNotFound
		}
		log.Printf("查询通行密钥失败: %v", err)
		return ErrInternal
	}

	if passkey.UserID != userID {
		return ErrPasskeyNotFound
	}

	return s.passkeyRepo.UpdateName(ctx, passkeyID, name)
}

// Delete 删除通行密钥，需校验所有权。
func (s *passkeyService) Delete(ctx context.Context, userID uint, passkeyID uint) error {
	passkey, err := s.passkeyRepo.FindByID(ctx, passkeyID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPasskeyNotFound
		}
		log.Printf("查询通行密钥失败: %v", err)
		return ErrInternal
	}

	if passkey.UserID != userID {
		return ErrPasskeyNotFound
	}

	return s.passkeyRepo.Delete(ctx, passkeyID)
}

// creationToMap 将 go-webauthn 的创建选项通过 JSON 序列化/反序列化转为 map，
// 便于直接作为 JSON 响应返回给前端。
func creationToMap(creation any) map[string]any {
	data, _ := json.Marshal(creation)
	var result map[string]any
	json.Unmarshal(data, &result)
	return result
}

// assertionToMap 将 go-webauthn 的断言选项通过 JSON 序列化/反序列化转为 map。
func assertionToMap(assertion any) map[string]any {
	data, _ := json.Marshal(assertion)
	var result map[string]any
	json.Unmarshal(data, &result)
	return result
}
