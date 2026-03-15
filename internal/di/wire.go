//go:build wireinject
// +build wireinject

package di

import (
	"easydrop/internal/config"
	"easydrop/internal/pkg/database"
	"easydrop/internal/pkg/email"
	"easydrop/internal/pkg/jwt"
	"easydrop/internal/pkg/redis"

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
		NewApp,
	)
	return &App{}, nil
}
