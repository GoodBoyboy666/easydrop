package middleware

import "github.com/google/wire"

var MiddlewareSet = wire.NewSet(
	NewAuth,
	NewSecurityHeaders,
	NewRateLimit,
	NewRequestBodyLimit,
)