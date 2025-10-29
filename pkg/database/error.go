package database

import (
	"net/http"

	"gorm.io/gorm"

	"gitee.com/keion8620/go-dango-gin/pkg/errors"
)

// 数据库相关错误码定义
var (
	RsonRecordNotFound = errors.ReasonEnum{
		Reason: "GormRecordNotFound",
		Msg:    "记录未找到",
	}

	RsonInvalidTransaction = errors.ReasonEnum{
		Reason: "GormInvalidTransaction",
		Msg:    "事务处理错误",
	}

	RsonNotImplemented = errors.ReasonEnum{
		Reason: "GormNotImplemented",
		Msg:    "功能未实现",
	}

	RsonMissingWhereClause = errors.ReasonEnum{
		Reason: "GormMissingWhereClause",
		Msg:    "缺少where条件",
	}

	RsonUnsupportedRelation = errors.ReasonEnum{
		Reason: "GormUnsupportedRelation",
		Msg:    "关联关系不支持",
	}

	RsonPrimaryKeyRequired = errors.ReasonEnum{
		Reason: "GormPrimaryKeyRequired",
		Msg:    "主键未设置",
	}

	RsonModelValueRequired = errors.ReasonEnum{
		Reason: "GormModelValueRequired",
		Msg:    "模型值未设置",
	}

	RsonModelAccessibleFieldsRequired = errors.ReasonEnum{
		Reason: "GormModelAccessibleFieldsRequired",
		Msg:    "模型字段不可访问",
	}

	RsonSubQueryRequired = errors.ReasonEnum{
		Reason: "GormSubQueryRequired",
		Msg:    "子查询未设置",
	}

	RsonInvalidData = errors.ReasonEnum{
		Reason: "GormInvalidData",
		Msg:    "无效的数据",
	}

	RsonUnsupportedDriver = errors.ReasonEnum{
		Reason: "GormUnsupportedDriver",
		Msg:    "不支持的数据库驱动",
	}

	RsonRegistered = errors.ReasonEnum{
		Reason: "GormRegistered",
		Msg:    "模型已注册",
	}

	RsonInvalidField = errors.ReasonEnum{
		Reason: "GormInvalidField",
		Msg:    "无效的字段",
	}

	RsonEmptySlice = errors.ReasonEnum{
		Reason: "GormEmptySlice",
		Msg:    "数组不能为空",
	}

	RsonDryRunModeUnsupported = errors.ReasonEnum{
		Reason: "GormDryRunModeUnsupported",
		Msg:    "不支持干运行模式",
	}

	RsonInvalidDB = errors.ReasonEnum{
		Reason: "GormInvalidDB",
		Msg:    "无效的数据库连接",
	}

	RsonInvalidValue = errors.ReasonEnum{
		Reason: "GormInvalidValue",
		Msg:    "无效的数据类型",
	}

	RsonInvalidValueOfLength = errors.ReasonEnum{
		Reason: "GormInvalidValueOfLength",
		Msg:    "关联值无效, 长度不匹配",
	}

	RsonPreloadNotAllowed = errors.ReasonEnum{
		Reason: "GormPreloadNotAllowed",
		Msg:    "使用计数时不允许预加载",
	}

	RsonDuplicatedKey = errors.ReasonEnum{
		Reason: "GormDuplicatedKey",
		Msg:    "唯一性约束冲突",
	}

	RsonForeignKeyViolated = errors.ReasonEnum{
		Reason: "GormForeignKeyViolated",
		Msg:    "外键约束冲突",
	}

	RsonCheckConstraintViolated = errors.ReasonEnum{
		Reason: "GormCheckConstraintViolated",
		Msg:    "检查约束冲突",
	}
	RsonModelIsNil = errors.ReasonEnum{
		Reason: "GormModelIsNil",
		Msg:    "数据库模型不能为空",
	}
)

var gormErrorsMap = map[string]errors.ReasonEnum{
	gorm.ErrRecordNotFound.Error():                RsonRecordNotFound,
	gorm.ErrInvalidTransaction.Error():            RsonInvalidTransaction,
	gorm.ErrNotImplemented.Error():                RsonNotImplemented,
	gorm.ErrMissingWhereClause.Error():            RsonMissingWhereClause,
	gorm.ErrUnsupportedRelation.Error():           RsonUnsupportedRelation,
	gorm.ErrPrimaryKeyRequired.Error():            RsonPrimaryKeyRequired,
	gorm.ErrModelValueRequired.Error():            RsonModelValueRequired,
	gorm.ErrModelAccessibleFieldsRequired.Error(): RsonModelAccessibleFieldsRequired,
	gorm.ErrSubQueryRequired.Error():              RsonSubQueryRequired,
	gorm.ErrInvalidData.Error():                   RsonInvalidData,
	gorm.ErrUnsupportedDriver.Error():             RsonUnsupportedDriver,
	gorm.ErrRegistered.Error():                    RsonRegistered,
	gorm.ErrInvalidField.Error():                  RsonInvalidField,
	gorm.ErrEmptySlice.Error():                    RsonEmptySlice,
	gorm.ErrDryRunModeUnsupported.Error():         RsonDryRunModeUnsupported,
	gorm.ErrInvalidDB.Error():                     RsonInvalidDB,
	gorm.ErrInvalidValue.Error():                  RsonInvalidValue,
	gorm.ErrInvalidValueOfLength.Error():          RsonInvalidValueOfLength,
	gorm.ErrPreloadNotAllowed.Error():             RsonPreloadNotAllowed,
	gorm.ErrDuplicatedKey.Error():                 RsonDuplicatedKey,
	gorm.ErrForeignKeyViolated.Error():            RsonForeignKeyViolated,
	gorm.ErrCheckConstraintViolated.Error():       RsonCheckConstraintViolated,
}

func NewGormError(err error, tmpData map[string]any) *errors.Error {
	errMsg := err.Error()
	value, ok := gormErrorsMap[errMsg]
	if tmpData == nil {
		tmpData = make(map[string]any)
	}
	if !ok {
		eErr := errors.New(
			http.StatusInternalServerError,
			errors.RsonUnknown.Reason,
			errors.RsonUnknown.Msg,
			tmpData,
		)
		return eErr.WithCause(err)
	}
	rErr := errors.New(http.StatusBadRequest, value.Reason, value.Msg, tmpData)
	return rErr.WithCause(err)
}

func GormModelIsNil(model string) *errors.Error {
	return errors.New(
		http.StatusBadRequest,
		RsonModelIsNil.Reason,
		RsonModelIsNil.Msg,
		map[string]any{"model": model},
	)
}
