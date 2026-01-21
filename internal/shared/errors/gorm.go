package errors

import (
	"net/http"

	"gorm.io/gorm"
)

// 数据库相关错误码定义
var (
	ErrRecordNotFound                = New(http.StatusNotFound, "GormRecordNotFound", "记录未找到", nil)
	ErrInvalidTransaction            = New(http.StatusBadRequest, "GormInvalidTransaction", "事务处理错误", nil)
	ErrNotImplemented                = New(http.StatusNotImplemented, "GormNotImplemented", "功能未实现", nil)
	ErrMissingWhereClause            = New(http.StatusBadRequest, "GormMissingWhereClause", "缺少where条件", nil)
	ErrUnsupportedRelation           = New(http.StatusBadRequest, "GormUnsupportedRelation", "关联关系不支持", nil)
	ErrPrimaryKeyRequired            = New(http.StatusBadRequest, "GormPrimaryKeyRequired", "主键未设置", nil)
	ErrModelValueRequired            = New(http.StatusBadRequest, "GormModelValueRequired", "模型值未设置", nil)
	ErrModelAccessibleFieldsRequired = New(http.StatusBadRequest, "GormModelAccessibleFieldsRequired", "模型字段不可访问", nil)
	ErrSubQueryRequired              = New(http.StatusBadRequest, "GormSubQueryRequired", "子查询未设置", nil)
	ErrInvalidData                   = New(http.StatusBadRequest, "GormInvalidData", "无效的数据", nil)
	ErrUnsupportedDriver             = New(http.StatusInternalServerError, "GormUnsupportedDriver", "不支持的数据库驱动", nil)
	ErrRegistered                    = New(http.StatusBadRequest, "GormRegistered", "模型已注册", nil)
	ErrInvalidField                  = New(http.StatusBadRequest, "GormInvalidField", "无效的字段", nil)
	ErrEmptySlice                    = New(http.StatusBadRequest, "GormEmptySlice", "数组不能为空", nil)
	ErrDryRunModeUnsupported         = New(http.StatusBadRequest, "GormDryRunModeUnsupported", "不支持干运行模式", nil)
	ErrInvalidDB                     = New(http.StatusInternalServerError, "GormInvalidDB", "无效的数据库连接", nil)
	ErrInvalidValue                  = New(http.StatusBadRequest, "GormInvalidValue", "无效的数据类型", nil)
	ErrInvalidValueOfLength          = New(http.StatusBadRequest, "GormInvalidValueOfLength", "关联值无效, 长度不匹配", nil)
	ErrPreloadNotAllowed             = New(http.StatusBadRequest, "GormPreloadNotAllowed", "使用计数时不允许预加载", nil)
	ErrDuplicatedKey                 = New(http.StatusConflict, "GormDuplicatedKey", "唯一性约束冲突", nil)
	ErrForeignKeyViolated            = New(http.StatusConflict, "GormForeignKeyViolated", "外键约束冲突", nil)
	ErrCheckConstraintViolated       = New(http.StatusBadRequest, "GormCheckConstraintViolated", "检查约束冲突", nil)
	ErrModelIsNil                    = New(http.StatusBadRequest, "GormModelIsNil", "数据库模型不能为空", nil)
)

var gormErrorsMap = map[string]*Error{
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

func NewGormError(err error, tmpData map[string]any) *Error {
	if tmpData == nil {
		tmpData = make(map[string]any)
	}
	errMsg := err.Error()
	value, ok := gormErrorsMap[errMsg]
	if !ok {
		rErr := FromError(err)
		return rErr.WithData(tmpData)
	}
	return value.WithData(tmpData)
}

func GormModelIsNil(model string) *Error {
	return New(
		http.StatusBadRequest,
		ErrModelIsNil.Reason,
		ErrModelIsNil.Msg,
		map[string]any{"model": model},
	)
}
