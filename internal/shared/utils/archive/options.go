package archive

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gin-artweb/internal/shared/errors"
)

// ArchiveOptions 压缩/解压选项
type ArchiveOptions struct {
	// 上下文用于取消操作
	Context context.Context

	// 最大文件大小限制 (字节)，0表示无限制
	MaxFileSize int64

	// 最大文件数量限制，0表示无限制
	MaxFiles int

	// 排除文件模式
	ExcludePatterns []string

	// 包含文件模式
	IncludeOnly []string

	// 是否跟随符号链接
	FollowSymlinks bool

	// 复制缓冲区大小
	BufferSize int
}

// DefaultArchiveOptions 默认选项
var DefaultArchiveOptions = ArchiveOptions{
	Context:        context.Background(),
	MaxFileSize:    100 << 20, // 100MB
	MaxFiles:       10000,     // 10000个文件
	BufferSize:     32 * 1024, // 32KB
	FollowSymlinks: false,
}

// ArchiveOption 函数选项模式
type ArchiveOption func(*ArchiveOptions)

// WithContext 设置上下文
func WithContext(ctx context.Context) ArchiveOption {
	return func(opts *ArchiveOptions) {
		opts.Context = ctx
	}
}

// WithMaxFileSize 设置最大文件大小
func WithMaxFileSize(size int64) ArchiveOption {
	return func(opts *ArchiveOptions) {
		opts.MaxFileSize = size
	}
}

// WithMaxFiles 设置最大文件数量
func WithMaxFiles(count int) ArchiveOption {
	return func(opts *ArchiveOptions) {
		opts.MaxFiles = count
	}
}

// WithBufferSize 设置缓冲区大小
func WithBufferSize(size int) ArchiveOption {
	return func(opts *ArchiveOptions) {
		opts.BufferSize = size
	}
}

// applyOptions 应用选项
func applyOptions(opts ...ArchiveOption) ArchiveOptions {
	options := DefaultArchiveOptions
	for _, opt := range opts {
		opt(&options)
	}
	return options
}

// safeCopy 安全复制，带大小限制和上下文检查
func safeCopy(ctx context.Context, dst io.Writer, src io.Reader, maxSize int64, bufferSize int) (int64, error) {
	var written int64
	buf := make([]byte, bufferSize)

	for {
		if err := errors.CheckContext(ctx); err != nil {
			return written, err
		}

		n, err := src.Read(buf)
		if n > 0 {
			if maxSize > 0 && written+int64(n) > maxSize {
				return written, fmt.Errorf("文件大小超过限制: %d bytes", maxSize)
			}

			nw, writeErr := dst.Write(buf[:n])
			written += int64(nw)

			if writeErr != nil {
				return written, writeErr
			}

			if nw != n {
				return written, io.ErrShortWrite
			}
		}

		if err != nil {
			if err == io.EOF {
				break
			}
			return written, err
		}
	}

	return written, nil
}

// isPathSafe 检查路径是否安全
func isPathSafe(target, base string) bool {
	cleanTarget := filepath.Clean(target)
	rel, err := filepath.Rel(base, cleanTarget)
	if err != nil {
		return false
	}

	// 检查是否跳出基础目录
	return !strings.HasPrefix(rel, ".."+string(os.PathSeparator)) &&
		!strings.HasPrefix(rel, "..") &&
		rel != ".."
}
