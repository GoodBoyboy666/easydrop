package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"easydrop/internal/pkg/ratelimit"

	"github.com/gin-gonic/gin"
)

type fakeLimiter struct {
	backend string
	allowFn func(ctx context.Context, key string, rule ratelimit.Rule) (*ratelimit.Decision, error)
}

func (f *fakeLimiter) Backend() string {
	if f.backend == "" {
		return ratelimit.BackendMemory
	}
	return f.backend
}

func (f *fakeLimiter) Allow(ctx context.Context, key string, rule ratelimit.Rule) (*ratelimit.Decision, error) {
	if f.allowFn == nil {
		return &ratelimit.Decision{Allowed: true}, nil
	}
	return f.allowFn(ctx, key, rule)
}

func TestRateLimitUsesUserIDBeforeIP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	limiter := &fakeLimiter{
		allowFn: func(ctx context.Context, key string, rule ratelimit.Rule) (*ratelimit.Decision, error) {
			if key != "user:42" {
				t.Fatalf("expected key user:42, got %q", key)
			}
			if rule.Name != RuleNameProfileWrite {
				t.Fatalf("expected rule name %q, got %q", RuleNameProfileWrite, rule.Name)
			}
			return &ratelimit.Decision{Allowed: true, Remaining: 3, ResetAt: time.Unix(100, 0)}, nil
		},
	}
	middleware := NewRateLimit(&ratelimit.Config{
		Enabled: true,
		Rules: map[string]ratelimit.RuleConfig{
			RuleNameProfileWrite: {
				Interval: time.Minute,
				Limit:    5,
			},
		},
	}, limiter)

	router := gin.New()
	router.POST("/limited", func(c *gin.Context) {
		c.Set(ContextUserIDKey, uint(42))
		c.Next()
	}, middleware.Window(RuleNameProfileWrite), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodPost, "/limited", nil)
	req.RemoteAddr = "198.51.100.1:12345"
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", recorder.Code)
	}
	if got := recorder.Header().Get("X-RateLimit-Remaining"); got != "3" {
		t.Fatalf("expected remaining header 3, got %q", got)
	}
}

func TestRateLimitFallsBackToClientIP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	limiter := &fakeLimiter{
		allowFn: func(ctx context.Context, key string, rule ratelimit.Rule) (*ratelimit.Decision, error) {
			if key != "ip:203.0.113.5" {
				t.Fatalf("expected key ip:203.0.113.5, got %q", key)
			}
			return &ratelimit.Decision{Allowed: true}, nil
		},
	}
	middleware := NewRateLimit(&ratelimit.Config{
		Enabled: true,
		Rules: map[string]ratelimit.RuleConfig{
			RuleNameAuthWrite: {
				Interval: time.Second,
			},
		},
	}, limiter)

	router := gin.New()
	router.POST("/limited", middleware.Cooldown(RuleNameAuthWrite), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodPost, "/limited", nil)
	req.RemoteAddr = "203.0.113.5:54321"
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", recorder.Code)
	}
}

func TestRateLimitRejectsTooManyRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)

	now := time.Unix(200, 0).UTC()
	middleware := NewRateLimit(&ratelimit.Config{
		Enabled: true,
		Rules: map[string]ratelimit.RuleConfig{
			RuleNameCommentWrite: {
				Interval: time.Minute,
				Limit:    2,
			},
		},
	}, &fakeLimiter{
		allowFn: func(ctx context.Context, key string, rule ratelimit.Rule) (*ratelimit.Decision, error) {
			return &ratelimit.Decision{
				Allowed:    false,
				Remaining:  0,
				RetryAfter: 1500 * time.Millisecond,
				ResetAt:    now.Add(1500 * time.Millisecond),
			}, nil
		},
	})

	router := gin.New()
	router.POST("/limited", middleware.Window(RuleNameCommentWrite), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/limited", nil))

	if recorder.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", recorder.Code)
	}
	if recorder.Body.String() != "{\"message\":\"请求过于频繁，请稍后再试\"}" {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
	if got := recorder.Header().Get("Retry-After"); got != "2" {
		t.Fatalf("expected Retry-After 2, got %q", got)
	}
	if got := recorder.Header().Get("X-RateLimit-Reset"); got != strconv.FormatInt(now.Add(1500*time.Millisecond).Unix(), 10) {
		t.Fatalf("unexpected X-RateLimit-Reset header: %q", got)
	}
}

func TestRateLimitUsesConfiguredRule(t *testing.T) {
	gin.SetMode(gin.TestMode)

	middleware := NewRateLimit(&ratelimit.Config{
		Enabled: true,
		Rules: map[string]ratelimit.RuleConfig{
			RuleNameCommentWrite: {
				Interval: time.Minute,
				Limit:    30,
			},
		},
	}, &fakeLimiter{
		allowFn: func(ctx context.Context, key string, rule ratelimit.Rule) (*ratelimit.Decision, error) {
			if rule.Name != RuleNameCommentWrite {
				t.Fatalf("expected rule name %q, got %q", RuleNameCommentWrite, rule.Name)
			}
			if rule.Mode != ratelimit.ModeWindow {
				t.Fatalf("expected mode %q, got %q", ratelimit.ModeWindow, rule.Mode)
			}
			if rule.Interval != time.Minute {
				t.Fatalf("expected interval 1m, got %s", rule.Interval)
			}
			if rule.Limit != 30 {
				t.Fatalf("expected limit 30, got %d", rule.Limit)
			}
			return &ratelimit.Decision{Allowed: true}, nil
		},
	})

	router := gin.New()
	router.POST("/limited", middleware.Window(RuleNameCommentWrite), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/limited", nil))

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", recorder.Code)
	}
}

func TestRateLimitSkipsUnconfiguredRule(t *testing.T) {
	gin.SetMode(gin.TestMode)

	middleware := NewRateLimit(&ratelimit.Config{
		Enabled: true,
		Rules: map[string]ratelimit.RuleConfig{
			RuleNameAuthWrite: {
				Interval: time.Second,
			},
		},
	}, &fakeLimiter{
		allowFn: func(ctx context.Context, key string, rule ratelimit.Rule) (*ratelimit.Decision, error) {
			t.Fatalf("unexpected allow call for unconfigured rule")
			return nil, nil
		},
	})

	router := gin.New()
	router.POST("/limited", middleware.Window(RuleNameProfileWrite), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/limited", nil))

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected 204 for unconfigured rule, got %d", recorder.Code)
	}
}

var _ ratelimit.Limiter = (*fakeLimiter)(nil)
