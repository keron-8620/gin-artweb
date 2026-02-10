package errors

type ErrorReason string

const (
	// 通用错误
	ReasonUnknown           ErrorReason = "ERROR_UNKNOWN"             // 未知错误
	ReasonValidationFailed  ErrorReason = "ERROR_VALIDATION_FAILED"   // 参数验证错误
	ReasonRequestTimeout    ErrorReason = "ERROR_REQUEST_TIMEOUT"     // 请求超时
	ReasonRateLimitExceeded ErrorReason = "ERROR_RATE_LIMIT_EXCEEDED" // 请求过于频繁

	// 上下文
	ReasonNoContext        ErrorReason = "ERROR_CTX_NO_CONTEXT"        // 上下文为空
	ReasonCanceled         ErrorReason = "ERROR_CTX_CANCELED"          // ctx取消
	ReasonDeadlineExceeded ErrorReason = "ERROR_CTX_DEADLINE_EXCEEDED" // ctx超时

	// 安全认证
	ReasonHostHeaderInvalid      ErrorReason = "SEC_HOST_HEADER_INVALID"      // Host头无效
	ReasonNonceNotFound          ErrorReason = "SEC_NONCE_NOT_FOUND"          // 请求头缺少随机数
	ReasonReplayAttack           ErrorReason = "SEC_REPLAY_ATTACK"            // 检测为重放攻击
	ReasonTimestampNotFound      ErrorReason = "SEC_TIMESTAMP_NOT_FOUND"      // 请求头缺少时间戳
	ReasonTimestampInvalid       ErrorReason = "SEC_TIMESTAMP_INVALID"        // 无效的时间戳
	ReasonTimestampExpired       ErrorReason = "SEC_TIMESTAMP_EXPIRED"        // 时间戳已过期
	ReasonPasswordStrengthFailed ErrorReason = "SEC_PASSWORD_STRENGTH_FAILED" // 密码强度不足

	// 身份权限认证
	ReasonUnauthorized      ErrorReason = "AUTH_UNAUTHORIZED"        // 未授权操作
	ReasonTokenExpired      ErrorReason = "AUTH_TOKEN_EXPIRED"       // 登录已过期，请重新登录
	ReasonTokenInvalid      ErrorReason = "AUTH_TOKEN_INVALID"       // 无效的登录凭证
	ReasonMissingAuth       ErrorReason = "AUTH_MISSING_AUTH"        // 缺少认证信息
	ReasonTokenTypeMismatch ErrorReason = "AUTH_TOKEN_TYPE_MISMATCH" // 令牌类型不匹配
	ReasonAuthFailed        ErrorReason = "AUTH_FAILED"              // 身份认证失败
	ReasonAccountLocked     ErrorReason = "AUTH_ACCOUNT_LOCKED"      // 账号已被锁定
	ReasonForbidden         ErrorReason = "AUTH_FORBIDDEN"           // 禁止访问

	// 数据库服务
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

	// ssh服务
	ReasonSSHConnectionFailed ErrorReason = "SSH_CONNECTION_FAILED"     // ssh连接失败
	ReasonSSHKeyDeployFailed  ErrorReason = "SSH_KEY_DEPLOYMENT_FAILED" // ssh密钥部署失败

	// 上传下载文件
	ReasonUploadFileNotFound            ErrorReason = "UPLOAD_FILE_NOT_FOUND"             // 上传的文件未找到
	ReasonUploadFileTooLarge            ErrorReason = "UPLOAD_FILE_TOO_LARGE"             // 上传的文件超出大小限制
	ReasonSaveUploadFileFailed          ErrorReason = "UPLOAD_FILE_SAVE_FAILED"           // 保存上传文件失败
	ReasonSetUploadFilePermissionFailed ErrorReason = "UPLOAD_FILE_SET_PERMISSION_FAILED" // 设置上传文件权限失败
	ReasonDownloadFileNotFound          ErrorReason = "DOWNLOAD_FILE_NOT_FOUND"           // 下载的文件未找到
	ReasonDownloadFilePermissionDenied  ErrorReason = "DOWNLOAD_FILE_PERMISSION_DENIED"   // 下载文件权限被拒绝
	ReasonDownloadFileFailed            ErrorReason = "DOWNLOAD_FILE_FAILED"              // 下载文件失败

	// 压缩解压文件
	ReasonUnZIPFailed       ErrorReason = "UNZIP_FAILED"          // 解压文件失败
	ReasonZIPFailed         ErrorReason = "ZIP_FAILED"            // 压缩文件失败
	ReasonZIPFileNotFound   ErrorReason = "ZIP_FILE_NOT_FOUND"    // 压缩文件未找到
	ReasonZIPFileIsEmpty    ErrorReason = "ZIP_FILE_IS_EMPTY"     // 压缩文件为空
	ReasonZIPFileIsNotValid ErrorReason = "ZIP_FILE_IS_NOT_VALID" // 压缩文件无效

	// 缓存文件相关
	ReasonExportCacheFileFailed ErrorReason = "EXPORT_CACHE_FILE_FAILED" // 导出缓存文件失败
	ReasonDeleteCacheFileFailed ErrorReason = "DELETE_CACHE_FILE_FAILED" // 删除缓存文件失败

	// 脚本相关
	ReasonScriptNotFound    ErrorReason = "SCRIPT_NOT_FOUND"     // 脚本未找到
	ReasonScriptIsBuiltin   ErrorReason = "SCRIPT_IS_BUILTIN"    // 脚本为内置脚本
	ReasonScriptIsDisabled  ErrorReason = "SCRIPT_IS_DISABLED"   // 脚本已禁用
	ReasonScriptLogNotFound ErrorReason = "SCRIPT_LOG_NOT_FOUND" // 脚本日志未找到
)
