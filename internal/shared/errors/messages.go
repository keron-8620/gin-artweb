package errors

// 默认错误消息映射
var defaultErrorMessages = map[ErrorReason]string{
	// 通用错误
	ReasonUnknown:            "未知错误",
	ReasonNoContext:          "上下文为空",
	ReasonCanceled:           "请求取消",
	ReasonDeadlineExceeded:   "请求超时",
	ReasonValidationFailed:   "参数验证错误",
	ReasonRequestTimeout:     "请求超时",
	ReasonNetworkError:       "网络错误",
	ReasonServiceUnavailable: "服务不可用",

	// 安全认证
	ReasonHostHeaderInvalid: "Host头无效",
	ReasonRateLimitExceeded: "请求过于频繁，超出请求频率限制",
	ReasonNonceNotFound:     "请求头缺少随机数",
	ReasonReplayAttack:      "检测为重放攻击",
	ReasonTimestampNotFound: "请求头缺少时间戳",
	ReasonTimestampInvalid:  "无效的时间戳",
	ReasonTimestampExpired:  "时间戳已过期",

	// 身份权限认证
	ReasonUnauthorized:      "未授权操作",
	ReasonTokenExpired:      "登录已过期，请重新登录",
	ReasonTokenInvalid:      "无效的登录凭证",
	ReasonMissingAuth:       "缺少认证信息",
	ReasonTokenTypeMismatch: "令牌类型不匹配",
	ReasonAuthFailed:        "用户名或密码错误",
	ReasonForbidden:         "禁止访问",

	// 上传下载文件
	ReasonUploadFileNotFound:            "上传的文件未找到",
	ReasonUploadFileTooLarge:            "上传的文件超出大小限制",
	ReasonSaveUploadFileFailed:          "保存上传文件失败",
	ReasonSetUploadFilePermissionFailed: "设置上传文件权限失败",
	ReasonDownloadFileNotFound:          "下载的文件未找到",
	ReasonDownloadFilePermissionDenied:  "下载文件权限被拒绝",
	ReasonDownloadFileFailed:            "下载文件失败",

	// GORM 标准错误
	ReasonRecordNotFound:                "记录未找到",
	ReasonInvalidTransaction:            "事务处理错误",
	ReasonNotImplemented:                "功能未实现",
	ReasonMissingWhereClause:            "缺少where条件",
	ReasonUnsupportedRelation:           "关联关系不支持",
	ReasonPrimaryKeyRequired:            "主键未设置",
	ReasonModelValueRequired:            "模型值未设置",
	ReasonModelAccessibleFieldsRequired: "模型字段不可访问",
	ReasonSubQueryRequired:              "子查询未设置",
	ReasonInvalidData:                   "无效的数据",
	ReasonUnsupportedDriver:             "不支持的数据库驱动",
	ReasonRegistered:                    "模型已注册",
	ReasonInvalidField:                  "无效的字段",
	ReasonEmptySlice:                    "数组不能为空",
	ReasonDryRunModeUnsupported:         "不支持干运行模式",
	ReasonInvalidDB:                     "无效的数据库连接",
	ReasonInvalidValue:                  "无效的数据类型",
	ReasonInvalidValueOfLength:          "关联值无效, 长度不匹配",
	ReasonPreloadNotAllowed:             "使用计数时不允许预加载",
	ReasonDuplicatedKey:                 "唯一性约束冲突",
	ReasonForeignKeyViolated:            "外键约束冲突",
	ReasonCheckConstraintViolated:       "检查约束冲突",
	ReasonModelIsNil:                    "数据库模型不能为空",
}
