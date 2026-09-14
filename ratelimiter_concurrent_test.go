package ratelimiter

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"
)

func TestConcurrentInitializationPreservesConcurrency(t *testing.T) {
	t.Parallel()

	const (
		path       = "/concurrent-initialization"
		iterations = 1000
		workers    = 16
	)

	for iteration := range iterations {
		var limiter Limiter
		start := make(chan struct{})
		acquired := make(chan *concurrencyLimiter, workers)
		var waitGroup sync.WaitGroup
		for worker := range workers {
			waitGroup.Go(func() {
				<-start
				if worker%2 == 0 {
					limiter.UpdateConcurrencyLimit(path, 1)
				} else {
					limiter.ensureConcurrencyLimiter(path, 1)
				}

				if slot, allowed := limiter.allowRequest(path); allowed {
					acquired <- slot
				}
			})
		}
		close(start)
		waitGroup.Wait()
		close(acquired)

		// Hold every admitted slot until all initializers and updaters have finished.
		limit, current := limiter.ConcurrencyLimitStatus(path)
		admitted := len(acquired)
		for slot := range acquired {
			slot.release()
		}
		require.Equal(t, 1, admitted, "iteration %d", iteration)
		require.Equal(t, uint64(1), limit)
		require.Equal(t, uint64(1), current)
		_, current = limiter.ConcurrencyLimitStatus(path)
		require.Zero(t, current)

		slot, allowed := limiter.allowRequest(path)
		require.True(t, allowed, "released capacity must be reusable")
		slot.release()
	}
}

func TestConcurrentInitializationPreservesTokens(t *testing.T) {
	t.Parallel()

	const (
		path       = "/concurrent-token-initialization"
		iterations = 1000
		workers    = 16
	)

	for iteration := range iterations {
		var limiter Limiter
		start := make(chan struct{})
		admitted := make(chan bool, workers)
		var waitGroup sync.WaitGroup
		for worker := range workers {
			waitGroup.Go(func() {
				<-start
				if worker%2 == 0 {
					limiter.UpdateRateLimit(path, 0, 1)
				} else {
					limiter.ensureRateLimiter(path, 0, 1)
				}

				_, allowed := limiter.allowRequest(path)
				admitted <- allowed
			})
		}
		close(start)
		waitGroup.Wait()
		close(admitted)

		allowedCount := 0
		for allowed := range admitted {
			if allowed {
				allowedCount++
			}
		}
		// A zero refill rate makes any second admission an actual bucket reset.
		require.Equal(t, 1, allowedCount, "iteration %d", iteration)
		_, allowed := limiter.allowRequest(path)
		require.False(t, allowed, "updates must not replenish a consumed token")
	}
}

func TestConcurrentRateLimitUpdates(t *testing.T) {
	t.Parallel()

	const (
		path       = "/concurrent-rate-updates"
		iterations = 1000
	)
	limiter := New()
	limiter.UpdateRateLimit(path, 100, 1)

	for iteration := range iterations {
		start := make(chan struct{})
		var waitGroup sync.WaitGroup
		waitGroup.Go(func() {
			<-start
			limiter.UpdateRateLimit(path, 100, 1)
		})
		waitGroup.Go(func() {
			<-start
			limiter.UpdateRateLimit(path, 1, 100)
		})
		close(start)
		waitGroup.Wait()

		limit, burst := limiter.RateLimitStatus(path)
		valid := (limit == 100 && burst == 1) || (limit == 1 && burst == 100)
		require.True(t, valid, "iteration %d left an unrequested pair (%v, %d)", iteration, limit, burst)
	}
}

func TestConcurrentRateLimitUpdatesKeepClosedRouteClosed(t *testing.T) {
	t.Parallel()

	const (
		path       = "/closed-route"
		iterations = 10000
		refillRate = rate.Limit(1000000)
		burstSize  = 100
	)
	limiter := New()
	limiter.UpdateRateLimit(path, 0, 0)
	start := make(chan struct{})
	var waitGroup sync.WaitGroup
	waitGroup.Go(func() {
		<-start
		for range iterations {
			limiter.UpdateRateLimit(path, 0, burstSize)
		}
	})
	waitGroup.Go(func() {
		<-start
		for range iterations {
			limiter.UpdateRateLimit(path, refillRate, 0)
		}
	})
	close(start)

	// Neither policy permits admission: one never refills the empty bucket, and
	// the other has no burst capacity. A mixed pair would open the route.
	admitted := 0
	mixedStatus := false
	for range iterations {
		if _, allowed := limiter.allowRequest(path); allowed {
			admitted++
		}
		limit, burst := limiter.RateLimitStatus(path)
		if limit != 0 && burst != 0 {
			mixedStatus = true
		}
	}
	waitGroup.Wait()
	require.Zero(t, admitted, "a request observed a partially updated policy")
	require.False(t, mixedStatus, "status observed a partially updated policy")
	_, allowed := limiter.allowRequest(path)
	require.False(t, allowed, "the final policy must still reject requests")
}
