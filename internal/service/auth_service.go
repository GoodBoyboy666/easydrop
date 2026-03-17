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
	Register(ctx context.Context, input dto.RegisterInput) (*dto.AuthResult, error)
	Login(ctx context.Context, input dto.LoginInput) (*dto.AuthResult, error)
}

type authService struct {
	userRepo repo.UserRepo
	dbConfig *config.DBConfig
	jwt      *jwt.Manager
	captcha  captcha.Verifier
}

func NewAuthService(userRepo repo.UserRepo, dbConfig *config.DBConfig, jwtManager *jwt.Manager, captchaVerifier captcha.Verifier) AuthService {
	return &authService{
		userRepo: userRepo,
		dbConfig: dbConfig,
		jwt:      jwtManager,
		captcha:  captchaVerifier,
	}
}

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
		log.Printf("auth.register: ensure register enabled failed: %v", err)
		return nil, ErrInternal
	}

	if err := s.verifyCaptcha(ctx, input.Captcha); err != nil {
		return nil, err
	}

	if err := s.ensureUserUnique(ctx, username, email); err != nil {
		if errors.Is(err, ErrUsernameExists) || errors.Is(err, ErrEmailExists) {
			return nil, err
		}
		log.Printf("auth.register: ensure user unique failed: %v", err)
		return nil, ErrInternal
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("auth.register: generate password failed: %v", err)
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
		log.Printf("auth.register: create user failed: %v", err)
		return nil, ErrInternal
	}

	result, err := s.buildAuthResult(user)
	if err != nil {
		log.Printf("auth.register: issue token failed: %v", err)
		return nil, ErrInternal
	}
	return result, nil
}

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
		log.Printf("auth.login: get user failed: %v", err)
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
		log.Printf("auth.login: issue token failed: %v", err)
		return nil, ErrInternal
	}
	return result, nil
}

func (s *authService) buildAuthResult(user *model.User) (*dto.AuthResult, error) {
	token, err := s.jwt.IssueAccessToken(user.ID, user.Username, user.Admin)
	if err != nil {
		return nil, err
	}

	return &dto.AuthResult{
		User:        toUserDTO(user),
		AccessToken: token,
	}, nil
}

func (s *authService) ensureRegisterEnabled(ctx context.Context) error {
	if s.dbConfig == nil {
		return ErrInvalidSiteSetting
	}

	value, ok, err := s.dbConfig.GetValue(ctx, "site.allow_register")
	if err != nil {
		log.Printf("auth.register: read site.allow_register failed: %v", err)
		return ErrInternal
	}
	if !ok {
		return nil
	}

	allow, err := strconv.ParseBool(strings.TrimSpace(value))
	if err != nil {
		log.Printf("auth.register: parse site.allow_register failed: %v", err)
		return ErrInvalidSiteSetting
	}
	if !allow {
		return ErrRegisterClosed
	}
	return nil
}

func (s *authService) ensureUserUnique(ctx context.Context, username, email string) error {
	if _, err := s.userRepo.GetByUsername(ctx, username); err == nil {
		return ErrUsernameExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("auth.register: get by username failed: %v", err)
		return ErrInternal
	}

	if _, err := s.userRepo.GetByEmail(ctx, email); err == nil {
		return ErrEmailExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("auth.register: get by email failed: %v", err)
		return ErrInternal
	}

	return nil
}

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
		log.Printf("auth.captcha: verify failed: %v", err)
		return ErrCaptchaFailed
	}
	return nil
}

func toUserDTO(user *model.User) dto.UserDTO {
	return dto.UserDTO{
		ID:            user.ID,
		Username:      user.Username,
		Nickname:      user.Nickname,
		Email:         user.Email,
		Admin:         user.Admin,
		Status:        user.Status,
		Avatar:        user.Avatar,
		EmailVerified: user.EmailVerified,
		StorageQuota:  user.StorageQuota,
		StorageUsed:   user.StorageUsed,
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
	}
}
