package middleware

import (
	"fmt"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"gitee.com/keion8620/go-dango-gin/pkg/errors"
)

// ErrorMiddleware 异常处理中间件
// 拦截所有panic和错误，进行统一处理和响应
func ErrorMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				// 记录panic信息和堆栈跟踪
				stack := debug.Stack()

				// 构造错误响应
				var errMsg string
				switch v := r.(type) {
				case error:
					errMsg = v.Error()
				case string:
					errMsg = v
				default:
					errMsg = fmt.Sprintf("%v", v)
				}

				logger.Error("panic recovered",
					zap.String("error", errMsg),
					zap.String("method", c.Request.Method),
					zap.String("url", c.Request.URL.String()),
					zap.String("client_ip", c.ClientIP()),
					zap.String("user_agent", c.Request.UserAgent()),
					zap.String("stack", string(stack)),
				)
				c.JSON(errors.UnknowError.Code, errors.UnknowError.Reply())
				// 中止请求处理
				c.Abort()
			}
		}()

		// 继续处理请求
		c.Next()

		// 只有在没有panic且存在错误时才处理Gin上下文中的错误
		if len(c.Errors) > 0 && !c.IsAborted() {
			err := c.Errors.Last()
			if err != nil {
				customErr := errors.FromError(err)
				logger.Error("server error",
					zap.String("error", customErr.Error()),
					zap.String("method", c.Request.Method),
					zap.String("url", c.Request.URL.String()),
				)
				c.JSON(customErr.Code, customErr.Reply())
				c.Abort()
			}
		}
	}
}
