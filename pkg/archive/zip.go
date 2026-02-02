package archive

import (
	"archive/zip"
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"gin-artweb/pkg/ctxutil"
)

// Zip 将文件或目录压缩为ZIP格式
func Zip(src, dst string, opts ...ArchiveOption) error {
	options := applyOptions(opts...)

	// 前置检查
	if err := ctxutil.CheckContext(options.Context); err != nil {
		return errors.WithMessage(err, "zip压缩:上下文检查失败")
	}
	if src == "" || dst == "" {
		return errors.New("源路径/目标路径不能为空")
	}

	// 路径安全检查
	cleanSrc := filepath.Clean(src)
	cleanDst := filepath.Clean(dst)

	// 验证源路径
	if !filepath.IsAbs(cleanSrc) {
		absSrc, err := filepath.Abs(cleanSrc)
		if err != nil {
			return errors.WithMessage(err, "获取源路径绝对路径失败")
		}
		cleanSrc = absSrc
	}

	// 创建目标文件前检查父目录
	dstDir := filepath.Dir(cleanDst)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return errors.WithMessagef(err, "创建目标目录失败, dir=%s", dstDir)
	}

	// 创建目标文件
	dstFile, err := os.Create(cleanDst)
	if err != nil {
		return errors.WithMessagef(err, "创建目标文件失败, dst=%s", cleanDst)
	}

	// 使用带缓冲的写入器提高性能
	bufferedWriter := bufio.NewWriterSize(dstFile, options.BufferSize)
	var closeErrors []error

	// 预先声明变量以便在defer中使用
	var zipWriter *zip.Writer

	// 改进的资源清理函数
	defer func() {
		// 先刷新缓冲区
		if err := bufferedWriter.Flush(); err != nil && len(closeErrors) == 0 {
			closeErrors = append(closeErrors, errors.WithMessage(err, "刷新缓冲区失败"))
		}

		// 关闭zip写入器
		if zipWriter != nil {
			if err := zipWriter.Close(); err != nil {
				closeErrors = append(closeErrors, errors.WithMessage(err, "关闭zip写入器失败"))
			}
		}

		// 关闭目标文件
		if err := dstFile.Close(); err != nil {
			closeErrors = append(closeErrors, errors.WithMessage(err, "关闭目标文件失败"))
		}

		// 如果有关闭错误且主操作成功，则返回第一个关闭错误
		if len(closeErrors) > 0 && err == nil {
			err = closeErrors[0]
		}
	}()

	// 初始化zip写入器
	zipWriter = zip.NewWriter(bufferedWriter)

	// 获取源信息
	srcInfo, err := os.Stat(cleanSrc)
	if err != nil {
		return errors.WithMessagef(err, "获取源信息失败, src=%s", cleanSrc)
	}

	// 处理文件/目录
	fileCount := 0
	var processErr error

	if srcInfo.IsDir() {
		processErr = filepath.Walk(cleanSrc, func(filePath string, info os.FileInfo, walkErr error) error {
			if walkErr != nil {
				return errors.WithMessagef(walkErr, "遍历目录失败, filepath=%s", filePath)
			}

			// 安全检查：确保文件路径在源目录内
			relPath, err := filepath.Rel(cleanSrc, filePath)
			if err != nil {
				return errors.WithMessagef(err, "计算相对路径失败, base=%s, target=%s", cleanSrc, filePath)
			}
			if strings.HasPrefix(relPath, "..") {
				return errors.Errorf("路径超出源目录范围: %s", filePath)
			}

			entryErr := processZipEntry(filePath, cleanSrc, info, zipWriter, &fileCount, options)
			if entryErr != nil {
				return errors.WithMessagef(entryErr, "处理zip条目失败, filepath=%s", filePath)
			}
			return nil
		})
	} else {
		// 处理单个文件
		parentDir := filepath.Dir(cleanSrc)
		processErr = processZipEntry(cleanSrc, parentDir, srcInfo, zipWriter, &fileCount, options)
	}

	if processErr != nil {
		return errors.WithMessage(processErr, "zip压缩失败")
	}

	// 显式刷新确保所有数据写入
	if err := bufferedWriter.Flush(); err != nil {
		return errors.WithMessage(err, "刷新写入缓冲区失败")
	}

	return nil
}

// processZipEntry 处理单个zip条目（解耦核心逻辑）
func processZipEntry(filePath, baseDir string, info os.FileInfo, zipWriter *zip.Writer, fileCount *int, options ArchiveOptions) error {
	// 上下文检查
	if err := ctxutil.CheckContext(options.Context); err != nil {
		return errors.WithMessage(err, "zip压缩:上下文检查失败")
	}

	// 跳过基础目录
	relPath, err := filepath.Rel(baseDir, filePath)
	if err != nil {
		return errors.WithMessagef(err, "计算相对路径失败, filepath=%s", filePath)
	}
	if relPath == "." {
		return nil
	}

	// 文件数量限制
	*fileCount++
	if options.MaxFiles > 0 && *fileCount > options.MaxFiles {
		return errors.Errorf("文件数量超过限制, max=%d, current=%d", options.MaxFiles, *fileCount)
	}

	// 创建zip头
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return errors.WithMessagef(err, "创建zip头失败, filepath=%s", filePath)
	}

	// 规范化路径（兼容跨平台）
	header.Name = filepath.ToSlash(relPath)
	if info.IsDir() {
		header.Name += "/"
	}

	// 写入zip头
	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return errors.WithMessagef(err, "创建zip条目失败, filepath=%s", filePath)
	}

	// 写入文件内容（仅普通文件）
	if !info.IsDir() && info.Mode().IsRegular() {
		// 大小限制
		if options.MaxFileSize > 0 && info.Size() > options.MaxFileSize {
			return errors.Errorf("文件大小超过限制, filepath=%s, max=%d, current=%d", filePath, options.MaxFileSize, info.Size())
		}

		// 读取并写入文件
		file, err := os.Open(filePath)
		if err != nil {
			return errors.WithMessagef(err, "打开文件失败, filepath=%s", filePath)
		}
		defer closeWithError(file, "关闭文件失败")

		_, err = safeCopy(options.Context, writer, file, options.MaxFileSize, options.BufferSize)
		if err != nil {
			return errors.WithMessagef(err, "复制文件内容失败, filepath=%s", filePath)
		}
	}

	return nil
}

// Unzip 解压ZIP文件到指定目录
// 优化点:解耦处理逻辑、批量上下文检查、强化资源安全
func Unzip(src, dst string, opts ...ArchiveOption) error {
	options := applyOptions(opts...)

	// 前置检查
	if err := ctxutil.CheckContext(options.Context); err != nil {
		return errors.WithMessage(err, "zip解压:上下文检查失败")
	}
	if src == "" || dst == "" {
		return errors.New("源路径/目标路径不能为空")
	}

	// 路径安全检查
	cleanSrc := filepath.Clean(src)
	cleanDst := filepath.Clean(dst)

	// 验证源文件是否存在
	if _, err := os.Stat(cleanSrc); err != nil {
		return errors.WithMessagef(err, "源文件不存在或无法访问, src=%s", cleanSrc)
	}

	// 打开zip文件
	reader, err := zip.OpenReader(cleanSrc)
	if err != nil {
		return errors.WithMessagef(err, "打开zip文件失败, src=%s", cleanSrc)
	}
	defer closeWithError(reader, "关闭zip读取器失败")

	// 创建目标目录
	if err := os.MkdirAll(cleanDst, 0755); err != nil {
		return errors.WithMessagef(err, "创建目标目录失败, dst=%s", cleanDst)
	}

	// 遍历zip条目
	fileCount := 0
	for i, file := range reader.File {
		// 批量上下文检查（每100个条目检查一次，减少开销）
		if i%100 == 0 {
			if err := ctxutil.CheckContext(options.Context); err != nil {
				return errors.WithMessage(err, "上下文检查失败")
			}
		}

		fileCount++
		if options.MaxFiles > 0 && fileCount > options.MaxFiles {
			return errors.Errorf("文件数量超过限制, max=%d, current=%d", options.MaxFiles, fileCount)
		}

		if err := processUnzipEntry(file, cleanDst, options); err != nil {
			return errors.WithMessagef(err, "处理zip条目失败, entry=%s", file.Name)
		}
	}

	return nil
}

// processUnzipEntry 处理单个解压条目（解耦核心逻辑）
func processUnzipEntry(zipFile *zip.File, dst string, options ArchiveOptions) error {
	// 构造目标路径并检查安全性
	target := filepath.Join(dst, filepath.FromSlash(zipFile.Name))

	// 更严格的路径安全检查
	if !isPathSafe(target, dst) {
		return errors.Errorf("非法路径(路径遍历攻击),target=%s, base=%s", target, dst)
	}

	// 额外的安全检查：确保目标路径在目标目录内
	relPath, err := filepath.Rel(dst, target)
	if err != nil {
		return errors.WithMessagef(err, "计算相对路径失败, target=%s, base=%s", target, dst)
	}
	if strings.HasPrefix(relPath, "..") {
		return errors.Errorf("路径超出目标目录范围: %s", target)
	}

	// 处理目录
	if zipFile.FileInfo().IsDir() {
		// 设置合适的目录权限
		dirMode := zipFile.Mode()
		if dirMode == 0 {
			dirMode = 0755
		}
		return os.MkdirAll(target, dirMode)
	}

	// 处理文件
	return unzipFile(zipFile, target, options)
}

// unzipFile 解压单个ZIP文件条目（优化资源释放）
func unzipFile(zipFile *zip.File, target string, options ArchiveOptions) error {
	// 大小限制
	if options.MaxFileSize > 0 && zipFile.FileInfo().Size() > options.MaxFileSize {
		return errors.Errorf("文件大小超过限制, filepath=%s, max=%d, current=%d", target, options.MaxFileSize, zipFile.FileInfo().Size())
	}

	// 创建父目录
	parentDir := filepath.Dir(target)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return errors.WithMessagef(err, "创建父目录失败, dir=%s", parentDir)
	}

	// 打开zip内文件
	srcFile, err := zipFile.Open()
	if err != nil {
		return errors.WithMessagef(err, "打开zip内文件失败, entry=%s", zipFile.Name)
	}
	defer closeWithError(srcFile, "关闭zip内文件失败")

	// 创建目标文件（使用更安全的权限模式）
	fileMode := zipFile.Mode()
	if fileMode == 0 {
		fileMode = 0644
	}
	// 清除特殊位以提高安全性
	fileMode &= ^(os.ModeSetuid | os.ModeSetgid | os.ModeSticky)

	targetFile, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR|os.O_TRUNC, fileMode)
	if err != nil {
		return errors.WithMessagef(err, "创建目标文件失败, target=%s", target)
	}
	defer closeWithError(targetFile, "关闭目标文件失败")

	// 复制内容
	_, err = safeCopy(options.Context, targetFile, srcFile, options.MaxFileSize, options.BufferSize)
	return errors.WithMessagef(err, "复制文件内容失败, target=%s", target)
}

// ValidateSingleDirZip 校验 ZIP 文件是否只包含一个顶层目录
// 优化点:提前终止、减少内存占用、统一错误格式
func ValidateSingleDirZip(src string, opts ...ArchiveOption) (string, error) {
	options := applyOptions(opts...)

	if err := ctxutil.CheckContext(options.Context); err != nil {
		return "", errors.WithMessage(err, "上下文检查失败")
	}

	// 路径安全检查
	cleanSrc := filepath.Clean(src)

	reader, err := zip.OpenReader(cleanSrc)
	if err != nil {
		return "", errors.WithMessagef(err, "打开zip文件失败, src=%s", cleanSrc)
	}
	defer closeWithError(reader, "关闭zip读取器失败")

	topLevelEntries := make(map[string]bool, 1) // 初始容量1
	var firstDirName string

	for i, file := range reader.File {
		// 批量上下文检查
		if i%100 == 0 {
			if err := ctxutil.CheckContext(options.Context); err != nil {
				return "", errors.WithMessage(err, "上下文检查失败")
			}
		}

		// 清理路径
		name := filepath.Clean(file.Name)
		name = strings.TrimPrefix(name, "./")
		name = strings.TrimSuffix(name, "/")
		if name == "" {
			continue
		}

		// 提取顶层目录
		topLevelName := name
		if strings.Contains(name, "/") {
			parts := strings.Split(name, "/")
			if len(parts) > 0 {
				topLevelName = parts[0]
			}
		}

		// 记录顶层目录
		topLevelEntries[topLevelName] = true
		if firstDirName == "" && file.FileInfo().IsDir() {
			firstDirName = topLevelName
		}

		// 提前终止
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
	if firstDirName == "" {
		return "", errors.New("压缩文件唯一条目不是目录")
	}

	return firstDirName, nil
}
