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
	RequestBodyLimit       middleware.RequestBodyLimit
	AuthHandler            *handler.AuthHandler
	CaptchaHandler         *handler.CaptchaHandler
	InitHandler            *handler.InitHandler
	UserHandler            *handler.UserHandler
	UserAdminHandler       *handler.UserAdminHandler
	AttachmentHandler      *handler.AttachmentHandler
	AttachmentAdminHandler *handler.AttachmentAdminHandler
	CommentHandler         *handler.CommentHandler
	CommentAdminHandler    *handler.CommentAdminHandler
	PostAdminHandler       *handler.PostAdminHandler
	PostHandler            *handler.PostHandler
	SettingAdminHandler    *handler.SettingAdminHandler
	TagHandler             *handler.TagHandler
}

// NewApp 构造 App 聚合对象。
func NewApp(
	cfg *config.StaticConfig,
	authMiddleware middleware.Auth,
	requestBodyLimit middleware.RequestBodyLimit,
	authHandler *handler.AuthHandler,
	captchaHandler *handler.CaptchaHandler,
	initHandler *handler.InitHandler,
	userHandler *handler.UserHandler,
	userAdminHandler *handler.UserAdminHandler,
	attachmentHandler *handler.AttachmentHandler,
	attachmentAdminHandler *handler.AttachmentAdminHandler,
	commentHandler *handler.CommentHandler,
	commentAdminHandler *handler.CommentAdminHandler,
	postAdminHandler *handler.PostAdminHandler,
	postHandler *handler.PostHandler,
	settingAdminHandler *handler.SettingAdminHandler,
	tagHandler *handler.TagHandler,

) *App {
	return &App{
		Config:                 cfg,
		Middleware:             authMiddleware,
		RequestBodyLimit:       requestBodyLimit,
		AuthHandler:            authHandler,
		CaptchaHandler:         captchaHandler,
		InitHandler:            initHandler,
		UserHandler:            userHandler,
		UserAdminHandler:       userAdminHandler,
		AttachmentHandler:      attachmentHandler,
		AttachmentAdminHandler: attachmentAdminHandler,
		CommentHandler:         commentHandler,
		CommentAdminHandler:    commentAdminHandler,
		PostAdminHandler:       postAdminHandler,
		PostHandler:            postHandler,
		SettingAdminHandler:    settingAdminHandler,
		TagHandler:             tagHandler,
	}
}
