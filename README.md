# gin-ratelimiter

[![Run Tests](https://github.com/rleungx/gin-ratelimiter/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/rleungx/gin-ratelimiter/actions/workflows/go.yml)
[![codecov](https://codecov.io/gh/rleungx/gin-ratelimiter/branch/main/graph/badge.svg)](https://codecov.io/gh/rleungx/gin-ratelimiter)
[![Go Report Card](https://goreportcard.com/badge/github.com/rleungx/gin-ratelimiter)](https://goreportcard.com/report/github.com/rleungx/gin-ratelimiter)
[![Go Reference](https://pkg.go.dev/badge/github.com/rleungx/gin-ratelimiter.svg)](https://pkg.go.dev/github.com/rleungx/gin-ratelimiter)

Gin middleware for per-route rate limiting and concurrency limiting.

## Usage

```go
package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	ratelimiter "github.com/rleungx/gin-ratelimiter"
)

func main() {
	router := gin.New()
	// Reuse one limiter instance so limits can be updated at runtime.
	limiter := ratelimiter.New()

	router.GET(
		"/ping",
		limiter.Middleware(
			// Allow 3 requests per second with a burst of 3.
			ratelimiter.WithRateLimit(3, 3),
			// Allow only 1 in-flight request for this route.
			ratelimiter.WithConcurrencyLimit(1),
		),
		func(c *gin.Context) {
			c.String(http.StatusOK, "pong")
		},
	)

	router.Run(":8880")
}
```

## Runtime Update

```go
limiter.UpdateRateLimit("/jobs", 20, 40)
limiter.UpdateConcurrencyLimit("/jobs", 10)
```

Use the Gin route pattern when updating limits, for example `/users/:id`.

See [`examples/main.go`](./examples/main.go) for a runnable example.
