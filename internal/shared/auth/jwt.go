package auth

import (
	"context"
	"time"

	emperror "emperror.dev/errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

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

// UserClaims 用户Claims
type UserClaims struct {
	jwt.RegisteredClaims
	UserInfo
	Type TokenType `json:"typ"` // 令牌类型
}

type JWTConfig struct {
	Issuer                 string            // 令牌签发者
	AccessTokenExpiration  time.Duration     // 访问令牌过期时间
	RefreshTokenExpiration time.Duration     // 刷新令牌过期时间
	AccessSecret           []byte            // 访问令牌密钥
	RefreshSecret          []byte            // 刷新令牌密钥
	AccessMethod           jwt.SigningMethod // 访问令牌签名方法
	RefreshMethod          jwt.SigningMethod // 刷新令牌签名方法
}

func NewJWTConfig(
	accessExpiration, refreshExpiration time.Duration,
	accessMethodstr, refreshMethodstr string,
	accessSecret, refreshSecret []byte,
) *JWTConfig {
	accessMethod := jwt.GetSigningMethod(accessMethodstr)
	if accessMethod == nil {
		panic("invalid access method")
	}
	refreshMethod := jwt.GetSigningMethod(refreshMethodstr)
	if refreshMethod == nil {
		panic("invalid refresh method")
	}
	if len(accessSecret) == 0 || len(refreshSecret) == 0 {
		panic("JWT_ACCESS_SECRET or JWT_REFRESH_SECRET is empty")
	}
	return &JWTConfig{
		AccessTokenExpiration:  accessExpiration,
		RefreshTokenExpiration: refreshExpiration,
		AccessSecret:           accessSecret,
		RefreshSecret:          refreshSecret,
		AccessMethod:           accessMethod,
		RefreshMethod:          refreshMethod,
	}
}

func newUserClaims(c *JWTConfig, u UserInfo, tt TokenType) UserClaims {
	now := time.Now()
	return UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    c.Issuer,
			Subject:   u.Username,
			ExpiresAt: jwt.NewNumericDate(now.Add(c.AccessTokenExpiration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ID:        uuid.NewString(),
		},
		UserInfo: u,
		Type:     tt,
	}
}

// NewJWT 创建JWT
func NewAccessJWT(ctx context.Context, c *JWTConfig, u UserInfo) (string, error) {
	if ctx.Err() != nil {
		return "", emperror.WrapIf(ctx.Err(), "上下文已取消/超时")
	}
	claims := newUserClaims(c, u, TokenTypeAccess)
	token := jwt.NewWithClaims(c.AccessMethod, claims)
	tokenString, err := token.SignedString(c.AccessSecret)
	if err != nil {
		return "", emperror.WrapIf(err, "创建jwt失败")
	}
	return tokenString, nil
}

// NewRefreshJWT 创建刷新JWT
func NewRefreshJWT(ctx context.Context, c *JWTConfig, u UserInfo) (string, error) {
	if ctx.Err() != nil {
		return "", emperror.WrapIf(ctx.Err(), "上下文已取消/超时")
	}
	claims := newUserClaims(c, u, TokenTypeRefresh)
	token := jwt.NewWithClaims(c.RefreshMethod, claims)
	tokenString, err := token.SignedString(c.RefreshSecret)
	if err != nil {
		return "", emperror.WrapIf(err, "创建刷新jwt失败")
	}
	return tokenString, nil
}

// ParseAccessToken 解析并验证JWT令牌
func ParseAccessToken(ctx context.Context, c *JWTConfig, tokenString string) (*UserClaims, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}
	token, err := jwt.ParseWithClaims(
		tokenString,
		&UserClaims{},
		func(token *jwt.Token) (any, error) {
			return c.AccessSecret, nil
		},
	)

	if err != nil {
		return nil, errors.ErrTokenInvalid
	}

	if claims, ok := token.Claims.(*UserClaims); ok && claims != nil && token.Valid {
		if claims.Type != TokenTypeAccess {
			return nil, errors.ErrTokenTypeMismatch
		}
		return claims, nil
	}

	return nil, errors.ErrTokenExpired
}

// ParseRefreshToken 解析并验证刷新JWT令牌
func ParseRefreshToken(ctx context.Context, c *JWTConfig, tokenString string) (*UserClaims, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}
	token, err := jwt.ParseWithClaims(
		tokenString,
		&UserClaims{},
		func(token *jwt.Token) (any, error) {
			return c.RefreshSecret, nil
		},
	)

	if err != nil {
		return nil, errors.ErrTokenInvalid
	}

	if claims, ok := token.Claims.(*UserClaims); ok && token.Valid {
		if claims.Type != TokenTypeRefresh {
			return nil, errors.ErrTokenTypeMismatch
		}
		return claims, nil
	}

	return nil, errors.ErrTokenExpired
}
