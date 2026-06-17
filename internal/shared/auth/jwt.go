package auth

import (
	"context"
	"time"

	emperror "emperror.dev/errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap/zapcore"

	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/errors"
)

type TokenType string

const (
	TokenTypeAccess  TokenType = "access"  // 访问令牌
	TokenTypeRefresh TokenType = "refresh" // 刷新令牌
)

type UserInfo struct {
	UserID   uint32 `json:"uid"` // 用户ID
	Username string `json:"un"`  // 用户名
	RoleID   uint32 `json:"rid"` // 角色
	IsStaff  bool   `json:"isf"` // 是否是工作人员
}

func (u *UserInfo) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddUint32("uid", u.UserID)
	enc.AddString("username", u.Username)
	enc.AddUint32("rid", u.RoleID)
	enc.AddBool("is_staff", u.IsStaff)
	return nil
}

// JwtClaims 用户Claims
type JwtClaims struct {
	jwt.RegisteredClaims
	UserInfo
	Type TokenType `json:"typ"` // 令牌类型
}

func NewUserClaims(c *config.JWTConfig, u UserInfo, tt TokenType) JwtClaims {
	now := time.Now()
	expiration := c.AccessTokenExpiration
	if tt == TokenTypeRefresh {
		expiration = c.RefreshTokenExpiration
	}
	return JwtClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    c.Issuer,
			Subject:   u.Username,
			ExpiresAt: jwt.NewNumericDate(now.Add(expiration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ID:        uuid.NewString(),
		},
		UserInfo: u,
		Type:     tt,
	}
}

// NewJWT 创建JWT
func NewAccessJWT(ctx context.Context, c *config.JWTConfig, u UserInfo) (string, error) {
	claims := NewUserClaims(c, u, TokenTypeAccess)
	token := jwt.NewWithClaims(c.AccessMethod, claims)
	tokenString, err := token.SignedString(c.AccessSecret)
	if err != nil {
		return "", emperror.WrapIf(err, "创建jwt失败")
	}
	return tokenString, nil
}

// NewRefreshJWT 创建刷新JWT
func NewRefreshJWT(ctx context.Context, c *config.JWTConfig, u UserInfo) (string, error) {
	claims := NewUserClaims(c, u, TokenTypeRefresh)
	token := jwt.NewWithClaims(c.RefreshMethod, claims)
	tokenString, err := token.SignedString(c.RefreshSecret)
	if err != nil {
		return "", emperror.WrapIf(err, "创建刷新jwt失败")
	}
	return tokenString, nil
}

// ParseAccessToken 解析并验证JWT令牌
func ParseAccessToken(ctx context.Context, c *config.JWTConfig, tokenString string) (*JwtClaims, *errors.Error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&JwtClaims{},
		func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.ErrTokenInvalid
			}
			return c.AccessSecret, nil
		},
	)

	if err != nil {
		return nil, errors.ErrTokenInvalid
	}

	if claims, ok := token.Claims.(*JwtClaims); ok && claims != nil && token.Valid {
		if claims.Type != TokenTypeAccess {
			return nil, errors.ErrTokenTypeMismatch
		}
		return claims, nil
	}

	return nil, errors.ErrTokenExpired
}

// ParseRefreshToken 解析并验证刷新JWT令牌
func ParseRefreshToken(ctx context.Context, c *config.JWTConfig, tokenString string) (*JwtClaims, *errors.Error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&JwtClaims{},
		func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.ErrTokenInvalid
			}
			return c.RefreshSecret, nil
		},
	)

	if err != nil {
		return nil, errors.ErrTokenInvalid
	}

	if claims, ok := token.Claims.(*JwtClaims); ok && token.Valid {
		if claims.Type != TokenTypeRefresh {
			return nil, errors.ErrTokenTypeMismatch
		}
		return claims, nil
	}

	return nil, errors.ErrTokenExpired
}
