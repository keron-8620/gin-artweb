package archive

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gin-artweb/internal/shared/errors"
)

// TarGz 将指定路径的文件或目录压缩为 tar.gz 格式
func TarGz(src, dst string, opts ...ArchiveOption) error {
	options := applyOptions(opts...)

	// 检查上下文
	if err := errors.CheckContext(options.Context); err != nil {
		return err
	}

	// 创建目标文件
	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("创建目标文件 %s 失败: %w", dst, err)
	}
	defer func() {
		if closeErr := dstFile.Close(); closeErr != nil {
			// 记录关闭错误
			panic(closeErr)
		}
	}()

	// 创建 gzip writer
	gzWriter := gzip.NewWriter(dstFile)
	defer func() {
		if closeErr := gzWriter.Close(); closeErr != nil {
			// 记录关闭错误
			panic(closeErr)
		}
	}()

	// 创建 tar writer
	tarWriter := tar.NewWriter(gzWriter)
	defer func() {
		if closeErr := tarWriter.Close(); closeErr != nil {
			// 记录关闭错误
		}
	}()

	// 获取源文件信息
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("获取源文件信息 %s 失败: %w", src, err)
	}

	fileCount := 0

	// 处理目录
	if srcInfo.IsDir() {
		return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// 检查上下文
			if err := errors.CheckContext(options.Context); err != nil {
				return err
			}

			// 计算相对路径
			relPath, err := filepath.Rel(filepath.Dir(src), path)
			if err != nil {
				return fmt.Errorf("计算相对路径失败 %s: %w", path, err)
			}

			// 跳过根目录本身
			if relPath == "." {
				return nil
			}

			// 检查文件数量限制
			fileCount++
			if options.MaxFiles > 0 && fileCount > options.MaxFiles {
				return fmt.Errorf("文件数量超过限制: %d", options.MaxFiles)
			}

			// 创建 tar header
			header, err := tar.FileInfoHeader(info, "")
			if err != nil {
				return fmt.Errorf("创建tar头信息失败 %s: %w", path, err)
			}
			header.Name = relPath

			// 写入 header
			if err := tarWriter.WriteHeader(header); err != nil {
				return fmt.Errorf("写入tar头信息失败 %s: %w", path, err)
			}

			// 如果是普通文件，写入内容
			if info.Mode().IsRegular() {
				// 检查文件大小限制
				if options.MaxFileSize > 0 && info.Size() > options.MaxFileSize {
					return fmt.Errorf("文件 %s 大小 %d 超过限制 %d", path, info.Size(), options.MaxFileSize)
				}

				file, err := os.Open(path)
				if err != nil {
					return fmt.Errorf("打开文件失败 %s: %w", path, err)
				}
				defer file.Close()

				if _, err := safeCopy(options.Context, tarWriter, file, options.MaxFileSize, options.BufferSize); err != nil {
					return fmt.Errorf("复制文件内容失败 %s: %w", path, err)
				}
			}

			return nil
		})
	} else {
		// 处理单个文件
		// 检查文件数量限制
		fileCount++
		if options.MaxFiles > 0 && fileCount > options.MaxFiles {
			return fmt.Errorf("文件数量超过限制: %d", options.MaxFiles)
		}

		// 检查文件大小限制
		if options.MaxFileSize > 0 && srcInfo.Size() > options.MaxFileSize {
			return fmt.Errorf("文件 %s 大小 %d 超过限制 %d", src, srcInfo.Size(), options.MaxFileSize)
		}

		header, err := tar.FileInfoHeader(srcInfo, "")
		if err != nil {
			return fmt.Errorf("创建tar头信息失败 %s: %w", src, err)
		}
		header.Name = filepath.Base(src)

		if err := tarWriter.WriteHeader(header); err != nil {
			return fmt.Errorf("写入tar头信息失败 %s: %w", src, err)
		}

		file, err := os.Open(src)
		if err != nil {
			return fmt.Errorf("打开文件失败 %s: %w", src, err)
		}
		defer file.Close()

		if _, err := safeCopy(options.Context, tarWriter, file, options.MaxFileSize, options.BufferSize); err != nil {
			return fmt.Errorf("复制文件内容失败 %s: %w", src, err)
		}

		return nil
	}
}

// UntarGz 解压 tar.gz 文件到指定目录
func UntarGz(src, dst string, opts ...ArchiveOption) error {
	options := applyOptions(opts...)

	// 检查上下文
	if err := errors.CheckContext(options.Context); err != nil {
		return err
	}

	// 打开源文件
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("打开源文件 %s 失败: %w", src, err)
	}
	defer srcFile.Close()

	// 创建 gzip reader
	gzReader, err := gzip.NewReader(srcFile)
	if err != nil {
		return fmt.Errorf("创建gzip reader失败: %w", err)
	}
	defer func() {
		if closeErr := gzReader.Close(); closeErr != nil {
			// 记录关闭错误
		}
	}()

	// 创建 tar reader
	tarReader := tar.NewReader(gzReader)

	// 确保目标目录存在
	if err := os.MkdirAll(dst, 0755); err != nil {
		return fmt.Errorf("创建目标目录 %s 失败: %w", dst, err)
	}

	fileCount := 0

	// 遍历 tar 中的每个文件
	for {
		// 检查上下文
		if err := errors.CheckContext(options.Context); err != nil {
			return err
		}

		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("读取tar header失败: %w", err)
		}

		// 检查文件数量限制
		fileCount++
		if options.MaxFiles > 0 && fileCount > options.MaxFiles {
			return fmt.Errorf("文件数量超过限制: %d", options.MaxFiles)
		}

		// 构造目标文件路径
		target := filepath.Join(dst, header.Name)

		// 防止路径遍历攻击
		if !isPathSafe(target, dst) {
			return fmt.Errorf("非法路径: %s", target)
		}

		// 处理不同类型的文件
		switch header.Typeflag {
		case tar.TypeDir:
			// 创建目录
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("创建目录 %s 失败: %w", target, err)
			}
		case tar.TypeReg:
			// 检查文件大小限制
			if options.MaxFileSize > 0 && header.Size > options.MaxFileSize {
				return fmt.Errorf("文件 %s 大小 %d 超过限制 %d", header.Name, header.Size, options.MaxFileSize)
			}

			// 确保父目录存在
			parentDir := filepath.Dir(target)
			if err := os.MkdirAll(parentDir, 0755); err != nil {
				return fmt.Errorf("创建父目录 %s 失败: %w", parentDir, err)
			}

			// 创建目标文件
			file, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("创建目标文件 %s 失败: %w", target, err)
			}

			func() {
				defer file.Close()

				// 复制文件内容
				if _, err := safeCopy(options.Context, file, tarReader, options.MaxFileSize, options.BufferSize); err != nil {
					panic(fmt.Errorf("复制文件内容失败 %s: %w", header.Name, err))
				}
			}()
		case tar.TypeSymlink:
			// 检查符号链接安全性
			if filepath.IsAbs(header.Linkname) {
				return fmt.Errorf("拒绝绝对路径符号链接: %s -> %s", header.Name, header.Linkname)
			}

			linkTarget := filepath.Join(filepath.Dir(target), header.Linkname)
			if !isPathSafe(linkTarget, dst) {
				return fmt.Errorf("符号链接指向目录外: %s -> %s", header.Name, header.Linkname)
			}

			// 创建符号链接
			if err := os.Symlink(header.Linkname, target); err != nil {
				return fmt.Errorf("创建符号链接 %s -> %s 失败: %w", target, header.Linkname, err)
			}
		default:
			// 忽略不支持的文件类型
		}
	}

	return nil
}

// TarGzWithTimeout 带超时的TarGz操作
func TarGzWithTimeout(src, dst string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return TarGz(src, dst, WithContext(ctx))
}

// ValidateSingleDirTarGz 校验 tar.gz 文件是否只包含一个顶层目录，并返回该目录名称
func ValidateSingleDirTarGz(src string, opts ...ArchiveOption) (string, error) {
	options := applyOptions(opts...)

	// 检查上下文
	if err := errors.CheckContext(options.Context); err != nil {
		return "", err
	}

	// 打开源文件
	srcFile, err := os.Open(src)
	if err != nil {
		return "", fmt.Errorf("打开源文件 %s 失败: %w", src, err)
	}
	defer srcFile.Close()

	// 创建 gzip reader
	gzReader, err := gzip.NewReader(srcFile)
	if err != nil {
		return "", fmt.Errorf("创建gzip reader失败: %w", err)
	}
	defer func() {
		if closeErr := gzReader.Close(); closeErr != nil {
			// 记录关闭错误
		}
	}()

	// 创建 tar reader
	tarReader := tar.NewReader(gzReader)

	// 用于存储所有顶层条目
	topLevelEntries := make(map[string]bool)
	var firstDirName string

	// 遍历 tar 中的每个文件
	for {
		// 检查上下文
		if err := errors.CheckContext(options.Context); err != nil {
			return "", err
		}

		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("读取tar header失败: %w", err)
		}

		// 获取条目名称
		name := header.Name

		// 清理路径名称
		name = filepath.Clean(name)
		name = strings.TrimPrefix(name, "./")

		// 提取顶层目录名
		var topLevelName string
		if strings.Contains(name, "/") {
			parts := strings.Split(name, "/")
			topLevelName = parts[0]
		} else {
			topLevelName = name
		}

		// 如果是空名称，跳过
		if topLevelName == "" {
			continue
		}

		// 记录顶层条目
		topLevelEntries[topLevelName] = true

		// 记录第一个目录名称
		if firstDirName == "" && header.Typeflag == tar.TypeDir {
			firstDirName = topLevelName
		}

		// 如果已经发现多个顶层条目，可以直接返回错误
		if len(topLevelEntries) > 1 {
			// 获取所有键名
			keys := make([]string, 0, len(topLevelEntries))
			for k := range topLevelEntries {
				keys = append(keys, k)
			}
			return "", fmt.Errorf("压缩文件包含多个顶层条目: %v", keys)
		}
	}

	// 检查结果
	if len(topLevelEntries) == 0 {
		return "", fmt.Errorf("压缩文件为空")
	}

	if len(topLevelEntries) > 1 {
		// 获取所有键名
		keys := make([]string, 0, len(topLevelEntries))
		for k := range topLevelEntries {
			keys = append(keys, k)
		}
		return "", fmt.Errorf("压缩文件包含多个顶层条目: %v", keys)
	}

	// 检查唯一的条目是否是目录
	if firstDirName == "" {
		return "", fmt.Errorf("压缩文件的唯一条目不是目录")
	}

	return firstDirName, nil
}
