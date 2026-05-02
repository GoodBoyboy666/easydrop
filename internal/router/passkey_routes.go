package router

import (
	"easydrop/internal/handler"

	"github.com/gin-gonic/gin"
)

type passkeyRouteDeps struct {
	v1                       *gin.RouterGroup
	loginGroup               *gin.RouterGroup
	passkeyHandler           *handler.PasskeyHandler
	ordinaryLimit            gin.HandlerFunc
	authWriteLimit           gin.HandlerFunc
	profileWriteLimit        gin.HandlerFunc
	csrfProtect              gin.HandlerFunc
	issueCSRFCookieOnSuccess gin.HandlerFunc
}

func registerPasskeyRoutes(deps *passkeyRouteDeps) {
	if deps == nil || deps.v1 == nil || deps.passkeyHandler == nil {
		return
	}

	passkeyAuthGroup := deps.v1.Group("/auth/passkey")
	passkeyAuthGroup.Use(deps.ordinaryLimit, deps.authWriteLimit)
	{
		passkeyAuthGroup.POST("/login/begin", deps.passkeyHandler.BeginLogin)
		passkeyAuthGroup.POST("/login/finish", deps.issueCSRFCookieOnSuccess, deps.passkeyHandler.FinishLogin)
		passkeyAuthGroup.POST("/register/begin", deps.passkeyHandler.BeginRegistration)
		passkeyAuthGroup.POST("/register/finish", deps.passkeyHandler.FinishRegistration)
	}

	passkeyMeGroup := deps.loginGroup.Group("/users/me/passkeys")
	passkeyMeGroup.Use(deps.ordinaryLimit)
	{
		passkeyMeGroup.GET("", deps.passkeyHandler.List)
		passkeyMeGroup.PATCH("/:id", deps.profileWriteLimit, deps.passkeyHandler.Rename)
		passkeyMeGroup.DELETE("/:id", deps.profileWriteLimit, deps.passkeyHandler.Delete)
	}
}
