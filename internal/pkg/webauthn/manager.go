package webauthn

import (
	"context"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	wa "github.com/go-webauthn/webauthn/webauthn"
	"github.com/redis/go-redis/v9"
)

// Config 是 WebAuthn 依赖方 (Relying Party) 的配置。
type Config struct {
	// RPDisplayName 是依赖方向用户展示的名称。
	RPDisplayName string `mapstructure:"rp_display_name" yaml:"rp_display_name"`
	// RPID 是依赖方 ID，通常为不含协议和端口的域名，如 "example.com"。
	RPID string `mapstructure:"rp_id" yaml:"rp_id"`
	// RPOrigin 是允许的依赖方来源列表，需填写完整的 origin，如 "https://example.com"。
	RPOrigin []string `mapstructure:"rp_origin" yaml:"rp_origin"`
	// Timeout 是注册和登录操作的超时时间。
	Timeout time.Duration `mapstructure:"timeout" yaml:"timeout"`
}

// SessionStore 定义了 WebAuthn 会话数据的存储接口。
// 会话数据在 Begin 和 Finish 步骤之间保存，用于验证认证器响应的合法性。
// 实现必须保证会话数据的完整性和不可篡改性。
type SessionStore interface {
	// Save 保存会话数据，ttl 指定过期时间。
	Save(ctx context.Context, sessionID string, data *wa.SessionData, ttl time.Duration) error
	// Get 获取会话数据，若不存在则返回错误。
	Get(ctx context.Context, sessionID string) (*wa.SessionData, error)
	// Delete 删除会话数据（通常在校验完成后调用）。
	Delete(ctx context.Context, sessionID string) error
}

// Manager 封装了 go-webauthn 库的核心操作，并提供会话管理。
// 所有 Begin* 方法会生成会话并返回会话 ID，Finish* 方法通过会话 ID 恢复会话数据并完成验证。
type Manager interface {
	// BeginRegistration 发起凭证创建流程，返回需发送给客户端的创建选项和会话 ID。
	BeginRegistration(user wa.User) (*protocol.CredentialCreation, string, error)
	// FinishRegistration 完成凭证创建流程，通过会话 ID 恢复会话数据并验证客户端响应。
	FinishRegistration(user wa.User, sessionID string, body []byte) (*wa.Credential, error)
	// BeginDiscoverableLogin 发起无用户名登录 (discoverable credential) 流程。
	// 返回需发送给客户端的断言选项和会话 ID。
	BeginDiscoverableLogin() (*protocol.CredentialAssertion, string, error)
	// FinishDiscoverableLogin 完成无用户名登录流程，通过 handler 查找凭证对应的用户。
	FinishDiscoverableLogin(handler wa.DiscoverableUserHandler, sessionID string, body []byte) (wa.User, *wa.Credential, error)
}

// manager 是 Manager 接口的具体实现。
type manager struct {
	wa           *wa.WebAuthn
	sessionStore SessionStore
	timeout      time.Duration
}

// NewManager 创建 WebAuthn 管理器实例。
// 初始化 go-webauthn 库并配置超时时间和服务端强制超时校验。
// 会话存储根据 redisClient 自动选择 Redis（多实例共享）或内存（单进程）实现。
func NewManager(cfg *Config, redisClient *redis.Client) (Manager, error) {
	waCfg := &wa.Config{
		RPID:          cfg.RPID,
		RPDisplayName: cfg.RPDisplayName,
		RPOrigins:     cfg.RPOrigin,
		Timeouts: wa.TimeoutsConfig{
			Login: wa.TimeoutConfig{
				Enforce:    true,
				Timeout:    cfg.Timeout,
				TimeoutUVD: cfg.Timeout,
			},
			Registration: wa.TimeoutConfig{
				Enforce:    true,
				Timeout:    cfg.Timeout,
				TimeoutUVD: cfg.Timeout,
			},
		},
	}

	w, err := wa.New(waCfg)
	if err != nil {
		return nil, err
	}

	return &manager{
		wa:           w,
		sessionStore: newSessionStore(redisClient),
		timeout:      cfg.Timeout,
	}, nil
}
