// Package ratelimiter provides Gin middleware for per-route rate and concurrency limiting.
package ratelimiter

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// Limiter manages per-route rate and concurrency limits.
//
// Routes are keyed by the pattern returned from gin.Context.FullPath.
type Limiter struct {
	rateLimiters        sync.Map
	concurrencyLimiters sync.Map
}

// New returns a limiter that can be shared across routes and updated at runtime.
func New() *Limiter {
	return &Limiter{}
}

// Middleware returns a Gin middleware that applies the provided options to the
// current route and rejects requests that exceed the configured thresholds.
func (limiter *Limiter) Middleware(opts ...Option) gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, opt := range opts {
			opt(c, limiter)
		}

		path := c.FullPath()
		concurrencyLimiter, allowed := limiter.allowRequest(path)
		if !allowed {
			c.AbortWithStatus(http.StatusTooManyRequests)
			return
		}

		if concurrencyLimiter != nil {
			defer concurrencyLimiter.release()
		}

		c.Next()
	}
}

// UpdateRateLimit updates the rate limiter for path. If the route does not yet
// have a rate limiter, one is created with the provided settings.
// Existing buckets retain their state, and rate and burst are updated together.
func (limiter *Limiter) UpdateRateLimit(path string, limit rate.Limit, burst int) {
	if rateLimiter, loaded := limiter.ensureRateLimiter(path, limit, burst); loaded {
		rateLimiter.updateLimit(limit, burst)
	}
}

// UpdateConcurrencyLimit updates the concurrency limiter for path. If the route
// does not yet have a concurrency limiter, one is created with the provided limit.
// Updating a route preserves its in-flight request count.
func (limiter *Limiter) UpdateConcurrencyLimit(path string, limit uint64) {
	if concurrencyLimiter, loaded := limiter.ensureConcurrencyLimiter(path, limit); loaded {
		concurrencyLimiter.updateLimit(limit)
	}
}

// RateLimitStatus returns the configured rate and burst for path.
// It returns zero values when the route has no rate limiter.
func (limiter *Limiter) RateLimitStatus(path string) (rate.Limit, int) {
	if rateLimiter, exists := limiter.loadRateLimiter(path); exists {
		return rateLimiter.status()
	}

	return 0, 0
}

// ConcurrencyLimitStatus returns the configured concurrency limit and current
// in-flight request count for path. It returns zero values when the route has
// no concurrency limiter.
func (limiter *Limiter) ConcurrencyLimitStatus(path string) (uint64, uint64) {
	if concurrencyLimiter, exists := limiter.loadConcurrencyLimiter(path); exists {
		return concurrencyLimiter.limitValue(), concurrencyLimiter.currentValue()
	}

	return 0, 0
}

func (limiter *Limiter) allowRequest(path string) (*concurrencyLimiter, bool) {
	concurrencyLimiter, exists := limiter.loadConcurrencyLimiter(path)
	if exists && !concurrencyLimiter.tryAcquire() {
		return nil, false
	}

	rateLimiter, exists := limiter.loadRateLimiter(path)
	if exists && !rateLimiter.allow() {
		if concurrencyLimiter != nil {
			concurrencyLimiter.release()
		}

		return nil, false
	}

	return concurrencyLimiter, true
}

// ensureRateLimiter returns the route's limiter and whether it already existed.
func (limiter *Limiter) ensureRateLimiter(path string, limit rate.Limit, burst int) (*rateLimiter, bool) {
	if existingLimiter, exists := limiter.loadRateLimiter(path); exists {
		return existingLimiter, true
	}

	newLimiter := newRateLimiter(limit, burst)
	actualLimiter, loaded := limiter.rateLimiters.LoadOrStore(path, newLimiter)
	return actualLimiter.(*rateLimiter), loaded
}

// ensureConcurrencyLimiter returns the route's limiter and whether it already existed.
func (limiter *Limiter) ensureConcurrencyLimiter(path string, limit uint64) (*concurrencyLimiter, bool) {
	if existingLimiter, exists := limiter.loadConcurrencyLimiter(path); exists {
		return existingLimiter, true
	}

	newLimiter := newConcurrencyLimiter(limit)
	actualLimiter, loaded := limiter.concurrencyLimiters.LoadOrStore(path, newLimiter)
	return actualLimiter.(*concurrencyLimiter), loaded
}

func (limiter *Limiter) loadRateLimiter(path string) (*rateLimiter, bool) {
	rawLimiter, exists := limiter.rateLimiters.Load(path)
	if !exists {
		return nil, false
	}

	return rawLimiter.(*rateLimiter), true
}

func (limiter *Limiter) loadConcurrencyLimiter(path string) (*concurrencyLimiter, bool) {
	rawLimiter, exists := limiter.concurrencyLimiters.Load(path)
	if !exists {
		return nil, false
	}

	return rawLimiter.(*concurrencyLimiter), true
}
