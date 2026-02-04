package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"

	"gin-artweb/internal/shared/errors"
)

// TimestampMiddleware 创建防重放攻击中间件
// logger: 日志记录器
// tolerance: 时间容忍度（毫秒），默认300000（5分钟）
// futureTolerance: 允许未来时间的容忍度（毫秒），默认60000（1分钟）
func TimestampMiddleware(nonceStore *cache.Cache, logger *zap.Logger, tolerance, futureTolerance int64, defaultExpiration time.Duration) gin.HandlerFunc {
	// 设置默认值
	if tolerance <= 0 {
		tolerance = 300000 // 默认5分钟
	}
	if futureTolerance <= 0 {
		futureTolerance = 60000 // 默认1分钟
	}

	return func(c *gin.Context) {
		// 检查是否是 API 请求
		if !strings.HasPrefix(c.Request.URL.Path, "/api") {
			c.Next()
			return
		}

		code := http.StatusBadRequest

		// 从请求头获取 X-Timestamp
		timestampStr := c.GetHeader("X-Timestamp")
		if timestampStr == "" {
			logger.Error("请求缺少 X-Timestamp 头")
			c.AbortWithStatusJSON(code, errors.ErrorResponse(code, errors.ErrTimestampNotFound))
			return
		}

		// 从请求头获取 X-Nonce
		nonce := c.GetHeader("X-Nonce")
		if nonce == "" {
			logger.Error("请求缺少 X-Nonce 头")
			c.AbortWithStatusJSON(code, errors.ErrorResponse(code, errors.ErrNonceNotFound))
			return
		}

		// 解析时间戳
		timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
		if err != nil {
			logger.Error(
				"请求时间戳解释失败",
				zap.String("timestamp", timestampStr),
				zap.String("error", err.Error()),
			)
			c.AbortWithStatusJSON(code, errors.ErrorResponse(code, errors.ErrTimestampInvalid))
			return
		}

		// 验证时间戳范围（防止异常大的时间戳）
		now := time.Now().UnixMilli()

		// 检查时间戳是否过期
		if abs(now-timestamp) > tolerance || timestamp > now+futureTolerance {
			logger.Error(
				"时间戳超出允许范围",
				zap.Int64("current", now),
				zap.Int64("received", timestamp),
				zap.Int64("tolerance", tolerance),
				zap.Int64("difference", abs(now-timestamp)),
			)
			c.AbortWithStatusJSON(code, errors.ErrorResponse(code, errors.ErrTimestampExpired))
			return
		}

		if _, exists := nonceStore.Get(nonce); exists {
			logger.Error(
				"检测到重复的请求，可能存在重放攻击",
				zap.String("nonce", nonce),
				zap.Int64("timestamp", timestamp),
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.String("client_ip", c.ClientIP()),
			)
			c.AbortWithStatusJSON(code, errors.ErrorResponse(code, errors.ErrReplayAttack))
			return
		}

		nonceStore.Set(nonce, true, defaultExpiration)
		c.Next()
	}
}

// abs 返回绝对值
func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}
