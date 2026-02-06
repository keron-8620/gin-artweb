package errors

type ErrorReason string

// 通用错误
const (
	ReasonUnknown            ErrorReason = "ERROR_UNKNOWN"               // 未知错误
	ReasonNoContext          ErrorReason = "ERROR_CTX_NO_CONTEXT"        // 上下文为空
	ReasonCanceled           ErrorReason = "ERROR_CTX_CANCELED"          // ctx取消
	ReasonDeadlineExceeded   ErrorReason = "ERROR_CTX_DEADLINE_EXCEEDED" // ctx超时
	ReasonValidationFailed   ErrorReason = "ERROR_VALIDATION_FAILED"     // 参数验证错误
	ReasonRequestTimeout     ErrorReason = "ERROR_REQUEST_TIMEOUT"       // 请求超时
	ReasonNetworkError       ErrorReason = "ERROR_NETWORK"               // 网络错误
	ReasonServiceUnavailable ErrorReason = "ERROR_SERVICE_UNAVAILABLE"   // 服务不可用
)

// 安全认证
const (
	ReasonHostHeaderInvalid ErrorReason = "SEC_HOST_HEADER_INVALID" // Host头无效
	ReasonRateLimitExceeded ErrorReason = "SEC_RATE_LIMIT_EXCEEDED" // 超出请求频率限制
	ReasonNonceNotFound     ErrorReason = "SEC_NONCE_NOT_FOUND"     // 请求头缺少随机数
	ReasonReplayAttack      ErrorReason = "SEC_REPLAY_ATTACK"       // 检测为重放攻击
	ReasonTimestampNotFound ErrorReason = "SEC_TIMESTAMP_NOT_FOUND" // 请求头缺少时间戳
	ReasonTimestampInvalid  ErrorReason = "SEC_TIMESTAMP_INVALID"   // 无效的时间戳
	ReasonTimestampExpired  ErrorReason = "SEC_TIMESTAMP_EXPIRED"   // 时间戳已过期
)

// 身份权限认证
const (
	ReasonUnauthorized      ErrorReason = "AUTH_UNAUTHORIZED"        // 未授权操作
	ReasonTokenExpired      ErrorReason = "AUTH_TOKEN_EXPIRED"       // 登录已过期，请重新登录
	ReasonTokenInvalid      ErrorReason = "AUTH_TOKEN_INVALID"       // 无效的登录凭证
	ReasonMissingAuth       ErrorReason = "AUTH_MISSING_AUTH"        // 缺少认证信息
	ReasonTokenTypeMismatch ErrorReason = "AUTH_TOKEN_TYPE_MISMATCH" // 令牌类型不匹配
	ReasonAuthFailed        ErrorReason = "AUTH_FAILED"              // 用户名或密码错误
	ReasonForbidden         ErrorReason = "AUTH_FORBIDDEN"           // 禁止访问
)

// 上传下载文件
const (
	ReasonUploadFileNotFound            ErrorReason = "UPLOAD_FILE_NOT_FOUND"             // 上传的文件未找到
	ReasonUploadFileTooLarge            ErrorReason = "UPLOAD_FILE_TOO_LARGE"             // 上传的文件超出大小限制
	ReasonSaveUploadFileFailed          ErrorReason = "UPLOAD_FILE_SAVE_FAILED"           // 保存上传文件失败
	ReasonSetUploadFilePermissionFailed ErrorReason = "UPLOAD_FILE_SET_PERMISSION_FAILED" // 设置上传文件权限失败
	ReasonDownloadFileNotFound          ErrorReason = "DOWNLOAD_FILE_NOT_FOUND"           // 下载的文件未找到
	ReasonDownloadFilePermissionDenied  ErrorReason = "DOWNLOAD_FILE_PERMISSION_DENIED"   // 下载文件权限被拒绝
	ReasonDownloadFileFailed            ErrorReason = "DOWNLOAD_FILE_FAILED"              // 下载文件失败
)

// 数据库操作
const (
	ReasonRecordNotFound                ErrorReason = "GORM_RECORD_NOT_FOUND"                 // 记录未找到
	ReasonInvalidTransaction            ErrorReason = "GORM_INVALID_TRANSACTION"              // 事务处理错误
	ReasonNotImplemented                ErrorReason = "GORM_NOT_IMPLEMENTED"                  // 功能未实现
	ReasonMissingWhereClause            ErrorReason = "GORM_MISSING_WHERE_CLAUSE"             // 缺少where条件
	ReasonUnsupportedRelation           ErrorReason = "GORM_UNSUPPORTED_RELATION"             // 关联关系不支持
	ReasonPrimaryKeyRequired            ErrorReason = "GORM_PRIMARY_KEY_REQUIRED"             // 主键未设置
	ReasonModelValueRequired            ErrorReason = "GORM_MODEL_VALUE_REQUIRED"             // 模型值未设置
	ReasonModelAccessibleFieldsRequired ErrorReason = "GORM_MODEL_ACCESSIBLE_FIELDS_REQUIRED" // 模型字段不可访问
	ReasonSubQueryRequired              ErrorReason = "GORM_SUB_QUERY_REQUIRED"               // 子查询未设置
	ReasonInvalidData                   ErrorReason = "GORM_INVALID_DATA"                     // 无效的数据
	ReasonUnsupportedDriver             ErrorReason = "GORM_UNSUPPORTED_DRIVER"               // 不支持的数据库驱动
	ReasonRegistered                    ErrorReason = "GORM_REGISTERED"                       // 模型已注册
	ReasonInvalidField                  ErrorReason = "GORM_INVALID_FIELD"                    // 无效的字段
	ReasonEmptySlice                    ErrorReason = "GORM_EMPTY_SLICE"                      // 数组不能为空
	ReasonDryRunModeUnsupported         ErrorReason = "GORM_DRY_RUN_MODE_UNSUPPORTED"         // 不支持干运行模式
	ReasonInvalidDB                     ErrorReason = "GORM_INVALID_DB"                       // 无效的数据库连接
	ReasonInvalidValue                  ErrorReason = "GORM_INVALID_VALUE"                    // 无效的数据类型
	ReasonInvalidValueOfLength          ErrorReason = "GORM_INVALID_VALUE_OF_LENGTH"          // 关联值无效, 长度不匹配
	ReasonPreloadNotAllowed             ErrorReason = "GORM_PRELOAD_NOT_ALLOWED"              // 使用计数时不允许预加载
	ReasonDuplicatedKey                 ErrorReason = "GORM_DUPLICATED_KEY"                   // 唯一性约束冲突
	ReasonForeignKeyViolated            ErrorReason = "GORM_FOREIGN_KEY_VIOLATED"             // 外键约束冲突
	ReasonCheckConstraintViolated       ErrorReason = "GORM_CHECK_CONSTRAINT_VIOLATED"        // 检查约束冲突
	ReasonModelIsNil                    ErrorReason = "GORM_MODEL_IS_NIL"                     // 数据库模型不能为空
)
