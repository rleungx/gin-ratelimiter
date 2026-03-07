package ratelimiter

import (
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// Option configures limits for the current route.
type Option func(*gin.Context, *Limiter)

// WithConcurrencyLimit ensures the current route has a concurrency limiter
// with the provided maximum number of in-flight requests.
func WithConcurrencyLimit(limit uint64) Option {
	return func(c *gin.Context, limiter *Limiter) {
		limiter.ensureConcurrencyLimiter(c.FullPath(), limit)
	}
}

// WithRateLimit ensures the current route has a token-bucket rate limiter with
// the provided rate and burst values.
func WithRateLimit(limit rate.Limit, burst int) Option {
	return func(c *gin.Context, limiter *Limiter) {
		limiter.ensureRateLimiter(c.FullPath(), limit, burst)
	}
}
