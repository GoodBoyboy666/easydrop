package di

import (
	"easydrop/internal/config"
	"easydrop/internal/handler"
	"easydrop/internal/middleware"
	"easydrop/internal/pkg/initsecret"
	"easydrop/internal/service"
)

// App 聚合应用运行所需的依赖。
type App struct {
	Config                 *config.StaticConfig
	InitService            service.InitService
	InitSecretGuard        initsecret.Guard
	Middleware             middleware.Auth
	CSRF                   middleware.CSRF
	SecurityHeaders        middleware.SecurityHeaders
	RateLimit              middleware.RateLimit
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
	OverviewAdminHandler   *handler.OverviewAdminHandler
	PostAdminHandler       *handler.PostAdminHandler
	PostHandler            *handler.PostHandler
	FeedHandler            *handler.FeedHandler
	SettingAdminHandler    *handler.SettingAdminHandler
	TagHandler             *handler.TagHandler
	PasskeyHandler         *handler.PasskeyHandler
	OAuthHandler           *handler.OAuthHandler
}

// NewApp 构造 App 聚合对象。
func NewApp(
	cfg *config.StaticConfig,
	initService service.InitService,
	initSecretGuard initsecret.Guard,
	authMiddleware middleware.Auth,
	csrf middleware.CSRF,
	securityHeaders middleware.SecurityHeaders,
	rateLimit middleware.RateLimit,
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
	overviewAdminHandler *handler.OverviewAdminHandler,
	postAdminHandler *handler.PostAdminHandler,
	postHandler *handler.PostHandler,
	feedHandler *handler.FeedHandler,
	settingAdminHandler *handler.SettingAdminHandler,
	tagHandler *handler.TagHandler,
	passkeyHandler *handler.PasskeyHandler,
	oauthHandler *handler.OAuthHandler,

) *App {
	return &App{
		Config:                 cfg,
		InitService:            initService,
		InitSecretGuard:        initSecretGuard,
		Middleware:             authMiddleware,
		CSRF:                   csrf,
		SecurityHeaders:        securityHeaders,
		RateLimit:              rateLimit,
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
		OverviewAdminHandler:   overviewAdminHandler,
		PostAdminHandler:       postAdminHandler,
		PostHandler:            postHandler,
		FeedHandler:            feedHandler,
		SettingAdminHandler:    settingAdminHandler,
		TagHandler:             tagHandler,
		PasskeyHandler:         passkeyHandler,
		OAuthHandler:           oauthHandler,
	}
}
