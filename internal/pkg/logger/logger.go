package logger

import (
	gormlogger "gorm.io/gorm/logger"
	"time"
)

// Define colors.
const (
	Reset       = "\033[0m"
	Red         = "\033[31m"
	Green       = "\033[32m"
	Yellow      = "\033[33m"
	Blue        = "\033[34m"
	Magenta     = "\033[35m"
	Cyan        = "\033[36m"
	White       = "\033[37m"
	BlueBold    = "\033[34;1m"
	MagentaBold = "\033[35;1m"
	RedBold     = "\033[31;1m"
	YellowBold  = "\033[33;1m"
)

// Define gorm log levels.
const (
	Silent gormlogger.LogLevel = iota + 1
	Error
	Warn
	Info
)

// Writer log writer interface.
type Writer interface {
	Printf(string, ...interface{})
}

// Config defines a gorm logger configuration.
type Config struct {
	SlowThreshold time.Duration
	Colorful      bool
	LogLevel      gormlogger.LogLevel
}
