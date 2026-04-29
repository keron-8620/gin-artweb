package log

import (
	"bytes"
	"strings"
	"testing"

	"gin-artweb/internal/shared/config"
	"go.uber.org/zap"
)

func TestNewLumLogger(t *testing.T) {
	testCases := []struct {
		name           string
		config         *config.LogConfig
		logPath        string
		wantMaxSize    int
		wantMaxAge     int
		wantMaxBackups int
		wantLocalTime  bool
		wantCompress   bool
	}{
		{
			name: "default config",
			config: &config.LogConfig{
				MaxSize:    100,
				MaxAge:     30,
				MaxBackups: 10,
				LocalTime:  true,
				Compress:   false,
			},
			logPath:        "/var/log/test.log",
			wantMaxSize:    100,
			wantMaxAge:     30,
			wantMaxBackups: 10,
			wantLocalTime:  true,
			wantCompress:   false,
		},
		{
			name: "minimal config",
			config: &config.LogConfig{
				MaxSize:    1,
				MaxAge:     1,
				MaxBackups: 1,
				LocalTime:  false,
				Compress:   true,
			},
			logPath:        "/tmp/app.log",
			wantMaxSize:    1,
			wantMaxAge:     1,
			wantMaxBackups: 1,
			wantLocalTime:  false,
			wantCompress:   true,
		},
		{
			name: "zero values",
			config: &config.LogConfig{
				MaxSize:    0,
				MaxAge:     0,
				MaxBackups: 0,
				LocalTime:  false,
				Compress:   false,
			},
			logPath:        "/dev/null",
			wantMaxSize:    0,
			wantMaxAge:     0,
			wantMaxBackups: 0,
			wantLocalTime:  false,
			wantCompress:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger := NewLumLogger(tc.config, tc.logPath)

			if logger == nil {
				t.Fatal("NewLumLogger returned nil")
			}

			if logger.Filename != tc.logPath {
				t.Errorf("expected Filename %q, got %q", tc.logPath, logger.Filename)
			}
			if logger.MaxSize != tc.wantMaxSize {
				t.Errorf("expected MaxSize %d, got %d", tc.wantMaxSize, logger.MaxSize)
			}
			if logger.MaxAge != tc.wantMaxAge {
				t.Errorf("expected MaxAge %d, got %d", tc.wantMaxAge, logger.MaxAge)
			}
			if logger.MaxBackups != tc.wantMaxBackups {
				t.Errorf("expected MaxBackups %d, got %d", tc.wantMaxBackups, logger.MaxBackups)
			}
			if logger.LocalTime != tc.wantLocalTime {
				t.Errorf("expected LocalTime %v, got %v", tc.wantLocalTime, logger.LocalTime)
			}
			if logger.Compress != tc.wantCompress {
				t.Errorf("expected Compress %v, got %v", tc.wantCompress, logger.Compress)
			}
		})
	}
}

func TestNewZapLogger(t *testing.T) {
	testCases := []struct {
		name    string
		level   string
		wantErr bool
	}{
		{name: "valid debug level", level: "debug", wantErr: false},
		{name: "valid info level", level: "info", wantErr: false},
		{name: "valid warn level", level: "warn", wantErr: false},
		{name: "valid error level", level: "error", wantErr: false},
		{name: "valid panic level", level: "panic", wantErr: false},
		{name: "valid fatal level", level: "fatal", wantErr: false},
		{name: "invalid level", level: "invalid", wantErr: true},
		{name: "empty level defaults to info", level: "", wantErr: false},
		{name: "uppercase level works", level: "DEBUG", wantErr: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger, err := NewZapLogger(tc.level, &buf)

			if tc.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if logger != nil {
					t.Error("expected nil logger when error occurs")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if logger == nil {
				t.Fatal("expected logger, got nil")
			}
		})
	}
}

func TestNewZapLogger_LogOutput(t *testing.T) {
	var buf bytes.Buffer
	logger, err := NewZapLogger("debug", &buf)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	logger.Debug("debug message")
	if !strings.Contains(buf.String(), "debug") {
		t.Errorf("expected log to contain 'debug', got: %s", buf.String())
	}
	if !strings.Contains(buf.String(), "debug message") {
		t.Errorf("expected log to contain 'debug message', got: %s", buf.String())
	}

	buf.Reset()
	logger.Info("info message")
	if !strings.Contains(buf.String(), "info") {
		t.Errorf("expected log to contain 'info', got: %s", buf.String())
	}
	if !strings.Contains(buf.String(), "info message") {
		t.Errorf("expected log to contain 'info message', got: %s", buf.String())
	}

	buf.Reset()
	logger.Warn("warn message")
	if !strings.Contains(buf.String(), "warn") {
		t.Errorf("expected log to contain 'warn', got: %s", buf.String())
	}

	buf.Reset()
	logger.Error("error message")
	if !strings.Contains(buf.String(), "error") {
		t.Errorf("expected log to contain 'error', got: %s", buf.String())
	}
}

func TestNewZapLogger_LogLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	logger, err := NewZapLogger("info", &buf)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	logger.Debug("debug message")
	if buf.Len() != 0 {
		t.Errorf("expected no output for debug level with info logger, got: %s", buf.String())
	}

	buf.Reset()
	logger.Info("info message")
	if buf.Len() == 0 {
		t.Error("expected output for info level")
	}
}

func TestNewZapLogger_LogFormat(t *testing.T) {
	var buf bytes.Buffer
	logger, err := NewZapLogger("info", &buf)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	logger.Info("test message")
	logOutput := buf.String()

	if !strings.Contains(logOutput, `"time"`) {
		t.Error("expected log to contain 'time' field")
	}
	if !strings.Contains(logOutput, `"level"`) {
		t.Error("expected log to contain 'level' field")
	}
	if !strings.Contains(logOutput, `"msg"`) {
		t.Error("expected log to contain 'msg' field")
	}
	if !strings.Contains(logOutput, `"caller"`) {
		t.Error("expected log to contain 'caller' field")
	}
	if !strings.Contains(logOutput, "test message") {
		t.Errorf("expected log to contain message, got: %s", logOutput)
	}
}

func TestNewZapLoggerMust(t *testing.T) {
	t.Run("valid level", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewZapLoggerMust("info", &buf)

		if logger == nil {
			t.Error("expected logger, got nil")
		}

		logger.Info("test")
		if buf.Len() == 0 {
			t.Error("expected log output")
		}
	})

	t.Run("invalid level should panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic for invalid level")
			}
		}()

		var buf bytes.Buffer
		NewZapLoggerMust("invalid", &buf)
	})
}

func TestNewZapLogger_WithFields(t *testing.T) {
	var buf bytes.Buffer
	logger, err := NewZapLogger("debug", &buf)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	logger.With(zap.String("key", "value")).Info("with fields")
	logOutput := buf.String()

	if !strings.Contains(logOutput, `"key":"value"`) {
		t.Errorf("expected log to contain 'key:value', got: %s", logOutput)
	}
}

func TestNewZapLogger_AtomicLevelChange(t *testing.T) {
	var buf bytes.Buffer
	logger, err := NewZapLogger("error", &buf)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	logger.Info("should not appear")
	if buf.Len() != 0 {
		t.Errorf("expected no output for info level with error logger, got: %s", buf.String())
	}

	zap.L().Info("should not appear either")
}
