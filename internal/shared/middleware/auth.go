package middleware

import (
	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"gin-artweb/internal/shared/auth"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/errors"
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

func JWTAuthMiddleware(c *auth.JWTConfig, logger *zap.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 从请求头获取token
		token := extractToken(ctx)
		if token == "" {
			errors.RespondWithError(ctx, errors.ErrUnauthorized)
			return
		}

		// 身份认证
		claims, pErr := auth.ParseAccessToken(ctx, c, token)
		if pErr != nil {
			logger.Error(
				"身份认证失败",
				zap.Error(pErr),
			)
			errors.RespondWithError(ctx, pErr)
			return
		}

		ctx.Set(ctxutil.UserClaimsKey, claims)
		ctx.Next()
	}
}

func CasbinAuthMiddleware(enforcer *casbin.Enforcer, logger *zap.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		claims, ucErr := ctxutil.GetUserClaims(ctx)
		if ucErr != nil {
			logger.Error(
				"获取用户声明失败",
				zap.Error(ucErr),
			)
			errors.RespondWithError(ctx, ucErr)
			return
		}

		role := auth.RoleToSubject(claims.RoleID)
		fullPath := ctx.FullPath()

		// 访问鉴权
		hasPerm, err := enforcer.Enforce(role, fullPath, ctx.Request.Method)
		if err != nil {
			logger.Error(
				"权限校验失败",
				zap.Error(err),
				zap.String(auth.SubKey, role),
				zap.String(auth.ObjKey, fullPath),
				zap.String(auth.ActKey, ctx.Request.Method),
			)
			errors.RespondWithError(ctx, errors.FromError(err))
			return
		}
		if !hasPerm {
			logger.Error(
				"权限被拒绝",
				zap.String(auth.SubKey, role),
				zap.String(auth.ObjKey, fullPath),
				zap.String(auth.ActKey, ctx.Request.Method),
			)
			errors.RespondWithError(ctx, errors.ErrForbidden)
			return
		}
		ctx.Next()
	}
}
