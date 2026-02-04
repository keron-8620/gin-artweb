package fileutil

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidatePath(t *testing.T) {
	ctx := context.Background()

	// 测试空路径
	err := ValidatePath(ctx, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "路径不能为空")

	// 测试空白路径
	err = ValidatePath(ctx, "   ")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "路径不能为空")

	// 测试安全路径
	err = ValidatePath(ctx, "/tmp/test")
	assert.NoError(t, err)

	// 测试包含..的路径（预期会失败，因为isPathSafe可能返回false）
	err = ValidatePath(ctx, "./../test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "路径包含不安全的父目录引用")
}

func TestGetFileInfo(t *testing.T) {
	ctx := context.Background()

	// 创建临时文件用于测试
	tempFile, err := os.CreateTemp("", "test")
	require.NoError(t, err)
	tempFilePath := tempFile.Name()
	defer os.Remove(tempFilePath)
	tempFile.Close()

	// 测试获取存在文件的信息
	info, err := GetFileInfo(ctx, tempFilePath)
	assert.NoError(t, err)
	assert.NotNil(t, info)
	assert.False(t, info.IsDir())

	// 测试获取不存在文件的信息
	info, err = GetFileInfo(ctx, "/tmp/nonexistent_file")
	assert.Error(t, err)
	assert.Nil(t, info)
}

func TestListFileInfo(t *testing.T) {
	ctx := context.Background()

	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// 创建文件
	testFile, err := os.CreateTemp(tempDir, "test")
	require.NoError(t, err)
	_, err = testFile.WriteString("test content")
	require.NoError(t, err)
	testFile.Close()

	// 创建子目录
	subdir := filepath.Join(tempDir, "subdir")
	err = os.Mkdir(subdir, 0755)
	require.NoError(t, err)

	// 在子目录中创建文件
	subFile, err := os.CreateTemp(subdir, "test")
	require.NoError(t, err)
	subFile.Close()

	// 测试获取文件信息
	info, err := ListFileInfo(ctx, tempDir)
	assert.NoError(t, err)
	assert.NotNil(t, info)
	assert.True(t, info.IsDir)
	assert.Greater(t, len(info.Children), 0)
}
