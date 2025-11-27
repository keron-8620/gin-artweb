package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"gin-artweb/internal/shared/config"
)

// CorsMiddleware 跨域中间件
func CorsMiddleware(cfg *config.AllowConfig) gin.HandlerFunc {
	// 如果配置为空，使用默认配置
	if cfg == nil {
		cfg = &config.AllowConfig{
			AllowOrigins:     []string{"*"},
			AllowCredentials: false,
			AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		}
	}

	// 预处理允许的源
	allowAllOrigins := false
	specificOrigins := make(map[string]bool)

	for _, origin := range cfg.AllowOrigins {
		if origin == "*" {
			allowAllOrigins = true
			break
		}
		specificOrigins[origin] = true
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// 如果没有 Origin 头，跳过 CORS 处理
		if origin == "" {
			c.Next()
			return
		}

		// 检查是否允许该源
		allowed := allowAllOrigins || specificOrigins[origin]

		// 如果不允许该源，跳过 CORS 头设置
		if !allowed {
			c.Next()
			return
		}

		// 设置 CORS 头
		if allowAllOrigins {
			c.Header("Access-Control-Allow-Origin", "*")
		} else {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		if len(cfg.AllowMethods) > 0 {
			c.Header("Access-Control-Allow-Methods", strings.Join(cfg.AllowMethods, ", "))
		}

		if len(cfg.AllowHeaders) > 0 {
			c.Header("Access-Control-Allow-Headers", strings.Join(cfg.AllowHeaders, ", "))
		}

		if cfg.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		// 处理预检请求
		if c.Request.Method == "OPTIONS" {
			c.Header("Access-Control-Max-Age", "86400")
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}