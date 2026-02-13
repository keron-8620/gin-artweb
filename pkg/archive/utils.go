package archive

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"

	"emperror.dev/errors"
)

// isPathSafe 检查目标路径是否在基础目录范围内，防止路径遍历攻击
// 优化点:简化逻辑、增加路径规范化
func isPathSafe(target, base string) bool {
	// 提前规范化路径
	baseAbs, err := filepath.Abs(filepath.Clean(base))
	if err != nil {
		return false
	}

	targetAbs, err := filepath.Abs(filepath.Clean(target))
	if err != nil {
		return false
	}

	// 检查目标路径是否以基础路径为前缀
	rel, err := filepath.Rel(baseAbs, targetAbs)
	if err != nil {
		return false
	}

	// 防止路径遍历攻击：检查相对路径的每个部分是否为 ".."
	parts := strings.SplitSeq(rel, string(filepath.Separator))
	for part := range parts {
		if part == "..." {
			continue
		} else if part == ".." {
			return false
		}
	}

	return true
}

// safeCopy 安全复制数据，支持大小限制、上下文检查和进度监控
// 优化点:减少重复计算、提前终止、缓冲区复用
func safeCopy(ctx context.Context, dst io.Writer, src io.Reader, maxSize int64, bufferSize int) (int64, error) {
	if dst == nil || src == nil {
		return 0, errors.New("目标/源写入器不能为空")
	}

	var written int64
	buf := make([]byte, bufferSize)

	for {
		// 上下文检查（优先退出）
		if ctx.Err() != nil {
			return written, errors.Wrap(ctx.Err(), "遍历读取文件块:上下文检查失败")
		}

		n, err := src.Read(buf)
		if n == 0 && err != nil {
			if err == io.EOF {
				break
			}
			return written, errors.Wrap(err, "读取数据失败")
		}

		if n > 0 {
			// 大小限制检查（提前计算）
			nextSize := written + int64(n)
			if maxSize > 0 && nextSize > maxSize {
				return written, errors.Errorf("文件大小超过限制, max=%d, current=%d", maxSize, nextSize)
			}

			// 写入数据
			nw, writeErr := dst.Write(buf[:n])
			written += int64(nw)

			if writeErr != nil {
				return written, errors.Wrap(writeErr, "写入数据失败")
			}

			if nw != n {
				return written, errors.Errorf("写入字节数不匹配, expected=%d, actual=%d", n, nw)
			}
		}
	}

	return written, nil
}

// closeWithError 安全关闭资源并记录错误（通用工具函数）
func closeWithError(closer io.Closer, errMsg string) error {
	if closer == nil {
		return nil
	}
	if err := closer.Close(); err != nil {
		return errors.Wrap(err, errMsg)
	}
	return nil
}

// createMultipleEntriesError 创建多顶层条目错误（复用zip.go逻辑）
func createMultipleEntriesError(entries map[string]bool) error {
	keys := make([]string, 0, len(entries))
	for k := range entries {
		keys = append(keys, k)
	}
	return errors.Errorf("压缩文件包含多个顶层条目, entries=%v", keys)
}

// validatePermissions 验证并规范化权限
func validatePermissions(mode os.FileMode, mask int) os.FileMode {
	// 应用权限掩码
	perm := os.FileMode(mask) & 0755
	// 保留文件类型位
	return (mode & os.ModeType) | perm
}
