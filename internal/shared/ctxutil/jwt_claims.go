package ctxutil

import (
	"context"

	"emperror.dev/errors"

	"gin-artweb/internal/shared/auth"
)

const (
	UserIDKey    = "user_id"
	JwtClaimsKey = "jwt_claims"
)

func GetJwtClaims(ctx context.Context) (*auth.JwtClaims, error) {
	if ctx == nil {
		return nil, errors.New("获取用户信息失败：context 不能为空")
	}
	value := ctx.Value(JwtClaimsKey)
	if value == nil {
		return nil, errors.New("获取用户信息失败：认证信息缺失")
	}
	jwtClaims, ok := value.(*auth.JwtClaims)
	if !ok {
		return nil, errors.New("获取用户信息失败：认证信息格式错误")
	}
	return jwtClaims, nil
}

func MustGetJwtClaims(ctx context.Context) *auth.JwtClaims {
	claims, err := GetJwtClaims(ctx)
	if err != nil {
		panic(err)
	}
	return claims
}

func SetJwtClaims(ctx context.Context, claims *auth.JwtClaims) context.Context {
	return context.WithValue(ctx, JwtClaimsKey, claims)
}
