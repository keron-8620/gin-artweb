package auth

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type UserClaims struct {
	jwt.RegisteredClaims
	IsStaff bool   `json:"isf"`  // 是否是工作人员
	UserId  uint32 `json:"uid"`  // 用户ID
	Role    string `json:"role"` // 角色
}

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
