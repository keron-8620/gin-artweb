package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
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

func AuthMiddleware(enforcer *AuthEnforcer, logger *zap.Logger, loginUrl string) gin.HandlerFunc {
	return func(c *gin.Context) {
		fullPath := c.FullPath()

		// 检查是否是 API 请求
		if !strings.HasPrefix(fullPath, "/api") {
			c.Next()
			return
		}

		// 检查是否是登陆请求
		if fullPath == loginUrl && c.Request.Method == http.MethodPost {
			c.Next()
			return
		}

		// 从请求头获取token
		token := extractToken(c)
		if token == "" {
			c.AbortWithStatusJSON(ErrNoAuthor.Code, ErrNoAuthor.ToMap())
			return
		}

		// 身份认证
		info, err := enforcer.Authentication(c, token)
		if err != nil {
			c.AbortWithStatusJSON(err.Code, err.ToMap())
			return
		}
		role := RoleToSubject(info.RoleID)

		// 访问鉴权
		hasPerm, err := enforcer.Authorization(c, role, fullPath, c.Request.Method)
		if err != nil {
			c.AbortWithStatusJSON(err.Code, err.ToMap())
			return
		}
		if !hasPerm {
			logger.Error(
				"权限校验失败",
				zap.String(SubKey, role),
				zap.String(ObjKey, fullPath),
				zap.String(ActKey, c.Request.Method),
			)
			c.AbortWithStatusJSON(ErrForbidden.Code, ErrForbidden.ToMap())
			return
		}
		SetUserClaims(c, *info)
		c.Next()
	}
}
