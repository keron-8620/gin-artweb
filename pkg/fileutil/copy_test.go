package fileutil

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCopyFile(t *testing.T) {
	ctx := context.Background()

	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// 创建源文件
	srcFile, err := os.CreateTemp(tempDir, "src")
	require.NoError(t, err)
	srcFilePath := srcFile.Name()
	srcContent := "test content"
	_, err = srcFile.WriteString(srcContent)
	require.NoError(t, err)
	srcFile.Close()

	// 测试复制到新文件
	dstFilePath := filepath.Join(tempDir, "dst.txt")
	err = CopyFile(ctx, srcFilePath, dstFilePath)
	assert.NoError(t, err)

	// 验证目标文件存在且内容正确
	dstContent, err := os.ReadFile(dstFilePath)
	assert.NoError(t, err)
	assert.Equal(t, srcContent, string(dstContent))

	// 测试复制到目录
	subdirPath := filepath.Join(tempDir, "subdir")
	err = os.Mkdir(subdirPath, 0755)
	require.NoError(t, err)

	err = CopyFile(ctx, srcFilePath, subdirPath)
	assert.NoError(t, err)

	// 验证文件已复制到目录
	copiedFilePath := filepath.Join(subdirPath, filepath.Base(srcFilePath))
	exists, err := pathExists(copiedFilePath)
	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestCopyDir(t *testing.T) {
	ctx := context.Background()

	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// 创建源目录结构
	srcDir := filepath.Join(tempDir, "src")
	err = os.Mkdir(srcDir, 0755)
	require.NoError(t, err)

	// 在源目录中创建文件
	srcFile, err := os.CreateTemp(srcDir, "test")
	require.NoError(t, err)
	_, err = srcFile.WriteString("test content")
	require.NoError(t, err)
	srcFile.Close()

	// 创建子目录
	srcSubdir := filepath.Join(srcDir, "subdir")
	err = os.Mkdir(srcSubdir, 0755)
	require.NoError(t, err)

	// 在子目录中创建文件
	srcSubfile, err := os.CreateTemp(srcSubdir, "test")
	require.NoError(t, err)
	_, err = srcSubfile.WriteString("subdir content")
	require.NoError(t, err)
	srcSubfile.Close()

	// 测试复制目录（不包含内容）
	dstDir := filepath.Join(tempDir, "dst")
	err = CopyDir(ctx, srcDir, dstDir, false)
	assert.NoError(t, err)

	// 验证目标目录结构
	// 当copyContents为false且dst不存在时，dest等于dst，而不是filepath.Join(dst, filepath.Base(src))
	exists, err := pathExists(dstDir)
	assert.NoError(t, err)
	assert.True(t, exists)

	// 测试复制目录到已存在的目录
	existingDir := filepath.Join(tempDir, "existing")
	err = os.Mkdir(existingDir, 0755)
	require.NoError(t, err)

	err = CopyDir(ctx, srcDir, existingDir, false)
	assert.NoError(t, err)

	// 验证目录已复制到已存在的目录下
	expectedSubDir := filepath.Join(existingDir, "src")
	exists, err = pathExists(expectedSubDir)
	assert.NoError(t, err)
	assert.True(t, exists)

	// 测试复制目录内容
	dstContentDir := filepath.Join(tempDir, "dst_content")
	err = os.Mkdir(dstContentDir, 0755)
	require.NoError(t, err)

	err = CopyDir(ctx, srcDir, dstContentDir, true)
	assert.NoError(t, err)

	// 验证目标目录中包含源目录的内容
	expectedFile := filepath.Join(dstContentDir, filepath.Base(srcFile.Name()))
	exists, err = pathExists(expectedFile)
	assert.NoError(t, err)
	assert.True(t, exists)
}
