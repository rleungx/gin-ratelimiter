package ratelimiter

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConcurrencyLimiter(t *testing.T) {
	t.Parallel()

	assertions := require.New(t)
	limiter := newConcurrencyLimiter(10)

	// Test allowing up to the limit
	for range 10 {
		assertions.True(limiter.tryAcquire())
	}
	assertions.False(limiter.tryAcquire())

	// Test releasing and allowing again
	limiter.release()
	assertions.True(limiter.tryAcquire())

	// Test getting and setting the limit
	assertions.Equal(uint64(10), limiter.limitValue())
	limiter.updateLimit(5)
	assertions.Equal(uint64(5), limiter.limitValue())

	// Test getting the current count
	assertions.Equal(uint64(10), limiter.currentValue())
	limiter.release()
	assertions.Equal(uint64(9), limiter.currentValue())

	// Test setting limit to zero
	limiter.updateLimit(0)
	assertions.False(limiter.tryAcquire())
	assertions.Equal(uint64(0), limiter.limitValue())

	// Test setting limit to a very high value
	limiter.updateLimit(^uint64(0)) // Max uint64 value
	for range 100 {
		assertions.True(limiter.tryAcquire())
	}
	assertions.Equal(uint64(109), limiter.currentValue()) // 9 from previous tests + 100

	// Test releasing all
	for range 109 {
		limiter.release()
	}
	assertions.Equal(uint64(0), limiter.currentValue())

	// Additional edge cases
	// Test releasing when current is zero
	limiter.release()
	assertions.Equal(uint64(0), limiter.currentValue())

	// Test setting limit close to the uint64 upper bound.
	limiter.updateLimit(^uint64(0) - 1)
	assertions.Equal(^uint64(0)-1, limiter.limitValue())
}
