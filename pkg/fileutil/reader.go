package fileutil

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"

	"emperror.dev/errors"
)

func WriteReaderToFile(
	ctx context.Context,
	fileReader io.Reader,
	filePath string,
	fileMode os.FileMode,
	overwrite bool,
) error {
	if err := ctx.Err(); err != nil {
		return errors.WrapIf(err, "上下文已取消")
	}

	if fileReader == nil {
		return errors.New("fileReader 不能为空")
	}
	if strings.TrimSpace(filePath) == "" {
		return errors.New("filePath 不能为空")
	}

	if !overwrite {
		if _, err := os.Stat(filePath); err == nil {
			return errors.NewWithDetails("文件已存在: %s", filePath)
		} else if !errors.Is(err, os.ErrNotExist) {
			return errors.WithMessagef(err, "检查文件状态失败: %s", filePath)
		}
	}

	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, os.FileMode(0o750)); err != nil {
		return errors.WithMessagef(err, "创建目录失败: %s", dir)
	}

	tempFile, err := os.CreateTemp(dir, ".tmp-save-reader-*")
	if err != nil {
		return errors.WithMessagef(err, "创建临时文件失败: %s", dir)
	}
	tempPath := tempFile.Name()

	if _, err = io.Copy(tempFile, fileReader); err != nil {
		return errors.WithMessagef(err, "复制数据到临时文件失败: %s", tempPath)
	}

	if err := tempFile.Sync(); err != nil {
		tempFile.Close()
		return errors.WithMessagef(err, "同步临时文件失败: %s", tempPath)
	}

	if err := tempFile.Close(); err != nil {
		return errors.WithMessagef(err, "关闭临时文件失败: %s", tempPath)
	}

	if err := os.Rename(tempPath, filePath); err != nil {
		return errors.WithMessagef(err, "重命名临时文件失败: %s -> %s", tempPath, filePath)
	}

	if err := os.Chmod(filePath, fileMode); err != nil {
		return errors.WithMessagef(err, "设置文件权限失败: %s", filePath)
	}

	return nil
}
