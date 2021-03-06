package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	ratelimiter "github.com/rleungx/gin-ratelimiter"
)

func main() {
	r := gin.New()

	l := ratelimiter.NewLimiter()
	// Example ping request.
	r.GET("/ping", l.SetLimiter(ratelimiter.WithConcurrencyLimiter(1), ratelimiter.WithQPSLimiter(1, 10)),
		func(c *gin.Context) {
			c.String(http.StatusOK, "pong "+fmt.Sprint(time.Now().UnixNano()))
		})

	// Listen and Server in 0.0.0.0:8080
	r.Run(":8880")
}
