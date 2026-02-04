package archive

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

// TarGz 将指定路径的文件或目录压缩为 tar.gz 格式
// 优化点:解耦核心逻辑、统一错误处理、减少重复代码、增强安全性
func TarGz(src, dst string, opts ...ArchiveOption) (resultErr error) {
	options := applyOptions(opts...)

	// 前置检查
	if options.Context.Err() != nil {
		return errors.Wrap(options.Context.Err(), "tar.gz压缩:上下文检查失败")
	}
	if src == "" || dst == "" {
		return errors.New("源路径/目标路径不能为空")
	}

	// 路径安全检查，防止路径遍历攻击
	cleanSrc := filepath.Clean(src)
	cleanDst := filepath.Clean(dst)

	// 验证路径是否在允许范围内（基础安全检查）
	if !filepath.IsAbs(cleanSrc) {
		absSrc, err := filepath.Abs(cleanSrc)
		if err != nil {
			return errors.Wrapf(err, "获取源路径绝对路径失败, src=%s", cleanSrc)
		}
		cleanSrc = absSrc
	}

	// 打开/创建文件
	srcInfo, err := os.Stat(cleanSrc)
	if err != nil {
		return errors.Wrapf(err, "获取源文件信息失败, src=%s", cleanSrc)
	}

	// 创建目标文件前检查父目录是否存在
	dstDir := filepath.Dir(cleanDst)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return errors.Wrapf(err, "创建目标目录失败, dir=%s", dstDir)
	}

	dstFile, err := os.Create(cleanDst)
	if err != nil {
		return errors.Wrapf(err, "创建目标文件失败, dst=%s", cleanDst)
	}

	// 使用带缓冲的写入器提高性能
	bufferedWriter := bufio.NewWriterSize(dstFile, options.BufferSize)
	var closeErrors []error

	// 预先声明变量以便在defer中使用
	var gzWriter *gzip.Writer
	var tarWriter *tar.Writer

	// 资源清理函数
	defer func() {
		// 先关闭 tarWriter，确保所有 tar 条目都被正确写入和结束
		if tarWriter != nil {
			if closeErr := tarWriter.Close(); closeErr != nil {
				closeErrors = append(closeErrors, errors.Wrap(closeErr, "关闭tar写入器失败"))
			}
		}

		// 再关闭 gzipWriter，确保所有压缩数据都被写入
		if gzWriter != nil {
			if closeErr := gzWriter.Close(); closeErr != nil {
				closeErrors = append(closeErrors, errors.Wrap(closeErr, "关闭gzip写入器失败"))
			}
		}

		// 再刷新缓冲区
		if flushErr := bufferedWriter.Flush(); flushErr != nil {
			closeErrors = append(closeErrors, errors.Wrap(flushErr, "刷新缓冲区失败"))
		}

		// 最后关闭目标文件
		if closeErr := dstFile.Close(); closeErr != nil {
			closeErrors = append(closeErrors, errors.Wrap(closeErr, "关闭目标文件失败"))
		}

		// 如果有关闭错误且主操作成功，则返回第一个关闭错误
		if len(closeErrors) > 0 && resultErr == nil {
			resultErr = closeErrors[0]
		}
	}()

	// 初始化压缩写入器
	gzWriter, wErr := gzip.NewWriterLevel(bufferedWriter, options.CompressionLevel)
	if wErr != nil {
		return errors.Wrap(wErr, "创建gzip写入器失败")
	}
	tarWriter = tar.NewWriter(gzWriter)

	// 统一处理文件/目录
	fileCount := 0
	totalSize := int64(0)

	var processErr error

	if srcInfo.IsDir() {
		processErr = filepath.Walk(cleanSrc, func(filePath string, info os.FileInfo, walkErr error) error {
			if walkErr != nil {
				return errors.Wrapf(walkErr, "遍历目录失败, filepath=%s", filePath)
			}

			// 安全检查：确保文件路径在源目录内
			relPath, err := filepath.Rel(cleanSrc, filePath)
			if err != nil {
				return errors.Wrapf(err, "计算相对路径失败, base=%s, target=%s", cleanSrc, filePath)
			}
			if strings.HasPrefix(relPath, "..") {
				return errors.Errorf("路径超出源目录范围: %s", filePath)
			}

			// 检查是否应该排除
			exclude, err := options.ShouldExclude(filePath)
			if err != nil {
				return err
			}
			if exclude {
				return nil
			}

			// 检查是否应该包含
			include, err := options.ShouldInclude(filePath)
			if err != nil {
				return err
			}
			if !include {
				return nil
			}

			entryErr := processTarEntry(filePath, cleanSrc, info, tarWriter, &fileCount, &totalSize, options)
			if entryErr != nil {
				return errors.Wrapf(entryErr, "处理文件条目失败, filepath=%s", filePath)
			}
			return nil
		})
	} else {
		// 检查是否应该排除
		exclude, err := options.ShouldExclude(cleanSrc)
		if err != nil {
			return err
		}
		if exclude {
			return nil
		}

		// 检查是否应该包含
		include, err := options.ShouldInclude(cleanSrc)
		if err != nil {
			return err
		}
		if !include {
			return nil
		}

		// 处理单个文件
		parentDir := filepath.Dir(cleanSrc)
		processErr = processTarEntry(cleanSrc, parentDir, srcInfo, tarWriter, &fileCount, &totalSize, options)
	}

	if processErr != nil {
		return errors.Wrap(processErr, "tar.gz压缩失败")
	}

	return nil
}

// processTarEntry 处理单个tar条目（解耦核心逻辑）
func processTarEntry(filePath, baseDir string, info os.FileInfo, tarWriter *tar.Writer, fileCount *int, totalSize *int64, options ArchiveOptions) error {
	// 上下文检查
	if options.Context.Err() != nil {
		return errors.Wrap(options.Context.Err(), "处理单个tar条目:上下文检查失败")
	}

	// 跳过基础目录
	if filePath == baseDir {
		return nil
	}

	// 文件数量限制
	*fileCount++
	if options.MaxFiles > 0 && *fileCount > options.MaxFiles {
		return errors.Errorf("文件数量超过限制, max=%d, current= %d", options.MaxFiles, *fileCount)
	}

	// 文件大小限制
	if options.MaxFileSize > 0 && info.Size() > options.MaxFileSize {
		return errors.Errorf("文件大小超过限制, file_path=%s, max=%d, current=%d", filePath, options.MaxFileSize, info.Size())
	}

	// 创建tar头
	relPath, err := filepath.Rel(baseDir, filePath)
	if err != nil {
		return errors.Wrapf(err, "计算相对路径失败, filepath=%s", filePath)
	}

	// 对于符号链接，获取目标路径
	linkTarget := ""
	if info.Mode()&os.ModeSymlink != 0 {
		linkTarget, err = os.Readlink(filePath)
		if err != nil {
			return errors.Wrapf(err, "读取符号链接目标失败, filepath=%s", filePath)
		}
	}

	header, err := tar.FileInfoHeader(info, linkTarget)
	if err != nil {
		return errors.Wrapf(err, "创建tar头失败, filepath=%s", filePath)
	}
	header.Name = relPath

	// 写入tar头
	if err := tarWriter.WriteHeader(header); err != nil {
		return errors.Wrapf(err, "写入tar头失败, filepath=%s", filePath)
	}

	// 写入文件内容（仅普通文件）
	if info.Mode().IsRegular() {
		file, err := os.Open(filePath)
		if err != nil {
			return errors.Wrapf(err, "打开文件失败, filepath=%s", filePath)
		}
		defer closeWithError(file, "关闭文件失败")

		written, err := safeCopy(options.Context, tarWriter, file, options.MaxFileSize, options.BufferSize)
		if err != nil {
			return errors.Wrapf(err, "复制文件内容失败, filepath=%s", filePath)
		}

		*totalSize += written
	}

	return nil
}

// UntarGz 解压 tar.gz 文件到指定目录
// 优化点:解耦处理逻辑、强化资源释放、统一错误格式
func UntarGz(src, dst string, opts ...ArchiveOption) error {
	options := applyOptions(opts...)

	// 前置检查
	if options.Context.Err() != nil {
		return errors.Wrap(options.Context.Err(), "tar.gz解压:上下文检查失败")
	}
	if src == "" || dst == "" {
		return errors.New("源路径/目标路径不能为空")
	}

	// 打开源文件
	srcFile, err := os.Open(src)
	if err != nil {
		return errors.Wrapf(err, "打开源文件失败, src=%s", src)
	}
	defer closeWithError(srcFile, "关闭源文件失败")

	// 初始化解压读取器
	gzReader, err := gzip.NewReader(srcFile)
	if err != nil {
		return errors.Wrapf(err, "创建gzip读取器失败, src=%s", src)
	}
	defer closeWithError(gzReader, "关闭gzip读取器失败")

	tarReader := tar.NewReader(gzReader)

	// 创建目标目录
	if err := os.MkdirAll(dst, 0755); err != nil {
		return errors.Wrapf(err, "创建目标目录失败, dst=%s", dst)
	}

	// 遍历tar条目
	fileCount := 0
	totalSize := int64(0)

	for {
		if options.Context.Err() != nil {
			return errors.Wrap(options.Context.Err(), "tar.gz解压遍历文件:上下文检查失败")
		}

		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return errors.Wrap(err, "读取tar条目失败")
		}

		fileCount++
		if options.MaxFiles > 0 && fileCount > options.MaxFiles {
			return errors.Errorf("文件数量超过限制, max=%d, current= %d", options.MaxFiles, fileCount)
		}

		entrySize, err := processUntarEntry(header, tarReader, dst, options)
		if err != nil {
			return errors.Wrapf(err, "处理tar条目失败, entry=%s", header.Name)
		}

		totalSize += entrySize
	}

	return nil

}

// processUntarEntry 处理单个解压条目（解耦核心逻辑）
func processUntarEntry(header *tar.Header, tarReader *tar.Reader, dst string, options ArchiveOptions) (int64, error) {
	// 构造目标路径并检查安全性
	target := filepath.Join(dst, header.Name)
	if !isPathSafe(target, dst) {
		return 0, errors.Errorf("非法路径（路径遍历攻击）, target=%s, base=%s", target, dst)
	}

	// 按类型处理
	switch header.Typeflag {
	case tar.TypeDir:
		// 设置合适的目录权限
		dirMode := os.FileMode(header.Mode)
		if dirMode == 0 {
			dirMode = 0755
		}
		// 应用权限掩码
		dirMode = validatePermissions(dirMode, options.PermissionsMask)

		if err := os.MkdirAll(target, dirMode); err != nil {
			return 0, err
		}
		return 0, nil

	case tar.TypeReg:
		// 大小限制
		if options.MaxFileSize > 0 && header.Size > options.MaxFileSize {
			return 0, errors.Errorf("文件大小超过限制, file_path=%s, size=%d, max=%d", header.Name, header.Size, options.MaxFileSize)
		}

		// 创建父目录
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return 0, errors.Wrap(err, "创建父目录失败")
		}

		// 写入文件
		fileMode := os.FileMode(header.Mode)
		if fileMode == 0 {
			fileMode = 0644
		}
		// 清除特殊位以提高安全性
		fileMode &= ^(os.ModeSetuid | os.ModeSetgid | os.ModeSticky)
		// 应用权限掩码
		fileMode = validatePermissions(fileMode, options.PermissionsMask)

		file, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, fileMode)
		if err != nil {
			return 0, errors.Wrap(err, "创建目标文件失败")
		}
		defer closeWithError(file, "关闭目标文件失败")

		written, err := safeCopy(options.Context, file, tarReader, options.MaxFileSize, options.BufferSize)
		if err != nil {
			return written, err
		}

		return written, nil

	case tar.TypeSymlink:
		// 符号链接安全检查
		if filepath.IsAbs(header.Linkname) {
			return 0, errors.New("拒绝绝对路径符号链接")
		}

		linkTarget := filepath.Join(filepath.Dir(target), header.Linkname)
		if !isPathSafe(linkTarget, dst) {
			return 0, errors.New("符号链接指向基础目录外")
		}

		// 确保父目录存在
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return 0, errors.Wrap(err, "创建符号链接父目录失败")
		}

		if !options.FollowSymlinks {
			if err := os.Symlink(header.Linkname, target); err != nil {
				return 0, err
			}
			return 0, nil
		}
		return 0, errors.New("不允许跟随符号链接")

	default:
		// 忽略不支持的类型
		return 0, nil
	}
}

// ValidateSingleDirTarGz 校验 tar.gz 文件是否只包含一个顶层目录
// 优化点:减少内存占用、提前终止检查
func ValidateSingleDirTarGz(src string, opts ...ArchiveOption) (string, error) {
	options := applyOptions(opts...)

	if options.Context.Err() != nil {
		return "", errors.Wrap(options.Context.Err(), "校验 tar.gz 文件是否只包含一个顶层目录:上下文检查失败")
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return "", errors.Wrapf(err, "打开源文件失败, src=%s", src)
	}
	defer closeWithError(srcFile, "关闭源文件失败")

	gzReader, err := gzip.NewReader(srcFile)
	if err != nil {
		return "", errors.Wrapf(err, "创建gzip读取器失败, src=%s", src)
	}
	defer closeWithError(gzReader, "关闭gzip读取器失败")

	tarReader := tar.NewReader(gzReader)
	topLevelEntries := make(map[string]bool, 1) // 初始容量1，减少扩容
	var firstDirName string

	for {
		if options.Context.Err() != nil {
			return "", errors.Wrap(options.Context.Err(), "遍历tar文件条目:上下文检查失败")
		}

		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", errors.Wrap(err, "读取tar条目失败")
		}

		// 提取顶层目录
		cleanName := filepath.Clean(header.Name)
		cleanName = strings.TrimPrefix(cleanName, "/")
		parts := strings.Split(cleanName, "/")
		if len(parts) == 0 || parts[0] == "" {
			continue
		}
		topLevelName := parts[0]

		// 记录顶层目录
		topLevelEntries[topLevelName] = true
		if firstDirName == "" {
			firstDirName = topLevelName
		}

		// 提前终止:超过1个顶层目录直接返回错误
		if len(topLevelEntries) > 1 {
			return "", createMultipleEntriesError(topLevelEntries)
		}
	}

	// 结果校验
	if len(topLevelEntries) == 0 {
		return "", errors.New("压缩文件为空")
	}
	if len(topLevelEntries) > 1 {
		return "", createMultipleEntriesError(topLevelEntries)
	}

	return firstDirName, nil
}

// TarGzStream 从流压缩到流
func TarGzStream(src io.Reader, dst io.Writer, fileName string, opts ...ArchiveOption) error {
	options := applyOptions(opts...)

	// 前置检查
	if options.Context.Err() != nil {
		return errors.Wrap(options.Context.Err(), "tar.gz流压缩:上下文检查失败")
	}
	if src == nil || dst == nil {
		return errors.New("源/目标流不能为空")
	}
	if fileName == "" {
		fileName = "data"
	}

	// 对于tar格式，需要先读取所有数据以计算大小
	var buffer bytes.Buffer
	_, err := safeCopy(options.Context, &buffer, src, options.MaxFileSize, options.BufferSize)
	if err != nil {
		return errors.Wrap(err, "读取流数据失败")
	}

	// 创建gzip写入器
	gzWriter, err := gzip.NewWriterLevel(dst, options.CompressionLevel)
	if err != nil {
		return errors.Wrap(err, "创建gzip写入器失败")
	}

	// 创建tar写入器
	tarWriter := tar.NewWriter(gzWriter)

	// 改进的资源清理
	defer func() {
		// 先关闭tar写入器
		if tarWriter != nil {
			closeWithError(tarWriter, "关闭tar写入器失败")
		}
		// 再关闭gzip写入器
		if gzWriter != nil {
			closeWithError(gzWriter, "关闭gzip写入器失败")
		}
	}()

	// 创建文件头
	header := &tar.Header{
		Name:     fileName,
		Size:     int64(buffer.Len()),
		Mode:     0644,
		Typeflag: tar.TypeReg,
	}

	// 写入tar头
	if err := tarWriter.WriteHeader(header); err != nil {
		return errors.Wrap(err, "写入tar头失败")
	}

	// 复制内容
	_, err = buffer.WriteTo(tarWriter)
	if err != nil {
		return errors.Wrap(err, "复制流内容失败")
	}

	return nil
}

// UntarGzStream 从流解压到流
func UntarGzStream(src io.Reader, dst io.Writer, opts ...ArchiveOption) error {
	options := applyOptions(opts...)

	// 前置检查
	if options.Context.Err() != nil {
		return errors.Wrap(options.Context.Err(), "tar.gz流解压:上下文检查失败")
	}
	if src == nil || dst == nil {
		return errors.New("源/目标流不能为空")
	}

	// 创建gzip读取器
	gzReader, err := gzip.NewReader(src)
	if err != nil {
		return errors.Wrap(err, "创建gzip读取器失败")
	}
	defer closeWithError(gzReader, "关闭gzip读取器失败")

	// 创建tar读取器
	tarReader := tar.NewReader(gzReader)

	// 只处理第一个文件
	_, err = tarReader.Next()
	if err == io.EOF {
		return errors.New("tar.gz流为空")
	}
	if err != nil {
		return errors.Wrap(err, "读取tar条目失败")
	}

	// 复制内容
	_, err = safeCopy(options.Context, dst, tarReader, options.MaxFileSize, options.BufferSize)
	if err != nil {
		return errors.Wrap(err, "复制流内容失败")
	}

	return nil
}
