package service

import (
	"context"
	"errors"
	"log"
	"strconv"
	"strings"
	"time"

	"easydrop/internal/consts"
	"easydrop/internal/dto"
	"easydrop/internal/model"
	"easydrop/internal/pkg/captcha"
	"easydrop/internal/pkg/jwt"
	"easydrop/internal/pkg/token"
	"easydrop/internal/pkg/validator"
	"easydrop/internal/repo"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrRegisterClosed       = errors.New("当前不允许注册")
	ErrUsernameExists       = errors.New("用户名已存在")
	ErrEmailExists          = errors.New("邮箱已存在")
	ErrEmailNotVerified     = errors.New("请先完成邮箱验证后再登录")
	ErrEmptyAccount         = errors.New("账号不能为空")
	ErrUserNotFound         = errors.New("用户不存在")
	ErrInvalidPassword      = errors.New("密码错误")
	ErrInvalidCredentials   = errors.New("账号或密码错误")
	ErrUserDisabled         = errors.New("用户状态异常")
	ErrCaptchaRequired      = errors.New("请完成验证码")
	ErrInvalidSiteSetting   = errors.New("站点配置异常")
	ErrInternal             = errors.New("服务异常，请稍后重试")
	ErrCaptchaFailed        = errors.New("验证码校验失败")
	ErrInvalidPasswordReset = errors.New("重置密码凭证无效或已过期")
	ErrInvalidEmailVerify   = errors.New("邮箱验证凭证无效或已过期")
)

const (
	passwordResetTokenTTL = 30 * time.Minute
	verifyEmailTokenTTL   = 24 * time.Hour
)

type AuthService interface {
	// Register 注册用户并发送邮箱验证邮件。
	Register(ctx context.Context, input dto.RegisterInput) (*dto.MessageResponse, error)
	// Login 使用用户名或邮箱登录并返回登录态信息。
	Login(ctx context.Context, input dto.LoginInput) (*dto.AuthResult, error)
	// RequestPasswordReset 发起忘记密码流程并发送重置邮件。
	RequestPasswordReset(ctx context.Context, input dto.PasswordResetRequestInput) error
	// ConfirmPasswordReset 校验重置 token 并更新密码。
	ConfirmPasswordReset(ctx context.Context, input dto.PasswordResetConfirmInput) error
	// ConfirmVerifyEmail 校验邮箱验证 token 并更新邮箱验证状态。
	ConfirmVerifyEmail(ctx context.Context, input dto.EmailVerifyConfirmInput) error
}

type authService struct {
	userRepo     repo.UserRepo
	settings     SettingService
	jwt          jwt.Manager
	captcha      captcha.Verifier
	tokenManager token.Manager
	emailService EmailService
}

// NewAuthService 创建认证服务实例。
func NewAuthService(userRepo repo.UserRepo, settings SettingService, jwtManager jwt.Manager, captchaVerifier captcha.Verifier, tokenManager token.Manager, emailService EmailService) AuthService {
	return &authService{
		userRepo:     userRepo,
		settings:     settings,
		jwt:          jwtManager,
		captcha:      captchaVerifier,
		tokenManager: tokenManager,
		emailService: emailService,
	}
}

// Register 校验注册参数、创建用户并发送邮箱验证邮件。
func (s *authService) Register(ctx context.Context, input dto.RegisterInput) (*dto.MessageResponse, error) {
	// 先完成基础输入校验与规范化。
	username := strings.TrimSpace(input.Username)
	if err := validator.ValidateUsername(username); err != nil {
		return nil, err
	}

	email := strings.TrimSpace(input.Email)
	if err := validator.ValidateEmail(email); err != nil {
		return nil, err
	}

	password := input.Password
	if err := validator.ValidatePassword(password); err != nil {
		return nil, err
	}

	nickname := strings.TrimSpace(input.Nickname)
	if nickname == "" {
		nickname = username
	}

	// 检查站点开关并校验验证码。
	if err := s.ensureRegisterEnabled(ctx); err != nil {
		if errors.Is(err, ErrRegisterClosed) || errors.Is(err, ErrInvalidSiteSetting) {
			return nil, err
		}
		log.Printf("检查注册开关失败: %v", err)
		return nil, ErrInternal
	}

	if err := s.verifyCaptcha(ctx, input.Captcha); err != nil {
		return nil, err
	}

	// 校验用户名和邮箱唯一性。
	if err := s.ensureUserUnique(ctx, username, email); err != nil {
		if errors.Is(err, ErrUsernameExists) || errors.Is(err, ErrEmailExists) {
			return nil, err
		}
		log.Printf("校验用户唯一性失败: %v", err)
		return nil, ErrInternal
	}

	// 生成密码哈希并创建用户。
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("生成密码哈希失败: %v", err)
		return nil, ErrInternal
	}

	user := &model.User{
		Username: username,
		Nickname: nickname,
		Email:    email,
		Password: string(hash),
		Status:   1,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		log.Printf("创建用户失败: %v", err)
		return nil, ErrInternal
	}

	// 异步发送验证邮件，避免阻塞注册响应。
	s.sendVerifyEmailAsync(ctx, user)

	return &dto.MessageResponse{Message: "注册成功，请先完成邮箱验证后登录"}, nil
}

// RequestPasswordReset 校验请求后发送密码重置邮件。
func (s *authService) RequestPasswordReset(ctx context.Context, input dto.PasswordResetRequestInput) error {
	// 先校验邮箱格式与验证码。
	email := strings.TrimSpace(input.Email)
	if err := validator.ValidateEmail(email); err != nil {
		return err
	}

	if err := s.verifyCaptcha(ctx, input.Captcha); err != nil {
		return err
	}

	if s.tokenManager == nil || s.emailService == nil {
		log.Printf("密码重置邮件服务未初始化")
		return nil
	}

	// 对外隐藏用户存在性：不存在或状态异常都直接返回成功。
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		log.Printf("按邮箱查询用户失败: %v", err)
		return nil
	}
	if user.Status != 1 {
		return nil
	}

	// 签发重置 token 并发送邮件。
	resetToken, err := s.tokenManager.Issue(ctx, user.ID, token.KindResetPassword, passwordResetTokenTTL, strings.TrimSpace(user.Email))
	if err != nil {
		log.Printf("签发密码重置 token 失败: %v", err)
		return nil
	}

	if err := s.emailService.SendPasswordResetEmail(ctx, user.Email, resetToken, passwordResetTokenTTL); err != nil {
		log.Printf("发送密码重置邮件失败: %v", err)
	}

	return nil
}

// ConfirmPasswordReset 校验 token 并重置用户密码，同时更新邮箱验证状态。
func (s *authService) ConfirmPasswordReset(ctx context.Context, input dto.PasswordResetConfirmInput) error {
	// 校验依赖与新密码格式。
	if s.tokenManager == nil {
		return ErrInternal
	}
	if err := validator.ValidatePassword(input.NewPassword); err != nil {
		return err
	}

	// 消费重置 token，并将 token 错误映射为统一业务错误。
	record, err := s.tokenManager.Consume(ctx, token.KindResetPassword, input.Token)
	if err != nil {
		switch {
		case errors.Is(err, token.ErrEmptyToken),
			errors.Is(err, token.ErrTokenNotFound),
			errors.Is(err, token.ErrTokenMismatch),
			errors.Is(err, token.ErrTokenExpired):
			return ErrInvalidPasswordReset
		default:
			log.Printf("消费密码重置 token 失败: %v", err)
			return ErrInternal
		}
	}

	// 校验用户与 token 载荷一致性。
	user, err := s.userRepo.GetByID(ctx, record.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		log.Printf("获取用户失败: %v", err)
		return ErrInternal
	}
	if !matchTokenEmailPayload(record.Payload, user.Email) {
		return ErrInvalidPasswordReset
	}

	// 生成新密码哈希并更新用户。
	hash, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("生成密码哈希失败: %v", err)
		return ErrInternal
	}

	user.Password = string(hash)
	user.EmailVerified = true
	if err := s.userRepo.Update(ctx, user); err != nil {
		log.Printf("更新重置密码失败: %v", err)
		return ErrInternal
	}

	return nil
}

// ConfirmVerifyEmail 校验邮箱验证 token 并更新用户邮箱验证状态。
func (s *authService) ConfirmVerifyEmail(ctx context.Context, input dto.EmailVerifyConfirmInput) error {
	// 校验依赖并消费验证 token。
	if s.tokenManager == nil {
		return ErrInternal
	}

	record, err := s.tokenManager.Consume(ctx, token.KindVerifyEmail, input.Token)
	if err != nil {
		switch {
		case errors.Is(err, token.ErrEmptyToken),
			errors.Is(err, token.ErrTokenNotFound),
			errors.Is(err, token.ErrTokenMismatch),
			errors.Is(err, token.ErrTokenExpired):
			return ErrInvalidEmailVerify
		default:
			log.Printf("消费邮箱验证 token 失败: %v", err)
			return ErrInternal
		}
	}

	// 校验用户存在且 token 载荷与当前邮箱匹配。
	user, err := s.userRepo.GetByID(ctx, record.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		log.Printf("获取用户失败: %v", err)
		return ErrInternal
	}
	if !matchTokenEmailPayload(record.Payload, user.Email) {
		return ErrInvalidEmailVerify
	}

	// 已验证用户幂等返回，否则更新验证状态。
	if user.EmailVerified {
		return nil
	}

	user.EmailVerified = true
	if err := s.userRepo.Update(ctx, user); err != nil {
		log.Printf("更新邮箱验证状态失败: %v", err)
		return ErrInternal
	}

	return nil
}

// Login 校验账号密码与验证码，并在成功后签发访问令牌。
func (s *authService) Login(ctx context.Context, input dto.LoginInput) (*dto.AuthResult, error) {
	// 基础输入校验。
	account := strings.TrimSpace(input.Account)
	if account == "" {
		return nil, ErrEmptyAccount
	}
	if strings.TrimSpace(input.Password) == "" {
		return nil, validator.ErrEmptyPassword
	}

	if err := s.verifyCaptcha(ctx, input.Captcha); err != nil {
		return nil, err
	}

	// 查询用户并统一处理凭证错误。
	user, err := s.userRepo.GetByUsernameOrEmail(ctx, account)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		log.Printf("获取用户失败: %v", err)
		return nil, ErrInternal
	}

	if user.Status != 1 {
		return nil, ErrUserDisabled
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// 根据站点配置决定是否强制邮箱已验证。
	if err := s.ensureLoginEmailVerified(ctx, user); err != nil {
		return nil, err
	}

	result, err := s.buildAuthResult(user)
	if err != nil {
		log.Printf("签发令牌失败: %v", err)
		return nil, ErrInternal
	}
	return result, nil
}

// buildAuthResult 仅签发访问令牌。
func (s *authService) buildAuthResult(user *model.User) (*dto.AuthResult, error) {
	token, err := s.jwt.IssueAccessToken(user.ID, user.Username, user.Admin)
	if err != nil {
		return nil, err
	}

	return &dto.AuthResult{
		AccessToken: token,
	}, nil
}

// ensureRegisterEnabled 检查站点是否允许新用户注册。
func (s *authService) ensureRegisterEnabled(ctx context.Context) error {
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

// ensureLoginEmailVerified 检查登录前是否需要邮箱验证。
func (s *authService) ensureLoginEmailVerified(ctx context.Context, user *model.User) error {
	if user == nil || user.EmailVerified || s.settings == nil {
		return nil
	}

	value, ok, err := s.settings.GetValue(ctx, consts.AuthRequireEmailVerificationSettingKey)
	if err != nil {
		log.Printf("读取登录邮箱验证配置失败: %v", err)
		return ErrInternal
	}
	if !ok {
		return nil
	}

	requireVerified, err := strconv.ParseBool(strings.TrimSpace(value))
	if err != nil {
		log.Printf("解析登录邮箱验证配置失败: %v", err)
		return ErrInvalidSiteSetting
	}
	if requireVerified {
		return ErrEmailNotVerified
	}

	return nil
}

// ensureUserUnique 检查用户名和邮箱是否都未被占用。
func (s *authService) ensureUserUnique(ctx context.Context, username, email string) error {
	if _, err := s.userRepo.GetByUsername(ctx, username); err == nil {
		return ErrUsernameExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("按用户名查询失败: %v", err)
		return ErrInternal
	}

	if _, err := s.userRepo.GetByEmail(ctx, email); err == nil {
		return ErrEmailExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("按邮箱查询失败: %v", err)
		return ErrInternal
	}

	return nil
}

// verifyCaptcha 在启用验证码时执行人机校验。
func (s *authService) verifyCaptcha(ctx context.Context, input *dto.CaptchaInput) error {
	if s.captcha == nil || !s.captcha.Enabled() {
		return nil
	}
	if input == nil {
		return ErrCaptchaRequired
	}

	payload := captcha.Payload{
		Token:         strings.TrimSpace(input.Token),
		RemoteIP:      strings.TrimSpace(input.RemoteIP),
		LotNumber:     strings.TrimSpace(input.LotNumber),
		CaptchaOutput: strings.TrimSpace(input.CaptchaOutput),
		PassToken:     strings.TrimSpace(input.PassToken),
		GenTime:       strings.TrimSpace(input.GenTime),
	}

	_, err := s.captcha.Verify(ctx, payload)
	if err != nil {
		log.Printf("验证码校验失败: %v", err)
		return ErrCaptchaFailed
	}
	return nil
}

// sendVerifyEmailAsync 异步签发并发送注册验证邮件。
func (s *authService) sendVerifyEmailAsync(ctx context.Context, user *model.User) {
	if s.tokenManager == nil || s.emailService == nil || user == nil || user.ID == 0 {
		return
	}

	verifyToken, err := s.tokenManager.Issue(ctx, user.ID, token.KindVerifyEmail, verifyEmailTokenTTL, strings.TrimSpace(user.Email))
	if err != nil {
		log.Printf("签发邮箱验证 token 失败: %v", err)
		return
	}

	if err := s.emailService.SendVerifyEmail(ctx, user.Email, verifyToken, verifyEmailTokenTTL); err != nil {
		log.Printf("发送注册验证邮件失败: %v", err)
	}
}

// matchTokenEmailPayload 校验 token 载荷邮箱与当前邮箱是否一致。
func matchTokenEmailPayload(payload, currentEmail string) bool {
	payload = strings.TrimSpace(payload)
	currentEmail = strings.TrimSpace(currentEmail)
	if payload == "" || currentEmail == "" {
		return false
	}
	if err := validator.ValidateEmail(payload); err != nil {
		return false
	}
	return strings.EqualFold(payload, currentEmail)
}
