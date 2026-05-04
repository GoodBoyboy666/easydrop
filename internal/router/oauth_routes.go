package router

import (
	"easydrop/internal/handler"

	"github.com/gin-gonic/gin"
)

type oauthRouteDeps struct {
	v1              *gin.RouterGroup
	loginGroup      *gin.RouterGroup
	oauthHandler    *handler.OAuthHandler
	ordinaryLimit   gin.HandlerFunc
	authWriteLimit  gin.HandlerFunc
	csrfProtect     gin.HandlerFunc
}

func registerOAuthRoutes(deps *oauthRouteDeps) {
	if deps == nil || deps.v1 == nil {
		return
	}

	oauthGroup := deps.v1.Group("/auth/oauth")
	oauthGroup.Use(deps.ordinaryLimit)
	{
		oauthGroup.GET("/providers", deps.oauthHandler.ListProviders)
		oauthGroup.GET("/:provider", deps.oauthHandler.Authorize)
		oauthGroup.POST("/:provider/callback", deps.authWriteLimit, deps.oauthHandler.Callback)
	}

	oauthBindGroup := deps.loginGroup.Group("/users/me/oauth-binds")
	oauthBindGroup.Use(deps.ordinaryLimit)
	{
		oauthBindGroup.GET("", deps.oauthHandler.ListBindings)
		oauthBindGroup.DELETE("/:id", deps.oauthHandler.Unbind)
		oauthBindGroup.POST("/:provider", deps.oauthHandler.BindManually)
	}
}
