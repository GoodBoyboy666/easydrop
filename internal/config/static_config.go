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

// StaticConfig 是应用的根配置结构。
type StaticConfig struct {
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
// strict 为 true 时，配置文件缺失会返回错误。
func Load(configDir string, strict bool) (*StaticConfig, error) {
	configDir = strings.TrimSpace(configDir)
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
	v.SetDefault("db.driver", database.DriverSQLite)
	v.SetDefault("db.sqlite_path", "data/easydrop.db")
	v.SetDefault("jwt.expire", time.Hour)
	v.SetDefault("email.tls_mode", email.TLSModeStartTLS)
	v.SetDefault("captcha.enabled", false)
	v.SetDefault("captcha.timeout", 5*time.Second)
	v.SetDefault("storage.backend", storage.BackendLocal)
	v.SetDefault("storage.local.base_path", "data/uploads")
	v.SetDefault("token.key_prefix", "token")

	if err := v.ReadInConfig(); err != nil {
		if strict {
			return nil, err
		}
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
