package serializer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/goccy/go-yaml"

	"gin-artweb/internal/shared/errors"
)

// ReadYAML 读取并解析 YAML 文件
func ReadYAML(filename string, v any, opts ...SerializerOption) (*ReadResult, error) {
	startTime := time.Now()

	options := applyOptions(opts...)

	// 检查上下文
	if err := errors.CheckContext(options.Context); err != nil {
		return nil, err
	}

	// 检查文件是否存在
	fileInfo, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("YAML文件不存在: %s", filename)
	}
	if err != nil {
		return nil, fmt.Errorf("获取文件信息 %s 失败: %w", filename, err)
	}

	// 检查文件大小限制
	if options.MaxFileSize > 0 && fileInfo.Size() > options.MaxFileSize {
		return nil, fmt.Errorf("YAML文件 %s 大小 %d 超过限制 %d", filename, fileInfo.Size(), options.MaxFileSize)
	}

	// 读取文件内容
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("读取YAML文件 %s 失败: %w", filename, err)
	}

	// 检查上下文
	if err := errors.CheckContext(options.Context); err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("YAML文件 %s 为空", filename)
	}

	// 解析YAML
	if err := yaml.Unmarshal(data, v); err != nil {
		return nil, fmt.Errorf("解析YAML文件 %s 失败: %w", filename, err)
	}

	result := &ReadResult{
		FileName: filename,
		Size:     int64(len(data)),
		Duration: time.Since(startTime),
	}

	return result, nil
}

// WriteYAML 将给定的数据写入指定路径的 YAML 文件
func WriteYAML(filename string, data any, opts ...SerializerOption) (*WriteResult, error) {
	startTime := time.Now()

	options := applyOptions(opts...)

	// 检查上下文
	if err := errors.CheckContext(options.Context); err != nil {
		return nil, err
	}

	if options.Atomic {
		return writeYAMLAtomic(filename, data, options, startTime)
	}

	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(filename), options.DirMode); err != nil {
		return nil, fmt.Errorf("创建目录失败: %w", err)
	}

	// 序列化为 YAML 字节流
	out, err := yaml.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("序列化YAML数据失败: %w", err)
	}

	// 检查文件大小限制
	if options.MaxFileSize > 0 && int64(len(out)) > options.MaxFileSize {
		return nil, fmt.Errorf("序列化后的YAML数据大小 %d 超过限制 %d", len(out), options.MaxFileSize)
	}

	// 写入文件（会覆盖已有内容）
	if err := os.WriteFile(filename, out, options.FileMode); err != nil {
		return nil, fmt.Errorf("写入YAML文件 %s 失败: %w", filename, err)
	}

	result := &WriteResult{
		FileName: filename,
		Size:     int64(len(out)),
		Duration: time.Since(startTime),
	}

	return result, nil
}

// writeYAMLAtomic 原子写入YAML
func writeYAMLAtomic(filename string, data any, options SerializerOptions, startTime time.Time) (*WriteResult, error) {
	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(filename), options.DirMode); err != nil {
		return nil, fmt.Errorf("创建目录失败: %w", err)
	}

	// 创建临时文件
	tmpFile := filename + ".tmp"

	// 先写入临时文件
	result, err := WriteYAML(tmpFile, data,
		WithContext(options.Context),
		WithFileMode(options.FileMode),
		WithDirMode(options.DirMode))

	if err != nil {
		// 清理临时文件
		os.Remove(tmpFile)
		return nil, err
	}

	// 检查上下文
	if err := errors.CheckContext(options.Context); err != nil {
		os.Remove(tmpFile)
		return nil, err
	}

	// 原子重命名
	if err := os.Rename(tmpFile, filename); err != nil {
		// 清理临时文件
		os.Remove(tmpFile)
		return nil, fmt.Errorf("原子重命名 %s -> %s 失败: %w", tmpFile, filename, err)
	}

	result.FileName = filename
	result.Duration = time.Since(startTime)
	return result, nil
}

// ReadYAMLWithTimeout 带超时的YAML读取
func ReadYAMLWithTimeout(filename string, v any, timeout time.Duration) (*ReadResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return ReadYAML(filename, v, WithContext(ctx))
}

// WriteYAMLWithTimeout 带超时的YAML写入
func WriteYAMLWithTimeout(filename string, data any, timeout time.Duration) (*WriteResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return WriteYAML(filename, data, WithContext(ctx))
}
