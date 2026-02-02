package serializer

import (
	"context"
	"os"
	"time"
)

// SerializerOptions 序列化选项
type SerializerOptions struct {
	// 上下文用于取消操作
	Context context.Context

	// 文件权限
	FileMode os.FileMode
	DirMode  os.FileMode

	// JSON缩进
	Indent uint8

	// 是否原子写入
	Atomic bool

	// 最大文件大小限制 (字节)，0表示无限制
	MaxFileSize int64

	// 超时时间
	Timeout time.Duration
}

// DefaultSerializerOptions 默认序列化选项
var DefaultSerializerOptions = SerializerOptions{
	Context:  context.Background(),
	FileMode: 0644,
	DirMode:  0755,
	Indent:   0,
	Atomic:   false,
	Timeout:  30 * time.Second,
}

// SerializerOption 函数选项模式
type SerializerOption func(*SerializerOptions)

// WithContext 设置上下文
func WithContext(ctx context.Context) SerializerOption {
	return func(opts *SerializerOptions) {
		if ctx != nil {
			opts.Context = ctx
		}
	}
}

// WithFileMode 设置文件权限
func WithFileMode(mode os.FileMode) SerializerOption {
	return func(opts *SerializerOptions) {
		if mode != 0 {
			opts.FileMode = mode
		}
	}
}

// WithDirMode 设置目录权限
func WithDirMode(mode os.FileMode) SerializerOption {
	return func(opts *SerializerOptions) {
		if mode != 0 {
			opts.DirMode = mode
		}
	}
}

// WithIndent 设置JSON缩进
func WithIndent(indent uint8) SerializerOption {
	return func(opts *SerializerOptions) {
		opts.Indent = indent
	}
}

// WithAtomic 设置原子写入
func WithAtomic(atomic bool) SerializerOption {
	return func(opts *SerializerOptions) {
		opts.Atomic = atomic
	}
}

// WithMaxFileSize 设置最大文件大小
func WithMaxFileSize(size int64) SerializerOption {
	return func(opts *SerializerOptions) {
		if size >= 0 {
			opts.MaxFileSize = size
		}
	}
}

// WithTimeout 设置超时时间
func WithTimeout(timeout time.Duration) SerializerOption {
	return func(opts *SerializerOptions) {
		if timeout > 0 {
			opts.Timeout = timeout
		}
	}
}

// applyOptions 应用选项
func applyOptions(opts ...SerializerOption) SerializerOptions {
	options := DefaultSerializerOptions
	for _, opt := range opts {
		opt(&options)
	}

	return options
}

// ReadResult 读取结果.
// ReadResult 读取操作结果
// 包含文件路径、大小、耗时和错误信息
// 适用于监控和日志记录
// 示例: {"file_path":"/path/to/file.json","size":1024,"duration":"10ms","success":true}
type ReadResult struct {
	FilePath string        `json:"file_path"`       // 文件路径
	Size     int64         `json:"size"`            // 文件大小(字节)
	Duration time.Duration `json:"duration"`        // 操作耗时
	Success  bool          `json:"success"`         // 是否成功
	Error    string        `json:"error,omitempty"` // 错误信息(如果有)
}

// WriteResult 写入操作结果
// 包含文件路径、大小、耗时和错误信息
// 适用于监控和日志记录
// 示例: {"file_path":"/path/to/file.json","size":2048,"duration":"20ms","success":true}
type WriteResult struct {
	FilePath string        `json:"file_path"`       // 文件路径
	Size     int64         `json:"size"`            // 文件大小(字节)
	Duration time.Duration `json:"duration"`        // 操作耗时
	Success  bool          `json:"success"`         // 是否成功
	Error    string        `json:"error,omitempty"` // 错误信息(如果有)
}

// SerializeResult 序列化操作结果
// 包含序列化数据大小和耗时信息
// 适用于内存序列化操作
// 示例: {"size":512,"duration":"5ms","success":true}
type SerializeResult struct {
	Size     int64         `json:"size"`            // 序列化数据大小(字节)
	Duration time.Duration `json:"duration"`        // 操作耗时
	Success  bool          `json:"success"`         // 是否成功
	Error    string        `json:"error,omitempty"` // 错误信息(如果有)
}
