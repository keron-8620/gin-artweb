package crontab

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestCronLog_Info(t *testing.T) {
	t.Run("info with message only", func(t *testing.T) {
		core, recorded := observer.New(zapcore.InfoLevel)
		logger := zap.New(core)

		cronLog := &cronLog{logger: logger}
		cronLog.Info("test message")

		assert.Equal(t, 1, recorded.Len())
		logEntry := recorded.All()[0]
		assert.Equal(t, "test message", logEntry.Message)
		assert.Equal(t, zapcore.InfoLevel, logEntry.Level)
	})

	t.Run("info with key-value pairs", func(t *testing.T) {
		core, recorded := observer.New(zapcore.InfoLevel)
		logger := zap.New(core)

		cronLog := &cronLog{logger: logger}
		cronLog.Info("test message", "key1", "value1", "key2", int64(123))

		assert.Equal(t, 1, recorded.Len())
		logEntry := recorded.All()[0]
		assert.Equal(t, "test message", logEntry.Message)
		assert.Equal(t, "value1", logEntry.ContextMap()["key1"])
		assert.Equal(t, int64(123), logEntry.ContextMap()["key2"])
	})

	t.Run("info with odd number of key-value pairs", func(t *testing.T) {
		core, recorded := observer.New(zapcore.InfoLevel)
		logger := zap.New(core)

		cronLog := &cronLog{logger: logger}
		cronLog.Info("test message", "key1", "value1", "key2")

		assert.Equal(t, 1, recorded.Len())
		logEntry := recorded.All()[0]
		assert.Equal(t, "test message", logEntry.Message)
		assert.Equal(t, "value1", logEntry.ContextMap()["key1"])
		assert.NotContains(t, logEntry.ContextMap(), "key2")
	})

	t.Run("info with nil logger should panic", func(t *testing.T) {
		cronLog := &cronLog{logger: nil}
		assert.Panics(t, func() {
			cronLog.Info("test message")
		})
	})

	t.Run("info with empty message", func(t *testing.T) {
		core, recorded := observer.New(zapcore.InfoLevel)
		logger := zap.New(core)

		cronLog := &cronLog{logger: logger}
		cronLog.Info("")

		assert.Equal(t, 1, recorded.Len())
		logEntry := recorded.All()[0]
		assert.Equal(t, "", logEntry.Message)
	})

	t.Run("info with non-string key", func(t *testing.T) {
		core, recorded := observer.New(zapcore.InfoLevel)
		logger := zap.New(core)

		cronLog := &cronLog{logger: logger}
		cronLog.Info("test message", 123, "value")

		assert.Equal(t, 1, recorded.Len())
		logEntry := recorded.All()[0]
		assert.Equal(t, "test message", logEntry.Message)
		assert.NotContains(t, logEntry.ContextMap(), 123)
	})
}

func TestCronLog_Error(t *testing.T) {
	testErr := errors.New("test error")

	t.Run("error with message and error", func(t *testing.T) {
		core, recorded := observer.New(zapcore.ErrorLevel)
		logger := zap.New(core)

		cronLog := &cronLog{logger: logger}
		cronLog.Error(testErr, "test error message")

		assert.Equal(t, 1, recorded.Len())
		logEntry := recorded.All()[0]
		assert.Equal(t, "test error message", logEntry.Message)
		assert.Equal(t, zapcore.ErrorLevel, logEntry.Level)
		assert.Contains(t, logEntry.ContextMap(), "error")
	})

	t.Run("error with key-value pairs", func(t *testing.T) {
		core, recorded := observer.New(zapcore.ErrorLevel)
		logger := zap.New(core)

		cronLog := &cronLog{logger: logger}
		cronLog.Error(testErr, "test error message", "key1", "value1", "key2", int64(456))

		assert.Equal(t, 1, recorded.Len())
		logEntry := recorded.All()[0]
		assert.Equal(t, "test error message", logEntry.Message)
		assert.Equal(t, "value1", logEntry.ContextMap()["key1"])
		assert.Equal(t, int64(456), logEntry.ContextMap()["key2"])
		assert.Contains(t, logEntry.ContextMap(), "error")
	})

	t.Run("error with nil error", func(t *testing.T) {
		core, recorded := observer.New(zapcore.ErrorLevel)
		logger := zap.New(core)

		cronLog := &cronLog{logger: logger}
		cronLog.Error(nil, "test message")

		assert.Equal(t, 1, recorded.Len())
		logEntry := recorded.All()[0]
		assert.Equal(t, "test message", logEntry.Message)
	})

	t.Run("error with nil logger should panic", func(t *testing.T) {
		cronLog := &cronLog{logger: nil}
		assert.Panics(t, func() {
			cronLog.Error(testErr, "test message")
		})
	})

	t.Run("error with empty message", func(t *testing.T) {
		core, recorded := observer.New(zapcore.ErrorLevel)
		logger := zap.New(core)

		cronLog := &cronLog{logger: logger}
		cronLog.Error(testErr, "")

		assert.Equal(t, 1, recorded.Len())
		logEntry := recorded.All()[0]
		assert.Equal(t, "", logEntry.Message)
	})
}

func TestCronLog_keyValuesToFields(t *testing.T) {
	t.Run("empty keyValues", func(t *testing.T) {
		cronLog := &cronLog{}
		fields := cronLog.keyValuesToFields()

		assert.Nil(t, fields)
	})

	t.Run("valid key-value pairs", func(t *testing.T) {
		cronLog := &cronLog{}
		fields := cronLog.keyValuesToFields("key1", "value1", "key2", 123)

		assert.NotNil(t, fields)
		assert.Len(t, fields, 2)
	})

	t.Run("odd number of key-value pairs", func(t *testing.T) {
		cronLog := &cronLog{}
		fields := cronLog.keyValuesToFields("key1", "value1", "key2")

		assert.NotNil(t, fields)
		assert.Len(t, fields, 1)
	})

	t.Run("non-string key", func(t *testing.T) {
		cronLog := &cronLog{}
		fields := cronLog.keyValuesToFields(123, "value", "key2", "value2")

		assert.NotNil(t, fields)
		assert.Len(t, fields, 1)
	})

	t.Run("nil values", func(t *testing.T) {
		cronLog := &cronLog{}
		fields := cronLog.keyValuesToFields("key1", nil, "key2", "value2")

		assert.NotNil(t, fields)
		assert.Len(t, fields, 2)
	})
}
