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

func (limiter *Limiter) allowRequest(path string) (*concurrencyLimiter, bool) {
	concurrencyLimiter, exists := limiter.loadConcurrencyLimiter(path)
	if exists && !concurrencyLimiter.tryAcquire() {
		return nil, false
	}

	rateLimiter, exists := limiter.loadRateLimiter(path)
	if exists && !rateLimiter.Allow() {
		if concurrencyLimiter != nil {
			concurrencyLimiter.release()
		}
		return nil, false
	}

	return concurrencyLimiter, true
}

// UpdateRateLimit updates the rate limiter for path. If the route does not yet
// have a rate limiter, one is created with the provided settings.
func (limiter *Limiter) UpdateRateLimit(path string, limit rate.Limit, burst int) {
	if rateLimiter, exists := limiter.loadRateLimiter(path); exists {
		rateLimiter.SetLimit(limit)
		rateLimiter.SetBurst(burst)
		return
	}

	limiter.rateLimiters.Store(path, rate.NewLimiter(limit, burst))
}

// UpdateConcurrencyLimit updates the concurrency limiter for path. If the route
// does not yet have a concurrency limiter, one is created with the provided limit.
func (limiter *Limiter) UpdateConcurrencyLimit(path string, limit uint64) {
	if concurrencyLimiter, exists := limiter.loadConcurrencyLimiter(path); exists {
		concurrencyLimiter.updateLimit(limit)
		return
	}

	limiter.concurrencyLimiters.Store(path, newConcurrencyLimiter(limit))
}

// RateLimitStatus returns the configured rate and burst for path.
// It returns zero values when the route has no rate limiter.
func (limiter *Limiter) RateLimitStatus(path string) (rate.Limit, int) {
	if rateLimiter, exists := limiter.loadRateLimiter(path); exists {
		return rateLimiter.Limit(), rateLimiter.Burst()
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

func (limiter *Limiter) ensureRateLimiter(path string, limit rate.Limit, burst int) *rate.Limiter {
	if existingLimiter, exists := limiter.loadRateLimiter(path); exists {
		return existingLimiter
	}

	newLimiter := rate.NewLimiter(limit, burst)
	actualLimiter, _ := limiter.rateLimiters.LoadOrStore(path, newLimiter)
	return actualLimiter.(*rate.Limiter)
}

func (limiter *Limiter) ensureConcurrencyLimiter(path string, limit uint64) *concurrencyLimiter {
	if existingLimiter, exists := limiter.loadConcurrencyLimiter(path); exists {
		return existingLimiter
	}

	newLimiter := newConcurrencyLimiter(limit)
	actualLimiter, _ := limiter.concurrencyLimiters.LoadOrStore(path, newLimiter)
	return actualLimiter.(*concurrencyLimiter)
}

func (limiter *Limiter) loadRateLimiter(path string) (*rate.Limiter, bool) {
	rawLimiter, exists := limiter.rateLimiters.Load(path)
	if !exists {
		return nil, false
	}

	return rawLimiter.(*rate.Limiter), true
}

func (limiter *Limiter) loadConcurrencyLimiter(path string) (*concurrencyLimiter, bool) {
	rawLimiter, exists := limiter.concurrencyLimiters.Load(path)
	if !exists {
		return nil, false
	}

	return rawLimiter.(*concurrencyLimiter), true
}
