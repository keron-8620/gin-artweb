package serializer

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"emperror.dev/errors"
	"github.com/goccy/go-yaml"
)

// ReadYAML 读取并解析 YAML 文件
func ReadYAML(filePath string, v any, opts ...SerializerOption) (*ReadResult, error) {
	startTime := time.Now()

	options := applyOptions(opts...)

	// 检查上下文是否已取消
	if ctxErr := options.Context.Err(); ctxErr != nil {
		return nil, errors.WithMessage(ctxErr, "上下文已取消")
	}

	// 验证参数
	if err := validatePath(filePath); err != nil {
		return nil, err
	}
	if v == nil {
		return nil, errors.New("接收变量不能为空")
	}

	// 检查文件是否存在
	fileInfo, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return nil, errors.Errorf("YAML文件不存在, 文件路径=%s", filePath)
	}
	if err != nil {
		return nil, errors.Errorf("获取文件信息失败, 文件路径=%s", filePath)
	}

	// 检查上下文是否已取消
	if ctxErr := options.Context.Err(); ctxErr != nil {
		return nil, errors.WithMessage(ctxErr, "上下文已取消")
	}

	// 检查文件大小限制
	if options.MaxFileSize > 0 && fileInfo.Size() > options.MaxFileSize {
		return nil, errors.Errorf("YAML文件大小超过限制, 文件路径=%s, 最大限制=%d, 当前大小=%d", filePath, options.MaxFileSize, fileInfo.Size())
	}

	// 读取文件内容
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, errors.Errorf("读取YAML文件失败, 文件路径=%s", filePath)
	}

	if len(data) == 0 {
		return nil, errors.Errorf("YAML文件为空, 文件路径=%s", filePath)
	}

	// 检查上下文是否已取消
	if ctxErr := options.Context.Err(); ctxErr != nil {
		return nil, errors.WithMessage(ctxErr, "上下文已取消")
	}

	// 解析YAML
	if err := yaml.Unmarshal(data, v); err != nil {
		return nil, errors.WithMessagef(err, "解析YAML文件失败, 文件路径=%s", filePath)
	}

	return &ReadResult{
		FilePath: filePath,
		Size:     int64(len(data)),
		Duration: time.Since(startTime),
		Success:  true,
	}, nil
}

// WriteYAML 将给定的数据写入指定路径的 YAML 文件
func WriteYAML(filePath string, data any, opts ...SerializerOption) (*WriteResult, error) {
	startTime := time.Now()

	options := applyOptions(opts...)

	// 检查上下文是否已取消
	if ctxErr := options.Context.Err(); ctxErr != nil {
		return nil, errors.WithMessage(ctxErr, "上下文已取消")
	}

	// 验证参数
	if err := validatePath(filePath); err != nil {
		return nil, err
	}

	if options.Atomic {
		return writeYAMLAtomic(filePath, data, options, startTime)
	}

	// 检查上下文是否已取消
	if ctxErr := options.Context.Err(); ctxErr != nil {
		return nil, errors.WithMessage(ctxErr, "上下文已取消")
	}

	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(filePath), options.DirMode); err != nil {
		return nil, errors.Errorf("创建目录失败, 目录路径=%s", filepath.Dir(filePath))
	}

	// 检查上下文是否已取消
	if ctxErr := options.Context.Err(); ctxErr != nil {
		return nil, errors.WithMessage(ctxErr, "上下文已取消")
	}

	// 序列化为 YAML 字节流
	out, err := yaml.Marshal(data)
	if err != nil {
		return nil, errors.Errorf("序列化YAML数据失败, 文件路径=%s", filePath)
	}

	// 检查文件大小限制
	if options.MaxFileSize > 0 && int64(len(out)) > options.MaxFileSize {
		return nil, errors.Errorf("YAML文件大小超过限制, 文件路径=%s, 最大限制=%d, 当前大小=%d", filePath, options.MaxFileSize, int64(len(out)))
	}

	// 检查上下文是否已取消
	if ctxErr := options.Context.Err(); ctxErr != nil {
		return nil, errors.WithMessage(ctxErr, "上下文已取消")
	}

	// 写入文件（会覆盖已有内容）
	if err := os.WriteFile(filePath, out, options.FileMode); err != nil {
		return nil, errors.WithMessagef(err, "写入YAML文件失败, 文件路径=%s", filePath)
	}

	return &WriteResult{
		FilePath: filePath,
		Size:     int64(len(out)),
		Duration: time.Since(startTime),
		Success:  true,
	}, nil
}

// writeYAMLAtomic 原子写入YAML
func writeYAMLAtomic(filePath string, data any, options SerializerOptions, startTime time.Time) (*WriteResult, error) {
	// 检查上下文是否已取消
	if ctxErr := options.Context.Err(); ctxErr != nil {
		return nil, errors.WithMessage(ctxErr, "上下文已取消")
	}

	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(filePath), options.DirMode); err != nil {
		return nil, errors.WithMessagef(err, "创建目录失败, 目录路径=%s", filepath.Dir(filePath))
	}

	// 检查上下文是否已取消
	if ctxErr := options.Context.Err(); ctxErr != nil {
		return nil, errors.WithMessage(ctxErr, "上下文已取消")
	}

	// 创建临时文件
	tmpFile := filePath + ".tmp"

	// 先写入临时文件，传递所有相关选项
	result, err := WriteYAML(tmpFile, data,
		WithContext(options.Context),
		WithFileMode(options.FileMode),
		WithDirMode(options.DirMode),
		WithMaxFileSize(options.MaxFileSize))

	if err != nil {
		// 清理临时文件
		os.Remove(tmpFile)
		return nil, errors.WithMessagef(err, "写入YAML临时文件失败, 文件路径=%s", filePath)
	}

	// 检查上下文是否已取消
	if ctxErr := options.Context.Err(); ctxErr != nil {
		// 清理临时文件
		os.Remove(tmpFile)
		return nil, errors.WithMessage(ctxErr, "上下文已取消")
	}

	// 原子重命名
	if err := os.Rename(tmpFile, filePath); err != nil {
		// 清理临时文件
		os.Remove(tmpFile)
		return nil, errors.WithMessagef(err, "将YAML临时文件重命名为指定的文件名失败, 临时文件=%s, 文件路径=%s", tmpFile, filePath)
	}

	result.FilePath = filePath
	result.Duration = time.Since(startTime)
	return result, nil
}

// ReadYAMLWithTimeout 带超时的YAML读取
func ReadYAMLWithTimeout(filePath string, v any, timeout time.Duration) (*ReadResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return ReadYAML(filePath, v, WithContext(ctx))
}

// WriteYAMLWithTimeout 带超时的YAML写入
func WriteYAMLWithTimeout(filePath string, data any, timeout time.Duration) (*WriteResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return WriteYAML(filePath, data, WithContext(ctx))
}
