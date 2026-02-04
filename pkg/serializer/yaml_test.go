package serializer

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
)

func TestReadYAML(t *testing.T) {
	// 创建临时测试文件
	tempFile, err := os.CreateTemp("", "test-*.yaml")
	assert.NoError(t, err)
	defer os.Remove(tempFile.Name())

	// 写入测试数据
	testData := map[string]string{"key": "value"}
	data, err := yaml.Marshal(testData)
	assert.NoError(t, err)
	err = os.WriteFile(tempFile.Name(), data, 0644)
	assert.NoError(t, err)

	// 测试正常读取
	var result map[string]string
	readResult, err := ReadYAML(tempFile.Name(), &result)
	assert.NoError(t, err)
	assert.NotNil(t, readResult)
	assert.Equal(t, tempFile.Name(), readResult.FilePath)
	assert.Greater(t, readResult.Size, int64(0))
	assert.Greater(t, readResult.Duration, time.Duration(0))
	assert.True(t, readResult.Success)
	assert.Equal(t, "value", result["key"])

	// 测试文件不存在
	_, err = ReadYAML("non-existent.yaml", &result)
	assert.Error(t, err)

	// 测试空文件
	emptyFile, err := os.CreateTemp("", "empty-*.yaml")
	assert.NoError(t, err)
	defer os.Remove(emptyFile.Name())

	_, err = ReadYAML(emptyFile.Name(), &result)
	assert.Error(t, err)

	// 测试上下文取消
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = ReadYAML(tempFile.Name(), &result, WithContext(ctx))
	assert.Error(t, err)
}

func TestWriteYAML(t *testing.T) {
	// 创建临时文件路径
	tempFile := filepath.Join(os.TempDir(), "test-write-"+time.Now().Format("20060102150405")+".yaml")
	defer os.Remove(tempFile)

	// 测试正常写入
	testData := map[string]string{"key": "value"}
	writeResult, err := WriteYAML(tempFile, testData)
	assert.NoError(t, err)
	assert.NotNil(t, writeResult)
	assert.Equal(t, tempFile, writeResult.FilePath)
	assert.Greater(t, writeResult.Size, int64(0))
	assert.Greater(t, writeResult.Duration, time.Duration(0))
	assert.True(t, writeResult.Success)

	// 验证文件内容
	var result map[string]string
	data, err := os.ReadFile(tempFile)
	assert.NoError(t, err)
	err = yaml.Unmarshal(data, &result)
	assert.NoError(t, err)
	assert.Equal(t, "value", result["key"])

	// 测试原子写入
	atomicFile := filepath.Join(os.TempDir(), "test-atomic-"+time.Now().Format("20060102150405")+".yaml")
	defer os.Remove(atomicFile)

	writeResult, err = WriteYAML(atomicFile, testData, WithAtomic(true))
	assert.NoError(t, err)
	assert.NotNil(t, writeResult)
	assert.Equal(t, atomicFile, writeResult.FilePath)

	// 测试上下文取消
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = WriteYAML(tempFile, testData, WithContext(ctx))
	assert.Error(t, err)
}

func TestReadYAMLWithTimeout(t *testing.T) {
	// 创建临时测试文件
	tempFile, err := os.CreateTemp("", "test-timeout-*.yaml")
	assert.NoError(t, err)
	defer os.Remove(tempFile.Name())

	// 写入测试数据
	testData := map[string]string{"key": "value"}
	data, err := yaml.Marshal(testData)
	assert.NoError(t, err)
	err = os.WriteFile(tempFile.Name(), data, 0644)
	assert.NoError(t, err)

	// 测试正常读取
	var result map[string]string
	readResult, err := ReadYAMLWithTimeout(tempFile.Name(), &result, 1*time.Second)
	assert.NoError(t, err)
	assert.NotNil(t, readResult)
}

func TestWriteYAMLWithTimeout(t *testing.T) {
	// 创建临时文件路径
	tempFile := filepath.Join(os.TempDir(), "test-write-timeout-"+time.Now().Format("20060102150405")+".yaml")
	defer os.Remove(tempFile)

	// 测试正常写入
	testData := map[string]string{"key": "value"}
	writeResult, err := WriteYAMLWithTimeout(tempFile, testData, 1*time.Second)
	assert.NoError(t, err)
	assert.NotNil(t, writeResult)
}
