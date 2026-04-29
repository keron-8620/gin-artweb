package ctxutil

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"gin-artweb/internal/shared/auth"
)

func TestGetJwtClaims(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		want    *auth.JwtClaims
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil context",
			ctx:     nil,
			want:    nil,
			wantErr: true,
			errMsg:  "获取用户信息失败:context 不能为空",
		},
		{
			name:    "empty context without claims",
			ctx:     context.Background(),
			want:    nil,
			wantErr: true,
			errMsg:  "获取用户信息失败:认证信息缺失",
		},
		{
			name:    "context with invalid claims type",
			ctx:     context.WithValue(context.Background(), JwtClaimsKey, "invalid"),
			want:    nil,
			wantErr: true,
			errMsg:  "获取用户信息失败:认证信息格式错误",
		},
		{
			name:    "success with valid claims",
			ctx:     SetJwtClaims(context.Background(), &auth.JwtClaims{}),
			want:    &auth.JwtClaims{},
			wantErr: false,
			errMsg:  "",
		},
		{
			name: "success with full claims",
			ctx: func() context.Context {
				claims := &auth.JwtClaims{
					UserInfo: auth.UserInfo{
						UserID:   123,
						Username: "testuser",
						RoleID:   1,
						IsStaff:  true,
					},
				}
				return SetJwtClaims(context.Background(), claims)
			}(),
			want: &auth.JwtClaims{
				UserInfo: auth.UserInfo{
					UserID:   123,
					Username: "testuser",
					RoleID:   1,
					IsStaff:  true,
				},
			},
			wantErr: false,
			errMsg:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetJwtClaims(tt.ctx)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				if tt.want != nil {
					assert.Equal(t, tt.want.UserInfo.UserID, got.UserInfo.UserID)
					assert.Equal(t, tt.want.UserInfo.Username, got.UserInfo.Username)
					assert.Equal(t, tt.want.UserInfo.RoleID, got.UserInfo.RoleID)
					assert.Equal(t, tt.want.UserInfo.IsStaff, got.UserInfo.IsStaff)
				}
			}
		})
	}
}

func TestMustGetJwtClaims(t *testing.T) {
	t.Run("success with valid claims", func(t *testing.T) {
		claims := &auth.JwtClaims{
			UserInfo: auth.UserInfo{UserID: 456},
		}
		ctx := SetJwtClaims(context.Background(), claims)
		got := MustGetJwtClaims(ctx)
		assert.NotNil(t, got)
		assert.Equal(t, uint32(456), got.UserInfo.UserID)
	})

	t.Run("panic with nil context", func(t *testing.T) {
		assert.Panics(t, func() {
			//nolint:staticcheck // intentionally testing nil context behavior
			MustGetJwtClaims(nil)
		})
	})

	t.Run("panic with missing claims", func(t *testing.T) {
		assert.Panics(t, func() {
			MustGetJwtClaims(context.Background())
		})
	})
}

func TestSetJwtClaims(t *testing.T) {
	t.Run("set and retrieve claims", func(t *testing.T) {
		ctx := context.Background()
		claims := &auth.JwtClaims{
			UserInfo: auth.UserInfo{UserID: 789},
		}

		newCtx := SetJwtClaims(ctx, claims)
		assert.NotEqual(t, ctx, newCtx)

		got, err := GetJwtClaims(newCtx)
		assert.NoError(t, err)
		assert.Equal(t, uint32(789), got.UserInfo.UserID)
	})

	t.Run("set nil claims", func(t *testing.T) {
		ctx := context.Background()
		newCtx := SetJwtClaims(ctx, nil)
		assert.NotEqual(t, ctx, newCtx)

		got, err := GetJwtClaims(newCtx)
		assert.NoError(t, err)
		assert.Nil(t, got)
	})
}

func TestGetTraceID(t *testing.T) {
	tests := []struct {
		name string
		ctx  context.Context
		want string
	}{
		{
			name: "nil context",
			ctx:  nil,
			want: DefaultTraceID,
		},
		{
			name: "empty context",
			ctx:  context.Background(),
			want: DefaultTraceID,
		},
		{
			name: "context with valid trace ID",
			ctx:  SetTraceID(context.Background(), "trace-12345"),
			want: "trace-12345",
		},
		{
			name: "context with empty string trace ID",
			ctx:  SetTraceID(context.Background(), ""),
			want: DefaultTraceID,
		},
		{
			name: "context with whitespace trace ID",
			ctx:  SetTraceID(context.Background(), "   "),
			want: DefaultTraceID,
		},
		{
			name: "context with valid trace ID containing spaces",
			ctx:  SetTraceID(context.Background(), "  trace-67890  "),
			want: "  trace-67890  ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetTraceID(tt.ctx)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSetTraceID(t *testing.T) {
	t.Run("set and retrieve trace ID", func(t *testing.T) {
		ctx := context.Background()
		traceID := "my-trace-id-123"

		newCtx := SetTraceID(ctx, traceID)
		assert.NotEqual(t, ctx, newCtx)

		got := GetTraceID(newCtx)
		assert.Equal(t, traceID, got)
	})

	t.Run("set empty trace ID", func(t *testing.T) {
		ctx := context.Background()
		newCtx := SetTraceID(ctx, "")
		assert.NotEqual(t, ctx, newCtx)

		got := GetTraceID(newCtx)
		assert.Equal(t, DefaultTraceID, got)
	})
}

func TestNewLogger(t *testing.T) {
	t.Run("logger with trace ID only", func(t *testing.T) {
		core, recorded := observer.New(zapcore.InfoLevel)
		logger := zap.New(core)

		ctx := SetTraceID(context.Background(), "test-trace-id")
		newLogger := NewLogger(logger, ctx)

		newLogger.Info("test message")
		assert.Equal(t, 1, recorded.Len())

		logEntry := recorded.All()[0]
		assert.Equal(t, "test-trace-id", logEntry.ContextMap()["trace_id"])
		assert.NotContains(t, logEntry.ContextMap(), "uid")
	})

	t.Run("logger with trace ID and user info", func(t *testing.T) {
		core, recorded := observer.New(zapcore.InfoLevel)
		logger := zap.New(core)

		claims := &auth.JwtClaims{
			UserInfo: auth.UserInfo{UserID: 1001},
		}
		ctx := SetJwtClaims(context.Background(), claims)
		ctx = SetTraceID(ctx, "trace-with-user")

		newLogger := NewLogger(logger, ctx)
		newLogger.Info("test with user")

		assert.Equal(t, 1, recorded.Len())
		logEntry := recorded.All()[0]
		assert.Equal(t, "trace-with-user", logEntry.ContextMap()["trace_id"])
		assert.Equal(t, uint32(1001), logEntry.ContextMap()["uid"])
	})

	t.Run("logger with nil context", func(t *testing.T) {
		core, recorded := observer.New(zapcore.InfoLevel)
		logger := zap.New(core)

		newLogger := NewLogger(logger, nil)
		newLogger.Info("test nil context")

		assert.Equal(t, 1, recorded.Len())
		logEntry := recorded.All()[0]
		assert.Equal(t, DefaultTraceID, logEntry.ContextMap()["trace_id"])
	})

	t.Run("logger with no jwt claims", func(t *testing.T) {
		core, recorded := observer.New(zapcore.DebugLevel)
		logger := zap.New(core)

		ctx := SetTraceID(context.Background(), "trace-no-claims")

		newLogger := NewLogger(logger, ctx)
		newLogger.Info("test no claims")

		assert.GreaterOrEqual(t, recorded.Len(), 1)

		var infoEntry *observer.LoggedEntry
		var debugEntry *observer.LoggedEntry
		for _, entry := range recorded.All() {
			if entry.Level == zapcore.InfoLevel {
				infoEntry = &entry
			} else if entry.Level == zapcore.DebugLevel {
				debugEntry = &entry
			}
		}

		assert.NotNil(t, infoEntry)
		assert.Equal(t, "trace-no-claims", infoEntry.ContextMap()["trace_id"])
		assert.NotContains(t, infoEntry.ContextMap(), "uid")

		assert.NotNil(t, debugEntry)
		assert.Contains(t, debugEntry.Message, "获取日志附加用户信息失败")
	})
}
