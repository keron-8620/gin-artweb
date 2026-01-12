package serializer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ReadJSON 读取并解析 JSON 文件
func ReadJSON(filename string, v any, opts ...SerializerOption) (*ReadResult, error) {
	startTime := time.Now()

	options := applyOptions(opts...)

	// 检查文件是否存在
	fileInfo, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("JSON文件不存在: %s", filename)
	}
	if err != nil {
		return nil, fmt.Errorf("获取文件信息 %s 失败: %w", filename, err)
	}

	// 检查文件大小限制
	if options.MaxFileSize > 0 && fileInfo.Size() > options.MaxFileSize {
		return nil, fmt.Errorf("JSON文件 %s 大小 %d 超过限制 %d", filename, fileInfo.Size(), options.MaxFileSize)
	}

	// 读取文件内容
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("读取JSON文件 %s 失败: %w", filename, err)
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("JSON文件 %s 为空", filename)
	}

	// 解析JSON
	if err := json.Unmarshal(data, v); err != nil {
		return nil, fmt.Errorf("解析JSON文件 %s 失败: %w", filename, err)
	}

	result := &ReadResult{
		FileName: filename,
		Size:     int64(len(data)),
		Duration: time.Since(startTime),
	}

	return result, nil
}

// WriteJSON 将数据序列化为JSON格式并写入文件
func WriteJSON(filename string, data any, opts ...SerializerOption) (*WriteResult, error) {
	startTime := time.Now()

	options := applyOptions(opts...)

	if options.Atomic {
		result, err := writeJSONAtomic(filename, data, options)
		if err != nil {
			return nil, err
		}
		result.Duration = time.Since(startTime)
		return result, nil
	}

	return writeJSON(filename, data, options, startTime)
}

// writeJSON 普通写入JSON
func writeJSON(filename string, data any, options SerializerOptions, startTime time.Time) (*WriteResult, error) {
	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(filename), options.DirMode); err != nil {
		return nil, fmt.Errorf("创建目录失败: %w", err)
	}

	// 创建或截断文件
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, options.FileMode)
	if err != nil {
		return nil, fmt.Errorf("创建文件 %s 失败: %w", filename, err)
	}
	defer file.Close()

	// 序列化为JSON并写入文件
	encoder := json.NewEncoder(file)

	if options.Indent > 0 {
		indentStr := strings.Repeat(" ", int(options.Indent))
		encoder.SetIndent("", indentStr)
	}

	if err := encoder.Encode(data); err != nil {
		return nil, fmt.Errorf("序列化数据到 %s 失败: %w", filename, err)
	}

	// 获取文件大小
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("获取文件信息失败: %w", err)
	}

	result := &WriteResult{
		FileName: filename,
		Size:     fileInfo.Size(),
		Duration: time.Since(startTime),
	}

	return result, nil
}

// writeJSONAtomic 原子写入JSON
func writeJSONAtomic(filename string, data any, options SerializerOptions) (*WriteResult, error) {
	startTime := time.Now()

	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(filename), options.DirMode); err != nil {
		return nil, fmt.Errorf("创建目录失败: %w", err)
	}

	// 创建临时文件
	tmpFile := filename + ".tmp"

	// 先写入临时文件
	result, err := writeJSON(tmpFile, data, options, startTime)
	if err != nil {
		// 清理临时文件
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

// ReadJSONWithTimeout 带超时的JSON读取
func ReadJSONWithTimeout(filename string, v any, timeout time.Duration) (*ReadResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return ReadJSON(filename, v, WithContext(ctx))
}

// WriteJSONWithTimeout 带超时的JSON写入
func WriteJSONWithTimeout(filename string, data any, timeout time.Duration, indent uint8) (*WriteResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return WriteJSON(filename, data, WithContext(ctx), WithIndent(indent))
}
