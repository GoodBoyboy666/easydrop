package router

import (
	"easydrop/internal/handler"

	"github.com/gin-gonic/gin"
)

type feedRouteDeps struct {
	v1            *gin.RouterGroup
	feedHandler   *handler.FeedHandler
	ordinaryLimit gin.HandlerFunc
}

func registerFeedRoutes(deps *feedRouteDeps) {
	if deps == nil || deps.v1 == nil {
		return
	}

	feed := deps.v1.Group("/feed")
	feed.Use(deps.ordinaryLimit)
	{
		feed.GET("/rss", deps.feedHandler.GetRSS)
		feed.GET("/atom", deps.feedHandler.GetAtom)
	}
}
