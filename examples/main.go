// Package main shows how to attach the ratelimiter middleware to a Gin router.
package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	ratelimiter "github.com/rleungx/gin-ratelimiter"
)

func main() {
	const (
		requestsPerSecond = 3
		burstSize         = 3
		maxInFlight       = 1
		serverAddress     = ":8880"
	)

	router := gin.New()
	// Keep a shared limiter so you can call UpdateRateLimit / UpdateConcurrencyLimit later.
	limiter := ratelimiter.New()

	router.GET(
		"/ping",
		limiter.Middleware(
			// 3 requests per second, burst size 3.
			ratelimiter.WithRateLimit(requestsPerSecond, burstSize),
			// Only 1 request can be in flight for /ping at a time.
			ratelimiter.WithConcurrencyLimit(maxInFlight),
		),
		func(c *gin.Context) {
			c.String(http.StatusOK, "pong")
		},
	)

	err := router.Run(serverAddress)
	if err != nil {
		panic(err)
	}
}
