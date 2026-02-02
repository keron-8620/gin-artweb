package archive


// 错误常量定义
const (
	ErrEmptyPath          = "源路径/目标路径不能为空"
	ErrContextCancelled   = "操作被取消"
	ErrFileTooLarge       = "文件大小超过限制"
	ErrTooManyFiles       = "文件数量超过限制"
	ErrInvalidBufferSize  = "缓冲区大小无效"
	ErrPathTraversal      = "路径遍历攻击检测"
	ErrMultipleTopEntries = "压缩文件包含多个顶层条目"
	ErrGetAbsPathFailed   = "获取绝对路径失败"
)
