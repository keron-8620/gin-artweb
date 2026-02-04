package fileutil

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

// Remove 删除文件或空目录。
// 路径不存在 → 返回 nil；目录非空 → 返回错误。
// 示例:
//
//	Remove(context.Background(), "/tmp/test.txt") // 删除文件
//	Remove(context.Background(), "/tmp/empty_dir") // 删除空目录
func Remove(ctx context.Context, filePath string) error {
	if err := ValidatePath(ctx, filePath); err != nil {
		return errors.WithMessage(err, "路径校验失败")
	}

	filePath = CleanPath(filePath)
	// 路径不存在 → 直接返回
	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return errors.WithMessagef(err, "检查路径状态失败, filepath=%s", filePath)
	}

	// 删除文件/空目录
	if err := os.Remove(filePath); err != nil {
		return errors.WithMessagef(err, "删除文件/空目录失败, filepath=%s", filePath)
	}

	return nil
}

// RemoveAll 删除路径及其所有子项（递归删除）。
// 路径不存在 → 返回 nil。
// 示例:
//
//	RemoveAll(context.Background(), "/tmp/test_dir") // 删除目录及所有内容
func RemoveAll(ctx context.Context, filePath string) error {
	if err := ValidatePath(ctx, filePath); err != nil {
		return errors.WithMessage(err, "路径校验失败")
	}

	filePath = CleanPath(filePath)
	// 路径不存在 → 直接返回
	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return errors.WithMessagef(err, "检查路径状态失败, filepath=%s", filePath)
	}

	// 递归删除
	if err := os.RemoveAll(filePath); err != nil {
		return errors.WithMessagef(err, "递归删除路径失败, filepath=%s", filePath)
	}

	return nil
}

// SafeRemoveAll 安全删除路径（防止误删系统/关键目录）。
// 增强安全检查:
//  1. 拒绝删除系统核心路径（/、/usr 等）
//  2. 拒绝删除当前工作目录或其父目录
//  3. 路径必须是绝对路径且非系统路径
//
// 示例:
//
//	SafeRemoveAll(context.Background(), "/tmp/test_dir") // 正常删除
//	SafeRemoveAll(context.Background(), "/usr") // 拒绝删除
func SafeRemoveAll(ctx context.Context, filePath string) error {
	if err := ValidatePath(ctx, filePath); err != nil {
		return errors.WithMessage(err, "路径校验失败")
	}

	// 解析绝对路径
	absPath, err := filepath.Abs(CleanPath(filePath))
	if err != nil {
		return errors.WithMessagef(err, "解析绝对路径失败, filepath=%s", filePath)
	}

	// 安全路径检查（精确匹配，避免前缀误判）
	unsafePaths := map[string]bool{
		"/":          true,
		"/usr":       true,
		"/usr/local": true,
		"/etc":       true,
		"/var":       true,
		"/lib":       true,
		"/lib64":     true,
		"/bin":       true,
		"/sbin":      true,
		"/boot":      true,
		"/dev":       true,
		"/proc":      true,
		"/sys":       true,
	}

	// 1. 拒绝精确匹配系统路径
	if unsafePaths[absPath] {
		return errors.Errorf("安全检查失败, 禁止删除系统核心路径, abs_path=%s", absPath)
	}

	// 2. 拒绝子路径（精确分隔符，避免 /usr-local 误判）
	for unsafePath := range unsafePaths {
		if strings.HasPrefix(absPath, unsafePath+"/") {
			return errors.Errorf("安全检查失败, 禁止删除系统路径子目录, abs_path=%s, unsafe_parent=%s", absPath, unsafePath)
		}
	}

	// 3. 拒绝删除当前工作目录/父目录
	cwd, err := os.Getwd()
	if err == nil {
		cwd = filepath.Clean(cwd)
		if absPath == cwd {
			return errors.Errorf("安全检查失败, 禁止删除当前工作目录, abs_path=%s, cwd=%s", absPath, cwd)
		}
		if strings.HasPrefix(cwd, absPath+"/") {
			return errors.Errorf("安全检查失败, 禁止删除当前工作目录的父目录, abs_path=%s, cmd=%s", absPath, cwd)
		}
	} else {
		return errors.WithMessage(err, "获取当前工作目录失败")
	}

	// 执行安全删除
	return RemoveAll(ctx, absPath)
}

// RemoveEmptyDir 仅删除空目录。
// 路径非目录/目录非空 → 返回错误；路径不存在 → 返回 nil。
// 示例:
//
//	RemoveEmptyDir(context.Background(), "/tmp/empty_dir") // 成功删除
//	RemoveEmptyDir(context.Background(), "/tmp/non_empty_dir") // 返回错误
func RemoveEmptyDir(ctx context.Context, filePath string) error {
	if err := ValidatePath(ctx, filePath); err != nil {
		return errors.WithMessage(err, "路径校验失败")
	}

	filePath = CleanPath(filePath)
	info, infoErr := GetFileInfo(ctx, filePath)
	if infoErr != nil {
		return errors.WithMessage(infoErr, "获取路径信息失败")
	}

	if !info.IsDir() {
		return errors.Errorf("路径不是目录，无法删除空目录, filepath=%s", filePath)
	}

	// 尝试删除（非空会失败）
	if err := os.Remove(filePath); err != nil {
		return errors.WithMessagef(err, "删除空目录失败（可能目录非空）, filepath=%s", filePath)
	}

	return nil
}
