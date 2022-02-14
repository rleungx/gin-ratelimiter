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
	return &Limiter{
		qpsLimiter:         sync.Map{},
		concurrencyLimiter: sync.Map{},
	}
}

func (l *Limiter) SetLimiter(opts ...Option) gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, opt := range opts {
			opt(c, l)
		}

		path := c.Request.URL.Path
		if !l.Allow(path) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, path)
			return
		}
		c.Next()
		cl, exist := l.concurrencyLimiter.Load(path)
		if exist {
			cl.(*concurrencyLimiter).release()
		}
	}
}

func (l *Limiter) Allow(path string) bool {
	v1, exist1 := l.qpsLimiter.Load(path)
	v2, exist2 := l.concurrencyLimiter.Load(path)

	if !exist1 && !exist2 {
		return false
	}
	if exist1 && exist2 {
		return v2.(*concurrencyLimiter).allow() && v1.(*rate.Limiter).Allow()
	}
	if exist1 {
		return v1.(*rate.Limiter).Allow()
	}
	if exist2 {
		return v2.(*concurrencyLimiter).allow()
	}

	return false
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
