package test

import (
	"context"

	"gin-artweb/internal/shared/auth"
	"gin-artweb/internal/shared/ctxutil"
)

// createTestContext 创建带有 JwtClaims 的测试上下文
func CreateTestContext() context.Context {
	// 创建测试用的 JwtClaims
	claims := &auth.JwtClaims{
		UserInfo: auth.UserInfo{
			UserID:   1,
			Username: "test_user",
			RoleID:   1,
			IsStaff:  true,
		},
	}
	// 将 JwtClaims 添加到上下文中
	ctx := context.Background()
	ctx = context.WithValue(ctx, ctxutil.JwtClaimsKey, claims)
	return ctx
}
