package router

import (
	"easydrop/internal/handler"

	"github.com/gin-gonic/gin"
)

type postRouteDeps struct {
	v1            *gin.RouterGroup
	postHandler   *handler.PostHandler
	ordinaryLimit gin.HandlerFunc
}

func registerPostRoutes(deps *postRouteDeps) {
	if deps == nil || deps.v1 == nil {
		return
	}

	posts := deps.v1.Group("/posts")
	posts.Use(deps.ordinaryLimit)
	{
		posts.GET("", deps.postHandler.List)
		posts.GET("/:id", deps.postHandler.Get)
	}
}
