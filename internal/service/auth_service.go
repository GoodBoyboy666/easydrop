package service

import (
	"context"
	"errors"
	"log"
	"strconv"
	"strings"

	"easydrop/internal/config"
	"easydrop/internal/dto"
	"easydrop/internal/model"
	"easydrop/internal/pkg/captcha"
	"easydrop/internal/pkg/jwt"
	"easydrop/internal/pkg/validator"
	"easydrop/internal/repo"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrRegisterClosed     = errors.New("当前不允许注册")
	ErrUsernameExists     = errors.New("用户名已存在")
	ErrEmailExists        = errors.New("邮箱已存在")
	ErrEmptyAccount       = errors.New("账号不能为空")
	ErrUserNotFound       = errors.New("用户不存在")
	ErrInvalidPassword    = errors.New("密码错误")
	ErrUserDisabled       = errors.New("用户状态异常")
	ErrCaptchaRequired    = errors.New("请完成验证码")
	ErrInvalidSiteSetting = errors.New("站点配置异常")
	ErrInternal           = errors.New("服务异常，请稍后重试")
	ErrCaptchaFailed      = errors.New("验证码校验失败")
)

type AuthService interface {
	// Register 注册用户并返回登录态信息。
	Register(ctx context.Context, input dto.RegisterInput) (*dto.AuthResult, error)
	// Login 使用用户名或邮箱登录并返回登录态信息。
	Login(ctx context.Context, input dto.LoginInput) (*dto.AuthResult, error)
}

type authService struct {
	userRepo repo.UserRepo
	dbConfig config.DBConfig
	jwt      jwt.Manager
	captcha  captcha.Verifier
}

// NewAuthService 创建认证服务实例。
func NewAuthService(userRepo repo.UserRepo, dbConfig config.DBConfig, jwtManager jwt.Manager, captchaVerifier captcha.Verifier) AuthService {
	return &authService{
		userRepo: userRepo,
		dbConfig: dbConfig,
		jwt:      jwtManager,
		captcha:  captchaVerifier,
	}
}

// Register 校验注册参数、创建用户并签发访问令牌。
func (s *authService) Register(ctx context.Context, input dto.RegisterInput) (*dto.AuthResult, error) {
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

	if err := s.ensureUserUnique(ctx, username, email); err != nil {
		if errors.Is(err, ErrUsernameExists) || errors.Is(err, ErrEmailExists) {
			return nil, err
		}
		log.Printf("校验用户唯一性失败: %v", err)
		return nil, ErrInternal
	}

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

	result, err := s.buildAuthResult(user)
	if err != nil {
		log.Printf("签发令牌失败: %v", err)
		return nil, ErrInternal
	}
	return result, nil
}

// Login 校验账号密码与验证码，并在成功后签发访问令牌。
func (s *authService) Login(ctx context.Context, input dto.LoginInput) (*dto.AuthResult, error) {
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

	user, err := s.userRepo.GetByUsernameOrEmail(ctx, account)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		log.Printf("获取用户失败: %v", err)
		return nil, ErrInternal
	}

	if user.Status != 1 {
		return nil, ErrUserDisabled
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		return nil, ErrInvalidPassword
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
	if s.dbConfig == nil {
		return ErrInvalidSiteSetting
	}

	value, ok, err := s.dbConfig.GetValue(ctx, "site.allow_register")
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
