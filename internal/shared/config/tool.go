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
	if _, ok := supportedJWTMethods[accessMethodstr]; !ok {
		panic("invalid access method")
	}
	if _, ok := supportedJWTMethods[refreshMethodstr]; !ok {
		panic("invalid refresh method")
	}
	accessMethod := jwt.GetSigningMethod(accessMethodstr)
	refreshMethod := jwt.GetSigningMethod(refreshMethodstr)
	if len(accessSecret) < 32 || len(refreshSecret) < 32 {
		panic("JWT_ACCESS_SECRET and JWT_REFRESH_SECRET must be at least 32 bytes")
	}
	if string(accessSecret) == string(refreshSecret) {
		panic("JWT access and refresh secrets must be different")
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
