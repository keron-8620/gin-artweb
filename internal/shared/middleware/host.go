package middleware

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"gin-artweb/internal/shared/errors"
)

// HostGuard 中间件用于检查请求的 Host 头是否被允许
func HostGuard(logger *zap.Logger, allowedHosts ...string) gin.HandlerFunc {
	allowed := make(map[string]bool)
	for _, host := range allowedHosts {
		allowed[host] = true
	}

	return func(c *gin.Context) {
		host := c.Request.Host
		if !allowed[host] {
			logger.Warn(
				"请求头不被允许",
				zap.String("host", host),
				zap.String("remote_ip", c.ClientIP()),
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.Strings("allowed_hosts", allowedHosts),
			)
			rErr := errors.ErrHostHeaderInvalid.WithField("host", host)
			errors.RespondWithError(c, rErr)
			return
		}
		c.Next()
	}
}
