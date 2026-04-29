package auth

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"

	"gin-artweb/internal/shared/config"
)

const (
	testAccessSecret  = "test-access-secret-key-1234567890123456"
	testRefreshSecret = "test-refresh-secret-key-1234567890123456"
)

func TestNewUserClaims(t *testing.T) {
	jwtConfig := &config.JWTConfig{
		Issuer:                "test-issuer",
		AccessTokenExpiration: time.Hour,
	}
	userInfo := UserInfo{
		UserID:   1,
		Username: "testuser",
		RoleID:   10,
		IsStaff:  true,
	}

	claims := NewUserClaims(jwtConfig, userInfo, TokenTypeAccess)

	assert.Equal(t, jwtConfig.Issuer, claims.Issuer)
	assert.Equal(t, userInfo.Username, claims.Subject)
	assert.Equal(t, userInfo.UserID, claims.UserID)
	assert.Equal(t, userInfo.Username, claims.Username)
	assert.Equal(t, userInfo.RoleID, claims.RoleID)
	assert.Equal(t, userInfo.IsStaff, claims.IsStaff)
	assert.Equal(t, TokenTypeAccess, claims.Type)
	assert.NotEmpty(t, claims.ID)
	assert.WithinDuration(t, time.Now().Add(jwtConfig.AccessTokenExpiration), claims.ExpiresAt.Time, time.Second)
}

func TestNewAccessJWT(t *testing.T) {
	ctx := context.Background()
	jwtConfig := &config.JWTConfig{
		Issuer:                "test-issuer",
		AccessTokenExpiration: 10 * 365 * 24 * time.Hour,
		AccessMethod:          jwt.SigningMethodHS256,
		AccessSecret:          []byte(testAccessSecret),
	}
	userInfo := UserInfo{
		UserID:   1,
		Username: "testuser",
		RoleID:   10,
		IsStaff:  true,
	}

	tokenString, err := NewAccessJWT(ctx, jwtConfig, userInfo)

	assert.Nil(t, err)
	assert.NotEmpty(t, tokenString)

	claims, err := ParseAccessToken(ctx, jwtConfig, tokenString)
	assert.Nil(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, userInfo.UserID, claims.UserID)
	assert.Equal(t, userInfo.Username, claims.Username)
	assert.Equal(t, TokenTypeAccess, claims.Type)
}

func TestNewRefreshJWT(t *testing.T) {
	ctx := context.Background()
	jwtConfig := &config.JWTConfig{
		Issuer:                 "test-issuer",
		AccessTokenExpiration:  10 * 365 * 24 * time.Hour,
		RefreshTokenExpiration: 10 * 365 * 24 * time.Hour,
		AccessMethod:           jwt.SigningMethodHS256,
		RefreshMethod:          jwt.SigningMethodHS256,
		AccessSecret:           []byte(testAccessSecret),
		RefreshSecret:          []byte(testRefreshSecret),
	}
	userInfo := UserInfo{
		UserID:   1,
		Username: "testuser",
		RoleID:   10,
		IsStaff:  true,
	}

	tokenString, err := NewRefreshJWT(ctx, jwtConfig, userInfo)

	assert.Nil(t, err)
	assert.NotEmpty(t, tokenString)

	claims, err := ParseRefreshToken(ctx, jwtConfig, tokenString)
	assert.Nil(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, userInfo.UserID, claims.UserID)
	assert.Equal(t, userInfo.Username, claims.Username)
	assert.Equal(t, TokenTypeRefresh, claims.Type)
}

func TestParseAccessToken_InvalidToken(t *testing.T) {
	ctx := context.Background()
	jwtConfig := &config.JWTConfig{
		Issuer:                "test-issuer",
		AccessTokenExpiration: time.Hour,
		AccessMethod:          jwt.SigningMethodHS256,
		AccessSecret:          []byte(testAccessSecret),
	}

	invalidTokens := []string{
		"invalid-token-string",
		"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
		"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ",
	}

	for _, token := range invalidTokens {
		claims, err := ParseAccessToken(ctx, jwtConfig, token)
		assert.Nil(t, claims)
		assert.NotNil(t, err)
	}
}

func TestParseAccessToken_TokenTypeMismatch(t *testing.T) {
	ctx := context.Background()
	jwtConfig := &config.JWTConfig{
		Issuer:                 "test-issuer",
		AccessTokenExpiration:  10 * 365 * 24 * time.Hour,
		RefreshTokenExpiration: 10 * 365 * 24 * time.Hour,
		AccessMethod:           jwt.SigningMethodHS256,
		RefreshMethod:          jwt.SigningMethodHS256,
		AccessSecret:           []byte(testAccessSecret),
		RefreshSecret:          []byte(testRefreshSecret),
	}
	userInfo := UserInfo{
		UserID:   1,
		Username: "testuser",
		RoleID:   10,
		IsStaff:  true,
	}

	refreshToken, _ := NewRefreshJWT(ctx, jwtConfig, userInfo)

	claims, err := ParseAccessToken(ctx, jwtConfig, refreshToken)

	assert.Nil(t, claims)
	assert.NotNil(t, err)
}

func TestParseRefreshToken_TokenTypeMismatch(t *testing.T) {
	ctx := context.Background()
	jwtConfig := &config.JWTConfig{
		Issuer:                 "test-issuer",
		AccessTokenExpiration:  10 * 365 * 24 * time.Hour,
		RefreshTokenExpiration: 10 * 365 * 24 * time.Hour,
		AccessMethod:           jwt.SigningMethodHS256,
		RefreshMethod:          jwt.SigningMethodHS256,
		AccessSecret:           []byte(testAccessSecret),
		RefreshSecret:          []byte(testRefreshSecret),
	}
	userInfo := UserInfo{
		UserID:   1,
		Username: "testuser",
		RoleID:   10,
		IsStaff:  true,
	}

	accessToken, _ := NewAccessJWT(ctx, jwtConfig, userInfo)

	claims, err := ParseRefreshToken(ctx, jwtConfig, accessToken)

	assert.Nil(t, claims)
	assert.NotNil(t, err)
}

func TestUserInfo_MarshalLogObject(t *testing.T) {
	userInfo := UserInfo{
		UserID:   1,
		Username: "testuser",
		RoleID:   10,
		IsStaff:  true,
	}

	encoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{})
	err := userInfo.MarshalLogObject(encoder)
	assert.Nil(t, err)
}
