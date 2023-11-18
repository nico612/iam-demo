package middleware

import (
	"errors"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// 限流器

// ErrLimitExceeded defines Limit exceeded error.
var ErrLimitExceeded = errors.New("Limit exceeded")

// Limit drops (HTTP status 429) the request if the limit is reached.
// maxEventsPerSec 限制速率（rate）：限制每秒事件个数， maxBurstSize： 令牌桶的最大容量
func Limit(maxEventsPerSec float64, maxBurstSize int) gin.HandlerFunc {

	// define limiter
	limiter := rate.NewLimiter(rate.Limit(maxEventsPerSec), maxBurstSize)

	return func(c *gin.Context) {
		if limiter.Allow() { // check reached
			c.Next()

			return
		}

		// Limit reached
		_ = c.Error(ErrLimitExceeded)
		c.AbortWithStatus(429)
	}
}
