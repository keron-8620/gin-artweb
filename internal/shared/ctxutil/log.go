package ctxutil

import (
	"context"

	"go.uber.org/zap"
)

// NewLogger 为日志实例附加上下文信息（traceID + 用户ID）
// 核心优化：增加错误日志、空值防护、提升可维护性
func NewLogger(log *zap.Logger, ctx context.Context) *zap.Logger {
	// 基础字段：必加 traceID（无论是否有用户信息）
	fields := []zap.Field{
		zap.String("trace_id", GetTraceID(ctx)),
	}

	// 尝试获取用户信息，失败时记录原因（仅 debug 级别，避免日志噪音）
	claims, err := GetJwtClaims(ctx)
	if err != nil {
		// 仅记录 debug 日志（排查问题用，不污染生产日志）
		log.Debug("获取日志附加用户信息失败", zap.Error(err))
		return log.With(fields...)
	}

	// 附加用户ID字段
	fields = append(fields, zap.Uint32("user_id", claims.UserID))
	return log.With(fields...)
}
