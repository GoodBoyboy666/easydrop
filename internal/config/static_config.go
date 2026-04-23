package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/wire"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"go.yaml.in/yaml/v3"

	"easydrop/internal/pkg/avatar"
	"easydrop/internal/pkg/captcha"
	cookiepkg "easydrop/internal/pkg/cookie"
	"easydrop/internal/pkg/database"
	"easydrop/internal/pkg/email"
	"easydrop/internal/pkg/jwt"
	"easydrop/internal/pkg/ratelimit"
	"easydrop/internal/pkg/redis"
	"easydrop/internal/pkg/storage"
	"easydrop/internal/pkg/token"
)

// GlobalConfigDir 保存运行时生效的配置目录，供需要定位配置同级资源的模块复用。
var GlobalConfigDir string

const (
	// ServerModeDevelopment 表示开发模式。
	ServerModeDevelopment = "development"
	// ServerModeProduction 表示生产模式。
	ServerModeProduction = "production"
)

// ServerConfig 是 HTTP 服务监听配置。
type ServerConfig struct {
	Mode            string        `mapstructure:"mode" yaml:"mode"`
	Addr            string        `mapstructure:"addr" yaml:"addr"`
	TrustedProxies  []string      `mapstructure:"trusted_proxies" yaml:"trusted_proxies"`
	RemoteIPHeaders []string      `mapstructure:"remote_ip_headers" yaml:"remote_ip_headers"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout" yaml:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout" yaml:"write_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout" yaml:"shutdown_timeout"`
	CSP             CSPConfig     `mapstructure:"csp" yaml:"csp"`
}

// CSPConfig 是 Content-Security-Policy 配置。
type CSPConfig struct {
	Enabled        bool     `mapstructure:"enabled" yaml:"enabled"`
	AllowedSources []string `mapstructure:"allowed_sources" yaml:"allowed_sources"`
}

// StaticConfig 是应用的根配置结构。
type StaticConfig struct {
	Server     ServerConfig             `mapstructure:"server" yaml:"server"`
	AuthCookie cookiepkg.Config         `mapstructure:"auth_cookie" yaml:"auth_cookie"`
	DB         database.Config          `mapstructure:"db" yaml:"db"`
	Redis      redis.Config             `mapstructure:"redis" yaml:"redis"`
	RateLimit  ratelimit.Config         `mapstructure:"rate_limit" yaml:"rate_limit"`
	Email      email.Config             `mapstructure:"email" yaml:"email"`
	JWT        jwt.Config               `mapstructure:"jwt" yaml:"jwt"`
	Captcha    captcha.AllCaptchaConfig `mapstructure:"captcha" yaml:"captcha"`
	Avatar     avatar.Config            `mapstructure:"avatar" yaml:"avatar"`
	Storage    storage.Config           `mapstructure:"storage" yaml:"storage"`
	Token      token.Config             `mapstructure:"token" yaml:"token"`
}

// StaticProviderSet 提供配置加载的 Wire 注入入口。
var StaticProviderSet = wire.NewSet(Load, ProvideDBConfig, ProvideRedisConfig, ProvideRateLimitConfig, ProvideEmailConfig, ProvideJWTConfig, ProvideAuthCookieConfig, ProvideCaptchaConfig, ProvideAvatarConfig, ProvideStorageConfig, ProvideTokenConfig, ProvideCSPConfig)

func newStaticConfigViper(configDir string, enableEnv bool) *viper.Viper {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	configDir = strings.TrimSpace(configDir)
	if configDir != "" {
		v.AddConfigPath(configDir)
	}

	if enableEnv {
		v.SetEnvPrefix("EASYDROP")
		v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		v.AutomaticEnv()
	}

	setStaticConfigDefaults(v)
	return v
}

func setStaticConfigDefaults(v *viper.Viper) {
	if v == nil {
		return
	}

	// 本地开发默认值。
	v.SetDefault("server.mode", ServerModeDevelopment)
	v.SetDefault("server.addr", ":8080")
	v.SetDefault("server.trusted_proxies", []string{})
	v.SetDefault("server.remote_ip_headers", []string{"X-Forwarded-For", "X-Real-IP"})
	v.SetDefault("server.read_timeout", "10s")
	v.SetDefault("server.write_timeout", "15s")
	v.SetDefault("server.shutdown_timeout", "5s")
	v.SetDefault("server.csp.enabled", true)
	v.SetDefault("server.csp.allowed_sources", []string{})
	v.SetDefault("auth_cookie.name", "easydrop_access_token")
	v.SetDefault("auth_cookie.path", "/")
	v.SetDefault("auth_cookie.domain", "")
	v.SetDefault("auth_cookie.same_site", "lax")
	v.SetDefault("db.driver", database.DriverSQLite)
	v.SetDefault("db.sqlite_path", "data/easydrop.db")
	v.SetDefault("rate_limit.enabled", false)
	v.SetDefault("rate_limit.key_prefix", "ratelimit")
	v.SetDefault("jwt.private_key_path", "data/jwt/private.pem")
	v.SetDefault("jwt.public_key_path", "data/jwt/public.pem")
	v.SetDefault("jwt.issuer", "easydrop")
	v.SetDefault("jwt.expire", "24h")
	v.SetDefault("email.enable", false)
	v.SetDefault("email.tls_mode", email.TLSModeStartTLS)
	v.SetDefault("captcha.enabled", false)
	v.SetDefault("captcha.timeout", "5s")
	v.SetDefault("avatar.gravatar_base_url", avatar.DefaultGravatarBaseURL)
	v.SetDefault("storage.backend", storage.BackendLocal)
	v.SetDefault("storage.local.base_path", "data/uploads")
	v.SetDefault("token.key_prefix", "token")
}

func readStaticConfig(v *viper.Viper) error {
	if err := v.ReadInConfig(); err != nil {
		var notFoundErr viper.ConfigFileNotFoundError
		if !errors.As(err, &notFoundErr) {
			return err
		}
	}

	return nil
}

func unmarshalStaticConfig(v *viper.Viper) (*StaticConfig, error) {
	cfg := &StaticConfig{}
	if err := v.Unmarshal(cfg, viper.DecodeHook(mapstructure.StringToTimeDurationHookFunc())); err != nil {
		return nil, err
	}

	return cfg, nil
}

// WriteDefaultConfigFile 将默认配置写入 configDir/config.yaml。
// 若目标文件已存在，将返回包含 os.ErrExist 的错误。
func WriteDefaultConfigFile(configDir string) error {
	configDir = strings.TrimSpace(configDir)
	if configDir == "" {
		return errors.New("配置目录不能为空")
	}

	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return fmt.Errorf("创建配置目录失败 (%s): %w", configDir, err)
	}

	v := newStaticConfigViper("", false)
	content, err := yaml.Marshal(v.AllSettings())
	if err != nil {
		return fmt.Errorf("序列化默认配置失败: %w", err)
	}

	configPath := filepath.Join(configDir, "config.yaml")
	f, err := os.OpenFile(configPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return fmt.Errorf("创建配置文件失败 (%s): %w", configPath, err)
	}
	defer f.Close()

	if _, err := f.Write(content); err != nil {
		return fmt.Errorf("写入默认配置文件失败 (%s): %w", configPath, err)
	}

	return nil
}

// Load 从 configDir/config.yaml 读取配置，并支持环境变量覆盖。
// 配置文件缺失时会回退到默认值与环境变量。
func Load(configDir string, strict bool) (*StaticConfig, error) {
	configDir = strings.TrimSpace(configDir)
	GlobalConfigDir = configDir
	if configDir == "" && strict {
		return nil, errors.New("配置目录不能为空")
	}

	v := newStaticConfigViper(configDir, true)

	if err := readStaticConfig(v); err != nil {
		return nil, err
	}

	return unmarshalStaticConfig(v)
}

// ProvideDBConfig 提供数据库配置。
func ProvideDBConfig(cfg *StaticConfig) *database.Config {
	return &cfg.DB
}

// ProvideRedisConfig 提供 Redis 配置。
func ProvideRedisConfig(cfg *StaticConfig) *redis.Config {
	return &cfg.Redis
}

// ProvideRateLimitConfig 提供限流配置。
func ProvideRateLimitConfig(cfg *StaticConfig) *ratelimit.Config {
	return &cfg.RateLimit
}

// ProvideEmailConfig 提供邮件配置。
func ProvideEmailConfig(cfg *StaticConfig) *email.Config {
	return &cfg.Email
}

// ProvideJWTConfig 提供 JWT 配置。
func ProvideJWTConfig(cfg *StaticConfig) *jwt.Config {
	return &cfg.JWT
}

// ProvideAuthCookieConfig 提供认证 Cookie 配置。
func ProvideAuthCookieConfig(cfg *StaticConfig) *cookiepkg.Config {
	if cfg == nil {
		return &cookiepkg.Config{}
	}

	authCookie := cfg.AuthCookie
	authCookie.Secure = cfg.Server.Mode == ServerModeProduction
	if cfg.JWT.Expire > 0 {
		authCookie.MaxAge = cfg.JWT.Expire
	}
	return &authCookie
}

// ProvideCaptchaConfig 提供验证码配置。
func ProvideCaptchaConfig(cfg *StaticConfig) *captcha.AllCaptchaConfig {
	return &cfg.Captcha
}

// ProvideAvatarConfig 提供头像配置。
func ProvideAvatarConfig(cfg *StaticConfig) *avatar.Config {
	return &cfg.Avatar
}

// ProvideStorageConfig 提供文件存储配置。
func ProvideStorageConfig(cfg *StaticConfig) *storage.Config {
	return &cfg.Storage
}

// ProvideTokenConfig 提供 token 配置。
func ProvideTokenConfig(cfg *StaticConfig) *token.Config {
	return &cfg.Token
}

// ProvideCSPConfig 提供 CSP 配置。
func ProvideCSPConfig(cfg *StaticConfig) *CSPConfig {
	return &cfg.Server.CSP
}
