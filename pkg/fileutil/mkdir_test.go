package fileutil

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMkdir(t *testing.T) {
	ctx := context.Background()

	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// 测试创建新目录
	newDir := filepath.Join(tempDir, "newdir")
	err = Mkdir(ctx, newDir, 0755)
	assert.NoError(t, err)

	// 验证目录已创建
	exists, err := pathExists(newDir)
	assert.NoError(t, err)
	assert.True(t, exists)

	// 测试创建已存在的目录
	err = Mkdir(ctx, newDir, 0755)
	assert.NoError(t, err)

	// 测试创建目录但路径已存在且不是目录
	tempFile, err := os.CreateTemp(tempDir, "file")
	require.NoError(t, err)
	filePath := tempFile.Name()
	tempFile.Close()

	err = Mkdir(ctx, filePath, 0755)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "路径已存在但不是目录")
}

func TestMkdirAll(t *testing.T) {
	ctx := context.Background()

	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// 测试创建多级目录
	deepDir := filepath.Join(tempDir, "a", "b", "c")
	err = MkdirAll(ctx, deepDir, 0755)
	assert.NoError(t, err)

	// 验证目录已创建
	exists, err := pathExists(deepDir)
	assert.NoError(t, err)
	assert.True(t, exists)

	// 测试创建已存在的多级目录
	err = MkdirAll(ctx, deepDir, 0755)
	assert.NoError(t, err)
}

func TestEnsureDir(t *testing.T) {
	ctx := context.Background()

	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// 测试确保文件路径的父目录存在
	filePath := filepath.Join(tempDir, "subdir1", "subdir2", "file.txt")
	err = EnsureDir(ctx, filePath)
	assert.NoError(t, err)

	// 验证父目录已创建
	parentDir := filepath.Dir(filePath)
	exists, err := pathExists(parentDir)
	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestEnsureParentDir(t *testing.T) {
	ctx := context.Background()

	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// 测试确保不存在的父目录
	testPath := filepath.Join(tempDir, "subdir", "file.txt")
	err = EnsureParentDir(ctx, testPath, 0755)
	assert.NoError(t, err)

	// 验证父目录已创建
	subdirPath := filepath.Join(tempDir, "subdir")
	exists, err := pathExists(subdirPath)
	assert.NoError(t, err)
	assert.True(t, exists)

	// 测试确保已存在的父目录
	err = EnsureParentDir(ctx, testPath, 0755)
	assert.NoError(t, err)
}
