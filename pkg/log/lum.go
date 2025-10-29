package log

import (
	"gopkg.in/natefinch/lumberjack.v2"
)

type LumberjackConfig interface {
	GetMaxSize() int
	GetMaxAge() int
	GetMaxBackUps() int
	GetLocalTime() bool
	GetCompress() bool
}

// NewlumLogger 根据配置初始化日志底层IO
func NewLumLogger(
	c LumberjackConfig,
	logPath string,
) *lumberjack.Logger {
	return &lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    c.GetMaxSize(),
		MaxAge:     c.GetMaxAge(),
		MaxBackups: c.GetMaxBackUps(),
		LocalTime:  c.GetLocalTime(),
		Compress:   c.GetCompress(),
	}
}
