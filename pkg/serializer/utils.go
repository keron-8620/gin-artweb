package serializer

import (
	"context"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

// checkContext 检查上下文状态
func checkContext(ctx context.Context) error {
	if ctx == nil {
		return nil
	}

	select {
	case <-ctx.Done():
		return errors.New(ErrContextCancelled)
	default:
		return nil
	}
}

// validatePath 校验文件路径
func validatePath(filePath string) error {
	if filePath == "" {
		return errors.New(ErrInvalidPath)
	}
	return nil
}

// checkFileSize 检查文件大小限制
func checkFileSize(fileSize int64, maxSize int64) error {
	if maxSize > 0 && fileSize > maxSize {
		return errors.Errorf("%s: max=%d, current=%d", ErrFileTooLarge, maxSize, fileSize)
	}
	return nil
}

// createTempFile 创建临时文件
func createTempFile(filePath string) (string, error) {
	dir := filepath.Dir(filePath)
	base := filepath.Base(filePath)

	tempFile, err := os.CreateTemp(dir, base+".*.tmp")
	if err != nil {
		return "", errors.WithMessage(err, "创建临时文件失败")
	}
	tempFile.Close()

	return tempFile.Name(), nil
}

// atomicRename 原子重命名文件
func atomicRename(oldPath, newPath string) error {
	if err := os.Rename(oldPath, newPath); err != nil {
		return errors.WithMessagef(err, "原子重命名失败: %s -> %s", oldPath, newPath)
	}
	return nil
}
