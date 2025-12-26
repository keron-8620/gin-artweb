package fileutil

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Remove 删除指定名称的文件或目录。
// 如果路径是目录，则只有在目录为空时才会被删除。
// 如果路径不存在，则 Remove 返回 nil（无错误）。
func Remove(path string) error {
	// 验证输入路径
	if path == "" {
		return fmt.Errorf("路径不能为空")
	}

	// 检查路径是否存在
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			// 路径不存在，无需删除
			return nil
		}
		return fmt.Errorf("检查路径失败: %w", err)
	}

	// 删除文件或空目录
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("删除 %s 失败: %w", path, err)
	}

	return nil
}

// RemoveAll 删除路径及其包含的所有子项。
// 它会删除所有能删除的内容，但返回遇到的第一个错误。
// 如果路径不存在，则 RemoveAll 返回 nil（无错误）。
func RemoveAll(path string) error {
	// 验证输入路径
	if path == "" {
		return fmt.Errorf("路径不能为空")
	}

	// 检查路径是否存在
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			// 路径不存在，无需删除
			return nil
		}
		return fmt.Errorf("检查路径失败: %w", err)
	}

	// 删除路径及其所有内容
	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("删除全部 %s 失败: %w", path, err)
	}

	return nil
}

// RemoveIfExists 如果文件或目录存在则删除它。
// 这是 Remove 的别名，用于语义上的清晰性。
func RemoveIfExists(path string) error {
	return Remove(path)
}

// SafeRemoveAll 删除路径及其包含的所有子项，
// 但包含安全检查以防止意外删除重要路径。
func SafeRemoveAll(path string) error {
	// 验证输入路径
	if path == "" {
		return fmt.Errorf("路径不能为空")
	}

	// 解析绝对路径以进行安全检查
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("解析绝对路径失败: %w", err)
	}

	// 安全检查以防止意外删除系统路径
	unsafePaths := []string{
		"/",
		"/usr",
		"/usr/local",
		"/etc",
		"/var",
		"/lib",
		"/lib64",
		"/bin",
		"/sbin",
		"/boot",
		"/dev",
		"/proc",
		"/sys",
	}

	for _, unsafePath := range unsafePaths {
		if strings.HasPrefix(absPath, unsafePath) && absPath == unsafePath {
			return fmt.Errorf("拒绝删除受保护的系统路径: %s", absPath)
		}
	}

	// 还要防止删除当前工作目录或父路径
	cwd, err := os.Getwd()
	if err == nil {
		if absPath == cwd || strings.HasPrefix(cwd, absPath+"/") {
			return fmt.Errorf("拒绝删除当前工作目录或父路径: %s", absPath)
		}
	}

	// 检查路径是否存在
	if _, err := os.Stat(absPath); err != nil {
		if os.IsNotExist(err) {
			// 路径不存在，无需删除
			return nil
		}
		return fmt.Errorf("检查路径失败: %w", err)
	}

	// 删除路径及其所有内容
	if err := os.RemoveAll(absPath); err != nil {
		return fmt.Errorf("安全删除全部 %s 失败: %w", absPath, err)
	}

	return nil
}

// RemoveEmptyDir 仅在目录为空时删除该目录。
// 如果路径不是目录或包含文件，则返回错误。
func RemoveEmptyDir(path string) error {
	// 验证输入路径
	if path == "" {
		return fmt.Errorf("路径不能为空")
	}

	// 检查路径是否存在
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			// 路径不存在，无需删除
			return nil
		}
		return fmt.Errorf("检查路径失败: %w", err)
	}

	// 检查路径是否为目录
	if !info.IsDir() {
		return fmt.Errorf("路径不是目录: %s", path)
	}

	// 尝试删除（如果目录不为空则会失败）
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("删除目录失败（可能不为空）: %w", err)
	}

	return nil
}
