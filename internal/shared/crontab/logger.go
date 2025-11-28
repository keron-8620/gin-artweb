package crontab

import (
	"go.uber.org/zap"
)

type cronLog struct {
	logger *zap.Logger
}

func (l *cronLog) Info(msg string, keysAndValues ...interface{}) {
	fields := l.keyValuesToFields(keysAndValues...)
	l.logger.Info(msg, fields...)
}

func (l *cronLog) Error(err error, msg string, keysAndValues ...interface{}) {
	fields := l.keyValuesToFields(keysAndValues...)
	fields = append(fields, zap.Error(err))
	l.logger.Error(msg, fields...)
}

func (l *cronLog) keyValuesToFields(keysAndValues ...interface{}) []zap.Field {
	if len(keysAndValues) == 0 {
		return nil
	}

	fields := make([]zap.Field, 0, len(keysAndValues)/2)
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			key, ok := keysAndValues[i].(string)
			if ok {
				fields = append(fields, zap.Any(key, keysAndValues[i+1]))
			}
		}
	}
	return fields
}
