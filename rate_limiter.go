package ratelimiter

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type rateLimiter struct {
	// Keep configuration pairs atomic with respect to updates, admission and status.
	mu     sync.Mutex
	bucket *rate.Limiter
}

func newRateLimiter(limit rate.Limit, burst int) *rateLimiter {
	return &rateLimiter{bucket: rate.NewLimiter(limit, burst)}
}

func (limiter *rateLimiter) allow() bool {
	limiter.mu.Lock()
	defer limiter.mu.Unlock()

	return limiter.bucket.Allow()
}

func (limiter *rateLimiter) updateLimit(limit rate.Limit, burst int) {
	limiter.mu.Lock()
	defer limiter.mu.Unlock()

	// Advance the bucket once, so no time accrues under a partially updated pair.
	now := time.Now()
	limiter.bucket.SetLimitAt(now, limit)
	limiter.bucket.SetBurstAt(now, burst)
}

func (limiter *rateLimiter) status() (rate.Limit, int) {
	limiter.mu.Lock()
	defer limiter.mu.Unlock()

	return limiter.bucket.Limit(), limiter.bucket.Burst()
}
