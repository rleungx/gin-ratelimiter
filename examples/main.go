package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	ratelimiter "github.com/rleungx/gin-ratelimiter"
)

func main() {
	r := gin.New()

	l := ratelimiter.NewLimiter()
	// Example ping request.
	r.GET("/ping", l.SetLimiter(ratelimiter.WithConcurrencyLimiter(1), ratelimiter.WithQPSLimiter(3, 3)),
		func(c *gin.Context) {
			c.String(http.StatusOK, "")
		})

	// Listen and Server in 0.0.0.0:8880
	r.Run(":8880")
}
