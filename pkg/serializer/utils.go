package serializer

import (
	"path/filepath"

	"emperror.dev/errors"
)

// validatePath 校验文件路径：
// 1. 不能为空
// 2. 必须是绝对路径
// 3. 防止路径遍历攻击
func validatePath(filePath string) error {
	// 1. 路径不能为空
	if filePath == "" {
		return errors.New("文件路径不能为空")
	}

	// 2. 必须是绝对路径
	if !filepath.IsAbs(filePath) {
		return errors.New("必须使用绝对路径")
	}

	// 3. 清理路径并检查是否存在路径遍历
	cleanPath := filepath.Clean(filePath)
	if !filepath.IsAbs(cleanPath) {
		return errors.New("路径清理后不是绝对路径，可能存在路径遍历攻击")
	}

	// 4. 检查清理后的路径是否与原始路径解析到同一位置
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return errors.Wrap(err, "获取绝对路径失败")
	}

	if cleanPath != absPath {
		return errors.New("路径可能包含路径遍历攻击")
	}

	return nil
}
