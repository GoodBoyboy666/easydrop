package main

import (
	"easydrop/internal/config"
	"easydrop/internal/di"
	"easydrop/internal/handler"
	"easydrop/internal/middleware"
	"easydrop/internal/pkg/captcha"
	cookiepkg "easydrop/internal/pkg/cookie"
	"easydrop/internal/router"

	"github.com/gin-gonic/gin"
)

type frontendAllowAuth struct{}

func (frontendAllowAuth) OptionalLogin(c *gin.Context) {
	c.Next()
}

func (frontendAllowAuth) RequireLogin(c *gin.Context) {
	c.Next()
}

func (frontendAllowAuth) RequireAdmin(c *gin.Context) {
	c.Next()
}

func newFrontendTestApp(mode string) *di.App {
	return &di.App{
		Config: &config.StaticConfig{
			Server: config.ServerConfig{Mode: mode},
		},
		Middleware:             frontendAllowAuth{},
		SecurityHeaders:        middleware.NewSecurityHeaders(&captcha.AllCaptchaConfig{}),
		RequestBodyLimit:       middleware.NewRequestBodyLimit(nil),
		AuthHandler:            handler.NewAuthHandler(nil, nil, cookiepkg.NewAuthCookie(nil)),
		CaptchaHandler:         handler.NewCaptchaHandler(nil),
		InitHandler:            handler.NewInitHandler(nil),
		UserHandler:            handler.NewUserHandler(nil),
		UserAdminHandler:       handler.NewUserAdminHandler(nil),
		AttachmentHandler:      handler.NewAttachmentHandler(nil, nil),
		AttachmentAdminHandler: handler.NewAttachmentAdminHandler(nil),
		CommentHandler:         handler.NewCommentHandler(nil),
		CommentAdminHandler:    handler.NewCommentAdminHandler(nil),
		OverviewAdminHandler:   handler.NewOverviewAdminHandler(nil),
		PostAdminHandler:       handler.NewPostAdminHandler(nil),
		PostHandler:            handler.NewPostHandler(nil),
		SettingAdminHandler:    handler.NewSettingAdminHandler(nil),
		TagHandler:             handler.NewTagHandler(nil),
	}
}

func newFrontendTestEngine(mode string) *gin.Engine {
	engine := router.BuildEngine(newFrontendTestApp(mode))
	registerFrontendRoutes(engine)
	return engine
}
