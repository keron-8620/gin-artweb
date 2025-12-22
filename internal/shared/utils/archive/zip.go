package archive

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gin-artweb/internal/shared/errors"
)

// Zip 压缩文件或目录为ZIP格式
func Zip(src, dst string, opts ...ArchiveOption) error {
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
			panic(fmt.Errorf("关闭文件%s失败: %v", dst, closeErr))
		}
	}()

	// 创建zip writer
	zipWriter := zip.NewWriter(dstFile)
	defer func() {
		if closeErr := zipWriter.Close(); closeErr != nil {
			panic(fmt.Errorf("关闭ZIP写入器失败: %w", closeErr))
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

			// 创建header
			header, err := zip.FileInfoHeader(info)
			if err != nil {
				return fmt.Errorf("创建文件头信息失败 %s: %w", path, err)
			}

			// 设置文件名
			header.Name = filepath.ToSlash(relPath)

			// 如果是目录，确保以 / 结尾
			if info.IsDir() {
				header.Name += "/"
			}

			// 写入header
			writer, err := zipWriter.CreateHeader(header)
			if err != nil {
				return fmt.Errorf("创建ZIP条目失败 %s: %w", path, err)
			}

			// 如果是普通文件，写入内容
			if !info.IsDir() && info.Mode().IsRegular() {
				// 检查文件大小限制
				if options.MaxFileSize > 0 && info.Size() > options.MaxFileSize {
					return fmt.Errorf("文件 %s 大小 %d 超过限制 %d", path, info.Size(), options.MaxFileSize)
				}

				file, err := os.Open(path)
				if err != nil {
					return fmt.Errorf("打开文件失败 %s: %w", path, err)
				}
				defer file.Close()

				if _, err := safeCopy(options.Context, writer, file, options.MaxFileSize, options.BufferSize); err != nil {
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

		header, err := zip.FileInfoHeader(srcInfo)
		if err != nil {
			return fmt.Errorf("创建文件头信息失败 %s: %w", src, err)
		}

		header.Name = filepath.Base(src)

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return fmt.Errorf("创建ZIP条目失败 %s: %w", src, err)
		}

		file, err := os.Open(src)
		if err != nil {
			return fmt.Errorf("打开文件失败 %s: %w", src, err)
		}
		defer file.Close()

		if _, err := safeCopy(options.Context, writer, file, options.MaxFileSize, options.BufferSize); err != nil {
			return fmt.Errorf("复制文件内容失败 %s: %w", src, err)
		}

		return nil
	}
}

// Unzip 解压ZIP文件
func Unzip(src, dst string, opts ...ArchiveOption) error {
	options := applyOptions(opts...)

	// 检查上下文
	if err := errors.CheckContext(options.Context); err != nil {
		return err
	}

	// 打开zip文件
	reader, err := zip.OpenReader(src)
	if err != nil {
		return fmt.Errorf("打开ZIP文件 %s 失败: %w", src, err)
	}
	defer reader.Close()

	// 确保目标目录存在
	if err := os.MkdirAll(dst, 0755); err != nil {
		return fmt.Errorf("创建目标目录 %s 失败: %w", dst, err)
	}

	fileCount := 0

	// 遍历zip中的文件
	for _, file := range reader.File {
		// 检查上下文
		if err := errors.CheckContext(options.Context); err != nil {
			return err
		}

		// 检查文件数量限制
		fileCount++
		if options.MaxFiles > 0 && fileCount > options.MaxFiles {
			return fmt.Errorf("文件数量超过限制: %d", options.MaxFiles)
		}

		// 构造目标路径
		target := filepath.Join(dst, filepath.FromSlash(file.Name))

		// 检查路径遍历漏洞
		if !isPathSafe(target, dst) {
			return fmt.Errorf("非法路径: %s", target)
		}

		// 处理目录
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(target, file.Mode()); err != nil {
				return fmt.Errorf("创建目录 %s 失败: %w", target, err)
			}
		} else {
			// 检查文件大小限制
			if options.MaxFileSize > 0 && file.FileInfo().Size() > options.MaxFileSize {
				return fmt.Errorf("文件 %s 大小 %d 超过限制 %d", file.Name, file.FileInfo().Size(), options.MaxFileSize)
			}

			// 确保父目录存在
			parentDir := filepath.Dir(target)
			if err := os.MkdirAll(parentDir, 0755); err != nil {
				return fmt.Errorf("创建父目录 %s 失败: %w", parentDir, err)
			}

			// 创建目标文件
			targetFile, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR|os.O_TRUNC, file.Mode())
			if err != nil {
				return fmt.Errorf("创建目标文件 %s 失败: %w", target, err)
			}

			// 确保文件被关闭
			func() {
				defer targetFile.Close()

				// 打开源文件
				srcFile, err := file.Open()
				if err != nil {
					panic(fmt.Errorf("打开ZIP内文件 %s 失败: %w", file.Name, err))
				}
				defer srcFile.Close()

				// 复制内容
				if _, err := safeCopy(options.Context, targetFile, srcFile, options.MaxFileSize, options.BufferSize); err != nil {
					panic(fmt.Errorf("复制文件内容失败 %s: %w", file.Name, err))
				}
			}()
		}
	}

	return nil
}

// ValidateSingleDirZip 校验 ZIP 文件是否只包含一个顶层目录，并返回该目录名称
func ValidateSingleDirZip(src string, opts ...ArchiveOption) (string, error) {
	options := applyOptions(opts...)

	// 检查上下文
	if err := errors.CheckContext(options.Context); err != nil {
		return "", err
	}

	// 打开zip文件
	reader, err := zip.OpenReader(src)
	if err != nil {
		return "", fmt.Errorf("打开ZIP文件 %s 失败: %w", src, err)
	}
	defer reader.Close()

	// 用于存储所有顶层条目
	topLevelEntries := make(map[string]bool)
	var firstDirName string

	// 遍历zip中的文件
	for _, file := range reader.File {
		// 检查上下文
		if err := errors.CheckContext(options.Context); err != nil {
			return "", err
		}

		// 获取条目名称
		name := file.Name

		// 清理路径名称
		name = filepath.Clean(name)
		name = strings.TrimPrefix(name, "./")
		name = strings.TrimSuffix(name, "/")

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
		if firstDirName == "" && file.FileInfo().IsDir() {
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
