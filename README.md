# gin-ratelimiter

[![Run Tests](https://github.com/rleungx/gin-ratelimiter/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/rleungx/gin-ratelimiter/actions/workflows/go.yml)
[![codecov](https://codecov.io/gh/rleungx/gin-ratelimiter/branch/main/graph/badge.svg)](https://codecov.io/gh/rleungx/gin-ratelimiter)
[![Go Report Card](https://goreportcard.com/badge/github.com/rleungx/gin-ratelimiter)](https://goreportcard.com/report/github.com/rleungx/gin-ratelimiter)
[![GoDoc](https://godoc.org/github.com/rleungx/gin-ratelimiter?status.svg)](https://godoc.org/github.com/rleungx/gin-ratelimiter)

Gin middleware for per-route rate limiting and concurrency limiting.

## Add Dependency

```bash
go get github.com/rleungx/gin-ratelimiter@latest
```

```go
import ratelimiter "github.com/rleungx/gin-ratelimiter"
```

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
```

## Runtime Update

```go
limiter.UpdateRateLimit("/jobs", 20, 40)
limiter.UpdateConcurrencyLimit("/jobs", 10)
```

Use the Gin route pattern when updating limits, for example `/users/:id`.

See `examples/main.go` for a runnable example.
