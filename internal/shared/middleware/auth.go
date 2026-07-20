package middleware

import (
	"context"
	"strings"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"gin-artweb/internal/shared/auth"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/errors"
)

// extractToken 从不同位置提取 token
func extractToken(c *gin.Context) string {
	// 1. 从 Authorization 头提取：必须严格是 Bearer
	auth := c.GetHeader("Authorization")
	if token := strictBearerExtract(auth); token != "" {
		return token
	}

	// 2. WebSocket 场景：Authorization 头宽松匹配，再回退到子协议
	if strings.EqualFold(c.GetHeader("Upgrade"), "websocket") {
		if token := looseTokenExtract(auth); token != "" {
			return token
		}
		protocols := strings.Split(c.GetHeader("Sec-WebSocket-Protocol"), ",")
		if len(protocols) > 0 {
			first := strings.TrimSpace(protocols[0])
			if token := looseTokenExtract(first); token != "" {
				return token
			}
		}
	}
	return ""
}

// strictBearerExtract 仅用于标准头：非 Bearer 一律视为无效
func strictBearerExtract(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	fields := strings.Fields(value)
	if len(fields) >= 2 && strings.EqualFold(fields[0], "Bearer") {
		return fields[1]
	}
	return ""
}

// looseTokenExtract 用于 WebSocket 回退：支持 Bearer，也支持纯 token
func looseTokenExtract(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	fields := strings.Fields(value)
	if len(fields) >= 2 && strings.EqualFold(fields[0], "Bearer") {
		return fields[1] // 带了 Bearer 前缀，取后面的
	}
	return value // 没有 Bearer 前缀，把整个字符串当作 token
}

func JWTAuthMiddleware(c *config.JWTConfig, logger *zap.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 从请求头获取token
		token := extractToken(ctx)
		if token == "" {
			errors.RespondWithError(ctx, errors.ErrMissingAuth)
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

		newCtx := context.WithValue(ctx.Request.Context(), ctxutil.JwtClaimsKey, claims)
		ctx.Request = ctx.Request.WithContext(newCtx)
		ctx.Next()
	}
}

func CasbinAuthMiddleware(enforcer *casbin.Enforcer, logger *zap.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		traceID := ctxutil.GetTraceID(ctx.Request.Context())
		claims, err := ctxutil.GetJwtClaims(ctx.Request.Context())
		if err != nil {
			logger.Error(
				"获取JWT claims失败",
				zap.Error(err),
				zap.String("trace_id", traceID),
			)
			errors.RespondWithError(ctx, errors.ErrAuthFailed)
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
				zap.String("sub", role),
				zap.String("obj", fullPath),
				zap.String("act", ctx.Request.Method),
				zap.String("trace_id", traceID),
			)
			errors.RespondWithError(ctx, errors.FromError(err))
			return
		}
		if !hasPerm {
			logger.Error(
				"权限被拒绝",
				zap.String("sub", role),
				zap.String("obj", fullPath),
				zap.String("act", ctx.Request.Method),
				zap.String("trace_id", traceID),
			)
			errors.RespondWithError(ctx, errors.ErrForbidden)
			return
		}
		ctx.Next()
	}
}
