package middleware

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"easydrop/internal/pkg/ratelimit"

	"github.com/gin-gonic/gin"
)

const (
	RuleNameAuthWrite         = "auth_write"
	RuleNameInitWrite         = "init_write"
	RuleNameProfileWrite      = "profile_write"
	RuleNameUserSecurityWrite = "user_security_write"
	RuleNameCommentWrite      = "comment_write"
	RuleNameAttachmentWrite   = "attachment_write"
)

// RateLimit 提供按规则名构造 Gin 限流中间件的能力。
type RateLimit interface {
	Window(name string) gin.HandlerFunc
	Cooldown(name string) gin.HandlerFunc
}

type rateLimit struct {
	cfg     *ratelimit.Config
	limiter ratelimit.Limiter
}

// NewRateLimit 创建限流中间件。
func NewRateLimit(cfg *ratelimit.Config, limiter ratelimit.Limiter) RateLimit {
	return &rateLimit{
		cfg:     cfg,
		limiter: limiter,
	}
}

func (m *rateLimit) Window(name string) gin.HandlerFunc {
	return m.buildMiddleware(name, ratelimit.ModeWindow)
}

func (m *rateLimit) Cooldown(name string) gin.HandlerFunc {
	return m.buildMiddleware(name, ratelimit.ModeCooldown)
}

func (m *rateLimit) buildMiddleware(name string, mode string) gin.HandlerFunc {
	if !m.enabled() {
		return passthroughRateLimit()
	}
	if m.limiter == nil {
		return invalidRateLimitMiddleware(http.StatusInternalServerError, "限流服务未正确初始化")
	}

	rule, ok := m.resolveRule(name, mode)
	if !ok {
		return passthroughRateLimit()
	}

	if err := ratelimit.ValidateRule(rule); err != nil {
		log.Printf("解析限流规则失败: %v", err)
		return invalidRateLimitMiddleware(http.StatusInternalServerError, "限流配置无效")
	}

	return func(c *gin.Context) {
		decision, err := m.limiter.Allow(c.Request.Context(), buildRateLimitKey(c), rule)
		if err != nil {
			log.Printf("执行限流失败: %v", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "限流服务异常"})
			return
		}

		writeRateLimitHeaders(c, decision)
		if !decision.Allowed {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"message": "请求过于频繁，请稍后再试"})
			return
		}

		c.Next()
	}
}

func (m *rateLimit) enabled() bool {
	return m != nil && m.cfg != nil && m.cfg.Enabled
}

func (m *rateLimit) resolveRule(name string, mode string) (ratelimit.Rule, bool) {
	if m == nil || m.cfg == nil {
		return ratelimit.Rule{}, false
	}

	normalizedName := strings.TrimSpace(name)
	if normalizedName == "" || m.cfg.Rules == nil {
		return ratelimit.Rule{}, false
	}

	ruleConfig, ok := m.cfg.Rules[normalizedName]
	if !ok {
		return ratelimit.Rule{}, false
	}

	return ratelimit.Rule{
		Name:     normalizedName,
		Mode:     mode,
		Interval: ruleConfig.Interval,
		Limit:    ruleConfig.Limit,
	}, true
}

func buildRateLimitKey(c *gin.Context) string {
	if userID, ok := GetUserID(c); ok && userID > 0 {
		return "user:" + strconv.FormatUint(uint64(userID), 10)
	}

	clientIP := strings.TrimSpace(c.ClientIP())
	if clientIP == "" {
		clientIP = "unknown"
	}
	return "ip:" + clientIP
}

func writeRateLimitHeaders(c *gin.Context, decision *ratelimit.Decision) {
	if c == nil || decision == nil {
		return
	}

	c.Header("X-RateLimit-Remaining", strconv.Itoa(maxHeaderInt(decision.Remaining)))
	if !decision.ResetAt.IsZero() {
		c.Header("X-RateLimit-Reset", strconv.FormatInt(decision.ResetAt.UTC().Unix(), 10))
	}
	if decision.RetryAfter > 0 {
		c.Header("Retry-After", strconv.FormatInt(retryAfterSeconds(decision.RetryAfter), 10))
	}
}

func retryAfterSeconds(duration time.Duration) int64 {
	seconds := int64(duration / time.Second)
	if duration%time.Second != 0 {
		seconds++
	}
	if seconds <= 0 {
		return 1
	}
	return seconds
}

func maxHeaderInt(value int) int {
	if value < 0 {
		return 0
	}
	return value
}

func passthroughRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}

func invalidRateLimitMiddleware(status int, message string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.AbortWithStatusJSON(status, gin.H{"message": message})
	}
}

var _ RateLimit = (*rateLimit)(nil)
