package archive

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"

	"emperror.dev/errors"
)

// ArchiveOptions 压缩/解压选项配置
type ArchiveOptions struct {
	// Context 上下文用于控制操作取消和超时
	Context context.Context

	// MaxFileSize 最大文件大小限制(字节)，0表示无限制
	MaxFileSize int64

	// MaxFiles 最大文件数量限制，0表示无限制
	MaxFiles int

	// ExcludePatterns 排除文件模式列表
	ExcludePatterns []string

	// IncludeOnly 包含文件模式列表
	IncludeOnly []string

	// FollowSymlinks 是否跟随符号链接
	FollowSymlinks bool

	// BufferSize 复制缓冲区大小(字节)
	BufferSize int
}

// DefaultArchiveOptions 默认压缩选项配置
var DefaultArchiveOptions = ArchiveOptions{
	Context:        context.Background(),
	MaxFileSize:    100 << 20, // 100MB
	MaxFiles:       10000,     // 10000个文件
	BufferSize:     32 * 1024, // 32KB
	FollowSymlinks: false,
}

// ArchiveOption 函数选项模式类型定义
type ArchiveOption func(*ArchiveOptions)

// WithContext 设置操作上下文
func WithContext(ctx context.Context) ArchiveOption {
	return func(opts *ArchiveOptions) {
		opts.Context = ctx
	}
}

// WithMaxFileSize 设置最大文件大小限制
func WithMaxFileSize(size int64) ArchiveOption {
	return func(opts *ArchiveOptions) {
		opts.MaxFileSize = size
	}
}

// WithMaxFiles 设置最大文件数量限制
func WithMaxFiles(count int) ArchiveOption {
	return func(opts *ArchiveOptions) {
		opts.MaxFiles = count
	}
}

// WithBufferSize 设置复制缓冲区大小
func WithBufferSize(size int) ArchiveOption {
	return func(opts *ArchiveOptions) {
		opts.BufferSize = size
	}
}

// applyOptions 应用选项配置并返回最终选项
func applyOptions(opts ...ArchiveOption) ArchiveOptions {
	options := DefaultArchiveOptions
	for _, opt := range opts {
		opt(&options)
	}
	return options
}

// checkContext 检查上下文状态，如果上下文已取消则返回相应错误
func checkContext(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

// safeCopy 安全复制数据，支持大小限制、上下文检查和进度监控
func safeCopy(ctx context.Context, dst io.Writer, src io.Reader, maxSize int64, bufferSize int) (int64, error) {
	var written int64
	buf := make([]byte, bufferSize)

	for {
		// 检查上下文是否已取消
		if err := checkContext(ctx); err != nil {
			return written, errors.WrapIf(err, "上下文检查失败")
		}

		n, err := src.Read(buf)
		if n > 0 {
			// 检查文件大小是否超出限制
			if maxSize > 0 && written+int64(n) > maxSize {
				return written, errors.NewWithDetails(
					"文件复制失败:文件大小超过限制",
					"max_size_bytes", maxSize,
					"current_size_bytes", written+int64(n),
				)
			}

			// 执行写入操作
			nw, writeErr := dst.Write(buf[:n])
			written += int64(nw)

			// 检查写入错误
			if writeErr != nil {
				return written, errors.WrapIf(writeErr, "文件写入失败")
			}

			// 验证写入字节数是否匹配
			if nw != n {
				return written, errors.NewWithDetails(
					"文件复制失败:写入字节数不匹配",
					"expected_bytes", n,
					"actual_bytes", nw,
				)
			}
		}

		// 处理读取结束或错误
		if err != nil {
			if err == io.EOF {
				break
			}
			return written, errors.WrapIf(err, "文件读取失败")
		}
	}

	return written, nil
}

// isPathSafe 检查目标路径是否在基础目录范围内，防止路径遍历攻击
func isPathSafe(target, base string) bool {
	cleanTarget := filepath.Clean(target)
	rel, err := filepath.Rel(base, cleanTarget)
	if err != nil {
		return false
	}

	// 检查是否尝试跳出基础目录
	return !strings.HasPrefix(rel, ".."+string(os.PathSeparator)) &&
		!strings.HasPrefix(rel, "..") &&
		rel != ".."
}
