package di

import (
	"easydrop/internal/config"
	emailpkg "easydrop/internal/pkg/email"
	jwtpkg "easydrop/internal/pkg/jwt"

	red "github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// App 聚合应用运行所需的依赖。
type App struct {
	Config   *config.StaticConfig
	DB       *gorm.DB
	DBConfig *config.DBConfig
	Redis    *red.Client
	Email    *emailpkg.Client
	JWT      *jwtpkg.Manager
}

// NewApp 构造 App 聚合对象。
func NewApp(
	cfg *config.StaticConfig,
	db *gorm.DB,
	dbConfig *config.DBConfig,
	redisClient *red.Client,
	emailClient *emailpkg.Client,
	jwtManager *jwtpkg.Manager,
) *App {
	return &App{
		Config:   cfg,
		DB:       db,
		DBConfig: dbConfig,
		Redis:    redisClient,
		Email:    emailClient,
		JWT:      jwtManager,
	}
}
