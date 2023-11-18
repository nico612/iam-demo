package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/mattn/go-isatty"
	"github.com/nico612/iam-demo/pkg/log"
	"io"
	"os"
	"time"
)

// 自定义 Gin 的日志输出中间件

// defaultLogFormatter is the default log format function Logger middleware uses.
var defaultLogFormatter = func(param gin.LogFormatterParams) string {
	var statusColor, methodColor, resetColor string

	if param.IsOutputColor() {
		statusColor = param.StatusCodeColor()
		methodColor = param.MethodColor()
		resetColor = param.ResetColor() // 重置 所有的逃逸属性？
	}

	if param.Latency > time.Minute { // 服务器处理某个请求所花费的时间。
		// Truncate in a golang < 1.8 safe way
		param.Latency = param.Latency - param.Latency%time.Second
	}

	return fmt.Sprintf("%s%3d%s - [%s] \"%v %s%s%s %s\" %s",
		// param.TimeStamp.Format("2006/01/02 - 15:04:05"),
		statusColor, param.StatusCode, resetColor,
		param.ClientIP,
		param.Latency,
		methodColor, param.Method, resetColor,
		param.Path,
		param.ErrorMessage,
	)
}

// Logger instances a Logger middleware that will write the logs to gin.DefaultWriter.
// default gin.DefaultWriter = os.Stdout.
func Logger() gin.HandlerFunc {
	return LoggerWithConfig(GetLoggerConfig(nil, nil, nil))
}

// LoggerWithFormatter instance a Logger middleware with the specified log format function.
func LoggerWithFormatter(f gin.LogFormatter) gin.HandlerFunc {
	return LoggerWithConfig(gin.LoggerConfig{
		Formatter: f,
	})
}

// LoggerWithWriter instance a Logger middleware with the specified writer buffer.
// Example: os.Stdout, a file opened in write mode, a socket...
func LoggerWithWriter(out io.Writer, notlogged ...string) gin.HandlerFunc {
	return LoggerWithConfig(gin.LoggerConfig{
		Output:    out,
		SkipPaths: notlogged,
	})
}

func LoggerWithConfig(conf gin.LoggerConfig) gin.HandlerFunc {
	formatter := conf.Formatter
	if formatter == nil {
		formatter = defaultLogFormatter
	}

	out := conf.Output
	if out == nil {
		out = gin.DefaultWriter // 默认 os.Stdout
	}

	// 一个 URL 路径数组，用于指定不记录日志的路径。可以配置一些特定的路径，当请求的路径匹配到这些路径时，将不会记录日志。这是一个可选的参数。
	notlogged := conf.SkipPaths

	isTerm := true // isTerminal

	// 判断是否是终端
	// os.Getenv("TERM") == "dumb" 非交互式终端
	// isatty.IsTerminal(w.Fd()) : 使用  isatty.IsTerminal  函数判断  w.Fd()  所代表的文件描述符是否为终端设备。
	// isatty.IsCygwinTerminal(w.Fd()) : 使用  isatty.IsCygwinTerminal  函数判断  w.Fd()  所代表的文件描述符是否为 Cygwin 终端设备。
	if w, ok := out.(*os.File); !ok || os.Getenv("TERM") == "dumb" || (!isatty.IsTerminal(w.Fd()) && !isatty.IsCygwinTerminal(w.Fd())) {
		isTerm = false
	}

	if isTerm {
		gin.ForceConsoleColor() // 强制在控制台中使用彩色输出。
	}

	var skip map[string]struct{}

	if length := len(notlogged); length > 0 {
		skip = make(map[string]struct{}, length)
		for _, path := range notlogged {
			skip[path] = struct{}{}
		}
	}

	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Log only when path is not being skipped
		if _, ok := skip[path]; !ok {
			param := gin.LogFormatterParams{
				Request: c.Request,
				Keys:    c.Keys,
			}

			// Stop timer
			param.TimeStamp = time.Now()
			param.Latency = param.TimeStamp.Sub(start)

			param.ClientIP = c.ClientIP()
			param.Method = c.Request.Method
			param.StatusCode = c.Writer.Status()
			param.ErrorMessage = c.Errors.ByType(gin.ErrorTypePrivate).String()

			param.BodySize = c.Writer.Size()

			if raw != "" {
				path = path + "?" + raw
			}

			param.Path = path

			log.L(c).Infow(formatter(param))
		}
	}

}
