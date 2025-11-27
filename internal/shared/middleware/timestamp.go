package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"gin-artweb/internal/shared/errors"
)

var (
	ErrNoTimestamp = errors.New(
		http.StatusBadRequest,
		"no_timestamp",
		"请求标头中缺少时间戳",
		nil,
	)
	ErrInvalidTimestamp = errors.New(
		http.StatusBadRequest,
		"invalid_timestamp",
		"无效的时间戳",
		nil,
	)
	ErrTimestampExpired = errors.New(
		http.StatusBadRequest,
		"timestamp_expired",
		"时间戳已过期, 请检查客户端时间同步",
		nil,
	)
	ErrTimestampInFuture = errors.New(
		http.StatusBadRequest,
		"timestamp_in_future",
		"时间戳指向未来时间，请检查客户端时间同步",
		nil,
	)
)

// TimestampMiddleware 创建防重放攻击中间件
// logger: 日志记录器
// tolerance: 时间容忍度（毫秒），默认300000（5分钟）
// futureTolerance: 允许未来时间的容忍度（毫秒），默认60000（1分钟）
func TimestampMiddleware(logger *zap.Logger, tolerance, futureTolerance int64) gin.HandlerFunc {
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

		// 从请求头获取 X-Timestamp
		timestampStr := c.GetHeader("X-Timestamp")
		if timestampStr == "" {
			logger.Error("请求缺少 X-Timestamp 头")
			c.JSON(ErrNoTimestamp.Code, ErrNoTimestamp.Reply())
			c.Abort()
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
			c.JSON(ErrInvalidTimestamp.Code, ErrInvalidTimestamp.Reply())
			c.Abort()
			return
		}

		// 获取当前时间（毫秒）
		now := time.Now().UnixMilli()

		// 检查时间戳是否指向过于遥远的未来
		if timestamp > now+futureTolerance {
			logger.Error("X-Timestamp is too far in the future",
				zap.Int64("current", now),
				zap.Int64("received", timestamp),
				zap.Int64("difference", timestamp-now))
			c.JSON(ErrTimestampInFuture.Code, ErrTimestampInFuture.Reply())
			c.Abort()
			return
		}

		// 检查时间戳是否过期
		if abs(now-timestamp) > tolerance {
			logger.Error("X-Timestamp expired",
				zap.Int64("current", now),
				zap.Int64("received", timestamp),
				zap.Int64("tolerance", tolerance),
				zap.Int64("difference", abs(now-timestamp)))
			c.JSON(ErrTimestampExpired.Code, ErrTimestampExpired.Reply())
			c.Abort()
			return
		}

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
