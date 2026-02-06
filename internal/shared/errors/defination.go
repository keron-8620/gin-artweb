package errors

func FromReason(reason ErrorReason) *Error {
	return New(reason, "", nil)
}

// 通用错误
var (
	ErrUnknown            = FromReason(ReasonUnknown)            // 未知错误
	ErrNoContext          = FromReason(ReasonNoContext)          // 上下文为空
	ErrCanceled           = FromReason(ReasonCanceled)           // 操作已取消
	ErrDeadlineExceeded   = FromReason(ReasonDeadlineExceeded)   // 操作超时
	ErrValidationFailed   = FromReason(ReasonValidationFailed)   // 验证失败
	ErrRequestTimeout     = FromReason(ReasonRequestTimeout)     // 请求超时
	ErrNetworkError       = FromReason(ReasonNetworkError)       // 网络错误
	ErrServiceUnavailable = FromReason(ReasonServiceUnavailable) // 服务不可用
)

// 安全认证
var (
	ErrHostHeaderInvalid = FromReason(ReasonHostHeaderInvalid) // Host头无效
	ErrRateLimitExceeded = FromReason(ReasonRateLimitExceeded) // 超出请求频率限制
	ErrNonceNotFound     = FromReason(ReasonNonceNotFound)     // Nonce不存在
	ErrReplayAttack      = FromReason(ReasonReplayAttack)      // 重复请求攻击
	ErrTimestampNotFound = FromReason(ReasonTimestampNotFound) // 时间戳不存在
	ErrTimestampInvalid  = FromReason(ReasonTimestampInvalid)  // 时间戳无效
	ErrTimestampExpired  = FromReason(ReasonTimestampExpired)  // 时间戳过期
)

// 身份权限认证
var (
	ErrUnauthorized      = FromReason(ReasonUnauthorized)      // 未授权
	ErrTokenExpired      = FromReason(ReasonTokenExpired)      // 令牌过期
	ErrTokenInvalid      = FromReason(ReasonTokenInvalid)      // 令牌无效
	ErrMissingAuth       = FromReason(ReasonMissingAuth)       // 缺少认证信息
	ErrTokenTypeMismatch = FromReason(ReasonTokenTypeMismatch) // 令牌类型不匹配
	ErrAuthFailed        = FromReason(ReasonAuthFailed)        // 认证失败
	ErrForbidden         = FromReason(ReasonForbidden)         // 禁止访问
)

// 上传下载文件
var (
	ErrUploadFileNotFound            = FromReason(ReasonUploadFileNotFound)            // 上传文件不存在
	ErrUploadFileTooLarge            = FromReason(ReasonUploadFileTooLarge)            // 上传文件过大
	ErrSaveUploadFileFailed          = FromReason(ReasonSaveUploadFileFailed)          // 保存上传文件失败
	ErrSetUploadFilePermissionFailed = FromReason(ReasonSetUploadFilePermissionFailed) // 设置上传文件权限失败
	ErrDownloadFileNotFound          = FromReason(ReasonDownloadFileNotFound)          // 下载文件不存在
	ErrDownloadFilePermissionDenied  = FromReason(ReasonDownloadFilePermissionDenied)  // 下载文件权限被拒绝
	ErrDownloadFileFailed            = FromReason(ReasonDownloadFileFailed)            // 下载文件失败
)
