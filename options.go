package ratelimiter

import (
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type Option func(*gin.Context, *Limiter)

func WithConcurrencyLimiter(limit uint64) Option {
	return func(c *gin.Context, l *Limiter) {
		// Ignore the return value since we don't care about it.
		l.concurrencyLimiter.LoadOrStore(c.Request.URL.Path, newConcurrencyLimiter(limit))
	}
}

func WithQPSLimiter(limit rate.Limit, burst int) Option {
	return func(c *gin.Context, l *Limiter) {
		// Ignore the return value since we don't care about it.
		l.qpsLimiter.LoadOrStore(c.Request.URL.Path, rate.NewLimiter(limit, burst))
	}
}
