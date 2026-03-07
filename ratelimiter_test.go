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

func TestUpdateConcurrencyLimit(t *testing.T) {
	t.Parallel()

	assertions := require.New(t)
	routePath := "/test/concurrency"
	router := gin.New()
	limiter := New()
	router.GET(routePath, limiter.Middleware(WithConcurrencyLimit(10)), func(c *gin.Context) {
		time.Sleep(2 * time.Second)
		c.JSON(http.StatusNoContent, nil)
	})

	request, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, routePath, nil)
	assertions.NoError(err)

	var waitGroup sync.WaitGroup
	for range 10 {
		waitGroup.Go(func() {
			assertResponseCode(assertions, router, request, http.StatusNoContent)
		})
	}
	time.Sleep(200 * time.Millisecond)
	assertResponseCode(assertions, router, request, http.StatusTooManyRequests)
	waitGroup.Wait()

	limit, current := limiter.ConcurrencyLimitStatus(routePath)
	assertions.Equal(uint64(10), limit)
	assertions.Equal(uint64(0), current)
	limiter.UpdateConcurrencyLimit(routePath, 5)
	limit, current = limiter.ConcurrencyLimitStatus(routePath)
	assertions.Equal(uint64(5), limit)
	assertions.Equal(uint64(0), current)

	for range 5 {
		waitGroup.Go(func() {
			assertResponseCode(assertions, router, request, http.StatusNoContent)
		})
	}
	time.Sleep(200 * time.Millisecond)
	assertResponseCode(assertions, router, request, http.StatusTooManyRequests)
	waitGroup.Wait()

	limiter.UpdateConcurrencyLimit(routePath, 0)
	limit, current = limiter.ConcurrencyLimitStatus(routePath)
	assertions.Equal(uint64(0), limit)
	assertions.Equal(uint64(0), current)
	assertResponseCode(assertions, router, request, http.StatusTooManyRequests)

	newRoutePath := "/test/new-concurrency"
	limit, current = limiter.ConcurrencyLimitStatus(newRoutePath)
	assertions.Equal(uint64(0), limit)
	assertions.Equal(uint64(0), current)
	limiter.UpdateConcurrencyLimit(newRoutePath, 3)
	router.GET(newRoutePath, limiter.Middleware(), func(c *gin.Context) {
		time.Sleep(2 * time.Second)
		c.JSON(http.StatusNoContent, nil)
	})

	request, err = http.NewRequestWithContext(context.TODO(), http.MethodGet, newRoutePath, nil)
	assertions.NoError(err)

	for range 3 {
		waitGroup.Go(func() {
			assertResponseCode(assertions, router, request, http.StatusNoContent)
		})
	}
	time.Sleep(200 * time.Millisecond)
	assertResponseCode(assertions, router, request, http.StatusTooManyRequests)
	waitGroup.Wait()
}

func TestUpdateRateLimit(t *testing.T) {
	t.Parallel()

	assertions := require.New(t)
	routePath := "/test/rate"
	router := gin.New()
	limiter := New()
	router.GET(routePath, limiter.Middleware(WithRateLimit(rate.Every(time.Second), 1)), func(c *gin.Context) {
		c.JSON(http.StatusNoContent, nil)
	})

	request, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, routePath, nil)
	assertions.NoError(err)
	assertResponseCode(assertions, router, request, http.StatusNoContent)
	assertResponseCode(assertions, router, request, http.StatusTooManyRequests)

	limit, burst := limiter.RateLimitStatus(routePath)
	//nolint:testifylint
	assertions.Equal(rate.Limit(1), limit)
	assertions.Equal(1, burst)
	limiter.UpdateRateLimit(routePath, 5, 5)
	limit, burst = limiter.RateLimitStatus(routePath)
	//nolint:testifylint
	assertions.Equal(rate.Limit(5), limit)
	assertions.Equal(5, burst)
	time.Sleep(time.Second)

	for index := range 10 {
		if index < 5 {
			assertResponseCode(assertions, router, request, http.StatusNoContent)
		} else {
			assertResponseCode(assertions, router, request, http.StatusTooManyRequests)
		}
	}
	assertResponseCode(assertions, router, request, http.StatusTooManyRequests)

	limiter.UpdateRateLimit(routePath, 0, 0)
	limit, burst = limiter.RateLimitStatus(routePath)
	//nolint:testifylint
	assertions.Equal(rate.Limit(0), limit)
	assertions.Equal(0, burst)
	assertResponseCode(assertions, router, request, http.StatusTooManyRequests)

	newRoutePath := "/test/new-rate"
	limit, burst = limiter.RateLimitStatus(newRoutePath)
	//nolint:testifylint
	assertions.Equal(rate.Limit(0), limit)
	assertions.Equal(0, burst)
	limiter.UpdateRateLimit(newRoutePath, 2, 2)
	router.GET(newRoutePath, limiter.Middleware(), func(c *gin.Context) {
		c.JSON(http.StatusNoContent, nil)
	})

	request, err = http.NewRequestWithContext(context.TODO(), http.MethodGet, newRoutePath, nil)
	assertions.NoError(err)

	for range 2 {
		assertResponseCode(assertions, router, request, http.StatusNoContent)
	}
	assertResponseCode(assertions, router, request, http.StatusTooManyRequests)
}

func TestRateLimit(t *testing.T) {
	t.Parallel()

	assertions := require.New(t)
	routePath := "/test/rate"
	router := gin.New()
	limiter := New()
	router.GET(routePath, limiter.Middleware(WithRateLimit(rate.Every(time.Second), 1)), func(c *gin.Context) {
		c.JSON(http.StatusNoContent, nil)
	})

	request, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, routePath, nil)
	assertions.NoError(err)
	assertResponseCode(assertions, router, request, http.StatusNoContent)
	assertResponseCode(assertions, router, request, http.StatusTooManyRequests)
	time.Sleep(time.Second)
	assertResponseCode(assertions, router, request, http.StatusNoContent)

	for range 5 {
		assertResponseCode(assertions, router, request, http.StatusTooManyRequests)
	}
}

func TestReleaseConcurrencyLimit(t *testing.T) {
	t.Parallel()

	assertions := require.New(t)
	routePath := "/test/concurrency"
	router := gin.New()
	limiter := New()
	router.GET(routePath, limiter.Middleware(WithConcurrencyLimit(10), WithRateLimit(3, 3)), func(c *gin.Context) {
		time.Sleep(2 * time.Second)
		c.JSON(http.StatusNoContent, nil)
	})

	request, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, routePath, nil)
	assertions.NoError(err)

	var waitGroup sync.WaitGroup
	for range 3 {
		waitGroup.Go(func() {
			assertResponseCode(assertions, router, request, http.StatusNoContent)
		})
	}
	time.Sleep(200 * time.Millisecond)
	assertResponseCode(assertions, router, request, http.StatusTooManyRequests)
	waitGroup.Wait()
}

func assertResponseCode(assertions *require.Assertions, handler http.Handler, request *http.Request, expectedCode int) {
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	assertions.Equal(expectedCode, response.Code)
}
