package oauth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	goog "golang.org/x/oauth2"
	googgithub "golang.org/x/oauth2/github"
	googgoogle "golang.org/x/oauth2/google"
	googmicrosoft "golang.org/x/oauth2/microsoft"

	"github.com/golang-jwt/jwt/v5"
)

// OAuth 包层错误定义。
var (
	// ErrProviderNotConfigured 表示该社交平台未在配置文件中声明。
	ErrProviderNotConfigured = errors.New("未配置该社交登录方式")
	// ErrProviderDisabled 表示该社交平台的 client_id 或 client_secret 为空，视为未启用。
	ErrProviderDisabled = errors.New("该社交登录方式未启用")
	// ErrTokenExchangeFailed 表示授权码换取 access_token 失败。
	ErrTokenExchangeFailed = errors.New("社交登录授权失败，请重试")
	// ErrFetchUserInfoFailed 表示从社交平台获取用户信息失败。
	ErrFetchUserInfoFailed = errors.New("获取社交账户信息失败")
	// ErrEmailNotReturned 表示社交平台未返回邮箱信息，无法完成登录/注册。
	ErrEmailNotReturned = errors.New("社交账户未提供邮箱")
)

// Config 是 OAuth 总配置，对应 config.yaml 中的 oauth 段。
type Config struct {
	// FrontendRedirectURL 既是 OAuth 回调 redirect_uri，也是登录完成后的前端地址。
	// OAuth 授权完成后提供方将回跳到该地址（附带 code 与 state），
	// 前端从中提取 code/state 后 POST 到后端回调接口完成登录。
	FrontendRedirectURL string `mapstructure:"frontend_redirect_url" yaml:"frontend_redirect_url"`
	// Providers 按提供商名称索引的社交平台配置。
	Providers map[string]ProviderConfig `mapstructure:"providers" yaml:"providers"`
}

// ProviderConfig 单个社交平台的 OAuth 配置。
type ProviderConfig struct {
	// ClientID 是社交平台分配的应用 ID。
	ClientID string `mapstructure:"client_id" yaml:"client_id"`
	// ClientSecret 是社交平台分配的应用密钥。
	ClientSecret string `mapstructure:"client_secret" yaml:"client_secret"`
	// 以下三个字段仅 Apple Sign In 需要：
	// TeamID 是 Apple Developer Team ID。
	TeamID string `mapstructure:"team_id" yaml:"team_id"`
	// KeyID 是 Apple 私钥的 Key ID。
	KeyID string `mapstructure:"key_id" yaml:"key_id"`
	// PrivateKey 是 Apple 私钥文件（PEM 格式）的路径。
	PrivateKey string `mapstructure:"private_key" yaml:"private_key"`
}

// ProviderUserInfo 是从社交平台获取到的用户信息。
type ProviderUserInfo struct {
	// ProviderUserID 是社交平台中的用户唯一标识。
	ProviderUserID string
	// Email 是社交平台返回的用户邮箱。
	Email string
	// Nickname 是社交平台返回的用户昵称或显示名。
	Nickname string
}

// Manager 管理各社交平台的 oauth2.Config 实例，提供授权 URL 生成、
// Token 交换和用户信息获取等底层能力。
type Manager interface {
	// IsProviderEnabled 判断指定社交平台是否已配置并启用。
	IsProviderEnabled(provider string) bool
	// GetEnabledProviders 返回所有已启用社交平台的名称列表。
	GetEnabledProviders() []string
	// AuthCodeURL 生成社交平台的 OAuth 授权页面 URL，使用传入的 state 参数。
	// 回调地址由 ServerBaseURL 与 provider 拼接而成。
	AuthCodeURL(provider, state string) (string, error)
	// Exchange 用授权码换取 access_token。
	// 回调地址由 ServerBaseURL 与 provider 拼接而成，OAuth 提供方会校验其与授权时一致。
	Exchange(ctx context.Context, provider, code string) (*goog.Token, error)
	// FetchUserInfo 使用 access_token 获取社交平台的用户信息。
	FetchUserInfo(ctx context.Context, provider string, token *goog.Token) (*ProviderUserInfo, error)
}

// manager 是 Manager 接口的实现。
type manager struct {
	config *Config
}

// NewManager 创建 OAuth Manager 实例。
func NewManager(cfg *Config) Manager {
	return &manager{config: cfg}
}

// IsProviderEnabled 判断指定社交平台是否已启用。
// 要求同时配置了 client_id 和 client_secret 才算已启用。
func (m *manager) IsProviderEnabled(provider string) bool {
	if m == nil || m.config == nil {
		return false
	}
	pc, ok := m.config.Providers[provider]
	if !ok {
		return false
	}
	return pc.ClientID != "" && pc.ClientSecret != ""
}

// GetEnabledProviders 返回所有已启用社交平台的名称列表。
func (m *manager) GetEnabledProviders() []string {
	if m == nil || m.config == nil {
		return nil
	}
	var enabled []string
	for name, pc := range m.config.Providers {
		if pc.ClientID != "" && pc.ClientSecret != "" {
			enabled = append(enabled, name)
		}
	}
	return enabled
}

// AuthCodeURL 生成社交平台的 OAuth 授权页面 URL。
// state 参数由调用方生成，用于 CSRF 防护。
// 回调地址由 ServerBaseURL 与 provider 拼接而成。
func (m *manager) AuthCodeURL(provider, state string) (string, error) {
	if !m.IsProviderEnabled(provider) {
		return "", ErrProviderDisabled
	}
	cfg, err := m.oauth2Config(provider)
	if err != nil {
		return "", err
	}
	return cfg.AuthCodeURL(state), nil
}

// Exchange 用授权码换取 access_token。
func (m *manager) Exchange(ctx context.Context, provider, code string) (*goog.Token, error) {
	cfg, err := m.oauth2Config(provider)
	if err != nil {
		return nil, err
	}
	tok, err := cfg.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrTokenExchangeFailed, err)
	}
	return tok, nil
}

// FetchUserInfo 使用 access_token 获取社交平台的用户信息。
// 根据 provider 名称自动路由到对应的平台 API。
func (m *manager) FetchUserInfo(ctx context.Context, provider string, token *goog.Token) (*ProviderUserInfo, error) {
	switch provider {
	case "google":
		return m.fetchGoogleUser(ctx, token)
	case "github":
		return m.fetchGitHubUser(ctx, token)
	case "twitter":
		return m.fetchTwitterUser(ctx, token)
	case "microsoft":
		return m.fetchMicrosoftUser(ctx, token)
	case "apple":
		return m.fetchAppleUser(token)
	default:
		return nil, ErrProviderDisabled
	}
}

// redirectURL 返回 OAuth 回调地址，格式为 FrontendRedirectURL/oauth/provider。
// 前端从中可提取 provider，再 POST 到 /api/v1/auth/oauth/{provider}/callback。
func (m *manager) redirectURL(provider string) string {
	base := m.config.FrontendRedirectURL
	if base == "" {
		base = "http://localhost:3000"
	}
	return base + "/oauth/" + provider
}

// oauth2Config 根据 provider 名称构建对应的 oauth2.Config。
// 回调地址由 ServerBaseURL 与 provider 拼接而成。
// Apple 的 client_secret 需要动态生成 JWT，其他平台直接使用配置中的值。
func (m *manager) oauth2Config(provider string) (*goog.Config, error) {
	pc, ok := m.config.Providers[provider]
	if !ok {
		return nil, ErrProviderNotConfigured
	}

	cfg := &goog.Config{
		ClientID:     pc.ClientID,
		ClientSecret: pc.ClientSecret,
		RedirectURL:  m.redirectURL(provider),
	}

	switch provider {
	case "google":
		cfg.Endpoint = googgoogle.Endpoint
		cfg.Scopes = []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"}
	case "github":
		cfg.Endpoint = googgithub.Endpoint
		cfg.Scopes = []string{"user:email"}
	case "twitter":
		cfg.Endpoint = goog.Endpoint{
			AuthURL:   "https://twitter.com/i/oauth2/authorize",
			TokenURL:  "https://api.twitter.com/2/oauth2/token",
			AuthStyle: goog.AuthStyleInHeader,
		}
		cfg.Scopes = []string{"tweet.read", "users.read", "offline.access"}
	case "microsoft":
		cfg.Endpoint = googmicrosoft.AzureADEndpoint("common")
		cfg.Scopes = []string{"https://graph.microsoft.com/User.Read"}
	case "apple":
		cfg.Endpoint = goog.Endpoint{
			AuthURL:   "https://appleid.apple.com/auth/authorize",
			TokenURL:  "https://appleid.apple.com/auth/token",
			AuthStyle: goog.AuthStyleInHeader,
		}
		cfg.Scopes = []string{"name", "email"}
		// Apple 的 client_secret 需要每次动态生成 JWT。
		secret, err := m.generateAppleClientSecret(pc)
		if err != nil {
			return nil, fmt.Errorf("生成 Apple client_secret 失败: %w", err)
		}
		cfg.ClientSecret = secret
	default:
		return nil, ErrProviderNotConfigured
	}

	return cfg, nil
}

// fetchGoogleUser 通过 Google userinfo 端点获取用户信息。
func (m *manager) fetchGoogleUser(ctx context.Context, token *goog.Token) (*ProviderUserInfo, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFetchUserInfoFailed, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var result struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFetchUserInfoFailed, err)
	}
	if result.Email == "" {
		return nil, ErrEmailNotReturned
	}
	return &ProviderUserInfo{
		ProviderUserID: result.ID,
		Email:          result.Email,
		Nickname:       result.Name,
	}, nil
}

// fetchGitHubUser 通过 GitHub API 获取用户信息。
// 优先使用公开邮箱，若为空则查 emails 列表取 primary verified 邮箱。
func (m *manager) fetchGitHubUser(ctx context.Context, token *goog.Token) (*ProviderUserInfo, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user", nil)
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFetchUserInfoFailed, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var userResult struct {
		ID    int64  `json:"id"`
		Login string `json:"login"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.Unmarshal(body, &userResult); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFetchUserInfoFailed, err)
	}
	providerUserID := fmt.Sprintf("%d", userResult.ID)
	email := userResult.Email
	nickname := userResult.Name
	if nickname == "" {
		nickname = userResult.Login
	}

	// 主接口未返回邮箱时，通过 emails 接口获取。
	if email == "" {
		email, _ = m.fetchGitHubPrimaryEmail(ctx, token)
	}
	if email == "" {
		return nil, ErrEmailNotReturned
	}

	return &ProviderUserInfo{
		ProviderUserID: providerUserID,
		Email:          email,
		Nickname:       nickname,
	}, nil
}

// fetchGitHubPrimaryEmail 通过 GitHub emails 接口获取已验证的主邮箱。
func (m *manager) fetchGitHubPrimaryEmail(ctx context.Context, token *goog.Token) (string, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user/emails", nil)
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	if err := json.Unmarshal(body, &emails); err != nil {
		return "", err
	}
	// 优先取 primary & verified，其次取任意 verified。
	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email, nil
		}
	}
	for _, e := range emails {
		if e.Verified {
			return e.Email, nil
		}
	}
	return "", nil
}

// fetchTwitterUser 通过 Twitter API v2 获取用户信息。
// 注意：Twitter OAuth 2.0 不直接返回邮箱，此处用 username@twitter.com 作为占位。
func (m *manager) fetchTwitterUser(ctx context.Context, token *goog.Token) (*ProviderUserInfo, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.twitter.com/2/users/me?user.fields=profile_image_url", nil)
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFetchUserInfoFailed, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var result struct {
		Data struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			Username string `json:"username"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFetchUserInfoFailed, err)
	}
	if result.Data.ID == "" {
		return nil, ErrFetchUserInfoFailed
	}

	email := result.Data.Username + "@twitter.com"
	return &ProviderUserInfo{
		ProviderUserID: result.Data.ID,
		Email:          email,
		Nickname:       result.Data.Name,
	}, nil
}

// fetchMicrosoftUser 通过 Microsoft Graph API 获取用户信息。
func (m *manager) fetchMicrosoftUser(ctx context.Context, token *goog.Token) (*ProviderUserInfo, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://graph.microsoft.com/v1.0/me", nil)
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFetchUserInfoFailed, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var result struct {
		ID                string `json:"id"`
		UserPrincipalName string `json:"userPrincipalName"`
		Mail              string `json:"mail"`
		DisplayName       string `json:"displayName"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFetchUserInfoFailed, err)
	}
	// 优先使用 Mail，其次使用 UserPrincipalName。
	email := result.Mail
	if email == "" {
		email = result.UserPrincipalName
	}
	if email == "" {
		return nil, ErrEmailNotReturned
	}
	return &ProviderUserInfo{
		ProviderUserID: result.ID,
		Email:          email,
		Nickname:       result.DisplayName,
	}, nil
}

// fetchAppleUser 从 Apple 返回的 id_token (JWT) 中解析用户信息。
// Apple 仅在用户首次授权时返回 email，后续授权可能不返回。
func (m *manager) fetchAppleUser(token *goog.Token) (*ProviderUserInfo, error) {
	idToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, fmt.Errorf("%w: 缺少 id_token", ErrFetchUserInfoFailed)
	}
	claims, _, err := new(jwt.Parser).ParseUnverified(idToken, jwt.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFetchUserInfoFailed, err)
	}
	mapClaims, ok := claims.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrFetchUserInfoFailed
	}
	sub, _ := mapClaims["sub"].(string)
	if sub == "" {
		return nil, ErrFetchUserInfoFailed
	}
	email, _ := mapClaims["email"].(string)
	name := ""
	// 首次授权时，Apple 还会在 token 的 user 字段中以 JSON 返回姓名。
	if email == "" {
		if userJSON, ok := token.Extra("user").(string); ok {
			var userData struct {
				Name struct {
					FirstName string `json:"firstName"`
					LastName  string `json:"lastName"`
				} `json:"name"`
				Email string `json:"email"`
			}
			if err := json.Unmarshal([]byte(userJSON), &userData); err == nil {
				email = userData.Email
				name = userData.Name.FirstName + " " + userData.Name.LastName
			}
		}
	}

	return &ProviderUserInfo{
		ProviderUserID: sub,
		Email:          email,
		Nickname:       name,
	}, nil
}

// generateAppleClientSecret 为 Apple Sign In 生成 client_secret (JWT)。
// Apple 要求 client_secret 是一个 ES256 签名的 JWT，有效期 10 分钟。
func (m *manager) generateAppleClientSecret(pc ProviderConfig) (string, error) {
	keyPath := pc.PrivateKey
	if keyPath == "" {
		return "", errors.New("Apple private_key 路径未配置")
	}
	keyBytes, err := os.ReadFile(keyPath)
	if err != nil {
		return "", fmt.Errorf("读取 Apple 私钥失败: %w", err)
	}
	key, err := jwt.ParseECPrivateKeyFromPEM(keyBytes)
	if err != nil {
		return "", fmt.Errorf("解析 Apple 私钥失败: %w", err)
	}

	now := time.Now()
	claims := jwt.MapClaims{
		"iss": pc.TeamID,
		"iat": now.Unix(),
		"exp": now.Add(10 * time.Minute).Unix(),
		"aud": "https://appleid.apple.com",
		"sub": pc.ClientID,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = pc.KeyID
	return token.SignedString(key)
}

// generateState 生成随机 state 参数，长度为 32 位十六进制字符串。
// 用于 OAuth 流程中防止 CSRF 攻击。
func generateState() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	hash := sha256.Sum256(b)
	return hex.EncodeToString(hash[:16])
}

// GenerateState 是 generateState 的公开版本，供 handler 层调用。
func GenerateState() string {
	return generateState()
}
