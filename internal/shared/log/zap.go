package log

import (
	"fmt"
	"io"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	DurationKey = "duration"
)

// NewZapLogger 根据配置初始化日志
func NewZapLogger(level string, w io.Writer) (*zap.Logger, error) {
	// 解析日志级别
	atomicLevel := zap.NewAtomicLevel()
	if err := atomicLevel.UnmarshalText([]byte(level)); err != nil {
		return nil, fmt.Errorf("无法解析日志级别: %v", err)
	}

	// 编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 创建核心
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(w),
		// zapcore.NewMultiWriteSyncer(
		// 	zapcore.AddSync(os.Stdout),
		// 	zapcore.AddSync(os.Stderr),
		// 	zapcore.AddSync(w),
		// ),
		atomicLevel,
	)

	// 添加调用者信息和堆栈跟踪
	return zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.DPanicLevel)), nil
}

func NewZapLoggerMust(level string, w io.Writer) *zap.Logger {
	logger, err := NewZapLogger(level, w)
	if err != nil {
		panic(fmt.Sprintf("初始化日志失败: %v", err))
	}
	return logger
}

type Loggers struct {
	Server  *zap.Logger
	Service *zap.Logger
	Biz     *zap.Logger
	Data    *zap.Logger
}
