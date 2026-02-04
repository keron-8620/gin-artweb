package archive

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

// ArchiveOptions 压缩/解压选项配置
type ArchiveOptions struct {
	Context          context.Context // 上下文用于控制操作取消和超时
	MaxFileSize      int64           // 最大文件大小限制(字节)，0表示无限制
	MaxFiles         int             // 最大文件数量限制，0表示无限制
	ExcludePatterns  []string        // 排除文件模式列表
	IncludeOnly      []string        // 包含文件模式列表
	FollowSymlinks   bool            // 是否跟随符号链接
	BufferSize       int             // 复制缓冲区大小(字节)
	CompressionLevel int             // 压缩级别(0-9)，0表示无压缩
	PermissionsMask  int             // 权限掩码，用于控制解压时的权限
	Concurrency      int             // 并发处理数量，0表示不使用并发
}

// DefaultArchiveOptions 默认压缩选项配置
var DefaultArchiveOptions = ArchiveOptions{
	Context:          context.Background(),
	MaxFileSize:      100 << 20, // 100MB
	MaxFiles:         10000,     // 10000个文件
	BufferSize:       64 * 1024, // 64KB
	FollowSymlinks:   false,
	CompressionLevel: 6,    // 默认压缩级别
	PermissionsMask:  0755, // 默认权限掩码
	Concurrency:      0,    // 默认不使用并发
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
		} else if size > 1<<20 {
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

// WithExcludePatterns 设置排除文件模式列表
func WithExcludePatterns(patterns ...string) ArchiveOption {
	return func(opts *ArchiveOptions) {
		opts.ExcludePatterns = append(opts.ExcludePatterns, patterns...)
	}
}

// WithIncludeOnly 设置包含文件模式列表
func WithIncludeOnly(patterns ...string) ArchiveOption {
	return func(opts *ArchiveOptions) {
		opts.IncludeOnly = append(opts.IncludeOnly, patterns...)
	}
}

// WithCompressionLevel 设置压缩级别
func WithCompressionLevel(level int) ArchiveOption {
	return func(opts *ArchiveOptions) {
		if level >= 0 && level <= 9 {
			opts.CompressionLevel = level
		}
	}
}

// WithPermissionsMask 设置权限掩码
func WithPermissionsMask(mask int) ArchiveOption {
	return func(opts *ArchiveOptions) {
		opts.PermissionsMask = mask
	}
}

// WithConcurrency 设置并发处理数量
func WithConcurrency(concurrency int) ArchiveOption {
	return func(opts *ArchiveOptions) {
		if concurrency >= 0 {
			opts.Concurrency = concurrency
		}
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

// ShouldExclude 检查文件是否应该被排除
func (o *ArchiveOptions) ShouldExclude(filePath string) (bool, error) {
	if len(o.ExcludePatterns) == 0 {
		return false, nil
	}

	baseName := filepath.Base(filePath)
	for _, pattern := range o.ExcludePatterns {
		match, err := filepath.Match(pattern, baseName)
		if err != nil {
			return false, errors.Wrapf(err, "无效的排除模式: %s", pattern)
		}
		if match {
			return true, nil
		}
	}

	return false, nil
}

// ShouldInclude 检查文件是否应该被包含
func (o *ArchiveOptions) ShouldInclude(filePath string) (bool, error) {
	if len(o.IncludeOnly) == 0 {
		return true, nil
	}

	baseName := filepath.Base(filePath)
	for _, pattern := range o.IncludeOnly {
		match, err := filepath.Match(pattern, baseName)
		if err != nil {
			return false, errors.Wrapf(err, "无效的包含模式: %s", pattern)
		}
		if match {
			return true, nil
		}
	}

	return false, nil
}

// IsPathAllowed 检查路径是否在允许范围内
func (o *ArchiveOptions) IsPathAllowed(filePath string, baseDir string) (bool, error) {
	relPath, err := filepath.Rel(baseDir, filePath)
	if err != nil {
		return false, errors.Wrap(err, "计算相对路径失败")
	}

	// 防止路径遍历攻击
	if strings.HasPrefix(relPath, "..") {
		return false, errors.New("检测到路径遍历攻击")
	}

	return true, nil
}
