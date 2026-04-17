package errors

import "net/http"

// #nosec G101 -- reasonToStatus 错误原因到HTTP状态码的映射
var reasonToStatus = map[ErrorReason]int{
	// 通用错误
	ReasonUnknown:           http.StatusInternalServerError,
	ReasonValidationFailed:  http.StatusBadRequest,
	ReasonRequestTimeout:    http.StatusRequestTimeout,
	ReasonRateLimitExceeded: http.StatusTooManyRequests,

	// #nosec G101 -- 上下文相关
	ReasonNoContext:        http.StatusBadRequest,
	ReasonCanceled:         http.StatusRequestTimeout,
	ReasonDeadlineExceeded: http.StatusRequestTimeout,

	// #nosec G101 -- 安全认证
	ReasonHostHeaderInvalid:      http.StatusBadRequest,
	ReasonNonceNotFound:          http.StatusBadRequest,
	ReasonReplayAttack:           http.StatusBadRequest,
	ReasonTimestampNotFound:      http.StatusBadRequest,
	ReasonTimestampInvalid:       http.StatusBadRequest,
	ReasonTimestampExpired:       http.StatusBadRequest,
	ReasonPasswordStrengthFailed: http.StatusBadRequest,

	// #nosec G101 -- 身份权限认证
	ReasonUnauthorized:      http.StatusUnauthorized,
	ReasonTokenExpired:      http.StatusUnauthorized,
	ReasonTokenInvalid:      http.StatusUnauthorized,
	ReasonMissingAuth:       http.StatusUnauthorized,
	ReasonTokenTypeMismatch: http.StatusUnauthorized,
	ReasonAuthFailed:        http.StatusUnauthorized,
	ReasonAccountLocked:     http.StatusForbidden,
	ReasonForbidden:         http.StatusForbidden,

	// #nosec G101 -- 数据库操作
	ReasonRecordNotFound:                http.StatusNotFound,
	ReasonInvalidTransaction:            http.StatusInternalServerError,
	ReasonNotImplemented:                http.StatusNotImplemented,
	ReasonMissingWhereClause:            http.StatusBadRequest,
	ReasonUnsupportedRelation:           http.StatusInternalServerError,
	ReasonPrimaryKeyRequired:            http.StatusBadRequest,
	ReasonModelValueRequired:            http.StatusBadRequest,
	ReasonModelAccessibleFieldsRequired: http.StatusInternalServerError,
	ReasonSubQueryRequired:              http.StatusBadRequest,
	ReasonInvalidData:                   http.StatusBadRequest,
	ReasonUnsupportedDriver:             http.StatusInternalServerError,
	ReasonRegistered:                    http.StatusInternalServerError,
	ReasonInvalidField:                  http.StatusBadRequest,
	ReasonEmptySlice:                    http.StatusBadRequest,
	ReasonDryRunModeUnsupported:         http.StatusInternalServerError,
	ReasonInvalidDB:                     http.StatusInternalServerError,
	ReasonInvalidValue:                  http.StatusBadRequest,
	ReasonInvalidValueOfLength:          http.StatusBadRequest,
	ReasonPreloadNotAllowed:             http.StatusBadRequest,
	ReasonDuplicatedKey:                 http.StatusConflict,
	ReasonForeignKeyViolated:            http.StatusConflict,
	ReasonCheckConstraintViolated:       http.StatusBadRequest,

	// #nosec G101 -- ssh链接
	ReasonSSHConnectionFailed: http.StatusBadRequest,
	ReasonSSHKeyDeployFailed:  http.StatusBadRequest,

	// #nosec G101 -- 上传下载文件
	ReasonDownloadFileNotFound:         http.StatusNotFound,
	ReasonDownloadFilePermissionDenied: http.StatusForbidden,
	ReasonDownloadFileFailed:           http.StatusInternalServerError,

	// #nosec G101 -- 压缩解压文件
	ReasonUnZIPFailed:       http.StatusInternalServerError,
	ReasonZIPFailed:         http.StatusInternalServerError,
	ReasonZIPFileNotFound:   http.StatusNotFound,
	ReasonZIPFileIsEmpty:    http.StatusBadRequest,
	ReasonZIPFileIsNotValid: http.StatusBadRequest,

	// #nosec G101 -- 缓存文件
	ReasonExportCacheFileFailed: http.StatusInternalServerError,
	ReasonDeleteCacheFileFailed: http.StatusInternalServerError,

	// #nosec G101 -- 脚本相关
	ReasonScriptNotFound:     http.StatusNotFound,
	ReasonScriptIsBuiltin:    http.StatusBadRequest,
	ReasonScriptIsDisabled:   http.StatusBadRequest,
	ReasonScriptLogNotFound:  http.StatusNotFound,
	ReasonScriptSaveFailed:   http.StatusInternalServerError,
	ReasonScriptRemoveFailed: http.StatusInternalServerError,

	// #nosec G101 -- 程序包相关
	ReasonPackageNotFound:     http.StatusNotFound,
	ReasonPackageSaveFailed:   http.StatusInternalServerError,
	ReasonPackageRemoveFailed: http.StatusInternalServerError,

	// #nosec G101 -- 计划任务相关
	ReasonCronSpecificationInvalid: http.StatusBadRequest,
}
