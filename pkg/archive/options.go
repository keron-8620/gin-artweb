package archive

import (
	"context"
)


// ArchiveOptions 压缩/解压选项配置
type ArchiveOptions struct {
	Context         context.Context // 上下文用于控制操作取消和超时
	MaxFileSize     int64           // 最大文件大小限制(字节)，0表示无限制
	MaxFiles        int             // 最大文件数量限制，0表示无限制
	ExcludePatterns []string        // 排除文件模式列表（原代码未使用，保留扩展）
	IncludeOnly     []string        // 包含文件模式列表（原代码未使用，保留扩展）
	FollowSymlinks  bool            // 是否跟随符号链接
	BufferSize      int             // 复制缓冲区大小(字节)
}

// DefaultArchiveOptions 默认压缩选项配置
var DefaultArchiveOptions = ArchiveOptions{
	Context:        context.Background(),
	MaxFileSize:    100 << 20, // 100MB
	MaxFiles:       10000,     // 10000个文件
	BufferSize:     64 * 1024, // 64KB
	FollowSymlinks: false,
}

// ArchiveOption 函数选项模式类型定义
type ArchiveOption func(*ArchiveOptions)

// WithContext 设置操作上下文
func WithContext(ctx context.Context) ArchiveOption {
	return func(opts *ArchiveOptions) {
		if ctx != nil {
			opts.Context = ctx
		}
	}
}

// WithMaxFileSize 设置最大文件大小限制（增加合法性校验）
func WithMaxFileSize(size int64) ArchiveOption {
	return func(opts *ArchiveOptions) {
		if size >= 0 {
			opts.MaxFileSize = size
		}
	}
}

// WithMaxFiles 设置最大文件数量限制（增加合法性校验）
func WithMaxFiles(count int) ArchiveOption {
	return func(opts *ArchiveOptions) {
		if count >= 0 {
			opts.MaxFiles = count
		}
	}
}

// WithBufferSize 设置复制缓冲区大小（增加合理范围校验）
func WithBufferSize(size int) ArchiveOption {
	return func(opts *ArchiveOptions) {
		if size <= 0 {
			size = 32 * 1024 // 默认32KB
		} else if size > 1<<20 { // 限制最大1MB，避免内存占用过高
			size = 1 << 20
		}
		opts.BufferSize = size
	}
}

// WithFollowSymlinks 设置是否跟随符号链接
func WithFollowSymlinks(follow bool) ArchiveOption {
	return func(opts *ArchiveOptions) {
		opts.FollowSymlinks = follow
	}
}

// applyOptions 应用选项配置并返回最终选项
func applyOptions(opts ...ArchiveOption) ArchiveOptions {
	options := DefaultArchiveOptions
	for _, opt := range opts {
		if opt != nil { // 防御性检查
			opt(&options)
		}
	}
	return options
}
