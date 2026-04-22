package router

import (
	"easydrop/internal/handler"

	"github.com/gin-gonic/gin"
)

type userRouteDeps struct {
	loginGroup             *gin.RouterGroup
	userHandler            *handler.UserHandler
	ordinaryLimit          gin.HandlerFunc
	uploadLimit            gin.HandlerFunc
	profileWriteLimit      gin.HandlerFunc
	userSecurityWriteLimit gin.HandlerFunc
}

func registerUserRoutes(deps *userRouteDeps) {
	if deps == nil || deps.loginGroup == nil {
		return
	}

	usersMe := deps.loginGroup.Group("/users/me")
	usersMe.Use(deps.ordinaryLimit)
	{
		usersMe.GET("", deps.userHandler.GetProfile)
		usersMe.PATCH("/profile", deps.profileWriteLimit, deps.userHandler.UpdateProfile)
		usersMe.PATCH("/password", deps.userSecurityWriteLimit, deps.userHandler.ChangePassword)
		usersMe.POST("/email-change", deps.userSecurityWriteLimit, deps.userHandler.RequestEmailChange)
		usersMe.DELETE("/avatar", deps.profileWriteLimit, deps.userHandler.DeleteAvatar)
	}

	usersMeUpload := deps.loginGroup.Group("/users/me")
	usersMeUpload.Use(deps.uploadLimit)
	{
		usersMeUpload.POST("/avatar", deps.profileWriteLimit, deps.userHandler.UploadAvatar)
	}
}
