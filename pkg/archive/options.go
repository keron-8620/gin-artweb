package archive

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type ArchiveOptions struct {
	Context             context.Context
	MaxFileSize         int64
	MaxFiles            int
	ExcludePatterns     []string
	IncludeOnly         []string
	FollowSymlinks      bool
	BufferSize          int
	MaxExtractPathDepth int
	AllowedExtensions   []string
	ExtractDirMode      os.FileMode
}

var DefaultArchiveOptions = ArchiveOptions{
	Context:             context.Background(),
	MaxFileSize:         100 << 20,
	MaxFiles:            10000,
	BufferSize:          32 * 1024,
	FollowSymlinks:      false,
	MaxExtractPathDepth: 100,
	AllowedExtensions:   nil,
}

type ArchiveOption func(*ArchiveOptions)

func WithContext(ctx context.Context) ArchiveOption {
	return func(opts *ArchiveOptions) {
		opts.Context = ctx
	}
}

func WithMaxFileSize(size int64) ArchiveOption {
	return func(opts *ArchiveOptions) {
		opts.MaxFileSize = size
	}
}

func WithMaxFiles(count int) ArchiveOption {
	return func(opts *ArchiveOptions) {
		opts.MaxFiles = count
	}
}

func WithBufferSize(size int) ArchiveOption {
	return func(opts *ArchiveOptions) {
		opts.BufferSize = size
	}
}

func WithFollowSymlinks(follow bool) ArchiveOption {
	return func(opts *ArchiveOptions) {
		opts.FollowSymlinks = follow
	}
}

func WithExcludePatterns(patterns []string) ArchiveOption {
	return func(opts *ArchiveOptions) {
		opts.ExcludePatterns = patterns
	}
}

func WithIncludeOnly(patterns []string) ArchiveOption {
	return func(opts *ArchiveOptions) {
		opts.IncludeOnly = patterns
	}
}

func WithMax解压PathDepth(depth int) ArchiveOption {
	return func(opts *ArchiveOptions) {
		opts.MaxExtractPathDepth = depth
	}
}

func WithAllowedExtensions(exts []string) ArchiveOption {
	return func(opts *ArchiveOptions) {
		opts.AllowedExtensions = exts
	}
}

// WithExtractDirMode overrides directory permissions while extracting.
// A zero value keeps the archive's original directory permissions.
func WithExtractDirMode(mode os.FileMode) ArchiveOption {
	return func(opts *ArchiveOptions) {
		opts.ExtractDirMode = mode.Perm()
	}
}

func applyOptions(opts ...ArchiveOption) ArchiveOptions {
	options := DefaultArchiveOptions
	for _, opt := range opts {
		opt(&options)
	}
	return options
}

func checkContext(ctx context.Context) error {
	if ctx == nil {
		return nil
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

func safeCopy(ctx context.Context, dst io.Writer, src io.Reader, maxSize int64, bufferSize int) (int64, error) {
	if bufferSize <= 0 {
		bufferSize = 32 * 1024
	}
	buf := make([]byte, bufferSize)
	var written int64
	for {
		if err := checkContext(ctx); err != nil {
			return written, fmt.Errorf("context check failed: %w", err)
		}
		n, err := src.Read(buf)
		if n > 0 {
			if maxSize > 0 && written+int64(n) > maxSize {
				return written, fmt.Errorf("file size exceeds limit: %d bytes", maxSize)
			}
			nw, writeErr := dst.Write(buf[:n])
			written += int64(nw)
			if writeErr != nil {
				return written, fmt.Errorf("write failed: %w", writeErr)
			}
			if nw != n {
				return written, io.ErrShortWrite
			}
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return written, fmt.Errorf("read failed: %w", err)
		}
	}
	return written, nil
}

func isPathSafe(target, base string) bool {
	if target == "" || base == "" {
		return false
	}
	cleanTarget := filepath.Clean(target)
	cleanBase := filepath.Clean(base)
	absTarget, err := filepath.Abs(cleanTarget)
	if err != nil {
		return false
	}
	absBase, err := filepath.Abs(cleanBase)
	if err != nil {
		return false
	}
	rel, err := filepath.Rel(absBase, absTarget)
	if err != nil {
		return false
	}
	if rel == "." {
		return true
	}
	if strings.HasPrefix(rel, ".."+string(filepath.Separator)) || rel == ".." {
		return false
	}
	if !filepath.IsAbs(cleanTarget) && strings.Contains(cleanTarget, "..") {
		return false
	}
	return true
}
