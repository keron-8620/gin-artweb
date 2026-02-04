package serializer

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestReadJSON(t *testing.T) {
	// 创建临时测试文件
	tempFile, err := os.CreateTemp("", "test-*.json")
	assert.NoError(t, err)
	defer os.Remove(tempFile.Name())

	// 写入测试数据
	testData := map[string]string{"key": "value"}
	data, err := json.Marshal(testData)
	assert.NoError(t, err)
	err = os.WriteFile(tempFile.Name(), data, 0644)
	assert.NoError(t, err)

	// 测试正常读取
	var result map[string]string
	readResult, err := ReadJSON(tempFile.Name(), &result)
	assert.NoError(t, err)
	assert.NotNil(t, readResult)
	assert.Equal(t, tempFile.Name(), readResult.FilePath)
	assert.Greater(t, readResult.Size, int64(0))
	assert.Greater(t, readResult.Duration, time.Duration(0))
	assert.True(t, readResult.Success)
	assert.Equal(t, "value", result["key"])

	// 测试文件不存在
	_, err = ReadJSON("non-existent.json", &result)
	assert.Error(t, err)

	// 测试空文件
	emptyFile, err := os.CreateTemp("", "empty-*.json")
	assert.NoError(t, err)
	defer os.Remove(emptyFile.Name())

	_, err = ReadJSON(emptyFile.Name(), &result)
	assert.Error(t, err)

	// 测试上下文取消
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = ReadJSON(tempFile.Name(), &result, WithContext(ctx))
	assert.Error(t, err)
}

func TestWriteJSON(t *testing.T) {
	// 创建临时文件路径
	tempFile := filepath.Join(os.TempDir(), "test-write-"+time.Now().Format("20060102150405")+".json")
	defer os.Remove(tempFile)

	// 测试正常写入
	testData := map[string]string{"key": "value"}
	writeResult, err := WriteJSON(tempFile, testData)
	assert.NoError(t, err)
	assert.NotNil(t, writeResult)
	assert.Equal(t, tempFile, writeResult.FilePath)
	assert.Greater(t, writeResult.Size, int64(0))
	assert.Greater(t, writeResult.Duration, time.Duration(0))
	assert.True(t, writeResult.Success)

	// 验证文件内容
	var result map[string]string
	err = json.Unmarshal(mustReadFile(t, tempFile), &result)
	assert.NoError(t, err)
	assert.Equal(t, "value", result["key"])

	// 测试原子写入
	atomicFile := filepath.Join(os.TempDir(), "test-atomic-"+time.Now().Format("20060102150405")+".json")
	defer os.Remove(atomicFile)

	writeResult, err = WriteJSON(atomicFile, testData, WithAtomic(true))
	assert.NoError(t, err)
	assert.NotNil(t, writeResult)
	assert.Equal(t, atomicFile, writeResult.FilePath)

	// 测试上下文取消
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = WriteJSON(tempFile, testData, WithContext(ctx))
	assert.Error(t, err)
}

func TestMarshalJSON(t *testing.T) {
	// 测试正常序列化
	testData := map[string]string{"key": "value"}
	data, result, err := MarshalJSON(testData)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Greater(t, result.Size, int64(0))
	assert.Greater(t, result.Duration, time.Duration(0))
	assert.True(t, result.Success)
	assert.NotEmpty(t, data)

	// 测试带缩进序列化
	data, result, err = MarshalJSON(testData, WithIndent(2))
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Greater(t, result.Size, int64(0))

	// 测试上下文取消
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, _, err = MarshalJSON(testData, WithContext(ctx))
	assert.Error(t, err)
}

func TestUnmarshalJSON(t *testing.T) {
	// 测试正常反序列化
	testData := map[string]string{"key": "value"}
	data, err := json.Marshal(testData)
	assert.NoError(t, err)

	var result map[string]string
	unmarshalResult, err := UnmarshalJSON(data, &result)
	assert.NoError(t, err)
	assert.NotNil(t, unmarshalResult)
	assert.Greater(t, unmarshalResult.Size, int64(0))
	assert.Greater(t, unmarshalResult.Duration, time.Duration(0))
	assert.True(t, unmarshalResult.Success)
	assert.Equal(t, "value", result["key"])

	// 测试上下文取消
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = UnmarshalJSON(data, &result, WithContext(ctx))
	assert.Error(t, err)

	// 测试参数验证
	_, err = UnmarshalJSON(nil, &result)
	assert.Error(t, err)

	_, err = UnmarshalJSON(data, nil)
	assert.Error(t, err)
}

func TestReadJSONWithTimeout(t *testing.T) {
	// 创建临时测试文件
	tempFile, err := os.CreateTemp("", "test-timeout-*.json")
	assert.NoError(t, err)
	defer os.Remove(tempFile.Name())

	// 写入测试数据
	testData := map[string]string{"key": "value"}
	data, err := json.Marshal(testData)
	assert.NoError(t, err)
	err = os.WriteFile(tempFile.Name(), data, 0644)
	assert.NoError(t, err)

	// 测试正常读取
	var result map[string]string
	readResult, err := ReadJSONWithTimeout(tempFile.Name(), &result, 1*time.Second)
	assert.NoError(t, err)
	assert.NotNil(t, readResult)
}

func TestWriteJSONWithTimeout(t *testing.T) {
	// 创建临时文件路径
	tempFile := filepath.Join(os.TempDir(), "test-write-timeout-"+time.Now().Format("20060102150405")+".json")
	defer os.Remove(tempFile)

	// 测试正常写入
	testData := map[string]string{"key": "value"}
	writeResult, err := WriteJSONWithTimeout(tempFile, testData, 1*time.Second, 2)
	assert.NoError(t, err)
	assert.NotNil(t, writeResult)
}

func mustReadFile(t *testing.T, path string) []byte {
	data, err := os.ReadFile(path)
	assert.NoError(t, err)
	return data
}
