package di

import (
	"easydrop/internal/config"
	"easydrop/internal/handler"
	"easydrop/internal/middleware"
)

// App 聚合应用运行所需的依赖。
type App struct {
	Config                 *config.StaticConfig
	Middleware             middleware.Auth
	AuthHandler            *handler.AuthHandler
	UserHandler            *handler.UserHandler
	UserAdminHandler       *handler.UserAdminHandler
	AttachmentHandler      *handler.AttachmentHandler
	AttachmentAdminHandler *handler.AttachmentAdminHandler
	CommentHandler         *handler.CommentHandler
	CommentAdminHandler    *handler.CommentAdminHandler
	PostAdminHandler       *handler.PostAdminHandler
	SettingAdminHandler    *handler.SettingAdminHandler
}

// NewApp 构造 App 聚合对象。
func NewApp(
	cfg *config.StaticConfig,
	middlewares middleware.Auth,
	authHandler *handler.AuthHandler,
	userHandler *handler.UserHandler,
	userAdminHandler *handler.UserAdminHandler,
	attachmentHandler *handler.AttachmentHandler,
	attachmentAdminHandler *handler.AttachmentAdminHandler,
	commentHandler *handler.CommentHandler,
	commentAdminHandler *handler.CommentAdminHandler,
	postAdminHandler *handler.PostAdminHandler,
	settingAdminHandler *handler.SettingAdminHandler,

) *App {
	return &App{
		Config:                 cfg,
		Middleware:             middlewares,
		AuthHandler:            authHandler,
		UserHandler:            userHandler,
		UserAdminHandler:       userAdminHandler,
		AttachmentHandler:      attachmentHandler,
		AttachmentAdminHandler: attachmentAdminHandler,
		CommentHandler:         commentHandler,
		CommentAdminHandler:    commentAdminHandler,
		PostAdminHandler:       postAdminHandler,
		SettingAdminHandler:    settingAdminHandler,
	}
}
