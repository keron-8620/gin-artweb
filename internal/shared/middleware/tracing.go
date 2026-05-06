// pkg/middleware/tracing.go
package middleware

import (
	"context"
	"gin-artweb/internal/shared/ctxutil"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func TracingMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 生成或获取请求ID
		traceID := c.GetHeader("X-Trace-ID")
		if traceID == "" {
			traceID = uuid.NewString()
		}
		ctx := context.WithValue(c.Request.Context(), ctxutil.TraceIDKey, traceID)
		c.Request = c.Request.WithContext(ctx)
		c.Header("X-Trace-ID", traceID)

		// 开始时间
		start := time.Now()

		// 记录请求开始
		logger.Debug(
			"请求开始",
			zap.String("trace_id", traceID),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("client_ip", c.ClientIP()),
		)

		// 处理请求
		c.Next()

		// 计算请求耗时
		duration := time.Since(start)
		c.Header("X-Cost-MS", strconv.FormatInt(duration.Milliseconds(), 10))

		statusCode := c.Writer.Status()
		endLog := logger.Info
		if statusCode >= http.StatusBadRequest {
			endLog = logger.Error
		}
		// 记录请求结束
		endLog(
			"请求结束",
			zap.String("trace_id", traceID),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status_code", c.Writer.Status()),
			zap.Duration("total_duration", duration),
		)
	}
}
