package database

import (
	"net/http"

	"gorm.io/gorm"

	"gin-artweb/internal/shared/errors"
)

// 数据库相关错误码定义
var (
	ErrRecordNotFound                = errors.New(http.StatusNotFound, "GormRecordNotFound", "记录未找到", nil)
	ErrInvalidTransaction            = errors.New(http.StatusBadRequest, "GormInvalidTransaction", "事务处理错误", nil)
	ErrNotImplemented                = errors.New(http.StatusNotImplemented, "GormNotImplemented", "功能未实现", nil)
	ErrMissingWhereClause            = errors.New(http.StatusBadRequest, "GormMissingWhereClause", "缺少where条件", nil)
	ErrUnsupportedRelation           = errors.New(http.StatusBadRequest, "GormUnsupportedRelation", "关联关系不支持", nil)
	ErrPrimaryKeyRequired            = errors.New(http.StatusBadRequest, "GormPrimaryKeyRequired", "主键未设置", nil)
	ErrModelValueRequired            = errors.New(http.StatusBadRequest, "GormModelValueRequired", "模型值未设置", nil)
	ErrModelAccessibleFieldsRequired = errors.New(http.StatusBadRequest, "GormModelAccessibleFieldsRequired", "模型字段不可访问", nil)
	ErrSubQueryRequired              = errors.New(http.StatusBadRequest, "GormSubQueryRequired", "子查询未设置", nil)
	ErrInvalidData                   = errors.New(http.StatusBadRequest, "GormInvalidData", "无效的数据", nil)
	ErrUnsupportedDriver             = errors.New(http.StatusInternalServerError, "GormUnsupportedDriver", "不支持的数据库驱动", nil)
	ErrRegistered                    = errors.New(http.StatusBadRequest, "GormRegistered", "模型已注册", nil)
	ErrInvalidField                  = errors.New(http.StatusBadRequest, "GormInvalidField", "无效的字段", nil)
	ErrEmptySlice                    = errors.New(http.StatusBadRequest, "GormEmptySlice", "数组不能为空", nil)
	ErrDryRunModeUnsupported         = errors.New(http.StatusBadRequest, "GormDryRunModeUnsupported", "不支持干运行模式", nil)
	ErrInvalidDB                     = errors.New(http.StatusInternalServerError, "GormInvalidDB", "无效的数据库连接", nil)
	ErrInvalidValue                  = errors.New(http.StatusBadRequest, "GormInvalidValue", "无效的数据类型", nil)
	ErrInvalidValueOfLength          = errors.New(http.StatusBadRequest, "GormInvalidValueOfLength", "关联值无效, 长度不匹配", nil)
	ErrPreloadNotAllowed             = errors.New(http.StatusBadRequest, "GormPreloadNotAllowed", "使用计数时不允许预加载", nil)
	ErrDuplicatedKey                 = errors.New(http.StatusConflict, "GormDuplicatedKey", "唯一性约束冲突", nil)
	ErrForeignKeyViolated            = errors.New(http.StatusConflict, "GormForeignKeyViolated", "外键约束冲突", nil)
	ErrCheckConstraintViolated       = errors.New(http.StatusBadRequest, "GormCheckConstraintViolated", "检查约束冲突", nil)
	ErrModelIsNil                    = errors.New(http.StatusBadRequest, "GormModelIsNil", "数据库模型不能为空", nil)
)

var gormErrorsMap = map[string]*errors.Error{
	gorm.ErrRecordNotFound.Error():                ErrRecordNotFound,
	gorm.ErrInvalidTransaction.Error():            ErrInvalidTransaction,
	gorm.ErrNotImplemented.Error():                ErrNotImplemented,
	gorm.ErrMissingWhereClause.Error():            ErrMissingWhereClause,
	gorm.ErrUnsupportedRelation.Error():           ErrUnsupportedRelation,
	gorm.ErrPrimaryKeyRequired.Error():            ErrPrimaryKeyRequired,
	gorm.ErrModelValueRequired.Error():            ErrModelValueRequired,
	gorm.ErrModelAccessibleFieldsRequired.Error(): ErrModelAccessibleFieldsRequired,
	gorm.ErrSubQueryRequired.Error():              ErrSubQueryRequired,
	gorm.ErrInvalidData.Error():                   ErrInvalidData,
	gorm.ErrUnsupportedDriver.Error():             ErrUnsupportedDriver,
	gorm.ErrRegistered.Error():                    ErrRegistered,
	gorm.ErrInvalidField.Error():                  ErrInvalidField,
	gorm.ErrEmptySlice.Error():                    ErrEmptySlice,
	gorm.ErrDryRunModeUnsupported.Error():         ErrDryRunModeUnsupported,
	gorm.ErrInvalidDB.Error():                     ErrInvalidDB,
	gorm.ErrInvalidValue.Error():                  ErrInvalidValue,
	gorm.ErrInvalidValueOfLength.Error():          ErrInvalidValueOfLength,
	gorm.ErrPreloadNotAllowed.Error():             ErrPreloadNotAllowed,
	gorm.ErrDuplicatedKey.Error():                 ErrDuplicatedKey,
	gorm.ErrForeignKeyViolated.Error():            ErrForeignKeyViolated,
	gorm.ErrCheckConstraintViolated.Error():       ErrCheckConstraintViolated,
}

func NewGormError(err error, tmpData map[string]any) *errors.Error {
	if tmpData == nil {
		tmpData = make(map[string]any)
	}
	errMsg := err.Error()
	value, ok := gormErrorsMap[errMsg]
	if !ok {
		rErr := errors.FromError(err)
		return rErr.WithData(tmpData)
	}
	return value.WithData(tmpData)
}

func GormModelIsNil(model string) *errors.Error {
	return errors.New(
		http.StatusBadRequest,
		ErrModelIsNil.Reason,
		ErrModelIsNil.Msg,
		map[string]any{"model": model},
	)
}
