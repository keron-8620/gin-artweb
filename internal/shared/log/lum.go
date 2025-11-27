package log

import (
	"gopkg.in/natefinch/lumberjack.v2"

	"gin-artweb/internal/shared/config"
)


// NewlumLogger 根据配置初始化日志底层IO
func NewLumLogger(
	c *config.LogConfig,
	logPath string,
) *lumberjack.Logger {
	return &lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    c.MaxSize,
		MaxAge:     c.MaxAge,
		MaxBackups: c.MaxBackups,
		LocalTime:  c.LocalTime,
		Compress:   c.Compress,
	}
}
