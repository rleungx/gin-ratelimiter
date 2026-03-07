package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	ratelimiter "github.com/rleungx/gin-ratelimiter"
)

func main() {
	router := gin.New()
	limiter := ratelimiter.New()

	router.GET(
		"/ping",
		limiter.Middleware(
			ratelimiter.WithRateLimit(3, 3),
			ratelimiter.WithConcurrencyLimit(1),
		),
		func(c *gin.Context) {
			c.String(http.StatusOK, "pong")
		},
	)

	router.Run(":8880")
}
