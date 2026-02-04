package fileutil

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRemove(t *testing.T) {
	ctx := context.Background()

	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// 创建临时文件
	tempFile, err := os.CreateTemp(tempDir, "test")
	require.NoError(t, err)
	tempFilePath := tempFile.Name()
	tempFile.Close()

	// 测试删除文件
	err = Remove(ctx, tempFilePath)
	assert.NoError(t, err)

	// 验证文件已删除
	exists, err := pathExists(tempFilePath)
	assert.NoError(t, err)
	assert.False(t, exists)

	// 测试删除不存在的文件
	err = Remove(ctx, tempFilePath)
	assert.NoError(t, err)

	// 创建空目录
	emptyDir := filepath.Join(tempDir, "empty")
	err = os.Mkdir(emptyDir, 0755)
	require.NoError(t, err)

	// 测试删除空目录
	err = Remove(ctx, emptyDir)
	assert.NoError(t, err)

	// 验证目录已删除
	exists, err = pathExists(emptyDir)
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestRemoveAll(t *testing.T) {
	ctx := context.Background()

	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// 创建目录结构
	testDir := filepath.Join(tempDir, "test")
	err = os.Mkdir(testDir, 0755)
	require.NoError(t, err)

	// 在目录中创建文件
	testFile, err := os.CreateTemp(testDir, "test")
	require.NoError(t, err)
	testFile.Close()

	// 创建子目录
	subdir := filepath.Join(testDir, "subdir")
	err = os.Mkdir(subdir, 0755)
	require.NoError(t, err)

	// 在子目录中创建文件
	subFile, err := os.CreateTemp(subdir, "test")
	require.NoError(t, err)
	subFile.Close()

	// 测试递归删除目录
	err = RemoveAll(ctx, testDir)
	assert.NoError(t, err)

	// 验证目录已删除
	exists, err := pathExists(testDir)
	assert.NoError(t, err)
	assert.False(t, exists)

	// 测试删除不存在的目录
	err = RemoveAll(ctx, testDir)
	assert.NoError(t, err)
}

func TestSafeRemoveAll(t *testing.T) {
	ctx := context.Background()

	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// 测试删除安全路径
	testDir := filepath.Join(tempDir, "test")
	err = os.Mkdir(testDir, 0755)
	require.NoError(t, err)

	err = SafeRemoveAll(ctx, testDir)
	assert.NoError(t, err)

	// 验证目录已删除
	exists, err := pathExists(testDir)
	assert.NoError(t, err)
	assert.False(t, exists)

	// 测试删除系统路径（应该失败）
	err = SafeRemoveAll(ctx, "/")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "安全检查失败, 禁止删除系统核心路径")

	// 测试删除系统路径的子目录（应该失败）
	err = SafeRemoveAll(ctx, "/usr/test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "安全检查失败, 禁止删除系统路径子目录")
}

func TestRemoveEmptyDir(t *testing.T) {
	ctx := context.Background()

	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// 创建空目录
	emptyDir := filepath.Join(tempDir, "empty")
	err = os.Mkdir(emptyDir, 0755)
	require.NoError(t, err)

	// 测试删除空目录
	err = RemoveEmptyDir(ctx, emptyDir)
	assert.NoError(t, err)

	// 验证目录已删除
	exists, err := pathExists(emptyDir)
	assert.NoError(t, err)
	assert.False(t, exists)

	// 创建非空目录
	nonEmptyDir := filepath.Join(tempDir, "nonempty")
	err = os.Mkdir(nonEmptyDir, 0755)
	require.NoError(t, err)

	// 在目录中创建文件
	_, err = os.CreateTemp(nonEmptyDir, "test")
	require.NoError(t, err)

	// 测试删除非空目录（应该失败）
	err = RemoveEmptyDir(ctx, nonEmptyDir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "删除空目录失败（可能目录非空）")
}
