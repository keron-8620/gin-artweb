package serializer

// 错误常量定义
const (
	ErrFileNotFound      = "文件不存在"
	ErrFileEmpty         = "文件为空"
	ErrFileTooLarge      = "文件大小超过限制"
	ErrInvalidPath       = "无效的文件路径"
	ErrReadFailed        = "读取文件失败"
	ErrWriteFailed       = "写入文件失败"
	ErrSerializeFailed   = "序列化失败"
	ErrDeserializeFailed = "反序列化失败"
	ErrContextCancelled  = "操作被取消"
	ErrTimeout           = "操作超时"
	ErrPermissionDenied  = "权限不足"
)