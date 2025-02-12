package ratelimiter

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConcurrencyLimiter(t *testing.T) {
	t.Parallel()

	re := require.New(t)
	cl := newConcurrencyLimiter(10)

	// Test allowing up to the limit
	for range 10 {
		re.True(cl.allow())
	}
	re.False(cl.allow())

	// Test releasing and allowing again
	cl.release()
	re.True(cl.allow())

	// Test getting and setting the limit
	re.Equal(uint64(10), cl.getLimit())
	cl.setLimit(5)
	re.Equal(uint64(5), cl.getLimit())

	// Test getting the current count
	re.Equal(uint64(10), cl.getCurrent())
	cl.release()
	re.Equal(uint64(9), cl.getCurrent())

	// Test setting limit to zero
	cl.setLimit(0)
	re.False(cl.allow())
	re.Equal(uint64(0), cl.getLimit())

	// Test setting limit to a very high value
	cl.setLimit(^uint64(0)) // Max uint64 value
	for range 100 {
		re.True(cl.allow())
	}
	re.Equal(uint64(109), cl.getCurrent()) // 9 from previous tests + 100

	// Test releasing all
	for range 109 {
		cl.release()
	}
	re.Equal(uint64(0), cl.getCurrent())

	// Additional edge cases
	// Test releasing when current is zero
	cl.release()
	re.Equal(uint64(0), cl.getCurrent())

	// Test setting limit to a negative value (should be handled gracefully)
	cl.setLimit(^uint64(0) - 1)
	re.Equal(^uint64(0)-1, cl.getLimit())
}
