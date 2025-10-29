package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// extractToken 从不同位置提取 token
func extractToken(c *gin.Context) string {
	// 检查是否为 WebSocket 升级请求
	if c.GetHeader("Connection") == "upgrade" &&
		c.GetHeader("Upgrade") == "websocket" {
		// WebSocket 请求优先从查询参数获取，其次从头部获取
		if token := c.Query("Authorization"); token != "" {
			return token
		}
		if token := c.GetHeader("Sec-WebSocket-Protocol"); token != "" {
			return token
		}
		return ""
	}

	// HTTP 请求从 Authorization 头部获取
	return c.GetHeader("Authorization")
}

func AuthMiddleware(enforcer *AuthEnforcer, loginUrl string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否是 API 请求
		if !strings.HasPrefix(c.Request.URL.Path, "/api") {
			c.Next()
			return
		}
		// 检查是否是登陆请求
		if c.Request.URL.Path == loginUrl && c.Request.Method == http.MethodPost {
			c.Next()
			return
		}
		// 从请求头获取token
		token := extractToken(c)
		if token == "" {
			c.JSON(ErrNoAuthor.Code, ErrNoAuthor.Reply())
			c.Abort()
			return
		}
		// 身份认证
		info, err := enforcer.Authentication(token)
		if err != nil {
			c.JSON(err.Code, err.Reply())
			c.Abort()
			return
		}
		// 访问鉴权
		hasPerm, err := enforcer.Authorization(info.Role, c.Request.URL.Path, c.Request.Method)
		if err != nil {
			c.JSON(err.Code, err.Reply())
			c.Abort()
			return
		}
		if !hasPerm {
			c.JSON(ErrForbidden.Code, ErrForbidden.Reply())
			c.Abort()
			return
		}
		SetUserClaims(c, *info)
		c.Next()
	}
}
