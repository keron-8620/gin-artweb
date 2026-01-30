package archive

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"

	"emperror.dev/errors"
)

// Zip 将文件或目录压缩为ZIP格式
// 支持目录递归压缩和单文件压缩，提供文件数量和大小限制功能
func Zip(src, dst string, opts ...ArchiveOption) error {
	options := applyOptions(opts...)

	// 检查操作上下文状态
	if err := checkContext(options.Context); err != nil {
		return errors.WrapIf(err, "Zip压缩:上下文检查失败")
	}

	// 验证输入参数
	if src == "" {
		return errors.New("Zip压缩:源路径不能为空")
	}
	if dst == "" {
		return errors.New("Zip压缩:目标路径不能为空")
	}

	// 创建目标ZIP文件
	dstFile, err := os.Create(dst)
	if err != nil {
		return errors.WrapIfWithDetails(
			err, "Zip压缩:创建目标ZIP文件失败",
			"dst", dst,
		)
	}
	defer dstFile.Close()

	// 创建ZIP写入器
	zipWriter := zip.NewWriter(dstFile)
	defer zipWriter.Close()

	// 获取源文件/目录信息
	srcInfo, err := os.Stat(src)
	if err != nil {
		return errors.WrapIfWithDetails(
			err, "Zip压缩:获取源路径信息失败",
			"src", src,
		)
	}

	// 初始化文件计数器
	fileCount := 0

	// 根据源路径类型分别处理
	if srcInfo.IsDir() {
		return zipDirectory(src, zipWriter, &fileCount, options)
	}
	return zipSingleFile(src, srcInfo, zipWriter, &fileCount, options)
}

// zipDirectory 递归压缩目录
func zipDirectory(src string, zipWriter *zip.Writer, fileCount *int, options ArchiveOptions) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return errors.WrapIfWithDetails(err, "Zip压缩:遍历目录失败", "path", path)
		}

		// 检查操作上下文状态
		if err := checkContext(options.Context); err != nil {
			return errors.WrapIf(err, "Zip压缩:上下文检查失败")
		}

		// 计算相对于源目录的路径
		relPath, err := filepath.Rel(filepath.Dir(src), path)
		if err != nil {
			return errors.WrapIfWithDetails(
				err, "Zip压缩:计算相对路径失败",
				"base_dir", filepath.Dir(src),
				"target_path", path,
			)
		}

		// 跳过根目录本身
		if relPath == "." {
			return nil
		}

		// 更新文件计数并检查限制
		*fileCount++
		if options.MaxFiles > 0 && *fileCount > options.MaxFiles {
			return errors.NewWithDetails(
				"Zip压缩:文件数量超过限制",
				"max_files", options.MaxFiles,
				"current_count", *fileCount,
			)
		}

		// 创建ZIP文件头信息
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return errors.WrapIfWithDetails(
				err, "Zip压缩:创建ZIP文件头失败",
				"file_path", path,
			)
		}

		// 设置ZIP中的文件路径（使用正斜杠）
		header.Name = filepath.ToSlash(relPath)

		// 目录项需要以斜杠结尾
		if info.IsDir() {
			header.Name += "/"
		}

		// 写入文件头信息
		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return errors.WrapIfWithDetails(
				err, "Zip压缩:写入ZIP文件头失败",
				"file_path", path,
			)
		}

		// 只处理普通文件的内容
		if !info.IsDir() && info.Mode().IsRegular() {
			// 检查单个文件大小限制
			if options.MaxFileSize > 0 && info.Size() > options.MaxFileSize {
				return errors.NewWithDetails(
					"Zip压缩:文件大小超过限制",
					"file_size", info.Size(),
					"max_size", options.MaxFileSize,
					"file_path", path,
				)
			}

			// 打开源文件
			file, err := os.Open(path)
			if err != nil {
				return errors.WrapIfWithDetails(
					err, "Zip压缩:打开源文件失败",
					"file_path", path,
				)
			}
			defer file.Close()

			// 安全复制文件内容
			if _, err := safeCopy(options.Context, writer, file, options.MaxFileSize, options.BufferSize); err != nil {
				return errors.WrapIfWithDetails(
					err, "Zip压缩:复制文件内容失败",
					"file_path", path,
				)
			}
		}

		return nil
	})
}

// zipSingleFile 压缩单个文件
func zipSingleFile(src string, srcInfo os.FileInfo, zipWriter *zip.Writer, fileCount *int, options ArchiveOptions) error {
	// 更新文件计数并检查限制
	*fileCount++
	if options.MaxFiles > 0 && *fileCount > options.MaxFiles {
		return errors.NewWithDetails(
			"Zip压缩:文件数量超过限制",
			"max_files", options.MaxFiles,
			"current_count", *fileCount,
		)
	}

	// 检查文件大小限制
	if options.MaxFileSize > 0 && srcInfo.Size() > options.MaxFileSize {
		return errors.NewWithDetails(
			"Zip压缩:文件大小超过限制",
			"file_size", srcInfo.Size(),
			"max_size", options.MaxFileSize,
			"file_path", src,
		)
	}

	// 创建ZIP文件头信息
	header, err := zip.FileInfoHeader(srcInfo)
	if err != nil {
		return errors.WrapIfWithDetails(
			err, "Zip压缩:创建ZIP文件头失败",
			"src_file", src,
		)
	}

	// 设置ZIP中的文件名（仅文件名，不含路径）
	header.Name = filepath.Base(src)

	// 创建ZIP条目
	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return errors.WrapIfWithDetails(
			err, "Zip压缩:创建ZIP条目失败",
			"src_file", src,
		)
	}

	// 打开源文件
	file, err := os.Open(src)
	if err != nil {
		return errors.WrapIfWithDetails(
			err, "Zip压缩:打开源文件失败",
			"src_file", src,
		)
	}
	defer file.Close()

	// 安全复制文件内容
	if _, err := safeCopy(options.Context, writer, file, options.MaxFileSize, options.BufferSize); err != nil {
		return errors.WrapIfWithDetails(
			err, "Zip压缩:复制文件内容失败",
			"src_file", src,
		)
	}

	return nil
}

// Unzip 解压ZIP文件到指定目录
// 支持路径安全检查、文件数量和大小限制，防止路径遍历攻击
func Unzip(src, dst string, opts ...ArchiveOption) error {
	options := applyOptions(opts...)

	// 检查操作上下文状态
	if err := checkContext(options.Context); err != nil {
		return errors.WrapIf(err, "ZIP解压:上下文检查失败")
	}

	// 验证输入参数
	if src == "" {
		return errors.New("ZIP解压:源ZIP文件路径不能为空")
	}
	if dst == "" {
		return errors.New("ZIP解压:目标解压目录路径不能为空")
	}

	// 打开ZIP文件进行读取
	reader, err := zip.OpenReader(src)
	if err != nil {
		return errors.WrapIfWithDetails(
			err, "ZIP解压:打开ZIP文件失败",
			"src_file", src,
		)
	}
	defer reader.Close()
	// 确保目标解压目录存在
	if err := os.MkdirAll(dst, 0755); err != nil {
		return errors.WrapIfWithDetails(
			err, "ZIP解压:创建目标目录失败",
			"target_dir", dst,
		)
	}

	// 初始化文件计数器
	fileCount := 0

	// 遍历ZIP文件中的所有条目
	for _, file := range reader.File {
		// 检查操作上下文状态
		if err := checkContext(options.Context); err != nil {
			return errors.WrapIf(err, "ZIP解压:上下文检查失败")
		}

		// 更新文件计数并检查数量限制
		fileCount++
		if options.MaxFiles > 0 && fileCount > options.MaxFiles {
			return errors.NewWithDetails(
				"ZIP解压:文件数量超过限制",
				"max_files", options.MaxFiles,
				"current_count", fileCount,
			)
		}

		// 构造解压后的目标文件路径
		target := filepath.Join(dst, filepath.FromSlash(file.Name))

		// 安全检查:防止路径遍历攻击
		if !isPathSafe(target, dst) {
			return errors.NewWithDetails(
				"ZIP解压:检测到路径遍历攻击风险",
				"unsafe_target", target,
				"base_dir", dst,
			)
		}

		// 根据文件类型分别处理
		if file.FileInfo().IsDir() {
			// 处理目录条目
			if err := os.MkdirAll(target, file.Mode()); err != nil {
				return errors.WrapIfWithDetails(
					err, "ZIP解压:创建目录失败",
					"target_dir", target,
				)
			}
		} else {
			// 处理文件条目
			if err := unzipFile(file, target, options); err != nil {
				return errors.WrapIfWithDetails(
					err, "ZIP解压:处理文件失败",
					"target_file", target,
				)
			}
		}
	}

	return nil
}

// unzipFile 解压单个ZIP文件条目
func unzipFile(zipFile *zip.File, target string, options ArchiveOptions) error {
	// 检查文件大小限制
	if options.MaxFileSize > 0 && zipFile.FileInfo().Size() > options.MaxFileSize {
		return errors.NewWithDetails(
			"ZIP解压:文件大小超过限制",
			"file_size", zipFile.FileInfo().Size(),
			"max_size", options.MaxFileSize,
			"target_file", target,
		)
	}

	// 确保目标文件的父目录存在
	parentDir := filepath.Dir(target)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return errors.WrapIfWithDetails(
			err, "ZIP解压:创建父目录失败",
			"parent_dir", parentDir,
		)
	}

	// 创建目标文件
	targetFile, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR|os.O_TRUNC, zipFile.Mode())
	if err != nil {
		return errors.WrapIfWithDetails(
			err, "ZIP解压:创建目标文件失败",
			"target_file", target,
		)
	}
	defer targetFile.Close()

	// 打开ZIP中的源文件
	srcFile, err := zipFile.Open()
	if err != nil {
		return errors.WrapIfWithDetails(
			err, "ZIP解压:打开ZIP内文件失败",
			"zip_entry", zipFile.Name,
		)
	}
	defer srcFile.Close()

	// 安全复制文件内容
	if _, err := safeCopy(options.Context, targetFile, srcFile, options.MaxFileSize, options.BufferSize); err != nil {
		return errors.WrapIfWithDetails(
			err, "ZIP解压:复制文件内容失败",
			"target_file", target,
		)
	}

	return nil
}

// ValidateSingleDirZip 校验 ZIP 文件是否只包含一个顶层目录，并返回该目录名称
func ValidateSingleDirZip(src string, opts ...ArchiveOption) (string, error) {
	options := applyOptions(opts...)

	// 检查上下文
	if err := checkContext(options.Context); err != nil {
		return "", errors.WrapIf(err, "上下文检查失败")
	}

	// 打开zip文件
	reader, err := zip.OpenReader(src)
	if err != nil {
		return "", errors.WrapIfWithDetails(
			err, "打开ZIP文件失败",
			"src", src,
		)
	}
	defer reader.Close()

	// 用于存储所有顶层条目
	topLevelEntries := make(map[string]bool)
	var firstDirName string

	// 遍历zip中的文件
	for i, file := range reader.File {
		// 每100个文件检查一次上下文，避免过于频繁的检查
		if i%100 == 0 {
			if err := checkContext(options.Context); err != nil {
				return "", errors.WrapIf(err, "上下文检查失败")
			}
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
			return createMultipleEntriesError(topLevelEntries)
		}
	}

	// 检查结果
	if len(topLevelEntries) == 0 {
		return "", errors.New("压缩文件为空")
	}

	if len(topLevelEntries) > 1 {
		return createMultipleEntriesError(topLevelEntries)
	}

	// 检查唯一的条目是否是目录
	if firstDirName == "" {
		return "", errors.NewWithDetails(
			"压缩文件的唯一条目不是目录",
			"first_dir_name", firstDirName,
		)
	}

	return firstDirName, nil
}

// createMultipleEntriesError 创建多顶层条目错误
func createMultipleEntriesError(entries map[string]bool) (string, error) {
	keys := make([]string, 0, len(entries))
	for k := range entries {
		keys = append(keys, k)
	}
	return "", errors.NewWithDetails(
		"压缩文件包含多个顶层条目",
		"top_level_entries", keys,
	)
}
