package auth

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

const (
	permissionSubjectFormat = "perm_%d"
	menuSubjectFormat       = "menu_%d"
	buttonSubjectFormat     = "button_%d"
	roleSubjectFormat       = "role_%d"
)

func PermissionToSubject(pk uint32) string {
	return fmt.Sprintf(permissionSubjectFormat, pk)
}

func MenuToSubject(pk uint32) string {
	return fmt.Sprintf(menuSubjectFormat, pk)
}

func ButtonToSubject(pk uint32) string {
	return fmt.Sprintf(buttonSubjectFormat, pk)
}

func RoleToSubject(pk uint32) string {
	return fmt.Sprintf(roleSubjectFormat, pk)
}

const (
	contextUserKey = "claims"
)

func GetUserClaims(ctx *gin.Context) *UserClaims {
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
