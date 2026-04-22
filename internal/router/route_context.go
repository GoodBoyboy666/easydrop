package router

import (
	"net/http"

	"easydrop/internal/di"
	"easydrop/internal/handler"
	"easydrop/internal/middleware"

	"github.com/gin-gonic/gin"
)

// routeContext 聚合路由注册需要的依赖与分组。
type routeContext struct {
	v1         *gin.RouterGroup
	loginGroup *gin.RouterGroup
	adminGroup *gin.RouterGroup

	authHandler            *handler.AuthHandler
	captchaHandler         *handler.CaptchaHandler
	initHandler            *handler.InitHandler
	userHandler            *handler.UserHandler
	userAdminHandler       *handler.UserAdminHandler
	attachmentHandler      *handler.AttachmentHandler
	attachmentAdminHandler *handler.AttachmentAdminHandler
	commentHandler         *handler.CommentHandler
	commentAdminHandler    *handler.CommentAdminHandler
	overviewAdminHandler   *handler.OverviewAdminHandler
	postAdminHandler       *handler.PostAdminHandler
	postHandler            *handler.PostHandler
	settingAdminHandler    *handler.SettingAdminHandler
	tagHandler             *handler.TagHandler

	requireLogin             gin.HandlerFunc
	requireAdmin             gin.HandlerFunc
	csrfProtect              gin.HandlerFunc
	issueCSRFCookieOnSuccess gin.HandlerFunc
	ordinaryLimit            gin.HandlerFunc
	uploadLimit              gin.HandlerFunc
	authWriteLimit           gin.HandlerFunc
	initWriteLimit           gin.HandlerFunc
	profileWriteLimit        gin.HandlerFunc
	userSecurityWriteLimit   gin.HandlerFunc
	commentWriteLimit        gin.HandlerFunc
	attachmentWriteLimit     gin.HandlerFunc
}

func newRouteContext(r *gin.Engine, app *di.App) *routeContext {
	if r == nil || app == nil {
		return nil
	}

	optionalLogin := passthroughMiddleware()
	ctx := &routeContext{
		authHandler:              app.AuthHandler,
		captchaHandler:           app.CaptchaHandler,
		initHandler:              app.InitHandler,
		userHandler:              app.UserHandler,
		userAdminHandler:         app.UserAdminHandler,
		attachmentHandler:        app.AttachmentHandler,
		attachmentAdminHandler:   app.AttachmentAdminHandler,
		commentHandler:           app.CommentHandler,
		commentAdminHandler:      app.CommentAdminHandler,
		overviewAdminHandler:     app.OverviewAdminHandler,
		postAdminHandler:         app.PostAdminHandler,
		postHandler:              app.PostHandler,
		settingAdminHandler:      app.SettingAdminHandler,
		tagHandler:               app.TagHandler,
		requireLogin:             fallbackMiddleware(http.StatusInternalServerError, "认证中间件未正确初始化"),
		requireAdmin:             fallbackMiddleware(http.StatusInternalServerError, "认证中间件未正确初始化"),
		csrfProtect:              passthroughMiddleware(),
		issueCSRFCookieOnSuccess: passthroughMiddleware(),
		ordinaryLimit:            passthroughMiddleware(),
		uploadLimit:              passthroughMiddleware(),
		authWriteLimit:           passthroughMiddleware(),
		initWriteLimit:           passthroughMiddleware(),
		profileWriteLimit:        passthroughMiddleware(),
		userSecurityWriteLimit:   passthroughMiddleware(),
		commentWriteLimit:        passthroughMiddleware(),
		attachmentWriteLimit:     passthroughMiddleware(),
	}

	if app.Middleware != nil {
		ctx.requireLogin = app.Middleware.RequireLogin
		ctx.requireAdmin = app.Middleware.RequireAdmin
		optionalLogin = app.Middleware.OptionalLogin
	}
	if app.CSRF != nil {
		ctx.csrfProtect = app.CSRF.Protect
		ctx.issueCSRFCookieOnSuccess = app.CSRF.IssueCSRFCookieOnSuccess
	}
	if app.RequestBodyLimit != nil {
		ctx.ordinaryLimit = app.RequestBodyLimit.Ordinary
		ctx.uploadLimit = app.RequestBodyLimit.Upload
	}
	if app.RateLimit != nil {
		ctx.authWriteLimit = app.RateLimit.Cooldown(middleware.RuleNameAuthWrite)
		ctx.initWriteLimit = app.RateLimit.Cooldown(middleware.RuleNameInitWrite)
		ctx.profileWriteLimit = app.RateLimit.Window(middleware.RuleNameProfileWrite)
		ctx.userSecurityWriteLimit = app.RateLimit.Cooldown(middleware.RuleNameUserSecurityWrite)
		ctx.commentWriteLimit = app.RateLimit.Window(middleware.RuleNameCommentWrite)
		ctx.attachmentWriteLimit = app.RateLimit.Window(middleware.RuleNameAttachmentWrite)
	}

	ctx.v1 = r.Group("/api/v1")
	ctx.v1.Use(optionalLogin)

	ctx.loginGroup = ctx.v1.Group("")
	ctx.loginGroup.Use(ctx.csrfProtect, ctx.requireLogin)

	ctx.adminGroup = ctx.v1.Group("/admin")
	ctx.adminGroup.Use(ctx.csrfProtect, ctx.requireAdmin)

	return ctx
}

func (ctx *routeContext) authDeps() *authRouteDeps {
	if ctx == nil || ctx.v1 == nil {
		return nil
	}

	return &authRouteDeps{
		v1:                       ctx.v1,
		authHandler:              ctx.authHandler,
		ordinaryLimit:            ctx.ordinaryLimit,
		authWriteLimit:           ctx.authWriteLimit,
		csrfProtect:              ctx.csrfProtect,
		issueCSRFCookieOnSuccess: ctx.issueCSRFCookieOnSuccess,
	}
}

func (ctx *routeContext) userDeps() *userRouteDeps {
	if ctx == nil || ctx.loginGroup == nil {
		return nil
	}

	return &userRouteDeps{
		loginGroup:             ctx.loginGroup,
		userHandler:            ctx.userHandler,
		ordinaryLimit:          ctx.ordinaryLimit,
		uploadLimit:            ctx.uploadLimit,
		profileWriteLimit:      ctx.profileWriteLimit,
		userSecurityWriteLimit: ctx.userSecurityWriteLimit,
	}
}

func (ctx *routeContext) postDeps() *postRouteDeps {
	if ctx == nil || ctx.v1 == nil {
		return nil
	}

	return &postRouteDeps{
		v1:            ctx.v1,
		postHandler:   ctx.postHandler,
		ordinaryLimit: ctx.ordinaryLimit,
	}
}

func (ctx *routeContext) commentDeps() *commentRouteDeps {
	if ctx == nil || ctx.v1 == nil || ctx.loginGroup == nil {
		return nil
	}

	return &commentRouteDeps{
		v1:                ctx.v1,
		loginGroup:        ctx.loginGroup,
		commentHandler:    ctx.commentHandler,
		ordinaryLimit:     ctx.ordinaryLimit,
		commentWriteLimit: ctx.commentWriteLimit,
	}
}

func (ctx *routeContext) attachmentDeps() *attachmentRouteDeps {
	if ctx == nil || ctx.loginGroup == nil {
		return nil
	}

	return &attachmentRouteDeps{
		loginGroup:           ctx.loginGroup,
		attachmentHandler:    ctx.attachmentHandler,
		ordinaryLimit:        ctx.ordinaryLimit,
		uploadLimit:          ctx.uploadLimit,
		attachmentWriteLimit: ctx.attachmentWriteLimit,
	}
}

func (ctx *routeContext) publicDeps() *publicRouteDeps {
	if ctx == nil || ctx.v1 == nil {
		return nil
	}

	return &publicRouteDeps{
		v1:                  ctx.v1,
		captchaHandler:      ctx.captchaHandler,
		initHandler:         ctx.initHandler,
		settingAdminHandler: ctx.settingAdminHandler,
		tagHandler:          ctx.tagHandler,
		ordinaryLimit:       ctx.ordinaryLimit,
		initWriteLimit:      ctx.initWriteLimit,
	}
}

func (ctx *routeContext) adminDeps() *adminRouteDeps {
	if ctx == nil || ctx.adminGroup == nil {
		return nil
	}

	return &adminRouteDeps{
		adminGroup:             ctx.adminGroup,
		overviewAdminHandler:   ctx.overviewAdminHandler,
		userAdminHandler:       ctx.userAdminHandler,
		postAdminHandler:       ctx.postAdminHandler,
		attachmentAdminHandler: ctx.attachmentAdminHandler,
		commentAdminHandler:    ctx.commentAdminHandler,
		settingAdminHandler:    ctx.settingAdminHandler,
		ordinaryLimit:          ctx.ordinaryLimit,
		uploadLimit:            ctx.uploadLimit,
	}
}
