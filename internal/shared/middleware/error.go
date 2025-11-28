package middleware

import (
	goerrors "errors"
	"fmt"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"gin-artweb/internal/shared/errors"
)

// ErrorMiddleware 异常处理中间件
// 拦截所有panic和错误，进行统一处理和响应
func ErrorMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				// 中止请求处理
				c.Abort()

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
					zap.Any("panic", r),
					zap.String("stack", string(stack)),
					zap.String("method", c.Request.Method),
					zap.String("url", c.Request.URL.String()),
					zap.String("client_ip", c.ClientIP()),
					zap.String("user_agent", c.Request.UserAgent()),
				)
				err := errors.FromError(goerrors.New(errMsg))
				c.JSON(err.Code, err.Reply())
			}
		}()

		// 继续处理请求
		c.Next()
	}
}
