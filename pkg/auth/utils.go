package auth

import "github.com/gin-gonic/gin"

const (
	contextUserKey = "claims"
)

func GetGinUserClaims(ctx *gin.Context) *UserClaims {
	if claims, exists := ctx.Get(contextUserKey); exists {
		if claims, ok := claims.(*UserClaims); ok {
			return claims
		}
	}
	return nil
}

func SetUserClaims(ctx *gin.Context, claims UserClaims) {
	ctx.Set(contextUserKey, &claims)
}
