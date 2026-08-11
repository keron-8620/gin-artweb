package common

import (
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	"gin-artweb/internal/shared/errors"
)

const (
	// NoPathTraversal 段校验: 用于校验按要求作为单层路径段接入的字符串,
	// 不允许包含路径分隔符、当前目录(.)、上级目录(..)等可造成路径穿透的字符。
	NoPathTraversal = "nopathtraversal"
)

func init() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		_ = v.RegisterValidation(NoPathTraversal, validateNoPathTraversal)
	}
}

// validateNoPathTraversal 仅允许可作为单个路径安全片段的字符串:
// 不允许为绝对路径、不允许包含路径分隔符、不允许为当前/上级目录。
func validateNoPathTraversal(fl validator.FieldLevel) bool {
	return isSafePathSegment(fl.Field().String())
}

// isSafePathSegment 检查 s 是否为一个安全的单层路径片段。
// 满足: 不包含路径分隔符(/ \)、不为 "." 或 "..",以此阻止路径穿透。
// 空串交给 required 等其他规则判断,此处视为安全。
func isSafePathSegment(s string) bool {
	if s == "" {
		return true
	}
	if strings.ContainsAny(s, `/\`) {
		return false
	}
	if s == "." || s == ".." {
		return false
	}
	return true
}

// IsSafePathSegment 导出供 handler 直接校验时使用。
func IsSafePathSegment(s string) bool {
	return isSafePathSegment(s)
}

// 1. 定义校验函数（白名单：只允许 字母数字 下划线 横杠 点）
func isPureFilename(filename string) bool {

	// 禁止包含路径分隔符
	if strings.ContainsAny(filename, "/\\") {
		return false
	}
	// 禁止包含相对路径跳转符（防御 .. 绕过）
	if strings.Contains(filename, "..") {
		return false
	}
	// 限定字符范围（防止特殊字符导致的意外）
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_.-]+$`, filename)
	return matched
}

func IsPureFilename(s string) bool {
	return isPureFilename(s)
}

func ShouldBind(ctx *gin.Context, logger *zap.Logger, v any, msg string) bool {
	if err := ctx.ShouldBind(v); err != nil {
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

func ShouldBindUri(ctx *gin.Context, logger *zap.Logger, v any, msg string) bool {
	if err := ctx.ShouldBindUri(v); err != nil {
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
	if err := ctx.ShouldBindQuery(v); err != nil {
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
