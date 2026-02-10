package errors

import (
	"emperror.dev/errors"
	"gorm.io/gorm"
)

// NewGormError 创建数据库错误
func NewGormError(err error, data map[string]any) *Error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrRecordNotFound.WithFields(data)
	}
	if errors.Is(err, gorm.ErrCheckConstraintViolated) {
		return ErrCheckConstraintViolated.WithFields(data)
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return ErrDuplicatedKey.WithFields(data)
	}
	if errors.Is(err, gorm.ErrRegistered) {
		return ErrRegistered.WithFields(data)
	}
	return FromError(err).WithFields(data)
}
