package archive

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"context"
	"gin-artweb/pkg/ctxutil"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

// checkContext 检查上下文状态
func checkContext(ctx context.Context) error {
	if ctx == nil {
		return nil
	}

	select {
	case <-ctx.Done():
		return errors.New(ErrContextCancelled)
	default:
		return nil
	}
}

// TarGz 将指定路径的文件或目录压缩为 tar.gz 格式
// 优化点:解耦核心逻辑、统一错误处理、减少重复代码、增强安全性
func TarGz(src, dst string, opts ...ArchiveOption) error {
	options := applyOptions(opts...)

	// 前置检查
	if err := ctxutil.CheckContext(options.Context); err != nil {
		return errors.Wrap(err, ErrContextCancelled)
	}
	if src == "" || dst == "" {
		return errors.New(ErrEmptyPath)
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

	// 改进的资源清理函数
	defer func() {
		// 先刷新缓冲区
		if err := bufferedWriter.Flush(); err != nil && len(closeErrors) == 0 {
			closeErrors = append(closeErrors, errors.Wrap(err, "刷新缓冲区失败"))
		}

		// 按顺序关闭资源
		if gzWriter != nil {
			if err := gzWriter.Close(); err != nil {
				closeErrors = append(closeErrors, errors.Wrap(err, "关闭gzip写入器失败"))
			}
		}
		if tarWriter != nil {
			if err := tarWriter.Close(); err != nil {
				closeErrors = append(closeErrors, errors.Wrap(err, "关闭tar写入器失败"))
			}
		}
		if err := dstFile.Close(); err != nil {
			closeErrors = append(closeErrors, errors.Wrap(err, "关闭目标文件失败"))
		}

		// 如果有关闭错误且主操作成功，则返回第一个关闭错误
		if len(closeErrors) > 0 && err == nil {
			err = closeErrors[0]
		}
	}()

	// 初始化压缩写入器
	gzWriter = gzip.NewWriter(bufferedWriter)
	tarWriter = tar.NewWriter(gzWriter)

	// 统一处理文件/目录
	fileCount := 0 // 保持int类型以匹配processTarEntry函数签名
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

			entryErr := processTarEntry(filePath, cleanSrc, info, tarWriter, &fileCount, options)
			if entryErr != nil {
				return errors.Wrapf(entryErr, "处理文件条目失败, filepath=%s", filePath)
			}
			return nil
		})
	} else {
		// 处理单个文件
		parentDir := filepath.Dir(cleanSrc)
		processErr = processTarEntry(cleanSrc, parentDir, srcInfo, tarWriter, &fileCount, options)
	}

	if processErr != nil {
		return errors.Wrap(processErr, "tar.gz压缩失败")
	}

	// 显式刷新确保所有数据写入
	if err := bufferedWriter.Flush(); err != nil {
		return errors.Wrap(err, "刷新写入缓冲区失败")
	}

	return nil
}

// validateInputs 验证输入参数
func validateInputs(src, dst string, options ArchiveOptions) error {
	if err := ctxutil.CheckContext(options.Context); err != nil {
		return errors.Wrap(err, "tar.gz压缩:上下文检查失败")
	}

	if src == "" || dst == "" {
		return errors.New("源路径/目标路径不能为空")
	}

	if src == dst {
		return errors.New("源路径和目标路径不能相同")
	}

	return nil
}

// sanitizePaths 路径标准化和安全检查
func sanitizePaths(src, dst string) (cleanSrc, cleanDst string, err error) {
	cleanSrc = filepath.Clean(src)
	cleanDst = filepath.Clean(dst)

	// 获取绝对路径
	absSrc, err := filepath.Abs(cleanSrc)
	if err != nil {
		return "", "", errors.Wrap(err, "获取源路径绝对路径失败")
	}
	cleanSrc = absSrc

	absDst, err := filepath.Abs(cleanDst)
	if err != nil {
		return "", "", errors.Wrap(err, "获取目标路径绝对路径失败")
	}
	cleanDst = absDst

	return cleanSrc, cleanDst, nil
}

// createDestinationFile 创建目标文件
func createDestinationFile(dst string) (*os.File, error) {
	// 创建目标目录
	dstDir := filepath.Dir(dst)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return nil, errors.Wrapf(err, "创建目标目录失败, dir=%s", dstDir)
	}

	// 检查目标文件是否已存在
	if _, err := os.Stat(dst); err == nil {
		return nil, errors.Errorf("目标文件已存在: %s", dst)
	}

	dstFile, err := os.Create(dst)
	if err != nil {
		return nil, errors.Wrapf(err, "创建目标文件失败, dst=%s", dst)
	}

	return dstFile, nil
}

// compressFiles 处理文件压缩逻辑
func compressFiles(src string, srcInfo os.FileInfo, tarWriter *tar.Writer, options ArchiveOptions) error {
	fileCount := 0

	if srcInfo.IsDir() {
		return filepath.Walk(src, func(filePath string, info os.FileInfo, walkErr error) error {
			if walkErr != nil {
				return errors.Wrapf(walkErr, "遍历目录失败, filepath=%s", filePath)
			}

			// 安全检查：确保文件路径在源目录内
			if err := validatePathWithinBase(filePath, src); err != nil {
				return err
			}

			return processTarEntry(filePath, src, info, tarWriter, &fileCount, options)
		})
	}

	// 处理单个文件
	parentDir := filepath.Dir(src)
	return processTarEntry(src, parentDir, srcInfo, tarWriter, &fileCount, options)
}

// validatePathWithinBase 验证文件路径是否在基础目录内
func validatePathWithinBase(filePath, baseDir string) error {
	relPath, err := filepath.Rel(baseDir, filePath)
	if err != nil {
		return errors.Wrapf(err, "计算相对路径失败, base=%s, target=%s", baseDir, filePath)
	}

	if strings.HasPrefix(relPath, "..") {
		return errors.Errorf("路径超出源目录范围: %s", filePath)
	}

	return nil
}

// processTarEntry 处理单个tar条目（解耦核心逻辑）
func processTarEntry(filePath, baseDir string, info os.FileInfo, tarWriter *tar.Writer, fileCount *int, options ArchiveOptions) error {
	// 上下文检查
	if err := ctxutil.CheckContext(options.Context); err != nil {
		return errors.Wrap(err, "处理单个tar条目:上下文检查失败")
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

	header, err := tar.FileInfoHeader(info, "")
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

		if _, err := safeCopy(options.Context, tarWriter, file, options.MaxFileSize, options.BufferSize); err != nil {
			return errors.Wrapf(err, "复制文件内容失败, filepath=%s", filePath)
		}
	}

	return nil
}

// UntarGz 解压 tar.gz 文件到指定目录
// 优化点:解耦处理逻辑、强化资源释放、统一错误格式
func UntarGz(src, dst string, opts ...ArchiveOption) error {
	options := applyOptions(opts...)

	// 前置检查
	if err := ctxutil.CheckContext(options.Context); err != nil {
		return errors.Wrap(err, "tar.gz解压:上下文检查失败")
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
	for {
		if err := ctxutil.CheckContext(options.Context); err != nil {
			return errors.Wrap(err, "tar.gz解压遍历文件:上下文检查失败")
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

		if err := processUntarEntry(header, tarReader, dst, options); err != nil {
			return errors.Wrapf(err, "处理tar条目失败, entry=%s", header.Name)
		}
	}

	return nil
}

// processUntarEntry 处理单个解压条目（解耦核心逻辑）
func processUntarEntry(header *tar.Header, tarReader *tar.Reader, dst string, options ArchiveOptions) error {
	// 构造目标路径并检查安全性
	target := filepath.Join(dst, header.Name)
	if !isPathSafe(target, dst) {
		return errors.Errorf("非法路径（路径遍历攻击）, target=%s, base=%s", target, dst)
	}

	// 按类型处理
	switch header.Typeflag {
	case tar.TypeDir:
		return os.MkdirAll(target, os.FileMode(header.Mode))

	case tar.TypeReg:
		// 大小限制
		if options.MaxFileSize > 0 && header.Size > options.MaxFileSize {
			return errors.Errorf("文件大小超过限制, file_path=%s, size=%d, max=%d", header.Name, header.Size, options.MaxFileSize)
		}

		// 创建父目录
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return errors.Wrap(err, "创建父目录失败")
		}

		// 写入文件
		file, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
		if err != nil {
			return errors.Wrap(err, "创建目标文件失败")
		}
		defer closeWithError(file, "关闭目标文件失败")

		_, err = safeCopy(options.Context, file, tarReader, options.MaxFileSize, options.BufferSize)
		return err

	case tar.TypeSymlink:
		// 符号链接安全检查
		if filepath.IsAbs(header.Linkname) {
			return errors.New("拒绝绝对路径符号链接")
		}

		linkTarget := filepath.Join(filepath.Dir(target), header.Linkname)
		if !isPathSafe(linkTarget, dst) {
			return errors.New("符号链接指向基础目录外")
		}

		if !options.FollowSymlinks {
			return os.Symlink(header.Linkname, target)
		}
		return errors.New("不允许跟随符号链接")

	default:
		// 忽略不支持的类型
		return nil
	}
}

// ValidateSingleDirTarGz 校验 tar.gz 文件是否只包含一个顶层目录
// 优化点:减少内存占用、提前终止检查
func ValidateSingleDirTarGz(src string, opts ...ArchiveOption) (string, error) {
	options := applyOptions(opts...)

	if err := ctxutil.CheckContext(options.Context); err != nil {
		return "", errors.Wrap(err, "校验 tar.gz 文件是否只包含一个顶层目录:上下文检查失败")
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
		if err := ctxutil.CheckContext(options.Context); err != nil {
			return "", errors.Wrap(err, "遍历tar文件条目:上下文检查失败")
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
