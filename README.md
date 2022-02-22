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

	// Listen and Server in 0.0.0.0:8080
	r.Run(":8880")
}
```

The output with 11 requests:
```
2022/02/21 20:01:32 pong 1645444892293270491 200
2022/02/21 20:01:32 pong 1645444892293716732 200
2022/02/21 20:01:32 pong 1645444892293964251 200
2022/02/21 20:01:32 pong 1645444892294363113 200
2022/02/21 20:01:32 pong 1645444892294575123 200
2022/02/21 20:01:32 pong 1645444892294839949 200
2022/02/21 20:01:32 pong 1645444892295033787 200
2022/02/21 20:01:32 pong 1645444892295200719 200
2022/02/21 20:01:32 pong 1645444892295349608 200
2022/02/21 20:01:32 pong 1645444892295528560 200
2022/02/21 20:01:32  429
```