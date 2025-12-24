package auth

import (
	"context"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	UserIDKey     = "user_id"
	UserClaimsKey = "user_claims"
)

// UserClaims 用户Claims
type UserClaims struct {
	jwt.RegisteredClaims
	IsStaff bool   `json:"isf"` // 是否是工作人员
	UserID  uint32 `json:"uid"` // 用户ID
	RoleID  uint32 `json:"rid"` // 角色
}

// NewJWT 创建JWT
func NewJWT(secretKey []byte, u UserClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, u)
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// GenerateTokenID 生成令牌ID
// 返回UUID字符串作为令牌ID
func GenerateTokenID() string {
	return uuid.New().String()
}

// GetUserClaims 获取用户Claims
func GetUserClaims(ctx context.Context) *UserClaims {
	if userClaims, ok := ctx.Value(UserClaimsKey).(*UserClaims); ok {
        return userClaims
    }
    return nil
}
