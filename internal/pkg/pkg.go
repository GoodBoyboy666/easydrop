package pkg

import (
	"easydrop/internal/pkg/cache"
	"easydrop/internal/pkg/captcha"
	"easydrop/internal/pkg/cookie"
	"easydrop/internal/pkg/database"
	"easydrop/internal/pkg/email"
	"easydrop/internal/pkg/jwt"
	"easydrop/internal/pkg/ratelimit"
	"easydrop/internal/pkg/redis"
	"easydrop/internal/pkg/storage"
	"easydrop/internal/pkg/token"

	"github.com/google/wire"
)

var Pkgset = wire.NewSet(
	database.NewDB,
	redis.NewOptionalClient,
	ratelimit.NewLimiter,
	cache.NewCache,
	email.NewClient,
	jwt.NewManager,
	storage.NewManager,
	token.NewManager,
	captcha.CaptchaSet,
	cookie.NewAuthCookie,
)