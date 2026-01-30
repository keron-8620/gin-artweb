package archive

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"

	"emperror.dev/errors"
)

// TarGz 将指定路径的文件或目录压缩为 tar.gz 格式
// 支持文件大小限制、文件数量限制、上下文取消等安全特性
func TarGz(src, dst string, opts ...ArchiveOption) error {
	options := applyOptions(opts...)

	// 检查上下文状态
	if err := checkContext(options.Context); err != nil {
		return errors.WrapIf(err, "TarGz压缩:上下文检查失败")
	}

	// 创建目标文件
	dstFile, err := os.Create(dst)
	if err != nil {
		return errors.WrapIfWithDetails(
			err, "TarGz压缩:创建目标文件失败",
			"src", src,
			"dst", dst,
		)
	}
	defer dstFile.Close()

	// 创建 gzip writer
	gzWriter := gzip.NewWriter(dstFile)
	defer gzWriter.Close()

	// 创建 tar writer
	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	// 获取源文件信息
	srcInfo, err := os.Stat(src)
	if err != nil {
		return errors.WrapIfWithDetails(
			err, "TarGz压缩:获取源文件信息失败",
			"src", src,
		)
	}

	fileCount := 0

	// 处理目录
	if srcInfo.IsDir() {
		return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return errors.WrapIfWithDetails(
					err, "TarGz压缩:遍历目录失败",
					"dir", src,
					"path", path,
				)
			}

			// 检查上下文状态
			if err := checkContext(options.Context); err != nil {
				return errors.WrapIf(err, "TarGz压缩:上下文检查失败")
			}

			// 跳过目录本身
			if path == src {
				return nil
			}

			// 检查文件数量限制
			fileCount++
			if options.MaxFiles > 0 && fileCount > options.MaxFiles {
				return errors.NewWithDetails(
					"TarGz压缩失败:文件数量超过限制",
					"max_files", options.MaxFiles,
					"current_count", fileCount,
				)
			}

			// 检查文件大小限制
			if options.MaxFileSize > 0 && info.Size() > options.MaxFileSize {
				return errors.NewWithDetails(
					"TarGz压缩失败:文件大小超过限制",
					"file", path,
					"size_bytes", info.Size(),
					"max_size_bytes", options.MaxFileSize,
				)
			}

			// 创建 tar 头信息
			header, err := tar.FileInfoHeader(info, "")
			if err != nil {
				return errors.WrapIfWithDetails(
					err, "TarGz压缩:创建tar头信息失败",
					"src", src,
					"path", path,
				)
			}

			// 设置相对路径
			relPath, err := filepath.Rel(src, path)
			if err != nil {
				return errors.WrapIfWithDetails(
					err, "TarGz压缩:计算相对路径失败",
					"src", src,
					"path", path,
				)
			}
			header.Name = relPath

			// 写入 tar 头信息
			if err := tarWriter.WriteHeader(header); err != nil {
				return errors.WrapIfWithDetails(
					err, "TarGz压缩:写入tar头信息失败",
					"src", src,
					"path", path,
				)
			}

			// 如果是普通文件，复制内容
			if info.Mode().IsRegular() {
				file, err := os.Open(path)
				if err != nil {
					return errors.WrapIfWithDetails(
						err, "TarGz压缩:打开文件失败",
						"src", src,
						"path", path,
					)
				}
				defer file.Close()

				if _, err := safeCopy(options.Context, tarWriter, file, options.MaxFileSize, options.BufferSize); err != nil {
					return errors.WrapIfWithDetails(
						err, "TarGz压缩:复制文件内容失败",
						"src", src,
						"path", path,
					)
				}
			}

			return nil
		})
	} else {
		// 处理单个文件
		// 检查文件数量限制
		fileCount++
		if options.MaxFiles > 0 && fileCount > options.MaxFiles {
			return errors.NewWithDetails(
				"TarGz压缩失败:文件数量超过限制",
				"max_files", options.MaxFiles,
				"current_count", fileCount,
			)
		}

		// 检查文件大小限制
		if options.MaxFileSize > 0 && srcInfo.Size() > options.MaxFileSize {
			return errors.NewWithDetails(
				"TarGz压缩失败:文件大小超过限制",
				"file", src,
				"size_bytes", srcInfo.Size(),
				"max_size_bytes", options.MaxFileSize,
			)
		}

		header, err := tar.FileInfoHeader(srcInfo, "")
		if err != nil {
			return errors.WrapIfWithDetails(
				err, "TarGz压缩:创建tar头信息失败",
				"src", src,
			)
		}
		header.Name = filepath.Base(src)

		if err := tarWriter.WriteHeader(header); err != nil {
			return errors.WrapIfWithDetails(
				err, "TarGz压缩:写入tar头信息失败",
				"src", src,
			)
		}

		file, err := os.Open(src)
		if err != nil {
			return errors.WrapIfWithDetails(
				err, "TarGz压缩:打开文件失败",
				"src", src,
			)
		}
		defer file.Close()

		if _, err := safeCopy(options.Context, tarWriter, file, options.MaxFileSize, options.BufferSize); err != nil {
			return errors.WrapIfWithDetails(
				err, "TarGz压缩:复制文件内容失败",
				"src", src,
			)
		}

		return nil
	}
}

// UntarGz 解压 tar.gz 文件到指定目录
// 支持路径安全检查、文件大小限制、符号链接安全处理等功能
func UntarGz(src, dst string, opts ...ArchiveOption) error {
	options := applyOptions(opts...)

	// 检查上下文状态
	if err := checkContext(options.Context); err != nil {
		return errors.WrapIf(err, "UntarGz解压:上下文检查失败")
	}

	// 打开源文件
	srcFile, err := os.Open(src)
	if err != nil {
		return errors.WrapIfWithDetails(
			err, "UntarGz解压:打开源文件失败",
			"src", src,
			"dst", dst,
		)
	}
	defer srcFile.Close()

	// 创建 gzip reader
	gzReader, err := gzip.NewReader(srcFile)
	if err != nil {
		return errors.WrapIfWithDetails(
			err, "UntarGz解压:创建gzip reader失败",
			"src", src,
		)
	}
	defer gzReader.Close()

	// 创建 tar reader
	tarReader := tar.NewReader(gzReader)

	// 确保目标目录存在
	if err := os.MkdirAll(dst, 0755); err != nil {
		return errors.WrapIfWithDetails(
			err, "UntarGz解压:创建目标目录失败",
			"src", src,
			"dst", dst,
		)
	}

	fileCount := 0

	// 遍历 tar 中的每个文件
	for {
		// 检查上下文状态
		if err := checkContext(options.Context); err != nil {
			return errors.WrapIf(err, "UntarGz解压:上下文检查失败")
		}

		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return errors.WrapIfWithDetails(
				err, "UntarGz解压:读取tar header失败",
				"src", src,
			)
		}

		// 检查文件数量限制
		fileCount++
		if options.MaxFiles > 0 && fileCount > options.MaxFiles {
			return errors.NewWithDetails(
				"UntarGz解压失败:文件数量超过限制",
				"max_files", options.MaxFiles,
				"current_count", fileCount,
			)
		}

		// 构造目标文件路径
		target := filepath.Join(dst, header.Name)

		// 防止路径遍历攻击
		if !isPathSafe(target, dst) {
			return errors.NewWithDetails(
				"UntarGz解压失败:检测到非法路径",
				"target", target,
				"base", dst,
			)
		}

		// 处理不同类型的文件
		switch header.Typeflag {
		case tar.TypeDir:
			// 创建目录
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return errors.WrapIfWithDetails(
					err, "UntarGz解压:创建目录失败",
					"target", target,
					"mode", header.Mode,
				)
			}
		case tar.TypeReg:
			// 检查文件大小限制
			if options.MaxFileSize > 0 && header.Size > options.MaxFileSize {
				return errors.NewWithDetails(
					"UntarGz解压失败:文件大小超过限制",
					"file", header.Name,
					"size_bytes", header.Size,
					"max_size_bytes", options.MaxFileSize,
				)
			}

			// 确保父目录存在
			parentDir := filepath.Dir(target)
			if err := os.MkdirAll(parentDir, 0755); err != nil {
				return errors.WrapIfWithDetails(
					err, "UntarGz解压:创建父目录失败",
					"parent_dir", parentDir,
					"target", target,
				)
			}

			// 创建目标文件
			file, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return errors.WrapIfWithDetails(
					err, "UntarGz解压:创建目标文件失败",
					"target", target,
					"mode", header.Mode,
				)
			}
			defer func() {
				if closeErr := file.Close(); closeErr != nil {
					// 记录关闭错误但不中断主流程
				}
			}()

			// 复制目标文件内容
			if _, err := safeCopy(options.Context, file, tarReader, options.MaxFileSize, options.BufferSize); err != nil {
				return errors.WrapIfWithDetails(
					err, "UntarGz解压:复制文件内容失败",
					"file", header.Name,
					"target", target,
				)
			}

		case tar.TypeSymlink:
			// 检查符号链接安全性
			if filepath.IsAbs(header.Linkname) {
				return errors.NewWithDetails(
					"UntarGz解压失败:拒绝绝对路径符号链接",
					"name", header.Name,
					"linkname", header.Linkname,
				)
			}

			linkTarget := filepath.Join(filepath.Dir(target), header.Linkname)
			if !isPathSafe(linkTarget, dst) {
				return errors.NewWithDetails(
					"UntarGz解压失败:符号链接指向目录外",
					"name", header.Name,
					"linkname", header.Linkname,
					"target", linkTarget,
				)
			}

			// 检查是否允许跟随符号链接
			if !options.FollowSymlinks {
				// 创建符号链接
				if err := os.Symlink(header.Linkname, target); err != nil {
					return errors.WrapIfWithDetails(
						err, "UntarGz解压:创建符号链接失败",
						"target", target,
						"linkname", header.Linkname,
					)
				}
			} else {
				// 如果跟随符号链接，需要额外的安全检查
				// 这里我们只是简单地拒绝创建符号链接，因为跟随符号链接可能存在安全风险
				return errors.NewWithDetails(
					"UntarGz解压失败:不允许跟随符号链接",
					"name", header.Name,
					"linkname", header.Linkname,
				)
			}
		default:
			// 忽略不支持的文件类型
		}
	}

	return nil
}

// ValidateSingleDirTarGz 校验 tar.gz 文件是否只包含一个顶层目录，并返回该目录名称
// 用于确保压缩包结构符合预期，防止解压后产生混乱的文件结构
func ValidateSingleDirTarGz(src string, opts ...ArchiveOption) (string, error) {
	options := applyOptions(opts...)

	// 检查上下文状态
	if err := checkContext(options.Context); err != nil {
		return "", errors.WrapIf(err, "ValidateSingleDirTarGz校验:上下文检查失败")
	}

	// 打开源文件
	srcFile, err := os.Open(src)
	if err != nil {
		return "", errors.WrapIfWithDetails(
			err, "ValidateSingleDirTarGz校验:打开源文件失败",
			"src", src,
		)
	}
	defer func() {
		if closeErr := srcFile.Close(); closeErr != nil {
			// 记录关闭错误但不中断主流程
		}
	}()

	// 创建 gzip reader
	gzReader, err := gzip.NewReader(srcFile)
	if err != nil {
		return "", errors.WrapIfWithDetails(
			err, "ValidateSingleDirTarGz校验:创建gzip reader失败",
			"src", src,
		)
	}
	defer func() {
		if closeErr := gzReader.Close(); closeErr != nil {
			// 记录关闭错误但不中断主流程
		}
	}()

	// 创建 tar reader
	tarReader := tar.NewReader(gzReader)

	// 用于存储所有顶层条目
	topLevelEntries := make(map[string]bool)
	var firstDirName string

	// 遍历 tar 中的每个文件
	for {
		// 检查上下文状态
		if err := checkContext(options.Context); err != nil {
			return "", errors.WrapIf(err, "ValidateSingleDirTarGz校验:上下文检查失败")
		}

		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", errors.WrapIfWithDetails(
				err, "ValidateSingleDirTarGz校验:读取tar header失败",
				"src", src,
			)
		}

		// 获取顶层目录名
		relPath := header.Name
		if filepath.IsAbs(relPath) {
			relPath = relPath[1:] // 去掉前导斜杠
		}

		// 分割路径获取第一级目录
		parts := strings.Split(strings.Trim(relPath, "/"), "/")
		if len(parts) > 0 && parts[0] != "" {
			topLevelEntries[parts[0]] = true
			if firstDirName == "" {
				firstDirName = parts[0]
			}
		}
	}

	// 检查是否只有一个顶层目录
	if len(topLevelEntries) != 1 {
		keys := make([]string, 0, len(topLevelEntries))
		for k := range topLevelEntries {
			keys = append(keys, k)
		}
		return "", errors.NewWithDetails(
			"ValidateSingleDirTarGz校验失败:tar.gz文件不包含恰好一个顶层目录",
			"top_level_entries", keys,
			"count", len(topLevelEntries),
		)
	}

	return firstDirName, nil
}
