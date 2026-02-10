package errors

var (
	// 通用错误
	ErrUnknown           = FromReason(ReasonUnknown)           // 未知错误
	ErrValidationFailed  = FromReason(ReasonValidationFailed)  // 验证失败
	ErrRequestTimeout    = FromReason(ReasonRequestTimeout)    // 请求超时
	ErrRateLimitExceeded = FromReason(ReasonRateLimitExceeded) // 请求过于频繁

	// 上下文
	ErrNoContext        = FromReason(ReasonNoContext)        // 上下文为空
	ErrCanceled         = FromReason(ReasonCanceled)         // 操作已取消
	ErrDeadlineExceeded = FromReason(ReasonDeadlineExceeded) // 操作超时

	// 安全认证
	ErrHostHeaderInvalid      = FromReason(ReasonHostHeaderInvalid)      // Host头无效
	ErrNonceNotFound          = FromReason(ReasonNonceNotFound)          // Nonce不存在
	ErrReplayAttack           = FromReason(ReasonReplayAttack)           // 重复请求攻击
	ErrTimestampNotFound      = FromReason(ReasonTimestampNotFound)      // 时间戳不存在
	ErrTimestampInvalid       = FromReason(ReasonTimestampInvalid)       // 时间戳无效
	ErrTimestampExpired       = FromReason(ReasonTimestampExpired)       // 时间戳过期
	ErrPasswordStrengthFailed = FromReason(ReasonPasswordStrengthFailed) // 密码强度不足

	// 身份权限认证
	ErrUnauthorized      = FromReason(ReasonUnauthorized)      // 未授权
	ErrTokenExpired      = FromReason(ReasonTokenExpired)      // 令牌过期
	ErrTokenInvalid      = FromReason(ReasonTokenInvalid)      // 令牌无效
	ErrMissingAuth       = FromReason(ReasonMissingAuth)       // 缺少认证信息
	ErrTokenTypeMismatch = FromReason(ReasonTokenTypeMismatch) // 令牌类型不匹配
	ErrAuthFailed        = FromReason(ReasonAuthFailed)        // 认证失败
	ErrAccountLocked     = FromReason(ReasonAccountLocked)     // 账号已被锁定
	ErrForbidden         = FromReason(ReasonForbidden)         // 禁止访问

	// 数据库
	ErrRecordNotFound                = FromReason(ReasonRecordNotFound)                // 记录不存在
	ErrUnsupportedRelation           = FromReason(ReasonUnsupportedRelation)           // 关联关系不支持
	ErrPrimaryKeyRequired            = FromReason(ReasonPrimaryKeyRequired)            // 主键未设置
	ErrModelValueRequired            = FromReason(ReasonModelValueRequired)            // 模型值未设置
	ErrModelAccessibleFieldsRequired = FromReason(ReasonModelAccessibleFieldsRequired) // 模型字段不可访问
	ErrSubQueryRequired              = FromReason(ReasonSubQueryRequired)              // 子查询未设置
	ErrInvalidData                   = FromReason(ReasonInvalidData)                   // 无效的数据
	ErrUnsupportedDriver             = FromReason(ReasonUnsupportedDriver)             // 不支持的数据库驱动
	ErrRegistered                    = FromReason(ReasonRegistered)                    // 模型已注册
	ErrInvalidField                  = FromReason(ReasonInvalidField)                  // 无效的字段
	ErrEmptySlice                    = FromReason(ReasonEmptySlice)                    // 数组不能为空
	ErrDryRunModeUnsupported         = FromReason(ReasonDryRunModeUnsupported)         // 不支持干运行模式
	ErrInvalidDB                     = FromReason(ReasonInvalidDB)                     // 无效的数据库连接
	ErrInvalidValue                  = FromReason(ReasonInvalidValue)                  // 无效的数据类型
	ErrInvalidValueOfLength          = FromReason(ReasonInvalidValueOfLength)          // 关联值无效, 长度不匹配
	ErrPreloadNotAllowed             = FromReason(ReasonPreloadNotAllowed)             // 使用计数时不允许预加载
	ErrDuplicatedKey                 = FromReason(ReasonDuplicatedKey)                 // 唯一性约束冲突
	ErrForeignKeyViolated            = FromReason(ReasonForeignKeyViolated)            // 外键约束冲突
	ErrCheckConstraintViolated       = FromReason(ReasonCheckConstraintViolated)       // 检查约束冲突

	// ssh服务
	ErrSSHConnectionFailed = FromReason(ReasonSSHConnectionFailed) // ssh连接失败
	ErrSSHKeyDeployFailed  = FromReason(ReasonSSHKeyDeployFailed)  // ssh密钥部署失败

	// 上传下载文件
	ErrUploadFileNotFound            = FromReason(ReasonUploadFileNotFound)            // 上传文件不存在
	ErrUploadFileTooLarge            = FromReason(ReasonUploadFileTooLarge)            // 上传文件过大
	ErrSaveUploadFileFailed          = FromReason(ReasonSaveUploadFileFailed)          // 保存上传文件失败
	ErrSetUploadFilePermissionFailed = FromReason(ReasonSetUploadFilePermissionFailed) // 设置上传文件权限失败
	ErrDownloadFileNotFound          = FromReason(ReasonDownloadFileNotFound)          // 下载文件不存在
	ErrDownloadFilePermissionDenied  = FromReason(ReasonDownloadFilePermissionDenied)  // 下载文件权限被拒绝
	ErrDownloadFileFailed            = FromReason(ReasonDownloadFileFailed)            // 下载文件失败

	// 压缩解压文件
	ErrUnZIPFailed       = FromReason(ReasonUnZIPFailed)       // 解压文件失败
	ErrZIPFailed         = FromReason(ReasonZIPFailed)         // 压缩文件失败
	ErrZIPFileNotFound   = FromReason(ReasonZIPFileNotFound)   // 压缩文件未找到
	ErrZIPFileIsEmpty    = FromReason(ReasonZIPFileIsEmpty)    // 压缩文件为空
	ErrZIPFileIsNotValid = FromReason(ReasonZIPFileIsNotValid) // 压缩文件无效

	// 缓存文件
	ErrExportCacheFileFailed = FromReason(ReasonExportCacheFileFailed) // 缓存文件导出失败
	ErrDeleteCacheFileFailed = FromReason(ReasonDeleteCacheFileFailed) // 缓存文件删除失败

	// 脚本相关
	ErrScriptNotFound    = FromReason(ReasonScriptNotFound)    // 脚本不存在
	ErrScriptIsBuiltin   = FromReason(ReasonScriptIsBuiltin)   // 脚本为内置脚本
	ErrScriptIsDisabled  = FromReason(ReasonScriptIsDisabled)  // 脚本已禁用
	ErrScriptLogNotFound = FromReason(ReasonScriptLogNotFound) // 脚本日志不存在
)
