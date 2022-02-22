package ratelimiter

import (
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// Option is used to create a limiter with the optional settings.
type Option func(*gin.Context, *Limiter)

// WithConcurrencyLimiter creates a concurrency limiter for a given path if it doesn't exist.
func WithConcurrencyLimiter(limit uint64) Option {
	return func(c *gin.Context, l *Limiter) {
		// Ignore the return value since we don't care about it.
		l.concurrencyLimiter.LoadOrStore(c.FullPath(), newConcurrencyLimiter(limit))
	}
}

// WithConcurrencyLimiter creates a QPS limiter for a given path if it doesn't exist.
func WithQPSLimiter(limit rate.Limit, burst int) Option {
	return func(c *gin.Context, l *Limiter) {
		// Ignore the return value since we don't care about it.
		l.qpsLimiter.LoadOrStore(c.FullPath(), rate.NewLimiter(limit, burst))
	}
}
