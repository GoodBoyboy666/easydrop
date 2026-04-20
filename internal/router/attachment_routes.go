package router

import (
	"easydrop/internal/handler"

	"github.com/gin-gonic/gin"
)

type attachmentRouteDeps struct {
	loginGroup           *gin.RouterGroup
	attachmentHandler    *handler.AttachmentHandler
	ordinaryLimit        gin.HandlerFunc
	uploadLimit          gin.HandlerFunc
	attachmentWriteLimit gin.HandlerFunc
}

func registerAttachmentRoutes(deps *attachmentRouteDeps) {
	if deps == nil || deps.loginGroup == nil {
		return
	}

	attachments := deps.loginGroup.Group("/attachments")
	attachments.Use(deps.ordinaryLimit)
	{
		attachments.GET("", deps.attachmentHandler.List)
		attachments.GET("/:id", deps.attachmentHandler.Get)
		attachments.DELETE("/:id", deps.attachmentWriteLimit, deps.attachmentHandler.Delete)
	}

	attachmentUploads := deps.loginGroup.Group("/attachments")
	attachmentUploads.Use(deps.uploadLimit)
	{
		attachmentUploads.POST("", deps.attachmentWriteLimit, deps.attachmentHandler.Upload)
	}
}
