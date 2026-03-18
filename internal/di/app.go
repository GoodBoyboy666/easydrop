package di

import (
	"easydrop/internal/config"
	"easydrop/internal/middleware"
	"easydrop/internal/pkg/email"
	"easydrop/internal/pkg/jwt"
	"easydrop/internal/pkg/storage"
	"easydrop/internal/service"

	red "github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// App 聚合应用运行所需的依赖。
type App struct {
	Config     *config.StaticConfig
	DB         *gorm.DB
	DBConfig   *config.DBConfig
	Redis      *red.Client
	Email      *email.Client
	JWT        *jwt.Manager
	Storage    *storage.Manager
	Middleware *middleware.Auth
	Attachment service.AttachmentService
	Auth       service.AuthService
	Comment    service.CommentService
	Post       service.PostService
	Tag        service.TagService
	User       service.UserService
}

// NewApp 构造 App 聚合对象。
func NewApp(
	cfg *config.StaticConfig,
	db *gorm.DB,
	dbConfig *config.DBConfig,
	redisClient *red.Client,
	emailClient *email.Client,
	jwtManager *jwt.Manager,
	storageManager *storage.Manager,
	middlewares *middleware.Auth,
	authService service.AuthService,
	attachmentService service.AttachmentService,
	commentService service.CommentService,
	postService service.PostService,
	tagService service.TagService,
	userService service.UserService,
) *App {
	return &App{
		Config:     cfg,
		DB:         db,
		DBConfig:   dbConfig,
		Redis:      redisClient,
		Email:      emailClient,
		JWT:        jwtManager,
		Storage:    storageManager,
		Middleware: middlewares,
		Attachment: attachmentService,
		Auth:       authService,
		Comment:    commentService,
		Post:       postService,
		Tag:        tagService,
		User:       userService,
	}
}
