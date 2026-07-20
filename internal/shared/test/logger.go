package test

import "go.uber.org/zap"

// NewTestZapLogger 默认返回静默日志器，避免正常的错误分支测试输出大量堆栈。
func NewTestZapLogger() *zap.Logger {
	return zap.NewNop()
}
