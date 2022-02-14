package ratelimiter

import "sync"

type concurrencyLimiter struct {
	mu      sync.RWMutex
	current uint64
	limit   uint64
}

func newConcurrencyLimiter(limit uint64) *concurrencyLimiter {
	return &concurrencyLimiter{limit: limit}
}

func (l *concurrencyLimiter) allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.current+1 <= l.limit {
		l.current++
		return true
	}
	return false
}

func (l *concurrencyLimiter) release() {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.current > 0 {
		l.current--
	}
}

func (l *concurrencyLimiter) getLimit() uint64 {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.limit
}

func (l *concurrencyLimiter) setLimit(limit uint64) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.limit = limit
}

func (l *concurrencyLimiter) getCurrent() uint64 {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.current
}
