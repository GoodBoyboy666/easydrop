package config

import (
	"errors"
	"strings"
	"time"

	"github.com/google/wire"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"

	"easydrop/internal/pkg/captcha"
	"easydrop/internal/pkg/database"
	"easydrop/internal/pkg/email"
	"easydrop/internal/pkg/jwt"
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
}

// StaticConfig 是应用的根配置结构。
type StaticConfig struct {
	Server  ServerConfig             `mapstructure:"server" yaml:"server"`
	DB      database.Config          `mapstructure:"db" yaml:"db"`
	Redis   redis.Config             `mapstructure:"redis" yaml:"redis"`
	Email   email.Config             `mapstructure:"email" yaml:"email"`
	JWT     jwt.Config               `mapstructure:"jwt" yaml:"jwt"`
	Captcha captcha.AllCaptchaConfig `mapstructure:"captcha" yaml:"captcha"`
	Storage storage.Config           `mapstructure:"storage" yaml:"storage"`
	Token   token.Config             `mapstructure:"token" yaml:"token"`
}

// StaticProviderSet 提供配置加载的 Wire 注入入口。
var StaticProviderSet = wire.NewSet(Load, ProvideDBConfig, ProvideRedisConfig, ProvideEmailConfig, ProvideJWTConfig, ProvideCaptchaConfig, ProvideStorageConfig, ProvideTokenConfig)

// Load 从 configDir/config.yaml 读取配置，并支持环境变量覆盖。
// 配置文件缺失时会回退到默认值与环境变量。
func Load(configDir string, strict bool) (*StaticConfig, error) {
	configDir = strings.TrimSpace(configDir)
	GlobalConfigDir = configDir
	if configDir == "" && strict {
		return nil, errors.New("config dir is required")
	}

	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	if configDir != "" {
		v.AddConfigPath(configDir)
	}

	v.SetEnvPrefix("EASYDROP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// 本地开发默认值。
	v.SetDefault("server.mode", ServerModeDevelopment)
	v.SetDefault("server.addr", ":8080")
	v.SetDefault("server.trusted_proxies", []string{})
	v.SetDefault("server.remote_ip_headers", []string{"X-Forwarded-For", "X-Real-IP"})
	v.SetDefault("server.read_timeout", 10*time.Second)
	v.SetDefault("server.write_timeout", 15*time.Second)
	v.SetDefault("server.shutdown_timeout", 5*time.Second)
	v.SetDefault("db.driver", database.DriverSQLite)
	v.SetDefault("db.sqlite_path", "data/easydrop.db")
	v.SetDefault("jwt.private_key_path", "data/jwt/private.pem")
	v.SetDefault("jwt.public_key_path", "data/jwt/public.pem")
	v.SetDefault("jwt.issuer", "easydrop")
	v.SetDefault("jwt.expire", 24*time.Hour)
	v.SetDefault("email.enable", false)
	v.SetDefault("email.tls_mode", email.TLSModeStartTLS)
	v.SetDefault("captcha.enabled", false)
	v.SetDefault("captcha.timeout", 5*time.Second)
	v.SetDefault("storage.backend", storage.BackendLocal)
	v.SetDefault("storage.local.base_path", "data/uploads")
	v.SetDefault("token.key_prefix", "token")

	if err := v.ReadInConfig(); err != nil {
		var notFoundErr viper.ConfigFileNotFoundError
		if !errors.As(err, &notFoundErr) {
			return nil, err
		}
	}

	cfg := &StaticConfig{}
	if err := v.Unmarshal(cfg, viper.DecodeHook(mapstructure.StringToTimeDurationHookFunc())); err != nil {
		return nil, err
	}

	return cfg, nil
}

// ProvideDBConfig 提供数据库配置。
func ProvideDBConfig(cfg *StaticConfig) *database.Config {
	return &cfg.DB
}

// ProvideRedisConfig 提供 Redis 配置。
func ProvideRedisConfig(cfg *StaticConfig) *redis.Config {
	return &cfg.Redis
}

// ProvideEmailConfig 提供邮件配置。
func ProvideEmailConfig(cfg *StaticConfig) *email.Config {
	return &cfg.Email
}

// ProvideJWTConfig 提供 JWT 配置。
func ProvideJWTConfig(cfg *StaticConfig) *jwt.Config {
	return &cfg.JWT
}

// ProvideCaptchaConfig 提供验证码配置。
func ProvideCaptchaConfig(cfg *StaticConfig) *captcha.AllCaptchaConfig {
	return &cfg.Captcha
}

// ProvideStorageConfig 提供文件存储配置。
func ProvideStorageConfig(cfg *StaticConfig) *storage.Config {
	return &cfg.Storage
}

// ProvideTokenConfig 提供 token 配置。
func ProvideTokenConfig(cfg *StaticConfig) *token.Config {
	return &cfg.Token
}
