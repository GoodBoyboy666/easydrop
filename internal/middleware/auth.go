package middleware

import (
	"errors"
	"net/http"
	"strings"

	"easydrop/internal/model"
	"easydrop/internal/pkg/jwt"
	"easydrop/internal/repo"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const ContextUserIDKey = "userID"

// Auth 持有认证中间件所需依赖。
type Auth interface {
	RequireLogin(c *gin.Context)
	RequireAdmin(c *gin.Context)
}

type auth struct {
	jwtManager jwt.Manager
	userRepo   repo.UserRepo
}

// NewAuth 构造认证中间件对象。
func NewAuth(jwtManager jwt.Manager, userRepo repo.UserRepo) Auth {
	return &auth{
		jwtManager: jwtManager,
		userRepo:   userRepo,
	}
}

// UserIDFromGin 从 gin context 中读取用户 ID。
func GetUserID(c *gin.Context) (uint, bool) {
	v, ok := c.Get(ContextUserIDKey)
	if !ok {
		return 0, false
	}
	uid, ok := v.(uint)
	return uid, ok
}

// RequireLogin 校验登录态、JWT、有无用户及用户状态，并将用户 ID 注入 gin context。
func (a *auth) RequireLogin(c *gin.Context) {
	user, ok := a.authenticateUser(c)
	if !ok {
		return
	}

	c.Set(ContextUserIDKey, user.ID)

	c.Next()
}

// RequireAdmin 校验管理员身份，并将用户 ID 注入 gin context。
func (a *auth) RequireAdmin(c *gin.Context) {
	user, ok := a.authenticateUser(c)
	if !ok {
		return
	}

	if !user.Admin {
		abortWithMessage(c, http.StatusForbidden, "需要管理员权限")
		return
	}

	c.Set(ContextUserIDKey, user.ID)

	c.Next()
}

func (a *auth) authenticateUser(c *gin.Context) (*model.User, bool) {
	if a == nil || a.jwtManager == nil || a.userRepo == nil {
		abortWithMessage(c, http.StatusInternalServerError, "认证服务未正确初始化")
		return nil, false
	}

	token, ok := extractBearerToken(c.GetHeader("Authorization"))
	if !ok {
		abortWithMessage(c, http.StatusUnauthorized, "未登录或登录已失效")
		return nil, false
	}

	claims, err := a.jwtManager.ParseToken(token)
	if err != nil {
		abortWithMessage(c, http.StatusUnauthorized, "登录已失效，请重新登录")
		return nil, false
	}

	if claims.UserID == 0 {
		abortWithMessage(c, http.StatusUnauthorized, "登录态无效")
		return nil, false
	}

	user, err := a.userRepo.GetByID(c.Request.Context(), claims.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			abortWithMessage(c, http.StatusUnauthorized, "用户不存在或已被删除")
			return nil, false
		}
		abortWithMessage(c, http.StatusInternalServerError, "查询用户失败")
		return nil, false
	}

	if user.Status != 1 {
		abortWithMessage(c, http.StatusForbidden, "用户状态异常")
		return nil, false
	}

	return user, true
}

func extractBearerToken(header string) (string, bool) {
	header = strings.TrimSpace(header)
	if header == "" {
		return "", false
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", false
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", false
	}

	return token, true
}

func abortWithMessage(c *gin.Context, status int, message string) {
	c.AbortWithStatusJSON(status, gin.H{"message": message})
}

var _ Auth = (*auth)(nil)
