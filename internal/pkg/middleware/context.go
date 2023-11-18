package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/nico612/iam-demo/pkg/log"
)

const UsernameKey = "username"

func Context() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(log.KeyRequestID, c.GetString(XRequestIDKey))
		c.Set(log.KeyUsername, c.GetString(UsernameKey))
		c.Next()
	}
}
