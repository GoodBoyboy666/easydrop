package router

import (
	"log"
	"path/filepath"
	"strings"

	_ "easydrop/docs"
	"easydrop/internal/config"
	"easydrop/internal/di"
	"easydrop/internal/middleware"
	"easydrop/internal/pkg/storage"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// BuildEngine 构建并返回应用的 HTTP 路由引擎。
func BuildEngine(app *di.App) *gin.Engine {
	if app != nil && app.Config != nil {
		if app.Config.Server.Mode == config.ServerModeProduction {
			gin.SetMode(gin.ReleaseMode)
		} else {
			gin.SetMode(gin.DebugMode)
		}
	}

	r := gin.New()
	r.MaxMultipartMemory = middleware.OrdinaryMaxRequestBodyBytes
	if err := r.SetTrustedProxies(nil); err != nil {
		log.Printf("禁用可信代理失败: %v", err)
	}
	if app != nil && app.Config != nil {
		if len(app.Config.Server.RemoteIPHeaders) > 0 {
			r.RemoteIPHeaders = app.Config.Server.RemoteIPHeaders
		}
		if len(app.Config.Server.TrustedProxies) > 0 {
			if err := r.SetTrustedProxies(app.Config.Server.TrustedProxies); err != nil {
				log.Printf("设置可信代理失败，回退为关闭: %v", err)
				if fallbackErr := r.SetTrustedProxies(nil); fallbackErr != nil {
					log.Printf("回退关闭可信代理失败: %v", fallbackErr)
				}
			}
		}
	}
	securityHeaders := passthroughMiddleware()
	if app != nil && app.SecurityHeaders != nil {
		securityHeaders = app.SecurityHeaders.Apply
	}
	r.Use(gin.Recovery(), securityHeaders)
	if shouldRegisterSwagger(app) {
		r.GET("/api/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	if app == nil {
		return r
	}

	registerLocalStorageStaticRoutes(r, app.Config)
	routeCtx := newRouteContext(r, app)
	registerAuthRoutes(routeCtx.authDeps())
	registerUserRoutes(routeCtx.userDeps())
	registerPostRoutes(routeCtx.postDeps())
	registerCommentRoutes(routeCtx.commentDeps())
	registerAttachmentRoutes(routeCtx.attachmentDeps())
	registerPublicRoutes(routeCtx.publicDeps())
	registerFeedRoutes(routeCtx.feedDeps())
	registerAdminRoutes(routeCtx.adminDeps())
	registerPasskeyRoutes(routeCtx.passkeyDeps())
	registerOAuthRoutes(routeCtx.oauthDeps())

	return r
}

func shouldRegisterSwagger(app *di.App) bool {
	if app == nil || app.Config == nil {
		return true
	}

	return app.Config.Server.Mode == config.ServerModeDevelopment
}

func fallbackMiddleware(status int, message string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.AbortWithStatusJSON(status, gin.H{"message": message})
	}
}

func passthroughMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}

func registerLocalStorageStaticRoutes(r *gin.Engine, cfg *config.StaticConfig) {
	if r == nil || cfg == nil {
		return
	}

	if strings.TrimSpace(cfg.Storage.Backend) != storage.BackendLocal {
		return
	}

	basePath := strings.TrimSpace(cfg.Storage.Local.BasePath)
	baseURL := normalizeLocalStorageRoutePrefix(cfg.Storage.Local.URLPrefix)
	if basePath == "" || baseURL == "" {
		return
	}

	absBasePath, err := filepath.Abs(basePath)
	if err != nil {
		log.Printf("解析本地存储目录失败: %v", err)
		return
	}

	registerLocalStorageCategoryRoute(r, baseURL, storage.CategoryFile, filepath.Join(absBasePath, storage.CategoryFile))
	registerLocalStorageCategoryRoute(r, baseURL, storage.CategoryAvatar, filepath.Join(absBasePath, storage.CategoryAvatar))
}

func normalizeLocalStorageRoutePrefix(value string) string {
	trimmed := strings.Trim(strings.TrimSpace(value), "/")
	if trimmed == "" {
		return "/api"
	}
	return "/api/" + trimmed
}

func registerLocalStorageCategoryRoute(r *gin.Engine, baseURL string, category string, categoryPath string) {
	if r == nil {
		return
	}

	absCategoryPath, err := filepath.Abs(categoryPath)
	if err != nil {
		log.Printf("解析本地存储分类目录失败: %v", err)
		return
	}

	mountPath := strings.TrimRight(baseURL, "/") + "/" + strings.Trim(category, "/")
	r.StaticFS(mountPath, gin.Dir(absCategoryPath, false))
}
