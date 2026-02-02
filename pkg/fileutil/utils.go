package fileutil

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

// ValidatePath 校验路径非空
func ValidatePath(filePath string) error {
	if strings.TrimSpace(filePath) == "" {
		return errors.New("路径不能为空")
	}
	return nil
}

// GetFileInfo 获取文件信息，封装通用错误（使用 WrapIf 避免堆栈重复）
func GetFileInfo(filePath string) (os.FileInfo, error) {
	if err := ValidatePath(filePath); err != nil {
		return nil, errors.WithMessagef(err, "路径校验失败, filepath=%s", filePath)
	}

	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.WithMessagef(err, "路径不存在, filepath=%s", filePath)
		}
		return nil, errors.WithMessagef(err, "获取路径信息失败, filepath=%s", filePath)
	}

	return info, nil
}

// IsSameFile 检查两个路径是否指向同一个文件/目录
func IsSameFile(srcInfo, dstInfo os.FileInfo) bool {
	return os.SameFile(srcInfo, dstInfo)
}

// EnsureParentDir 确保目标路径的父目录存在，继承源路径权限（默认0755）
func EnsureParentDir(dstPath string, perm ...os.FileMode) error {
	dir := filepath.Dir(dstPath)
	p := os.FileMode(0755)
	if len(perm) > 0 {
		p = perm[0]
	}

	if err := Mkdir(dir, p); err != nil {
		return errors.WithMessagef(err, "创建父目录失败, parent_dir=%s", dir)
	}

	return nil
}

// CleanPath 清理路径（去除冗余分隔符、解析相对路径）
func CleanPath(filePath string) string {
	return filepath.Clean(filepath.FromSlash(filePath))
}
