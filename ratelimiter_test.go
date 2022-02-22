package ratelimiter

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
)

func TestUpdateConcurrencyLimiter(t *testing.T) {
	t.Parallel()

	testPath := "/test/concurrency"
	r := gin.New()
	l := NewLimiter()
	context.Background()
	r.GET(testPath, l.SetLimiter(WithConcurrencyLimiter(10)), func(c *gin.Context) {
		time.Sleep(2 * time.Second)
		c.JSON(http.StatusNoContent, nil)
	})

	req, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, testPath, nil)
	assert.NoError(t, err)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			checkResponseCode(t, r, req, http.StatusNoContent)
		}()
	}
	time.Sleep(200 * time.Millisecond)
	checkResponseCode(t, r, req, http.StatusTooManyRequests)
	wg.Wait()

	limit, current := l.GetConcurrencyLimiterStatus(testPath)
	assert.Equal(t, uint64(10), limit)
	assert.Equal(t, uint64(0), current)
	l.UpdateConcurrencyLimiter(testPath, 5)
	limit, current = l.GetConcurrencyLimiterStatus(testPath)
	assert.Equal(t, uint64(5), limit)
	assert.Equal(t, uint64(0), current)

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			checkResponseCode(t, r, req, http.StatusNoContent)
		}()
	}
	time.Sleep(200 * time.Millisecond)
	checkResponseCode(t, r, req, http.StatusTooManyRequests)
	wg.Wait()
}

func TestUpdateQPSLimiter(t *testing.T) {
	t.Parallel()

	testPath := "/test/qps"
	r := gin.New()
	l := NewLimiter()
	r.GET(testPath, l.SetLimiter(WithQPSLimiter(rate.Every(time.Second), 1)), func(c *gin.Context) {
		c.JSON(http.StatusNoContent, nil)
	})

	req, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, testPath, nil)
	assert.NoError(t, err)
	checkResponseCode(t, r, req, http.StatusNoContent)

	checkResponseCode(t, r, req, http.StatusTooManyRequests)

	limit, burst := l.GetQPSLimiterStatus(testPath)
	assert.Equal(t, rate.Limit(1), limit)
	assert.Equal(t, 1, burst)
	l.UpdateQPSLimiter(testPath, 5, 5)
	limit, burst = l.GetQPSLimiterStatus(testPath)
	assert.Equal(t, rate.Limit(5), limit)
	assert.Equal(t, 5, burst)
	time.Sleep(time.Second)

	for i := 0; i < 10; i++ {
		if i < 5 {
			checkResponseCode(t, r, req, http.StatusNoContent)
		} else {
			checkResponseCode(t, r, req, http.StatusTooManyRequests)
		}
	}
	checkResponseCode(t, r, req, http.StatusTooManyRequests)
}

func TestQPSLimiter(t *testing.T) {
	t.Parallel()

	testPath := "/test/qps"
	r := gin.New()
	l := NewLimiter()
	r.GET(testPath, l.SetLimiter(WithQPSLimiter(rate.Every(time.Second), 1)), func(c *gin.Context) {
		c.JSON(http.StatusNoContent, nil)
	})

	req, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, testPath, nil)
	assert.NoError(t, err)
	checkResponseCode(t, r, req, http.StatusNoContent)

	checkResponseCode(t, r, req, http.StatusTooManyRequests)
	time.Sleep(time.Second)
	checkResponseCode(t, r, req, http.StatusNoContent)
}

func checkResponseCode(t *testing.T, r http.Handler, req *http.Request, expectCode int) {
	t.Helper()

	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)
	assert.Equal(t, expectCode, res.Code)
}
