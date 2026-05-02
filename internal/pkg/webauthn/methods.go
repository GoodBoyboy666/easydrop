package webauthn

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	cryptorand "crypto/rand"

	"github.com/go-webauthn/webauthn/protocol"
	wa "github.com/go-webauthn/webauthn/webauthn"
)

// sessionIDBytes 是会话 ID 的随机字节数 (32 字节 → 43 字符 base64url)。
const sessionIDBytes = 32

// BeginRegistration 发起通行密钥注册流程。
// 调用 go-webauthn 生成凭证创建选项和会话数据，服务端保存会话后将会话 ID 一并返回。
// 要求创建常驻凭证 (discoverable credential) 并强制用户验证，确保无用户名登录可用。
func (m *manager) BeginRegistration(user wa.User) (*protocol.CredentialCreation, string, error) {
	creation, session, err := m.wa.BeginRegistration(
		user,
		wa.WithResidentKeyRequirement(protocol.ResidentKeyRequirementRequired),
		wa.WithAuthenticatorSelection(protocol.AuthenticatorSelection{
			UserVerification: protocol.VerificationRequired,
		}),
	)
	if err != nil {
		return nil, "", fmt.Errorf("开始注册: %w", err)
	}

	sessionID, err := generateSessionID()
	if err != nil {
		return nil, "", err
	}

	if err := m.sessionStore.Save(context.Background(), sessionID, session, m.timeout); err != nil {
		return nil, "", fmt.Errorf("保存会话: %w", err)
	}

	return creation, sessionID, nil
}

// FinishRegistration 完成通行密钥注册流程。
// 通过会话 ID 恢复会话数据，解析客户端响应并调用 go-webauthn 验证凭证合法性。
// 验证成功后会自动删除对应的会话数据。
func (m *manager) FinishRegistration(user wa.User, sessionID string, body []byte) (*wa.Credential, error) {
	session, err := m.sessionStore.Get(context.Background(), sessionID)
	if err != nil {
		return nil, fmt.Errorf("会话无效或已过期")
	}

	defer m.sessionStore.Delete(context.Background(), sessionID)

	parsed, err := protocol.ParseCredentialCreationResponseBytes(body)
	if err != nil {
		return nil, fmt.Errorf("解析注册响应: %w", err)
	}

	credential, err := m.wa.CreateCredential(user, *session, parsed)
	if err != nil {
		return nil, fmt.Errorf("验证注册凭证: %w", err)
	}

	return credential, nil
}

// BeginDiscoverableLogin 发起无用户名通行密钥登录流程。
// 使用 discoverable credential 方式，无需预先知道用户身份。
// 返回断言选项和会话 ID，服务端保存会话数据供后续验证使用。
// 要求用户验证 (PIN/生物特征)，提高安全性。
func (m *manager) BeginDiscoverableLogin() (*protocol.CredentialAssertion, string, error) {
	assertion, session, err := m.wa.BeginDiscoverableLogin(
		wa.WithUserVerification(protocol.VerificationRequired),
	)
	if err != nil {
		return nil, "", fmt.Errorf("开始登录: %w", err)
	}

	sessionID, err := generateSessionID()
	if err != nil {
		return nil, "", err
	}

	if err := m.sessionStore.Save(context.Background(), sessionID, session, m.timeout); err != nil {
		return nil, "", fmt.Errorf("保存会话: %w", err)
	}

	return assertion, sessionID, nil
}

// FinishDiscoverableLogin 完成无用户名通行密钥登录流程。
// 通过会话 ID 恢复会话数据，解析客户端断言响应，调用 handler 查找凭证所有者后验证。
// 验证成功后会自动删除对应的会话数据。
func (m *manager) FinishDiscoverableLogin(handler wa.DiscoverableUserHandler, sessionID string, body []byte) (wa.User, *wa.Credential, error) {
	session, err := m.sessionStore.Get(context.Background(), sessionID)
	if err != nil {
		return nil, nil, fmt.Errorf("会话无效或已过期")
	}

	defer m.sessionStore.Delete(context.Background(), sessionID)

	parsed, err := protocol.ParseCredentialRequestResponseBytes(body)
	if err != nil {
		return nil, nil, fmt.Errorf("解析登录响应: %w", err)
	}

	user, credential, err := m.wa.ValidatePasskeyLogin(handler, *session, parsed)
	if err != nil {
		return nil, nil, fmt.Errorf("验证登录凭证: %w", err)
	}

	return user, credential, nil
}

// generateSessionID 生成加密安全的随机会话 ID，编码为 base64url 格式。
func generateSessionID() (string, error) {
	buf := make([]byte, sessionIDBytes)
	if _, err := cryptorand.Read(buf); err != nil {
		return "", fmt.Errorf("生成会话ID失败: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// serializeSessionData 将 SessionData 序列化为 JSON 字节，用于存储。
func serializeSessionData(data *wa.SessionData) ([]byte, error) {
	return json.Marshal(data)
}

// deserializeSessionData 从 JSON 字节反序列化为 SessionData。
func deserializeSessionData(raw []byte) (*wa.SessionData, error) {
	var data wa.SessionData
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, err
	}
	return &data, nil
}
