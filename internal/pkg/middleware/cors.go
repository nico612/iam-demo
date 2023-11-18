package middleware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"time"
)

// 跨域中间件设置

const (
	maxAge = 12
)

func Cors() gin.HandlerFunc {
	return cors.New(cors.Config{
		// 必须：允许域名请求， 如果不返回这个头部。浏览器会抛出跨域错误（注：服务端不会报错）
		AllowOrigins: []string{"*"},
		AllowOriginFunc: func(origin string) bool {
			return origin == "https://github.com"
		},
		// 必选，逗号分隔的字符串，表明服务器支持的所有跨域请求的方法
		AllowMethods: []string{"PUT", "PATCH", "GET", "POST", "OPTIONS", "DELETE"},
		// 表明服务器支持的所有头信息字段，不限于浏览器在"预检"中请求的字段。如果浏览器请求包括 Access-Control-Request-Headers 字段，则此字段是必选的
		AllowHeaders: []string{"Origin", "Authorization", "Content-Type", "Accept"},
		// 可选，布尔值，默认是false，表示不允许发送 Cookie
		AllowCredentials: true,
		ExposeHeaders:    []string{"Content-Length"},
		// 指定本次预检请求的有效期，单位为秒。可以避免频繁的预检请求
		MaxAge: maxAge * time.Hour,
	})
}
