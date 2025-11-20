package middleware

import (
	"context"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func isWebSocketRequest(c *gin.Context) bool {
	// 检查Upgrade头部
	upgrade := c.GetHeader("Upgrade")
	connection := c.GetHeader("Connection")

	// WebSocket请求特征
	return strings.ToLower(upgrade) == "websocket" &&
		strings.Contains(strings.ToLower(connection), "upgrade")
}

func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isWebSocketRequest(c) {
			// 创建带超时的 context
			ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
			defer cancel()

			// 替换请求的 context
			c.Request = c.Request.WithContext(ctx)
		}
		c.Next()
	}
}
