package router

import (
	"easydrop/internal/handler"

	"github.com/gin-gonic/gin"
)

type publicRouteDeps struct {
	v1                  *gin.RouterGroup
	captchaHandler      *handler.CaptchaHandler
	initHandler         *handler.InitHandler
	settingAdminHandler *handler.SettingAdminHandler
	tagHandler          *handler.TagHandler
	ordinaryLimit       gin.HandlerFunc
	initWriteLimit      gin.HandlerFunc
}

func registerPublicRoutes(deps *publicRouteDeps) {
	if deps == nil || deps.v1 == nil {
		return
	}

	captchaGroup := deps.v1.Group("/captcha")
	captchaGroup.Use(deps.ordinaryLimit)
	{
		captchaGroup.GET("/config", deps.captchaHandler.GetConfig)
	}

	initGroup := deps.v1.Group("/init")
	initGroup.Use(deps.ordinaryLimit)
	{
		initGroup.GET("/status", deps.initHandler.Status)
		initGroup.POST("", deps.initWriteLimit, deps.initHandler.Initialize)
	}

	settings := deps.v1.Group("/settings")
	settings.Use(deps.ordinaryLimit)
	{
		settings.GET("/public", deps.settingAdminHandler.Public)
	}

	tags := deps.v1.Group("/tags")
	tags.Use(deps.ordinaryLimit)
	{
		tags.GET("", deps.tagHandler.List)
	}
}
