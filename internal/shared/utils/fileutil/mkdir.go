package fileutil

import (
	"fmt"
	"os"
	"path/filepath"
)

// Mkdir 使用指定的权限位创建名为 path 的目录。
// 如果目录已经存在，则不执行任何操作并返回 nil。
func Mkdir(path string, perm os.FileMode) error {
	// 验证输入路径
	if path == "" {
		return fmt.Errorf("路径不能为空")
	}

	// 检查目录是否已存在
	if info, err := os.Stat(path); err == nil {
		if info.IsDir() {
			// 目录已存在
			return nil
		}
		// 路径存在但不是目录
		return fmt.Errorf("路径存在但不是目录: %s", path)
	} else if !os.IsNotExist(err) {
		// 其他状态错误
		return fmt.Errorf("检查路径失败: %w", err)
	}

	// 创建目录
	if err := os.Mkdir(path, perm); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	return nil
}

// MkdirAll 创建名为 path 的目录以及任何必要的父目录，
// 使用指定的权限位。
// 如果目录已经存在，则不执行任何操作并返回 nil。
func MkdirAll(path string, perm os.FileMode) error {
	// 验证输入路径
	if path == "" {
		return fmt.Errorf("路径不能为空")
	}

	// 清理路径
	path = filepath.Clean(path)

	// 检查目录是否已存在
	if info, err := os.Stat(path); err == nil {
		if info.IsDir() {
			// 目录已存在
			return nil
		}
		// 路径存在但不是目录
		return fmt.Errorf("路径存在但不是目录: %s", path)
	} else if !os.IsNotExist(err) {
		// 其他状态错误
		return fmt.Errorf("检查路径失败: %w", err)
	}

	// 创建目录及任何必要的父目录
	if err := os.MkdirAll(path, perm); err != nil {
		return fmt.Errorf("创建目录路径失败: %w", err)
	}

	return nil
}

// EnsureDir 确保包含给定文件路径的目录存在。
// 当您想在创建文件之前确保其父目录存在时，这很有用。
func EnsureDir(filePath string) error {
	// 验证输入路径
	if filePath == "" {
		return fmt.Errorf("文件路径不能为空")
	}

	// 获取文件路径的目录部分
	dir := filepath.Dir(filePath)

	// 如果目录不存在，则使用默认权限创建目录
	return MkdirAll(dir, 0755)
}