package common

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"gin-artweb/internal/shared/errors"
)

func ShouldBind(ctx *gin.Context, logger *zap.Logger, v any, msg string) bool {
	if err := ctx.ShouldBind(&v); err != nil {
		logger.Error(
			msg,
			zap.Error(err),
			zap.String("request_uri", ctx.Request.RequestURI),
			zap.String("request_method", ctx.Request.Method),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return false
	}
	return true
}

func ShouldBindQuery(ctx *gin.Context, logger *zap.Logger, v any, msg string) bool {
	if err := ctx.ShouldBindQuery(&v); err != nil {
		logger.Error(
			msg,
			zap.Error(err),
			zap.String("request_uri", ctx.Request.RequestURI),
			zap.String("request_method", ctx.Request.Method),
		)
		rErr := errors.ErrValidationFailed.WithCause(err)
		errors.RespondWithError(ctx, rErr)
		return false
	}
	return true
}
