package ctxutil

import (
	"context"

	"gin-artweb/internal/shared/auth"
	"gin-artweb/internal/shared/errors"
)

const (
	UserIDKey     = "user_id"
	UserClaimsKey = "user_claims"
)

func GetUserClaims(ctx context.Context) (*auth.UserClaims, *errors.Error) {
	if ctx == nil {
		return nil, errors.ErrNoContext
	}
	value := ctx.Value(UserClaimsKey)
	if value == nil {
		return nil, errors.ErrMissingAuth
	}
	if userClaims, ok := value.(*auth.UserClaims); ok {
		return userClaims, nil
	}
	return nil, errors.ErrMissingAuth
}
