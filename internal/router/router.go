package router

import (
	"net/http"

	"easydrop/internal/di"

	"github.com/gin-gonic/gin"
)

// BuildEngine 构建并返回应用的 HTTP 路由引擎。
func BuildEngine(app *di.App) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	if app == nil {
		return r
	}

	authHandler := app.AuthHandler
	userHandler := app.UserHandler
	userAdminHandler := app.UserAdminHandler
	attachmentHandler := app.AttachmentHandler
	attachmentAdminHandler := app.AttachmentAdminHandler
	commentHandler := app.CommentHandler
	commentAdminHandler := app.CommentAdminHandler
	postAdminHandler := app.PostAdminHandler
	settingAdminHandler := app.SettingAdminHandler

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
			}

			attachments := loginGroup.Group("/attachments")
			{
				attachments.POST("", attachmentHandler.Upload)
				attachments.GET("", attachmentHandler.List)
				attachments.GET("/:id", attachmentHandler.Get)
				attachments.DELETE("/:id", attachmentHandler.Delete)
			}

			comments := loginGroup.Group("/comments")
			{
				comments.POST("", commentHandler.Create)
				comments.GET("", commentHandler.List)
				comments.GET("/:id", commentHandler.Get)
				comments.PATCH("/:id", commentHandler.Update)
				comments.DELETE("/:id", commentHandler.Delete)
			}
		}

		settings := v1.Group("/settings")
		{
			settings.GET("/public", settingAdminHandler.Public)
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
