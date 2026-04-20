package router

import (
	"easydrop/internal/handler"

	"github.com/gin-gonic/gin"
)

type adminRouteDeps struct {
	adminGroup             *gin.RouterGroup
	overviewAdminHandler   *handler.OverviewAdminHandler
	userAdminHandler       *handler.UserAdminHandler
	postAdminHandler       *handler.PostAdminHandler
	attachmentAdminHandler *handler.AttachmentAdminHandler
	commentAdminHandler    *handler.CommentAdminHandler
	settingAdminHandler    *handler.SettingAdminHandler
	ordinaryLimit          gin.HandlerFunc
	uploadLimit            gin.HandlerFunc
}

func registerAdminRoutes(deps *adminRouteDeps) {
	if deps == nil || deps.adminGroup == nil {
		return
	}

	overview := deps.adminGroup.Group("/overview")
	overview.Use(deps.ordinaryLimit)
	{
		overview.GET("", deps.overviewAdminHandler.Get)
	}

	users := deps.adminGroup.Group("/users")
	users.Use(deps.ordinaryLimit)
	{
		users.GET("", deps.userAdminHandler.List)
		users.POST("", deps.userAdminHandler.Create)
		users.PATCH("/:id", deps.userAdminHandler.Update)
		users.DELETE("/:id", deps.userAdminHandler.Delete)
		users.DELETE("/:id/avatar", deps.userAdminHandler.DeleteAvatar)
	}
	userAvatarUploads := deps.adminGroup.Group("/users")
	userAvatarUploads.Use(deps.uploadLimit)
	{
		userAvatarUploads.POST("/:id/avatar", deps.userAdminHandler.UploadAvatar)
	}

	posts := deps.adminGroup.Group("/posts")
	posts.Use(deps.ordinaryLimit)
	{
		posts.GET("", deps.postAdminHandler.List)
		posts.GET("/:id", deps.postAdminHandler.Get)
		posts.POST("", deps.postAdminHandler.Create)
		posts.PATCH("/:id", deps.postAdminHandler.Update)
		posts.DELETE("/:id", deps.postAdminHandler.Delete)
	}

	attachments := deps.adminGroup.Group("/attachments")
	attachments.Use(deps.ordinaryLimit)
	{
		attachments.GET("", deps.attachmentAdminHandler.List)
		attachments.DELETE("/:id", deps.attachmentAdminHandler.Delete)
		attachments.POST("/batch-delete", deps.attachmentAdminHandler.BatchDelete)
	}

	comments := deps.adminGroup.Group("/comments")
	comments.Use(deps.ordinaryLimit)
	{
		comments.GET("", deps.commentAdminHandler.List)
		comments.GET("/:id", deps.commentAdminHandler.Get)
		comments.PATCH("/:id", deps.commentAdminHandler.Update)
		comments.DELETE("/:id", deps.commentAdminHandler.Delete)
	}

	settings := deps.adminGroup.Group("/settings")
	settings.Use(deps.ordinaryLimit)
	{
		settings.GET("", deps.settingAdminHandler.List)
		settings.PATCH("/:key", deps.settingAdminHandler.Update)
	}
}
