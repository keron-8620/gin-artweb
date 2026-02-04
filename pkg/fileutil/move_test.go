package fileutil

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMove(t *testing.T) {
	ctx := context.Background()

	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// 创建源文件
	srcFile, err := os.CreateTemp(tempDir, "src")
	require.NoError(t, err)
	srcFilePath := srcFile.Name()
	_, err = srcFile.WriteString("test content")
	require.NoError(t, err)
	srcFile.Close()

	// 测试移动文件到新位置
	dstFilePath := filepath.Join(tempDir, "dst.txt")
	err = Move(ctx, srcFilePath, dstFilePath)
	assert.NoError(t, err)

	// 验证源文件不存在，目标文件存在
	srcExists, err := pathExists(srcFilePath)
	assert.NoError(t, err)
	assert.False(t, srcExists)

	dstExists, err := pathExists(dstFilePath)
	assert.NoError(t, err)
	assert.True(t, dstExists)

	// 测试移动文件到目录
	subdirPath := filepath.Join(tempDir, "subdir")
	err = os.Mkdir(subdirPath, 0755)
	require.NoError(t, err)

	err = Move(ctx, dstFilePath, subdirPath)
	assert.NoError(t, err)

	// 验证文件已移动到目录
	movedFilePath := filepath.Join(subdirPath, "dst.txt")
	movedExists, err := pathExists(movedFilePath)
	assert.NoError(t, err)
	assert.True(t, movedExists)
}
