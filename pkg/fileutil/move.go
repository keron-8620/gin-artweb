package fileutil

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

// Move 将文件或目录从源路径移动到目标路径。
// 特性:
//  1. 目标是已存在目录 → 源移动到该目录下
//  2. 目标是文件/不存在 → 源重命名为目标路径
//  3. 跨文件系统时自动降级为「复制+删除」
//
// 示例:
//
//	Move("/tmp/src.txt", "/tmp/dst.txt") // 重命名
//	Move("/tmp/src.txt", "/tmp/dir/")    // 移动到目录
//	Move("/tmp/src", "/tmp/dst")         // 移动目录
func Move(src, dst string) error {
	// 公共校验
	if err := ValidatePath(src); err != nil {
		return errors.WithMessage(err, "源路径校验失败")
	}
	if err := ValidatePath(dst); err != nil {
		return errors.WithMessage(err, "目标路径校验失败")
	}

	// 获取源信息
	srcInfo, err := GetFileInfo(src)
	if err != nil {
		return errors.WithMessage(err, "获取源路径信息失败")
	}

	// 处理目标路径
	dst = CleanPath(dst)
	dstInfo, err := GetFileInfo(dst)
	if err == nil {
		if IsSameFile(srcInfo, dstInfo) {
			return errors.New("源和目标路径相同")
		}
		// 目标是目录 → 拼接文件名
		if dstInfo.IsDir() {
			dst = filepath.Join(dst, filepath.Base(src))
			// 再次检查拼接后的路径是否相同
			if newDstInfo, err := GetFileInfo(dst); err == nil && IsSameFile(srcInfo, newDstInfo) {
				return errors.New("拼接后源和目标路径相同")
			}
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return errors.WithMessage(err, "检查目标路径失败")
	}

	// 确保父目录存在
	if err := EnsureParentDir(dst); err != nil {
		return errors.WithMessage(err, "确保目标路径父目录存在失败")
	}

	// 尝试原子重命名（同文件系统）
	if err := os.Rename(src, dst); err == nil {
		return nil
	}

	// 跨文件系统 → 复制+删除
	if srcInfo.IsDir() {
		// 复制目录
		if err := CopyDir(src, dst, false); err != nil {
			return errors.WithMessage(err, "跨文件系统复制目录失败")
		}
		// 删除源目录
		if err := RemoveAll(src); err != nil {
			return errors.WithMessage(err, "跨文件系统删除源目录失败")
		}
	} else {
		// 复制文件
		if err := CopyFile(src, dst); err != nil {
			return errors.WithMessage(err, "跨文件系统复制文件失败")
		}
		// 删除源文件
		if err := Remove(src); err != nil {
			return errors.WithMessage(err, "跨文件系统删除源文件失败")
		}
	}

	return nil
}
