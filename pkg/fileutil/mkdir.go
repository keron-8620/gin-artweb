package fileutil

import (
	"context"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

// Mkdir 创建单个目录（不创建父目录），权限由 perm 指定。
// 目录已存在 → 返回 nil；路径存在但非目录 → 返回 ErrPathExist。
// 示例:
//
//	Mkdir(context.Background(), "/tmp/test", 0755) // 创建 /tmp/test（父目录必须存在）
func Mkdir(ctx context.Context, dirPath string, perm os.FileMode) error {
	if err := ValidatePath(ctx, dirPath); err != nil {
		return errors.WithMessage(err, "路径校验失败")
	}

	dirPath = CleanPath(dirPath)
	info, err := GetFileInfo(ctx, dirPath)
	if err == nil {
		if info.IsDir() {
			return nil // 目录已存在
		}
		return errors.Errorf("路径已存在但不是目录, filepath=%s", dirPath)
	} else if !errors.Is(err, os.ErrNotExist) {
		return errors.WithMessage(err, "获取路径信息失败")
	}

	// 仅创建单个目录
	if err := os.Mkdir(dirPath, perm); err != nil {
		return errors.WithMessagef(err, "创建单个目录失败, dirpath=%s", dirPath)
	}

	return nil
}

// MkdirAll 创建目录及所有必要的父目录，权限由 perm 指定。
// 目录已存在 → 返回 nil；路径存在但非目录 → 返回 ErrPathExist。
// 示例:
//
//	MkdirAll(context.Background(), "/tmp/a/b/c", 0755) // 递归创建 a/b/c
func MkdirAll(ctx context.Context, dirPath string, perm os.FileMode) error {
	if err := ValidatePath(ctx, dirPath); err != nil {
		return errors.WithMessage(err, "路径校验失败")
	}

	dirPath = CleanPath(dirPath)
	info, err := GetFileInfo(ctx, dirPath)
	if err == nil {
		if info.IsDir() {
			return nil // 目录已存在
		}
		return errors.Errorf("路径已存在但不是目录, dirpath=%s", dirPath)
	} else if !errors.Is(err, os.ErrNotExist) {
		return errors.WithMessage(err, "获取路径信息失败")
	}

	// 递归创建目录
	if err := os.MkdirAll(dirPath, perm); err != nil {
		return errors.WithMessagef(err, "创建目录树失败, dirpath=%s", dirPath)
	}

	return nil
}

// EnsureDir 确保文件路径的父目录存在（默认权限 0755）。
// 常用于创建文件前预检查父目录。
// 示例:
//
//	EnsureDir(context.Background(), "/tmp/a/b/c.txt") // 确保 /tmp/a/b 存在
func EnsureDir(ctx context.Context, dirPath string) error {
	if err := ValidatePath(ctx, dirPath); err != nil {
		return errors.WithMessage(err, "路径校验失败")
	}

	dir := CleanPath(filepath.Dir(dirPath))
	if err := MkdirAll(ctx, dir, 0755); err != nil {
		return errors.WithMessage(err, "确保父目录存在失败")
	}
	return nil
}
