package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	ratelimiter "github.com/rleungx/gin-ratelimiter"
)

func main() {
	router := gin.New()
	// Keep a shared limiter so you can call UpdateRateLimit / UpdateConcurrencyLimit later.
	limiter := ratelimiter.New()

	router.GET(
		"/ping",
		limiter.Middleware(
			// 3 requests per second, burst size 3.
			ratelimiter.WithRateLimit(3, 3),
			// Only 1 request can be in flight for /ping at a time.
			ratelimiter.WithConcurrencyLimit(1),
		),
		func(c *gin.Context) {
			c.String(http.StatusOK, "pong")
		},
	)

	router.Run(":8880")
}
