package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"easydrop/internal/consts"
	"easydrop/internal/dto"
	"easydrop/internal/model"
	"easydrop/internal/pkg/jwt"
	"easydrop/internal/pkg/oauth"
	"easydrop/internal/repo"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// OAuth 服务层错误定义。
var (
	// ErrOAuthNotConfigured 表示该社交登录方式未在配置中启用。
	ErrOAuthNotConfigured = errors.New("该社交登录方式未配置")
	// ErrOAuthStateMismatch 表示 OAuth state 参数校验失败，可能存在 CSRF 攻击。
	ErrOAuthStateMismatch = errors.New("社交登录状态校验失败，请重新尝试")
	// ErrOAuthEmailExistsUnbound 表示社交账户邮箱已注册但未绑定任何社交平台，
	// 用户需先使用密码登录，再到设置中手动绑定。
	ErrOAuthEmailExistsUnbound = errors.New("该邮箱已被注册但未绑定此社交账号，请先使用密码登录后在设置中手动绑定")
	// ErrOAuthBindAlreadyExists 表示该社交平台账户已绑定到其他本地用户。
	ErrOAuthBindAlreadyExists = errors.New("该社交账号已绑定其他用户")
	// ErrOAuthProviderBindAlreadyExists 表示当前用户已绑定该社交平台，无法重复绑定。
	ErrOAuthProviderBindAlreadyExists = errors.New("该社交平台已绑定")
	// ErrOAuthBindNotFound 表示未找到指定的绑定记录。
	ErrOAuthBindNotFound = errors.New("未找到该社交账号绑定")
	// ErrOAuthBindFailed 表示社交账号绑定过程中发生未知错误。
	ErrOAuthBindFailed = errors.New("社交账号绑定失败")
)

// OAuthService 定义社交登录业务逻辑接口。
type OAuthService interface {
	// GetAuthURL 生成社交平台的 OAuth 授权页面 URL。
	GetAuthURL(ctx context.Context, provider, state string) (string, error)
	// HandleCallback 处理社交平台回调：
	//   - 若第三方账户已绑定本地用户 → 直接签发 JWT 登录。
	//   - 若第三方账户邮箱不存在于数据库 → 静默注册用户并完成绑定。
	//   - 若邮箱已存在但未绑定 → 返回 ErrOAuthEmailExistsUnbound 要求手动绑定。
	HandleCallback(ctx context.Context, provider, code, stateFromQuery, stateFromCookie string) (*dto.AuthResult, error)
	// GetEnabledProviders 返回所有已启用且已配置的社交登录方式列表。
	GetEnabledProviders() []dto.OAuthProviderItem
	// GetUserBindings 查询当前用户已绑定的所有社交账号。
	GetUserBindings(ctx context.Context, userID uint) ([]dto.OAuthBindDTO, error)
	// Unbind 解除当前用户的一条社交账号绑定。
	Unbind(ctx context.Context, userID uint, bindID uint) error
	// BindManually 已登录用户主动绑定一个社交平台账户。
	BindManually(ctx context.Context, userID uint, provider, code, stateFromQuery, stateFromCookie string) error
}

// oauthService 是 OAuthService 的实现。
type oauthService struct {
	oauthManager  oauth.Manager
	oauthBindRepo repo.OAuthBindRepo
	userRepo      repo.UserRepo
	jwtManager    jwt.Manager
	settings      SettingService
}

// NewOAuthService 创建社交登录服务实例。
func NewOAuthService(
	oauthManager oauth.Manager,
	oauthBindRepo repo.OAuthBindRepo,
	userRepo repo.UserRepo,
	jwtManager jwt.Manager,
	settings SettingService,
) OAuthService {
	return &oauthService{
		oauthManager:  oauthManager,
		oauthBindRepo: oauthBindRepo,
		userRepo:      userRepo,
		jwtManager:    jwtManager,
		settings:      settings,
	}
}

// GetAuthURL 生成社交平台的 OAuth 授权页面 URL。
func (s *oauthService) GetAuthURL(ctx context.Context, provider, state string) (string, error) {
	return s.oauthManager.AuthCodeURL(provider, state)
}

// GetEnabledProviders 返回所有已启用且已配置的社交登录方式列表。
// 每一项包含 provider 标识和对应的前端授权跳转路径。
func (s *oauthService) GetEnabledProviders() []dto.OAuthProviderItem {
	providers := s.oauthManager.GetEnabledProviders()
	items := make([]dto.OAuthProviderItem, 0, len(providers))
	for _, p := range providers {
		items = append(items, dto.OAuthProviderItem{
			Provider: p,
			AuthURL:  "/api/v1/auth/oauth/" + p,
		})
	}
	return items
}

// HandleCallback 处理社交平台 OAuth 回调。
//
// 处理流程：
//  1. 校验 state 参数防止 CSRF 攻击。
//  2. 用授权码换取 access_token。
//  3. 获取社交平台用户信息（邮箱、昵称、平台唯一ID）。
//  4. 查找 (provider, providerUserID) 绑定：
//     - 已绑定 → 校验用户状态 → 签发 JWT。
//     - 未绑定 → 按邮箱查找用户：
//     - 邮箱不存在 → 静默注册用户（自动生成用户名、随机密码）→ 创建绑定 → 签发 JWT。
//     - 邮箱已存在 → 返回 ErrOAuthEmailExistsUnbound，提示需手动绑定。
func (s *oauthService) HandleCallback(ctx context.Context, provider, code, stateFromQuery, stateFromCookie string) (*dto.AuthResult, error) {
	// 校验社交登录方式是否启用。
	if !s.oauthManager.IsProviderEnabled(provider) {
		return nil, ErrOAuthNotConfigured
	}
	// 校验 state 参数防止 CSRF。
	if stateFromQuery == "" || stateFromQuery != stateFromCookie {
		return nil, ErrOAuthStateMismatch
	}

	// 用授权码换取 access_token。
	token, err := s.oauthManager.Exchange(ctx, provider, code)
	if err != nil {
		return nil, fmt.Errorf("社交登录授权失败: %w", err)
	}

	// 获取社交平台用户信息。
	info, err := s.oauthManager.FetchUserInfo(ctx, provider, token)
	if err != nil {
		return nil, err
	}
	if info.Email == "" {
		return nil, oauth.ErrEmailNotReturned
	}

	// 情况一：已绑定 → 直接登录。
	bind, err := s.oauthBindRepo.FindByProviderAndUID(ctx, provider, info.ProviderUserID)
	if err == nil {
		user, err := s.userRepo.GetByID(ctx, bind.UserID)
		if err != nil {
			log.Printf("查询已绑定用户失败: %v", err)
			return nil, ErrInternal
		}
		if user.Status != 1 {
			return nil, ErrUserDisabled
		}
		return s.buildAuthResult(user)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("查询OAuth绑定失败: %v", err)
		return nil, ErrInternal
	}

	// 情况二：邮箱已存在但未绑定 → 拒绝登录，要求手动绑定。
	_, err = s.userRepo.GetByEmail(ctx, info.Email)
	if err == nil {
		return nil, ErrOAuthEmailExistsUnbound
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("按邮箱查询用户失败: %v", err)
		return nil, ErrInternal
	}

	// 情况三：新用户 → 静默注册并绑定。
	// 检查站点是否允许注册。

	if err := s.ensureRegisterEnabled(ctx); err != nil {
		return nil, err
	}

	username := generateOAuthUsername(provider)
	nickname := info.Nickname
	if nickname == "" {
		nickname = username
	}
	if len(nickname) > 100 {
		nickname = nickname[:100]
	}

	// 为 OAuth 用户生成随机密码，用户可通过"忘记密码"设置真实密码。
	randomPass := generateRandomString(32)
	hash, err := bcrypt.GenerateFromPassword([]byte(randomPass), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("生成密码哈希失败: %v", err)
		return nil, ErrInternal
	}

	// 创建本地用户，OAuth 注册的邮箱视为已验证。
	user := &model.User{
		Username:      username,
		Nickname:      nickname,
		Email:         info.Email,
		Password:      string(hash),
		Status:        1,
		EmailVerified: true,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		log.Printf("创建OAuth用户失败: %v", err)
		return nil, ErrInternal
	}

	// 创建社交账号绑定记录。
	newBind := &model.OAuthBind{
		UserID:         user.ID,
		Provider:       provider,
		ProviderUserID: info.ProviderUserID,
		ProviderEmail:  info.Email,
	}
	if err := s.oauthBindRepo.Create(ctx, newBind); err != nil {
		log.Printf("创建OAuth绑定失败: %v", err)
		return nil, ErrInternal
	}

	return s.buildAuthResult(user)
}

// GetUserBindings 查询当前用户已绑定的所有社交账号。
func (s *oauthService) GetUserBindings(ctx context.Context, userID uint) ([]dto.OAuthBindDTO, error) {
	binds, err := s.oauthBindRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	result := make([]dto.OAuthBindDTO, 0, len(binds))
	for _, b := range binds {
		result = append(result, dto.OAuthBindDTO{
			ID:            b.ID,
			Provider:      b.Provider,
			ProviderEmail: b.ProviderEmail,
		})
	}
	return result, nil
}

// Unbind 解除当前用户的一条社交账号绑定。
func (s *oauthService) Unbind(ctx context.Context, userID uint, bindID uint) error {
	binds, err := s.oauthBindRepo.FindByUserID(ctx, userID)
	if err != nil {
		return err
	}
	for _, b := range binds {
		if b.ID == bindID {
			return s.oauthBindRepo.Delete(ctx, bindID)
		}
	}
	return ErrOAuthBindNotFound
}

// BindManually 已登录用户主动绑定一个社交平台账户。
//
// 处理流程：
//  1. 校验 state 参数。
//  2. 用授权码换取 access_token。
//  3. 获取社交平台用户信息。
//  4. 校验该社交账户未被其他用户绑定。
//  5. 校验当前用户未绑定该社交平台。
//  6. 创建绑定记录。
func (s *oauthService) BindManually(ctx context.Context, userID uint, provider, code, stateFromQuery, stateFromCookie string) error {
	if !s.oauthManager.IsProviderEnabled(provider) {
		return ErrOAuthNotConfigured
	}
	if stateFromQuery == "" || stateFromQuery != stateFromCookie {
		return ErrOAuthStateMismatch
	}

	token, err := s.oauthManager.Exchange(ctx, provider, code)
	if err != nil {
		return fmt.Errorf("社交登录授权失败: %w", err)
	}

	info, err := s.oauthManager.FetchUserInfo(ctx, provider, token)
	if err != nil {
		return err
	}

	// 校验该社交账户未被其他用户绑定。
	if _, err := s.oauthBindRepo.FindByProviderAndUID(ctx, provider, info.ProviderUserID); err == nil {
		return ErrOAuthBindAlreadyExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("检查社交账户绑定关系失败")
	}

	// 校验当前用户未绑定该社交平台。
	if _, err := s.oauthBindRepo.FindByUserIDAndProvider(ctx, userID, provider); err == nil {
		return ErrOAuthProviderBindAlreadyExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("检查用户社交平台绑定状态失败")
	}

	bind := &model.OAuthBind{
		UserID:         userID,
		Provider:       provider,
		ProviderUserID: info.ProviderUserID,
		ProviderEmail:  info.Email,
	}
	if err := s.oauthBindRepo.Create(ctx, bind); err != nil {
		return ErrOAuthBindFailed
	}
	return nil
}

// buildAuthResult 为用户签发 JWT 访问令牌。
func (s *oauthService) buildAuthResult(user *model.User) (*dto.AuthResult, error) {
	token, err := s.jwtManager.IssueAccessToken(user.ID, user.Username, user.Admin)
	if err != nil {
		return nil, ErrInternal
	}
	return &dto.AuthResult{AccessToken: token}, nil
}

// ensureRegisterEnabled 检查站点是否允许新用户注册。
func (s *oauthService) ensureRegisterEnabled(ctx context.Context) error {
	if s.settings == nil {
		return ErrInvalidSiteSetting
	}

	value, ok, err := s.settings.GetValue(ctx, consts.SiteAllowRegisterSettingKey)
	if err != nil {
		log.Printf("读取注册配置失败: %v", err)
		return ErrInternal
	}
	if !ok {
		return nil
	}

	allow, err := strconv.ParseBool(strings.TrimSpace(value))
	if err != nil {
		log.Printf("解析注册配置失败: %v", err)
		return ErrInvalidSiteSetting
	}
	if !allow {
		return ErrRegisterClosed
	}
	return nil
}

// generateOAuthUsername 为 OAuth 注册用户生成唯一用户名，格式为 provider_xxxxxxxx。
func generateOAuthUsername(provider string) string {
	return strings.ToLower(provider) + "_" + generateRandomString(8)
}

// generateRandomString 生成指定长度的随机十六进制字符串。
func generateRandomString(length int) string {
	b := make([]byte, length)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)[:length]
}
