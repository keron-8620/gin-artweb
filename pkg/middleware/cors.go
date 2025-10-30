package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"gin-artweb/pkg/config"
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

	// 预处理允许的源，提高运行时性能
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

		// 如果没有 Origin 头或者 Origin 为空，跳过 CORS 处理
		if origin == "" {
			c.Next()
			return
		}

		// 检查是否允许该源
		allowed := allowAllOrigins || specificOrigins[origin]

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
		} else if len(specificOrigins) > 0 {
			// 检查是否匹配通配符模式（如果有需要的话）
			for allowedOrigin := range specificOrigins {
				if matchOriginPattern(allowedOrigin, origin) {
					c.Header("Access-Control-Allow-Origin", origin)
					allowed = true
					break
				}
			}
		}

		// 如果不允许该源，跳过 CORS 头设置
		if !allowed {
			c.Next()
			return
		}

		// 设置其他 CORS 头
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

// matchOriginPattern 简单的源模式匹配（可选）
func matchOriginPattern(pattern, origin string) bool {
	// 这里可以实现更复杂的模式匹配逻辑
	// 例如支持 *.example.com 这样的通配符
	return pattern == origin
}
