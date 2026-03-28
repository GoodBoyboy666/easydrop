//go:build wireinject
// +build wireinject

package di

import (
	"easydrop/internal/config"
	"easydrop/internal/handler"
	"easydrop/internal/middleware"
	"easydrop/internal/repo"
	"easydrop/internal/service"
	"easydrop/internal/pkg"

	"github.com/google/wire"
)

// Initialize 通过 Wire 组装应用依赖。
func Initialize(configDir string, strict bool) (*App, error) {
	wire.Build(
		config.StaticProviderSet,
		pkg.Pkgset,
		repo.RepositorySet,
		service.ServiceSet,
		middleware.MiddlewareSet,
		handler.HandlerSet,
		NewApp,
	)
	return &App{}, nil
}
