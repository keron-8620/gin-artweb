package middleware

import (
	"fmt"
	"runtime/debug"

	emperrors "emperror.dev/errors"
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
				var err error
				switch v := r.(type) {
				case error:
					err = emperrors.WithStack(v)
				case string:
					err = emperrors.New(v)
				default:
					err = emperrors.New(fmt.Sprintf("%v", v))
				}
				err = emperrors.WrapWithDetails(
					err,
					"panic recovered",
					"method", c.Request.Method,
					"url", c.Request.URL.String(),
					"client_ip", c.ClientIP(),
					"user_agent", c.Request.UserAgent(),
				)

				logger.Error("panic recovered",
					zap.Error(err),
					zap.Any("panic", r),
					zap.String("stack", string(stack)),
					zap.String("method", c.Request.Method),
					zap.String("url", c.Request.URL.String()),
					zap.String("client_ip", c.ClientIP()),
					zap.String("user_agent", c.Request.UserAgent()),
				)
				rErr := errors.FromError(err)
				errors.RespondWithError(c, rErr)
			}
		}()

		// 继续处理请求
		c.Next()
	}
}
