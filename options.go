package ratelimiter

import (
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type Option func(*gin.Context, *Limiter)

func WithConcurrencyLimiter(limit uint64) Option {
	return func(c *gin.Context, l *Limiter) {
		path := c.Request.URL.Path
		l.concurrencyLimiter.LoadOrStore(path, newConcurrencyLimiter(limit))
	}
}

func WithQPSLimiter(limit rate.Limit, burst int) Option {
	return func(c *gin.Context, l *Limiter) {
		path := c.Request.URL.Path
		l.qpsLimiter.LoadOrStore(path, rate.NewLimiter(limit, burst))
	}
}
