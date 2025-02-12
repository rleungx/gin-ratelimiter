package ratelimiter

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"
)

func TestUpdateConcurrencyLimiter(t *testing.T) {
	t.Parallel()

	re := require.New(t)
	testPath := "/test/concurrency"
	r := gin.New()
	l := NewLimiter()
	r.GET(testPath, l.SetLimiter(WithConcurrencyLimiter(10)), func(c *gin.Context) {
		time.Sleep(2 * time.Second)
		c.JSON(http.StatusNoContent, nil)
	})

	req, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, testPath, nil)
	re.NoError(err)

	var wg sync.WaitGroup
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			checkResponseCode(re, r, req, http.StatusNoContent)
		}()
	}
	time.Sleep(200 * time.Millisecond)
	checkResponseCode(re, r, req, http.StatusTooManyRequests)
	wg.Wait()

	limit, current := l.GetConcurrencyLimiterStatus(testPath)
	re.Equal(uint64(10), limit)
	re.Equal(uint64(0), current)
	l.UpdateConcurrencyLimiter(testPath, 5)
	limit, current = l.GetConcurrencyLimiterStatus(testPath)
	re.Equal(uint64(5), limit)
	re.Equal(uint64(0), current)

	for range 5 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			checkResponseCode(re, r, req, http.StatusNoContent)
		}()
	}
	time.Sleep(200 * time.Millisecond)
	checkResponseCode(re, r, req, http.StatusTooManyRequests)
	wg.Wait()

	l.UpdateConcurrencyLimiter(testPath, 0)
	limit, current = l.GetConcurrencyLimiterStatus(testPath)
	re.Equal(uint64(0), limit)
	re.Equal(uint64(0), current)
	checkResponseCode(re, r, req, http.StatusTooManyRequests)

	// // Store new concurrency limiter
	newPath := "/test/new-concurrency"
	limit, current = l.GetConcurrencyLimiterStatus(newPath)
	re.Equal(uint64(0), limit)
	re.Equal(uint64(0), current)
	l.UpdateConcurrencyLimiter(newPath, 3)
	r.GET(newPath, l.SetLimiter(), func(c *gin.Context) {
		time.Sleep(2 * time.Second)
		c.JSON(http.StatusNoContent, nil)
	})

	req, err = http.NewRequestWithContext(context.TODO(), http.MethodGet, newPath, nil)
	re.NoError(err)

	for range 3 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			checkResponseCode(re, r, req, http.StatusNoContent)
		}()
	}
	time.Sleep(200 * time.Millisecond)
	checkResponseCode(re, r, req, http.StatusTooManyRequests)
	wg.Wait()
}

func TestUpdateQPSLimiter(t *testing.T) {
	t.Parallel()

	re := require.New(t)
	testPath := "/test/qps"
	r := gin.New()
	l := NewLimiter()
	r.GET(testPath, l.SetLimiter(WithQPSLimiter(rate.Every(time.Second), 1)), func(c *gin.Context) {
		c.JSON(http.StatusNoContent, nil)
	})

	req, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, testPath, nil)
	re.NoError(err)
	checkResponseCode(re, r, req, http.StatusNoContent)

	checkResponseCode(re, r, req, http.StatusTooManyRequests)

	limit, burst := l.GetQPSLimiterStatus(testPath)
	//nolint:testifylint
	re.Equal(rate.Limit(1), limit)
	re.Equal(1, burst)
	l.UpdateQPSLimiter(testPath, 5, 5)
	limit, burst = l.GetQPSLimiterStatus(testPath)
	//nolint:testifylint
	re.Equal(rate.Limit(5), limit)
	re.Equal(5, burst)
	time.Sleep(time.Second)

	for i := range 10 {
		if i < 5 {
			checkResponseCode(re, r, req, http.StatusNoContent)
		} else {
			checkResponseCode(re, r, req, http.StatusTooManyRequests)
		}
	}
	checkResponseCode(re, r, req, http.StatusTooManyRequests)

	l.UpdateQPSLimiter(testPath, 0, 0)
	limit, burst = l.GetQPSLimiterStatus(testPath)
	//nolint:testifylint
	re.Equal(rate.Limit(0), limit)
	re.Equal(0, burst)
	checkResponseCode(re, r, req, http.StatusTooManyRequests)

	// Store new QPS limiter
	newPath := "/test/new-qps"
	limit, burst = l.GetQPSLimiterStatus(newPath)
	//nolint:testifylint
	re.Equal(rate.Limit(0), limit)
	re.Equal(0, burst)
	l.UpdateQPSLimiter(newPath, 2, 2)
	r.GET(newPath, l.SetLimiter(), func(c *gin.Context) {
		c.JSON(http.StatusNoContent, nil)
	})

	req, err = http.NewRequestWithContext(context.TODO(), http.MethodGet, newPath, nil)
	re.NoError(err)

	for range 2 {
		checkResponseCode(re, r, req, http.StatusNoContent)
	}
	checkResponseCode(re, r, req, http.StatusTooManyRequests)
}

func TestQPSLimiter(t *testing.T) {
	t.Parallel()

	re := require.New(t)
	testPath := "/test/qps"
	r := gin.New()
	l := NewLimiter()
	r.GET(testPath, l.SetLimiter(WithQPSLimiter(rate.Every(time.Second), 1)), func(c *gin.Context) {
		c.JSON(http.StatusNoContent, nil)
	})

	req, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, testPath, nil)
	re.NoError(err)
	checkResponseCode(re, r, req, http.StatusNoContent)

	checkResponseCode(re, r, req, http.StatusTooManyRequests)
	time.Sleep(time.Second)
	checkResponseCode(re, r, req, http.StatusNoContent)

	for range 5 {
		checkResponseCode(re, r, req, http.StatusTooManyRequests)
	}
}

func TestReleaseConcurrencyLimiter(t *testing.T) {
	t.Parallel()

	re := require.New(t)
	testPath := "/test/concurrency"
	r := gin.New()
	l := NewLimiter()
	r.GET(testPath, l.SetLimiter(WithConcurrencyLimiter(10), WithQPSLimiter(3, 3)), func(c *gin.Context) {
		time.Sleep(2 * time.Second)
		c.JSON(http.StatusNoContent, nil)
	})

	req, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, testPath, nil)
	re.NoError(err)

	var wg sync.WaitGroup
	for range 3 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			checkResponseCode(re, r, req, http.StatusNoContent)
		}()
	}
	time.Sleep(200 * time.Millisecond)
	checkResponseCode(re, r, req, http.StatusTooManyRequests)
	wg.Wait()
}

func checkResponseCode(re *require.Assertions, r http.Handler, req *http.Request, expectCode int) {
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)
	re.Equal(expectCode, res.Code)
}
