# gin-ratelimiter

[![Run Tests](https://github.com/rleungx/gin-ratelimiter/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/rleungx/gin-ratelimiter/actions/workflows/go.yml)
[![codecov](https://codecov.io/gh/rleungx/gin-ratelimiter/branch/main/graph/badge.svg)](https://codecov.io/gh/rleungx/gin-ratelimiter)
[![Go Report Card](https://goreportcard.com/badge/github.com/rleungx/gin-ratelimiter)](https://goreportcard.com/report/github.com/rleungx/gin-ratelimiter)
[![GoDoc](https://godoc.org/github.com/rleungx/gin-ratelimiter?status.svg)](https://godoc.org/github.com/rleungx/gin-ratelimiter)

The gin-ratelimiter is a middleware for limiting the request rate under [Gin framework](https://github.com/gin-gonic/gin) based on [golang.org/x/time/rate](golang.org/x/time/rate)

## Usage

```go
go get github.com/rleungx/gin-ratelimiter
```

And import it in your code:

```go
import "github.com/rleungx/gin-ratelimiter"
```

## Example

See the [example](examples/main.go).

```go
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
```

The output with 4 requests:
```
HTTP/1.1 200 OK
Content-Type: text/plain; charset=utf-8
Date: Wed, 12 Feb 2025 08:34:44 GMT
Content-Length: 0

HTTP/1.1 200 OK
Content-Type: text/plain; charset=utf-8
Date: Wed, 12 Feb 2025 08:34:44 GMT
Content-Length: 0

HTTP/1.1 200 OK
Content-Type: text/plain; charset=utf-8
Date: Wed, 12 Feb 2025 08:34:44 GMT
Content-Length: 0

HTTP/1.1 429 Too Many Requests
Date: Wed, 12 Feb 2025 08:34:44 GMT
Content-Length: 0
```
