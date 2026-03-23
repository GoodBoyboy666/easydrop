package router

import (
	"log"
	"net/http"

	_ "easydrop/docs"
	"easydrop/internal/config"
	"easydrop/internal/di"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// BuildEngine 构建并返回应用的 HTTP 路由引擎。
func BuildEngine(app *di.App) *gin.Engine {
	if app != nil && app.Config != nil {
		if app.Config.Server.Mode == config.ServerModeProduction {
			gin.SetMode(gin.ReleaseMode)
		} else {
			gin.SetMode(gin.DebugMode)
		}
	}

	r := gin.New()
	if err := r.SetTrustedProxies(nil); err != nil {
		log.Printf("禁用可信代理失败: %v", err)
	}
	if app != nil && app.Config != nil {
		if len(app.Config.Server.RemoteIPHeaders) > 0 {
			r.RemoteIPHeaders = app.Config.Server.RemoteIPHeaders
		}
		if len(app.Config.Server.TrustedProxies) > 0 {
			if err := r.SetTrustedProxies(app.Config.Server.TrustedProxies); err != nil {
				log.Printf("设置可信代理失败，回退为关闭: %v", err)
				if fallbackErr := r.SetTrustedProxies(nil); fallbackErr != nil {
					log.Printf("回退关闭可信代理失败: %v", fallbackErr)
				}
			}
		}
	}
	r.Use(gin.Recovery())
	r.GET("/api/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	if app == nil {
		return r
	}

	authHandler := app.AuthHandler
	captchaHandler := app.CaptchaHandler
	initHandler := app.InitHandler
	userHandler := app.UserHandler
	userAdminHandler := app.UserAdminHandler
	attachmentHandler := app.AttachmentHandler
	attachmentAdminHandler := app.AttachmentAdminHandler
	commentHandler := app.CommentHandler
	commentAdminHandler := app.CommentAdminHandler
	postAdminHandler := app.PostAdminHandler
	postHandler := app.PostHandler
	settingAdminHandler := app.SettingAdminHandler
	tagHandler := app.TagHandler

	requireLogin := fallbackMiddleware(http.StatusInternalServerError, "认证中间件未正确初始化")
	requireAdmin := fallbackMiddleware(http.StatusInternalServerError, "认证中间件未正确初始化")
	if app.Middleware != nil {
		requireLogin = app.Middleware.RequireLogin
		requireAdmin = app.Middleware.RequireAdmin
	}

	v1 := r.Group("/api/v1")
	{
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/register", authHandler.Register)
			authGroup.POST("/login", authHandler.Login)
		}

		captchaGroup := v1.Group("/captcha")
		{
			captchaGroup.GET("/config", captchaHandler.GetConfig)
		}

		initGroup := v1.Group("/init")
		{
			initGroup.GET("/status", initHandler.Status)
			initGroup.POST("", initHandler.Initialize)
		}

		loginGroup := v1.Group("")
		loginGroup.Use(requireLogin)
		{
			usersMe := loginGroup.Group("/users/me")
			{
				usersMe.GET("", userHandler.GetProfile)
				usersMe.PATCH("/profile", userHandler.UpdateProfile)
				usersMe.PATCH("/password", userHandler.ChangePassword)
				usersMe.POST("/email-change", userHandler.RequestEmailChange)
				usersMe.POST("/avatar", userHandler.UploadAvatar)
				usersMe.DELETE("/avatar", userHandler.DeleteAvatar)

				comments := usersMe.Group("/comments")
				{
					comments.GET("", commentHandler.List)
					comments.GET("/:id", commentHandler.Get)
					comments.PATCH("/:id", commentHandler.Update)
					comments.DELETE("/:id", commentHandler.Delete)
				}
			}

			posts := loginGroup.Group("/posts")
			{
				posts.POST("/:id/comments", commentHandler.Create)
			}

			attachments := loginGroup.Group("/attachments")
			{
				attachments.POST("", attachmentHandler.Upload)
				attachments.GET("", attachmentHandler.List)
				attachments.GET("/:id", attachmentHandler.Get)
				attachments.DELETE("/:id", attachmentHandler.Delete)
			}

		}

		comments := v1.Group("/comments")
		{
			comments.GET("", commentHandler.ListPublic)
		}

		settings := v1.Group("/settings")
		{
			settings.GET("/public", settingAdminHandler.Public)
		}

		posts := v1.Group("/posts")
		{
			posts.GET("", postHandler.List)
			posts.GET("/:id/comments", commentHandler.ListByPost)
		}

		tags := v1.Group("/tags")
		{
			tags.GET("", tagHandler.List)
		}

		adminGroup := v1.Group("/admin")
		adminGroup.Use(requireAdmin)
		{
			users := adminGroup.Group("/users")
			{
				users.GET("", userAdminHandler.List)
				users.POST("", userAdminHandler.Create)
				users.PATCH("/:id", userAdminHandler.Update)
				users.DELETE("/:id", userAdminHandler.Delete)
				users.POST("/:id/avatar", userAdminHandler.UploadAvatar)
				users.DELETE("/:id/avatar", userAdminHandler.DeleteAvatar)
			}

			posts := adminGroup.Group("/posts")
			{
				posts.GET("", postAdminHandler.List)
				posts.GET("/:id", postAdminHandler.Get)
				posts.POST("", postAdminHandler.Create)
				posts.PATCH("/:id", postAdminHandler.Update)
				posts.DELETE("/:id", postAdminHandler.Delete)
			}

			attachments := adminGroup.Group("/attachments")
			{
				attachments.GET("", attachmentAdminHandler.List)
				attachments.DELETE("/:id", attachmentAdminHandler.Delete)
				attachments.POST("/batch-delete", attachmentAdminHandler.BatchDelete)
			}

			comments := adminGroup.Group("/comments")
			{
				comments.GET("", commentAdminHandler.List)
				comments.GET("/:id", commentAdminHandler.Get)
				comments.PATCH("/:id", commentAdminHandler.Update)
				comments.DELETE("/:id", commentAdminHandler.Delete)
			}

			settings := adminGroup.Group("/settings")
			{
				settings.GET("", settingAdminHandler.List)
				settings.PATCH("/:key", settingAdminHandler.Update)
			}
		}
	}

	return r
}

func fallbackMiddleware(status int, message string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.AbortWithStatusJSON(status, gin.H{"message": message})
	}
}
