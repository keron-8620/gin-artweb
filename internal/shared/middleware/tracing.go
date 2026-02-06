// pkg/middleware/tracing.go
package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"gin-artweb/internal/shared/ctxutil"
)

func TracingMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 生成或获取请求ID
		traceID := uuid.NewString()
		c.Set(ctxutil.TraceIDKey, traceID)

		// 开始时间
		start := time.Now()

		// 记录请求开始
		logger.Info("请求开始",
			zap.String(ctxutil.TraceIDKey, traceID),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("client_ip", c.ClientIP()),
		)

		// 处理请求
		c.Next()

		// 记录请求结束
		logger.Info("请求结束",
			zap.String(ctxutil.TraceIDKey, traceID),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status_code", c.Writer.Status()),
			zap.Duration("latency", time.Since(start)),
		)
	}
}
