package ratelimiter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConcurrencyLimiter(t *testing.T) {
	t.Parallel()

	cl := newConcurrencyLimiter(10)

	// Test allowing up to the limit
	for range 10 {
		assert.True(t, cl.allow())
	}
	assert.False(t, cl.allow())

	// Test releasing and allowing again
	cl.release()
	assert.True(t, cl.allow())

	// Test getting and setting the limit
	assert.Equal(t, uint64(10), cl.getLimit())
	cl.setLimit(5)
	assert.Equal(t, uint64(5), cl.getLimit())

	// Test getting the current count
	assert.Equal(t, uint64(10), cl.getCurrent())
	cl.release()
	assert.Equal(t, uint64(9), cl.getCurrent())

	// Test setting limit to zero
	cl.setLimit(0)
	assert.False(t, cl.allow())
	assert.Equal(t, uint64(0), cl.getLimit())

	// Test setting limit to a very high value
	cl.setLimit(^uint64(0)) // Max uint64 value
	for range 100 {
		assert.True(t, cl.allow())
	}
	assert.Equal(t, uint64(109), cl.getCurrent()) // 9 from previous tests + 100

	// Test releasing all
	for range 109 {
		cl.release()
	}
	assert.Equal(t, uint64(0), cl.getCurrent())

	// Additional edge cases
	// Test releasing when current is zero
	cl.release()
	assert.Equal(t, uint64(0), cl.getCurrent())

	// Test setting limit to a negative value (should be handled gracefully)
	cl.setLimit(^uint64(0) - 1)
	assert.Equal(t, ^uint64(0)-1, cl.getLimit())
}
