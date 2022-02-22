package ratelimiter

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type Limiter struct {
	qpsLimiter         sync.Map
	concurrencyLimiter sync.Map
}

func NewLimiter() *Limiter {
	return &Limiter{}
}

func (l *Limiter) SetLimiter(opts ...Option) gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, opt := range opts {
			opt(c, l)
		}

		path := c.Request.URL.Path
		if !l.Allow(path) {
			c.AbortWithStatus(http.StatusTooManyRequests)
			return
		}
		c.Next()
		if limiter, exist := l.concurrencyLimiter.Load(path); exist {
			if cl, ok := limiter.(*concurrencyLimiter); ok {
				cl.release()
			}
		}
	}
}

func (l *Limiter) Allow(path string) bool {
	var cl *concurrencyLimiter
	var ok bool
	if limiter, exist := l.concurrencyLimiter.Load(path); exist {
		if cl, ok = limiter.(*concurrencyLimiter); ok && !cl.allow() {
			return false
		}
	}

	if limiter, exist := l.qpsLimiter.Load(path); exist {
		if ql, ok := limiter.(*rate.Limiter); ok && !ql.Allow() {
			if cl != nil {
				cl.release()
			}
			return false
		}
	}

	return true
}

func (l *Limiter) UpdateQPSLimiter(path string, limit rate.Limit, burst int) {
	if limiter, exist := l.qpsLimiter.Load(path); exist {
		limiter.(*rate.Limiter).SetLimit(limit)
		limiter.(*rate.Limiter).SetBurst(burst)
	} else {
		l.qpsLimiter.Store(path, rate.NewLimiter(limit, burst))
	}
}

func (l *Limiter) UpdateConcurrencyLimiter(path string, limit uint64) {
	if limiter, exist := l.concurrencyLimiter.Load(path); exist {
		limiter.(*concurrencyLimiter).setLimit(limit)
	} else {
		l.concurrencyLimiter.Store(path, newConcurrencyLimiter(limit))
	}
}

func (l *Limiter) GetQPSLimiterStatus(path string) (limit rate.Limit, burst int) {
	if limiter, exist := l.qpsLimiter.Load(path); exist {
		return limiter.(*rate.Limiter).Limit(), limiter.(*rate.Limiter).Burst()
	}

	return 0, 0
}

func (l *Limiter) GetConcurrencyLimiterStatus(path string) (uint64, uint64) {
	if limiter, exist := l.concurrencyLimiter.Load(path); exist {
		return limiter.(*concurrencyLimiter).getLimit(), limiter.(*concurrencyLimiter).getCurrent()
	}

	return 0, 0
}
