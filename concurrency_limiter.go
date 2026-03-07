package ratelimiter

import (
	"sync/atomic"
)

type concurrencyLimiter struct {
	current atomic.Uint64
	limit   atomic.Uint64
}

func newConcurrencyLimiter(limit uint64) *concurrencyLimiter {
	limiter := &concurrencyLimiter{}
	limiter.limit.Store(limit)
	return limiter
}

func (limiter *concurrencyLimiter) tryAcquire() bool {
	for {
		current := limiter.current.Load()
		limit := limiter.limit.Load()

		if current >= limit {
			return false
		}

		if limiter.current.CompareAndSwap(current, current+1) {
			if current+1 <= limiter.limit.Load() {
				return true
			}

			limiter.release()
			return false
		}
	}
}

func (limiter *concurrencyLimiter) release() {
	for {
		current := limiter.current.Load()
		if current == 0 {
			return
		}

		if limiter.current.CompareAndSwap(current, current-1) {
			return
		}
	}
}

func (limiter *concurrencyLimiter) limitValue() uint64 {
	return limiter.limit.Load()
}

func (limiter *concurrencyLimiter) updateLimit(limit uint64) {
	limiter.limit.Store(limit)
}

func (limiter *concurrencyLimiter) currentValue() uint64 {
	return limiter.current.Load()
}
