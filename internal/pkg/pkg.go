package pkg

import (
	"easydrop/internal/pkg/cache"
	"easydrop/internal/pkg/captcha"
	"easydrop/internal/pkg/cookie"
	"easydrop/internal/pkg/database"
	"easydrop/internal/pkg/email"
	"easydrop/internal/pkg/initsecret"
	"easydrop/internal/pkg/jwt"
	"easydrop/internal/pkg/oauth"
	"easydrop/internal/pkg/ratelimit"
	"easydrop/internal/pkg/redis"
	"easydrop/internal/pkg/storage"
	"easydrop/internal/pkg/token"
	"easydrop/internal/pkg/webauthn"

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
	initsecret.NewGuard,
	captcha.CaptchaSet,
	cookie.NewAuthCookie,
	webauthn.NewManager,
	oauth.NewManager,
)
