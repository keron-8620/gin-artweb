package fileutil

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// CopyFile 从源路径复制单个文件到目标路径。
// 它会保留文件权限和修改时间。
func CopyFile(src, dst string) error {
	// 验证输入路径
	if src == "" {
		return fmt.Errorf("源路径不能为空")
	}
	if dst == "" {
		return fmt.Errorf("目标路径不能为空")
	}

	// 检查源文件是否存在
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("获取源文件信息失败: %w", err)
	}

	// 检查源是否确实是一个文件
	if srcInfo.IsDir() {
		return fmt.Errorf("源是一个目录，不是文件")
	}

	// 处理目标路径是已存在目录的情况
	if dstInfo, err := os.Stat(dst); err == nil {
		if os.SameFile(srcInfo, dstInfo) {
			return nil // 相同文件，无需操作
		}
		if dstInfo.IsDir() {
			dst = filepath.Join(dst, filepath.Base(src))
		}
	} else if os.IsNotExist(err) {
		// 目标不存在，检查父目录是否存在
		if dirInfo, err := os.Stat(filepath.Dir(dst)); err == nil && dirInfo.IsDir() {
			// 父目录存在，这是正常的
		} else if os.IsNotExist(err) {
			// 需要创建父目录
		} else {
			return fmt.Errorf("检查父目录失败: %w", err)
		}
	} else {
		return fmt.Errorf("检查目标路径失败: %w", err)
	}

	// 确保目标目录存在
	dstDir := filepath.Dir(dst)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("创建目标目录失败: %w", err)
	}

	// 打开源文件
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("打开源文件失败: %w", err)
	}
	defer srcFile.Close()

	// 创建目标文件
	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("创建目标文件失败: %w", err)
	}
	defer dstFile.Close()

	// 使用缓冲区复制文件内容，提高大文件复制性能
	buffer := make([]byte, 64*1024) // 64KB buffer
	if _, err := io.CopyBuffer(dstFile, srcFile, buffer); err != nil {
		return fmt.Errorf("复制文件内容失败: %w", err)
	}

	// 同步到磁盘
	if err := dstFile.Sync(); err != nil {
		return fmt.Errorf("同步目标文件失败: %w", err)
	}

	// 保留修改时间
	if err := os.Chtimes(dst, time.Now(), srcInfo.ModTime()); err != nil {
		return fmt.Errorf("保留修改时间失败: %w", err)
	}

	return nil
}

// CopyDir 复制目录树从源到目标
// 如果 copyContents 为 true，行为类似于 `cp -r src/. dst`
// 如果 copyContents 为 false，行为类似于 `cp -r src dst`
func CopyDir(src, dst string, copyContents bool) error {
	// 验证输入路径
	if src == "" {
		return fmt.Errorf("源路径不能为空")
	}
	if dst == "" {
		return fmt.Errorf("目标路径不能为空")
	}

	// 获取源目录信息
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("获取源目录信息失败: %w", err)
	}

	// 检查源是否确实是一个目录
	if !srcInfo.IsDir() {
		return fmt.Errorf("源是文件，不是目录")
	}

	// 当复制内容时，我们不修改目标路径
	targetDst := dst
	if !copyContents {
		// 处理目标路径是已存在目录的情况
		if dstInfo, err := os.Stat(dst); err == nil {
			if os.SameFile(srcInfo, dstInfo) {
				return nil // 相同目录，无需操作
			}
			if dstInfo.IsDir() {
				targetDst = filepath.Join(dst, filepath.Base(src))
			}
		} else if os.IsNotExist(err) {
			// 目标不存在，这是正常的
		} else {
			return fmt.Errorf("检查目标路径失败: %w", err)
		}
	}

	// 创建具有相同权限的目标目录
	if err := os.MkdirAll(targetDst, srcInfo.Mode()); err != nil {
		return fmt.Errorf("创建目标目录失败: %w", err)
	}

	// 确定读取条目的源路径
	sourcePath := src
	if copyContents {
		// 当复制内容时，直接从源读取
		sourcePath = src
	}

	// 读取目录条目
	entries, err := os.ReadDir(sourcePath)
	if err != nil {
		return fmt.Errorf("读取源目录失败: %w", err)
	}

	// 处理每个条目
	for _, entry := range entries {
		srcPath := filepath.Join(sourcePath, entry.Name())
		dstPath := filepath.Join(targetDst, entry.Name())

		// 获取文件信息
		entryInfo, err := entry.Info()
		if err != nil {
			return fmt.Errorf("获取条目信息失败 %s: %w", entry.Name(), err)
		}

		if entryInfo.IsDir() {
			// 递归复制子目录
			// 对于内容复制，始终为子目录传递 copyContents=true
			if err := CopyDir(srcPath, dstPath, copyContents); err != nil {
				return fmt.Errorf("复制子目录 %s 到 %s 失败: %w", srcPath, dstPath, err)
			}
		} else {
			// 复制普通文件
			if err := CopyFile(srcPath, dstPath); err != nil {
				return fmt.Errorf("复制文件 %s 到 %s 失败: %w", srcPath, dstPath, err)
			}
		}
	}

	// 保留目录修改时间
	if err := os.Chtimes(targetDst, time.Now(), srcInfo.ModTime()); err != nil {
		return fmt.Errorf("保留目录修改时间失败: %w", err)
	}

	return nil
}
