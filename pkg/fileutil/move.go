package fileutil

import (
	"context"
	"os"
	"path/filepath"

	"emperror.dev/errors"
)

// Move 将文件或目录从源路径移动到目标路径。
// 特性:
//  1. 目标是已存在目录 → 源移动到该目录下
//  2. 目标是文件/不存在 → 源重命名为目标路径
//  3. 跨文件系统时自动降级为「复制+删除」
//
// 示例:
//
//	Move(context.Background(), "/tmp/src.txt", "/tmp/dst.txt") // 重命名
//	Move(context.Background(), "/tmp/src.txt", "/tmp/dir/")    // 移动到目录
//	Move(context.Background(), "/tmp/src", "/tmp/dst")         // 移动目录
func Move(ctx context.Context, src, dst string) error {
	// 公共校验
	if err := ValidatePath(ctx, src); err != nil {
		return errors.WithMessage(err, "源路径校验失败")
	}
	if err := ValidatePath(ctx, dst); err != nil {
		return errors.WithMessage(err, "目标路径校验失败")
	}

	// 获取源信息
	srcInfo, err := GetFileInfo(ctx, src)
	if err != nil {
		return errors.WithMessage(err, "获取源路径信息失败")
	}

	// 处理目标路径
	dst = CleanPath(dst)
	dstInfo, err := GetFileInfo(ctx, dst)
	if err == nil {
		if IsSameFile(srcInfo, dstInfo) {
			return errors.New("源和目标路径相同")
		}
		// 目标是目录 → 拼接文件名
		if dstInfo.IsDir() {
			dst = filepath.Join(dst, filepath.Base(src))
			// 再次检查拼接后的路径是否相同
			if newDstInfo, err := GetFileInfo(ctx, dst); err == nil && IsSameFile(srcInfo, newDstInfo) {
				return errors.New("拼接后源和目标路径相同")
			}
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return errors.WithMessage(err, "检查目标路径失败")
	}

	// 确保父目录存在
	if err := EnsureParentDir(ctx, dst); err != nil {
		return errors.WithMessage(err, "确保目标路径父目录存在失败")
	}

	// 尝试原子重命名（同文件系统）
	if err := os.Rename(src, dst); err == nil {
		return nil
	}

	// 跨文件系统 → 原子性复制+删除
	if srcInfo.IsDir() {
		// 先复制到临时目录，再重命名
		tempDst := dst + ".tmp"
		if err := CopyDir(ctx, src, tempDst, false); err != nil {
			// 清理临时文件
			_ = RemoveAll(ctx, tempDst)
			return errors.WithMessage(err, "跨文件系统复制目录失败")
		}

		// 重命名临时目录到目标
		if err := os.Rename(tempDst, dst); err != nil {
			// 清理临时文件
			_ = RemoveAll(ctx, tempDst)
			return errors.WithMessage(err, "重命名临时目录失败")
		}

		// 删除源目录
		if err := RemoveAll(ctx, src); err != nil {
			// 注意：此时目标已存在，源删除失败需要记录但不要回滚
			return errors.WithMessage(err, "跨文件系统删除源目录失败")
		}
	} else {
		// 先复制到临时文件，再重命名
		tempDst := dst + ".tmp"
		if err := CopyFile(ctx, src, tempDst); err != nil {
			// 清理临时文件
			_ = Remove(ctx, tempDst)
			return errors.WithMessage(err, "跨文件系统复制文件失败")
		}

		// 重命名临时文件到目标
		if err := os.Rename(tempDst, dst); err != nil {
			// 清理临时文件
			_ = Remove(ctx, tempDst)
			return errors.WithMessage(err, "重命名临时文件失败")
		}

		// 删除源文件
		if err := Remove(ctx, src); err != nil {
			// 注意：此时目标已存在，源删除失败需要记录但不要回滚
			return errors.WithMessage(err, "跨文件系统删除源文件失败")
		}
	}

	return nil
}
