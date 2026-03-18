//go:build wireinject
// +build wireinject

package di

import (
	"easydrop/internal/config"
	"easydrop/internal/middleware"
	"easydrop/internal/pkg/captcha"
	"easydrop/internal/pkg/database"
	"easydrop/internal/pkg/email"
	"easydrop/internal/pkg/jwt"
	"easydrop/internal/pkg/redis"
	"easydrop/internal/pkg/storage"
	"easydrop/internal/repo"
	"easydrop/internal/service"

	"github.com/google/wire"
)

// Initialize 通过 Wire 组装应用依赖。
func Initialize(configDir string, strict bool) (*App, error) {
	wire.Build(
		config.StaticProviderSet,
		database.NewDB,
		config.DBProviderSet,
		redis.NewClient,
		email.NewClient,
		jwt.NewManager,
		storage.NewManager,
		captcha.CaptchaSet,
		repo.RepositorySet,
		middleware.NewAuth,
		service.ServiceSet,
		NewApp,
	)
	return &App{}, nil
}
