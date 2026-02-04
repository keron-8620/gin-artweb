package middleware

import (
	goerrors "errors"
	"fmt"
	"net/http"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"

	"gin-artweb/internal/shared/auth"
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

func JWTAuthMiddleware(jwtKey string, logger *zap.Logger) gin.HandlerFunc {
	key := []byte(jwtKey)

	return func(c *gin.Context) {
		code := http.StatusUnauthorized
		// 从请求头获取token
		token := extractToken(c)
		if token == "" {
			c.AbortWithStatusJSON(code, errors.ErrorResponse(code, errors.ErrUnauthorized))
			return
		}

		// 身份认证
		parsedToken, err := jwt.ParseWithClaims(token, &auth.UserClaims{}, func(token *jwt.Token) (any, error) {
			return key, nil
		})
		if err != nil {
			if goerrors.Is(err, jwt.ErrTokenExpired) {
				c.AbortWithStatusJSON(code, errors.ErrorResponse(code, errors.ErrTokenExpired))
			} else {
				c.AbortWithStatusJSON(code, errors.ErrorResponse(code, errors.ErrTokenInvalid))
			}
			return
		}
		// 验证token有效性
		if !parsedToken.Valid {
			c.AbortWithStatusJSON(code, errors.ErrorResponse(code, errors.ErrTokenInvalid))
			return
		}

		// 类型断言获取claims
		info, ok := parsedToken.Claims.(*auth.UserClaims)
		if !ok {
			logger.Error(
				"用户声明类型断言失败",
				zap.Any(auth.UserClaimsKey, parsedToken.Claims),
				zap.String("type", fmt.Sprintf("%T", parsedToken.Claims)),
			)
			c.AbortWithStatusJSON(code, errors.ErrorResponse(code, errors.ErrTokenInvalid))
			return
		}

		c.Set(auth.UserClaimsKey, info)
		c.Next()
	}
}

func CasbinAuthMiddleware(enforcer *casbin.Enforcer, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		code := http.StatusUnauthorized
		userClaims, exists := c.Get(auth.UserClaimsKey)
		if !exists {
			c.AbortWithStatusJSON(code, errors.ErrorResponse(code, errors.ErrMissingAuth))
			return
		}

		info, ok := userClaims.(*auth.UserClaims)
		if !ok {
			logger.Error(
				"用户声明类型断言失败",
				zap.Any(auth.UserClaimsKey, userClaims),
				zap.String("type", fmt.Sprintf("%T", userClaims)),
			)
			c.AbortWithStatusJSON(code, errors.ErrorResponse(code, errors.ErrMissingAuth))
			return
		}

		role := auth.RoleToSubject(info.RoleID)
		fullPath := c.FullPath()

		// 访问鉴权
		hasPerm, err := enforcer.Enforce(role, fullPath, c.Request.Method)
		if err != nil {
			logger.Error(
				"权限校验失败",
				zap.Error(err),
				zap.String(auth.SubKey, role),
				zap.String(auth.ObjKey, fullPath),
				zap.String(auth.ActKey, c.Request.Method),
			)
			rErr := errors.FromError(err).WithCause(err)
			code = http.StatusInternalServerError
			c.AbortWithStatusJSON(code, errors.ErrorResponse(code, rErr))
			return
		}
		if !hasPerm {
			logger.Error(
				"权限被拒绝",
				zap.String(auth.SubKey, role),
				zap.String(auth.ObjKey, fullPath),
				zap.String(auth.ActKey, c.Request.Method),
			)
			code = http.StatusForbidden
			c.AbortWithStatusJSON(code, errors.ErrorResponse(code, errors.ErrForbidden))
			return
		}
		c.Next()
	}
}
