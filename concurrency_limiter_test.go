package ratelimiter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConcurrencyLimiter(t *testing.T) {
	t.Parallel()

	cl := newConcurrencyLimiter(10)

	for i := 0; i < 10; i++ {
		assert.True(t, cl.allow())
	}
	assert.False(t, cl.allow())
	cl.release()
	assert.True(t, cl.allow())
	assert.Equal(t, uint64(10), cl.getLimit())
	cl.setLimit(5)
	assert.Equal(t, uint64(5), cl.getLimit())
	assert.Equal(t, uint64(10), cl.getCurrent())
	cl.release()
	assert.Equal(t, uint64(9), cl.getCurrent())
}
