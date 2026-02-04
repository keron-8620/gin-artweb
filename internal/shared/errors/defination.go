package errors

func FromReason(reason ErrorReason) *Error {
	return New(reason, "", nil)
}

// 通用错误
var (
	ErrUnknown            = FromReason(ReasonUnknown)
	ErrCanceled           = FromReason(ReasonCanceled)
	ErrDeadlineExceeded   = FromReason(ReasonDeadlineExceeded)
	ErrValidationFailed   = FromReason(ReasonValidationFailed)
	ErrRequestTimeout     = FromReason(ReasonRequestTimeout)
	ErrNetworkError       = FromReason(ReasonNetworkError)
	ErrServiceUnavailable = FromReason(ReasonServiceUnavailable)
)

// 请求安全相关错误
var (
	ErrHostHeaderInvalid = FromReason(ReasonHostHeaderInvalid)
	ErrRateLimitExceeded = FromReason(ReasonRateLimitExceeded)
	ErrNonceNotFound     = FromReason(ReasonNonceNotFound)
	ErrReplayAttack      = FromReason(ReasonReplayAttack)
	ErrTimestampNotFound = FromReason(ReasonTimestampNotFound)
	ErrTimestampInvalid  = FromReason(ReasonTimestampInvalid)
	ErrTimestampExpired  = FromReason(ReasonTimestampExpired)
)

// 身份权限认证
var (
	ErrUnauthorized = FromReason(ReasonUnauthorized)
	ErrTokenExpired = FromReason(ReasonTokenExpired)
	ErrTokenInvalid = FromReason(ReasonTokenInvalid)
	ErrMissingAuth  = FromReason(ReasonMissingAuth)
	ErrAuthFailed   = FromReason(ReasonAuthFailed)
	ErrForbidden    = FromReason(ReasonForbidden)
)

// 上传下载文件
var (
	ErrUploadFileNotFound            = FromReason(ReasonUploadFileNotFound)
	ErrUploadFileTooLarge            = FromReason(ReasonUploadFileTooLarge)
	ErrSaveUploadFileFailed          = FromReason(ReasonSaveUploadFileFailed)
	ErrSetUploadFilePermissionFailed = FromReason(ReasonSetUploadFilePermissionFailed)
	ErrDownloadFileNotFound          = FromReason(ReasonDownloadFileNotFound)
	ErrDownloadFilePermissionDenied  = FromReason(ReasonDownloadFilePermissionDenied)
	ErrDownloadFileFailed            = FromReason(ReasonDownloadFileFailed)
)
