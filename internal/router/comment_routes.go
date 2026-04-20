package router

import (
	"easydrop/internal/handler"

	"github.com/gin-gonic/gin"
)

type commentRouteDeps struct {
	v1                *gin.RouterGroup
	loginGroup        *gin.RouterGroup
	commentHandler    *handler.CommentHandler
	ordinaryLimit     gin.HandlerFunc
	commentWriteLimit gin.HandlerFunc
}

func registerCommentRoutes(deps *commentRouteDeps) {
	if deps == nil || deps.v1 == nil || deps.loginGroup == nil {
		return
	}

	comments := deps.v1.Group("/comments")
	comments.Use(deps.ordinaryLimit)
	{
		comments.GET("", deps.commentHandler.ListPublic)
	}

	publicPostComments := deps.v1.Group("/posts")
	publicPostComments.Use(deps.ordinaryLimit)
	{
		publicPostComments.GET("/:id/comments", deps.commentHandler.ListByPost)
	}

	postComments := deps.loginGroup.Group("/posts")
	postComments.Use(deps.ordinaryLimit)
	{
		postComments.POST("/:id/comments", deps.commentWriteLimit, deps.commentHandler.Create)
	}

	meComments := deps.loginGroup.Group("/users/me/comments")
	meComments.Use(deps.ordinaryLimit)
	{
		meComments.GET("", deps.commentHandler.List)
		meComments.GET("/:id", deps.commentHandler.Get)
		meComments.PATCH("/:id", deps.commentWriteLimit, deps.commentHandler.Update)
		meComments.DELETE("/:id", deps.commentWriteLimit, deps.commentHandler.Delete)
	}
}
