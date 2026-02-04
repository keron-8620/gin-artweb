package serializer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// ReadJSON 读取并解析 JSON 文件
func ReadJSON(filePath string, v any, opts ...SerializerOption) (*ReadResult, error) {
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
		return nil, errors.Errorf("JSON文件不存在, 文件路径=%s", filePath)
	}
	if err != nil {
		return nil, errors.WithMessagef(err, "获取文件信息失败, 文件路径=%s", filePath)
	}

	// 检查上下文是否已取消
	if ctxErr := options.Context.Err(); ctxErr != nil {
		return nil, errors.WithMessage(ctxErr, "上下文已取消")
	}

	// 检查文件大小限制
	if options.MaxFileSize > 0 && fileInfo.Size() > options.MaxFileSize {
		return nil, errors.Errorf("JSON文件大小超过限制, 文件路径=%s, 最大限制=%d, 当前大小=%d", filePath, options.MaxFileSize, fileInfo.Size())
	}

	// 读取文件内容
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, errors.WithMessagef(err, "读取JSON文件失败, 文件路径=%s", filePath)
	}

	if len(data) == 0 {
		return nil, errors.Errorf("JSON文件为空, 文件路径=%s", filePath)
	}

	// 检查上下文是否已取消
	if ctxErr := options.Context.Err(); ctxErr != nil {
		return nil, errors.WithMessage(ctxErr, "上下文已取消")
	}

	// 解析JSON
	if err := json.Unmarshal(data, v); err != nil {
		return nil, errors.WithMessagef(err, "解析JSON文件失败, 文件路径=%s", filePath)
	}

	result := &ReadResult{
		FilePath: filePath,
		Size:     int64(len(data)),
		Duration: time.Since(startTime),
		Success:  true,
	}

	return result, nil
}

// WriteJSON 将数据序列化为JSON格式并写入文件
func WriteJSON(filePath string, data any, opts ...SerializerOption) (*WriteResult, error) {
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
		result, err := writeJSONAtomic(filePath, data, options)
		if err != nil {
			return nil, err
		}
		result.Duration = time.Since(startTime)
		return result, nil
	}

	return writeJSON(filePath, data, options, startTime)
}

// writeJSON 普通写入JSON
func writeJSON(filePath string, data any, options SerializerOptions, startTime time.Time) (*WriteResult, error) {
	// 检查上下文是否已取消
	if ctxErr := options.Context.Err(); ctxErr != nil {
		return nil, errors.WithMessage(ctxErr, "上下文已取消")
	}

	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(filePath), options.DirMode); err != nil {
		return nil, fmt.Errorf("创建目录失败: %w", err)
	}

	// 检查上下文是否已取消
	if ctxErr := options.Context.Err(); ctxErr != nil {
		return nil, errors.WithMessage(ctxErr, "上下文已取消")
	}

	// 创建或截断文件
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, options.FileMode)
	if err != nil {
		return nil, fmt.Errorf("创建文件 %s 失败: %w", filePath, err)
	}
	defer file.Close()

	// 序列化为JSON并写入文件
	encoder := json.NewEncoder(file)

	if options.Indent > 0 {
		indentStr := strings.Repeat(" ", int(options.Indent))
		encoder.SetIndent("", indentStr)
	}

	if err := encoder.Encode(data); err != nil {
		return nil, fmt.Errorf("序列化数据到 %s 失败: %w", filePath, err)
	}

	// 获取文件大小
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("获取文件信息失败: %w", err)
	}

	result := &WriteResult{
		FilePath: filePath,
		Size:     fileInfo.Size(),
		Duration: time.Since(startTime),
		Success:  true,
	}

	return result, nil
}

// writeJSONAtomic 原子写入JSON
func writeJSONAtomic(filePath string, data any, options SerializerOptions) (*WriteResult, error) {
	startTime := time.Now()

	// 检查上下文是否已取消
	if ctxErr := options.Context.Err(); ctxErr != nil {
		return nil, errors.WithMessage(ctxErr, "上下文已取消")
	}

	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(filePath), options.DirMode); err != nil {
		return nil, fmt.Errorf("创建目录失败: %w", err)
	}

	// 检查上下文是否已取消
	if ctxErr := options.Context.Err(); ctxErr != nil {
		return nil, errors.WithMessage(ctxErr, "上下文已取消")
	}

	// 创建临时文件
	tmpFile := filePath + ".tmp"

	// 先写入临时文件
	result, err := writeJSON(tmpFile, data, options, startTime)
	if err != nil {
		// 清理临时文件
		os.Remove(tmpFile)
		return nil, err
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
		return nil, fmt.Errorf("原子重命名 %s -> %s 失败: %w", tmpFile, filePath, err)
	}

	result.FilePath = filePath
	result.Duration = time.Since(startTime)
	return result, nil
}

// ReadJSONWithTimeout 带超时的JSON读取
func ReadJSONWithTimeout(filePath string, v any, timeout time.Duration) (*ReadResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return ReadJSON(filePath, v, WithContext(ctx))
}

// WriteJSONWithTimeout 带超时的JSON写入
func WriteJSONWithTimeout(filePath string, data any, timeout time.Duration, indent uint8) (*WriteResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return WriteJSON(filePath, data, WithContext(ctx), WithIndent(indent))
}

// MarshalJSON 将数据序列化为JSON字节数组
// 支持缩进和上下文控制
// 适用于内存序列化操作
func MarshalJSON(data any, opts ...SerializerOption) ([]byte, *SerializeResult, error) {
	startTime := time.Now()
	options := applyOptions(opts...)

	// 检查上下文是否已取消
	if ctxErr := options.Context.Err(); ctxErr != nil {
		serializeResult := &SerializeResult{
			Size:     0,
			Duration: time.Since(startTime),
			Success:  false,
			Error:    ctxErr.Error(),
		}
		return nil, serializeResult, errors.WithMessage(ctxErr, "上下文已取消")
	}

	var result []byte
	var err error

	if options.Indent > 0 {
		indentStr := strings.Repeat(" ", int(options.Indent))
		result, err = json.MarshalIndent(data, "", indentStr)
	} else {
		result, err = json.Marshal(data)
	}

	serializeResult := &SerializeResult{
		Size:     int64(len(result)),
		Duration: time.Since(startTime),
		Success:  err == nil,
	}

	if err != nil {
		serializeResult.Error = err.Error()
		return nil, serializeResult, errors.WithMessage(err, "序列化JSON数据失败")
	}

	return result, serializeResult, nil
}

// UnmarshalJSON 将JSON字节数组反序列化为数据
// 支持上下文控制
// 适用于内存反序列化操作
func UnmarshalJSON(data []byte, v any, opts ...SerializerOption) (*SerializeResult, error) {
	startTime := time.Now()
	options := applyOptions(opts...)

	// 检查上下文是否已取消
	if ctxErr := options.Context.Err(); ctxErr != nil {
		serializeResult := &SerializeResult{
			Size:     int64(len(data)),
			Duration: time.Since(startTime),
			Success:  false,
			Error:    ctxErr.Error(),
		}
		return serializeResult, errors.WithMessage(ctxErr, "上下文已取消")
	}

	// 验证参数
	if data == nil {
		serializeResult := &SerializeResult{
			Size:     0,
			Duration: time.Since(startTime),
			Success:  false,
			Error:    "JSON数据不能为空",
		}
		return serializeResult, errors.New("JSON数据不能为空")
	}
	if v == nil {
		serializeResult := &SerializeResult{
			Size:     int64(len(data)),
			Duration: time.Since(startTime),
			Success:  false,
			Error:    "接收变量不能为空",
		}
		return serializeResult, errors.New("接收变量不能为空")
	}

	if err := json.Unmarshal(data, v); err != nil {
		serializeResult := &SerializeResult{
			Size:     int64(len(data)),
			Duration: time.Since(startTime),
			Success:  false,
			Error:    err.Error(),
		}
		return serializeResult, errors.WithMessage(err, "反序列化JSON数据失败")
	}

	serializeResult := &SerializeResult{
		Size:     int64(len(data)),
		Duration: time.Since(startTime),
		Success:  true,
	}

	return serializeResult, nil
}
