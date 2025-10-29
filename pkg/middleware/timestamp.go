package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"gitee.com/keion8620/go-dango-gin/pkg/errors"
)

var (
	RsonNoTimestamp = errors.ReasonEnum{
		Reason: "no_timestamp",
		Msg:    "请求标头中缺少时间戳",
	}
	RsonInvalidTimestamp = errors.ReasonEnum{
		Reason: "invalid_timestamp",
		Msg:    "无效的时间戳",
	}
	RsonTimestampExpired = errors.ReasonEnum{
		Reason: "timestamp_expired",
		Msg:    "时间戳已过期, 请检查客户端时间同步",
	}
)

var (
	ErrNoTimestamp = errors.New(
		http.StatusBadRequest,
		RsonNoTimestamp.Reason,
		RsonNoTimestamp.Msg,
		nil,
	)
	ErrInvalidTimestamp = errors.New(
		http.StatusBadRequest,
		RsonInvalidTimestamp.Reason,
		RsonInvalidTimestamp.Msg,
		nil,
	)
	ErrTimestampExpired = errors.New(
		http.StatusBadRequest,
		RsonTimestampExpired.Reason,
		RsonTimestampExpired.Msg,
		nil,
	)
)

func AuthMiddleware(logger *zap.Logger, t int64) gin.HandlerFunc {
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

		// 检查时间戳是否过期
		now := time.Now().Unix()
		if abs(now-timestamp) > t { // 300 秒 = 5 分钟
			logger.Error("X-Timestamp expired",
				zap.Int64("current", now),
				zap.Int64("received", timestamp))
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
