package router

import (
	"easydrop/internal/handler"

	"github.com/gin-gonic/gin"
)

type authRouteDeps struct {
	v1                       *gin.RouterGroup
	authHandler              *handler.AuthHandler
	ordinaryLimit            gin.HandlerFunc
	authWriteLimit           gin.HandlerFunc
	csrfProtect              gin.HandlerFunc
	issueCSRFCookieOnSuccess gin.HandlerFunc
}

func registerAuthRoutes(deps *authRouteDeps) {
	if deps == nil || deps.v1 == nil {
		return
	}

	authGroup := deps.v1.Group("/auth")
	authGroup.Use(deps.ordinaryLimit, deps.authWriteLimit)
	{
		authGroup.POST("/register", deps.authHandler.Register)
		authGroup.POST("/login", deps.issueCSRFCookieOnSuccess, deps.authHandler.Login)
		authGroup.POST("/logout", deps.csrfProtect, deps.authHandler.Logout)
		authGroup.POST("/password-reset/request", deps.authHandler.RequestPasswordReset)
		authGroup.POST("/password-reset/confirm", deps.authHandler.ConfirmPasswordReset)
		authGroup.POST("/verify-email/confirm", deps.authHandler.ConfirmVerifyEmail)
		authGroup.POST("/email-change/confirm", deps.authHandler.ConfirmEmailChange)
	}
}
