package middleware

import (
	"github.com/gin-gonic/gin"
	gindump "github.com/tpkeeper/gin-dump"
	"net/http"
	"time"
)

// Middlewares storage registered middlewares.
var Middlewares = defaultMiddlewares()

// NoCache 是一个 Gin 中间件，用来禁止客户端缓存 HTTP 请求的返回结果..
func NoCache(c *gin.Context) {
	c.Header("Cache-Control", "no-cache, no-storage, max-age=0, must-revalidate, value")
	c.Header("Expires", "Thu, 01 Jan 1970 00:00:00 GMT")
	c.Header("Last-Modified", time.Now().UTC().Format(http.TimeFormat))
	c.Next()
}

// Options 是一个 Gin 中间件，用来设置 options 请求的返回头，然后退出中间件链，并结束请求(浏览器跨域设置).
// 复杂请求的 CORS 跨域处理
// 复杂请求的 CORS 请求，会在正式通信之前，增加一次 HTTP 查询请求，称为"预检"请求（preflight）。"预检"请求用的请求方法是OPTIONS，表示这个请求是用来询问请求能否安全送出的。
// 预检通过后，浏览器就正常发起请求和响应，流程和简单请求一致。
// 当后端收到预检请求后，可以设置跨域相关 Header 以完成跨域请求。支持的 Header 具体如下表所示：.
func Options(c *gin.Context) {
	if c.Request.Method != "OPTIONS" {
		c.Next()
	} else {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "authorization, origin, content-type, accept")
		c.Header("Allow", "HEAD,GET,POST,PUT,PATCH,DELETE,OPTIONS")
		c.Header("Content-Type", "application/json")
		c.AbortWithStatus(http.StatusOK)
	}
}

// Secure 是一个 Gin 中间件，用来添加一些安全和资源访问相关的 HTTP 头.
func Secure(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("X-Frame-Options", "DENY")
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("X-XSS-Protection", "1; mode=block")

	if c.Request.TLS != nil {
		c.Header("Strict-Transport-Security", "max-age=31536000")
	}
}

// 默认中间件配置
func defaultMiddlewares() map[string]gin.HandlerFunc {
	return map[string]gin.HandlerFunc{
		"recovery":  gin.Recovery(),
		"secure":    Secure,
		"options":   Options,
		"nocache":   NoCache,
		"cors":      Cors(),
		"requestid": RequestID(),
		"logger":    Logger(),
		"dump":      gindump.Dump(), // 用于 Gin 框架的中间件库，它用于将 HTTP 请求和响应的详细信息打印到控制台或日志文件中。
	}
}
