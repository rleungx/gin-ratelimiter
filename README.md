# gin-ratelimiter
The gin-ratelimiter is a middleware for limiting the request rate under [Gin framework](https://github.com/gin-gonic/gin) based on [golang.org/x/time/rate](golang.org/x/time/rate)

## Usage

```go
go get -u github.com/rleungx/gin-ratelimiter
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

	// Listen and Server in 0.0.0.0:8888
	r.Run(":8888")
}
```

The output with 11 requests:
```
status code: 200, response: pong 1645506503302297793
status code: 200, response: pong 1645506503302647317
status code: 200, response: pong 1645506503302812520
status code: 200, response: pong 1645506503303018293
status code: 200, response: pong 1645506503303311303
status code: 200, response: pong 1645506503303427116
status code: 200, response: pong 1645506503303540257
status code: 200, response: pong 1645506503303677963
status code: 200, response: pong 1645506503303826727
status code: 200, response: pong 1645506503303956919
status code: 429, response:
```
