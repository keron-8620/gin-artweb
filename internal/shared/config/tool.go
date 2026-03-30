package config

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

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
