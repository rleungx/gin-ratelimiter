package ratelimiter

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
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
			res := httptest.NewRecorder()
			r.ServeHTTP(res, req)
			assert.Equal(t, http.StatusNoContent, res.Code)
		}()
	}
	time.Sleep(200 * time.Millisecond)
	failTooManyRequests(t, r, req, testPath)
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
			res := httptest.NewRecorder()
			r.ServeHTTP(res, req)
			assert.Equal(t, http.StatusNoContent, res.Code)
		}()
	}
	time.Sleep(200 * time.Millisecond)
	failTooManyRequests(t, r, req, testPath)
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
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)
	assert.Equal(t, http.StatusNoContent, res.Code)

	failTooManyRequests(t, r, req, testPath)

	limit, burst := l.GetQPSLimiterStatus(testPath)
	assert.Equal(t, rate.Limit(1), limit)
	assert.Equal(t, 1, burst)
	l.UpdateQPSLimiter(testPath, 5, 5)
	limit, burst = l.GetQPSLimiterStatus(testPath)
	assert.Equal(t, rate.Limit(5), limit)
	assert.Equal(t, 5, burst)
	time.Sleep(time.Second)

	for i := 0; i < 10; i++ {
		res = httptest.NewRecorder()
		r.ServeHTTP(res, req)
		if i < 5 {
			assert.Equal(t, http.StatusNoContent, res.Code)
		} else {
			assert.Equal(t, http.StatusTooManyRequests, res.Code)
		}
	}
	failTooManyRequests(t, r, req, testPath)
}

func failTooManyRequests(t *testing.T, r http.Handler, req *http.Request, testPath string) {
	t.Helper()

	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)
	assert.Equal(t, http.StatusTooManyRequests, res.Code)
	actual, err := strconv.Unquote(res.Body.String())
	assert.NoError(t, err)
	assert.Equal(t, testPath, actual)
}
