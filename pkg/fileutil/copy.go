package fileutil

import (
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

// CopyFile 从源路径复制单个文件到目标路径。
// 保留文件权限、修改时间（mtime），访问时间（atime）使用默认值。
// 示例:
//
//	CopyFile("/tmp/src.txt", "/tmp/dst.txt") // 直接复制
//	CopyFile("/tmp/src.txt", "/tmp/dir/")    // 复制到目录（自动拼接文件名）
func CopyFile(src, dst string) error {
	// 公共校验
	if err := ValidatePath(src); err != nil {
		return errors.WithMessage(err, "源路径校验失败")
	}
	if err := ValidatePath(dst); err != nil {
		return errors.WithMessage(err, "目标路径校验失败")
	}

	// 获取源文件信息
	srcInfo, err := GetFileInfo(src)
	if err != nil {
		return errors.WithMessage(err, "获取源文件信息失败")
	}
	if srcInfo.IsDir() {
		return errors.Errorf("源路径是目录，无法复制为文件, src=%s", src)
	}

	// 处理目标路径（如果是目录，自动拼接文件名）
	dst = CleanPath(dst)
	dstInfo, err := GetFileInfo(dst)
	if err == nil {
		if IsSameFile(srcInfo, dstInfo) {
			return nil // 相同文件，无需操作
		}
		if dstInfo.IsDir() {
			dst = filepath.Join(dst, filepath.Base(src))
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return errors.WithMessage(err, "检查目标路径失败")
	}

	// 确保父目录存在（继承源文件权限）
	if err := EnsureParentDir(dst, srcInfo.Mode().Perm()); err != nil {
		return errors.WithMessage(err, "确保父目录存在失败")
	}

	// 打开源文件（只读）
	srcFile, err := os.Open(src)
	if err != nil {
		return errors.WithMessagef(err, "打开源文件失败, src=%s", src)
	}
	defer srcFile.Close()

	// 创建目标文件（继承源权限）
	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode().Perm())
	if err != nil {
		return errors.WithMessagef(err, "创建目标文件失败, dst=%s", dst)
	}
	defer dstFile.Close()

	// 带缓冲区复制（64KB 适配大多数场景）
	buffer := make([]byte, 64*1024)
	if _, err := io.CopyBuffer(dstFile, srcFile, buffer); err != nil {
		return errors.WithMessage(err, "复制文件内容失败")
	}

	// 同步到磁盘（确保数据落盘）
	if err := dstFile.Sync(); err != nil {
		return errors.WithMessage(err, "同步目标文件到磁盘失败")
	}

	// 保留修改时间（atime 用源文件的，无需设为当前时间）
	if err := os.Chtimes(dst, srcInfo.ModTime(), srcInfo.ModTime()); err != nil {
		return errors.WithMessagef(err, "设置文件时间失败, dst=%s", dst)
	}

	return nil
}

// CopyDir 复制目录树从源到目标
// copyContents 为 true: 复制 src 内的所有内容到 dst（类似 cp -r src/. dst）
// copyContents 为 false: 复制 src 目录本身到 dst（类似 cp -r src dst）
// 示例:
//
//	CopyDir("/tmp/src", "/tmp/dst", false) // 结果: /tmp/dst/src
//	CopyDir("/tmp/src", "/tmp/dst", true)  // 结果: /tmp/dst/[src内的文件]
func CopyDir(src, dst string, copyContents bool) error {
	// 公共校验
	if err := ValidatePath(src); err != nil {
		return errors.WithMessage(err, "源路径校验失败")
	}
	if err := ValidatePath(dst); err != nil {
		return errors.WithMessage(err, "目标路径校验失败")
	}

	// 获取源目录信息
	srcInfo, err := GetFileInfo(src)
	if err != nil {
		return errors.WithMessage(err, "获取源目录信息失败")
	}
	if !srcInfo.IsDir() {
		return errors.Errorf("源路径是文件，无法复制为目录, src=%s", src)
	}

	// 确定目标路径
	dest := dst
	if !copyContents {
		dstInfo, err := GetFileInfo(dst)
		if err == nil {
			if IsSameFile(srcInfo, dstInfo) {
				return nil // 相同目录，无需操作
			}
			if dstInfo.IsDir() {
				dest = filepath.Join(dst, filepath.Base(src))
			}
		} else if !errors.Is(err, os.ErrNotExist) {
			return errors.WithMessage(err, "检查目标路径失败")
		}
	}

	// 创建目标目录（继承源权限）
	if err := MkdirAll(dest, srcInfo.Mode().Perm()); err != nil {
		return errors.WithMessagef(err, "创建目标目录失败, dest=%s", dest)
	}

	// 读取源目录条目
	entries, err := os.ReadDir(src)
	if err != nil {
		return errors.WithMessagef(err, "读取源目录条目失败, src=%s", src)
	}

	// 遍历并复制所有条目
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dest, entry.Name())

		// 直接用 entry.Info()，避免重复 os.Stat
		entryInfo, err := entry.Info()
		if err != nil {
			return errors.WithMessagef(err, "获取条目信息失败, entry_name=%s", entry.Name())
		}

		if entryInfo.IsDir() {
			// 递归复制子目录（子目录始终复制内容）
			if err := CopyDir(srcPath, dstPath, true); err != nil {
				return errors.WithMessagef(err, "复制子目录失败, src_path=%s, dst_path=%s", srcPath, dstPath)
			}
		} else {
			// 复制文件
			if err := CopyFile(srcPath, dstPath); err != nil {
				return errors.WithMessagef(err, "复制文件失败, src_path=%s, dst_path=%s", srcPath, dstPath)
			}
		}
	}

	// 保留目录修改时间
	if err := os.Chtimes(dest, srcInfo.ModTime(), srcInfo.ModTime()); err != nil {
		return errors.WithMessagef(err, "设置目录时间失败, dest=%s", dest)
	}

	return nil
}
