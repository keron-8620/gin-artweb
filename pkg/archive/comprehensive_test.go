package archive

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProcessTarEntryComprehensive(t *testing.T) {
	tempDir := createTempDir(t)

	// Test symlink entry
	symlinkPath := filepath.Join(tempDir, "link.txt")

	// Create test file for symlink
	createTestFile(t, tempDir, "source.txt", "content")

	// Create symlink
	err := os.Symlink("source.txt", symlinkPath)
	assert.NoError(t, err)
}

func TestValidateTarSingleDirComprehensive(t *testing.T) {
	tempDir := createTempDir(t)

	// Test with single directory
	topDir := createTestDir(t, tempDir, "mydir")
	createTestFile(t, topDir, "file.txt", "content")

	// Create tar.gz archive
	archivePath := filepath.Join(tempDir, "test.tar.gz")
	err := TarGz(topDir, archivePath)
	assert.NoError(t, err)

	// Test validation
	dirName, err := ValidateSingleDirTarGz(archivePath)
	assert.NoError(t, err)
	assert.Equal(t, "mydir", dirName)
}

func TestValidateZipSingleDirComprehensive(t *testing.T) {
	tempDir := createTempDir(t)

	// Test with single directory
	topDir := createTestDir(t, tempDir, "mydir")
	createTestFile(t, topDir, "file.txt", "content")

	// Create zip archive
	archivePath := filepath.Join(tempDir, "test.zip")
	err := Zip(topDir, archivePath)
	assert.NoError(t, err)

	// Test validation
	dirName, err := ValidateSingleDirZip(archivePath)
	assert.NoError(t, err)
	assert.Equal(t, "mydir", dirName)
}

func TestUnzipComprehensive(t *testing.T) {
	tempDir := createTempDir(t)

	// Create test file
	testFile := createTestFile(t, tempDir, "test.txt", "content")

	// Create zip archive
	archivePath := filepath.Join(tempDir, "test.zip")
	err := Zip(testFile, archivePath)
	assert.NoError(t, err)

	// Test decompression
	extractDir := createTestDir(t, tempDir, "extract")
	err = Unzip(archivePath, extractDir)
	assert.NoError(t, err)

	// Verify extracted file
	assert.FileExists(t, filepath.Join(extractDir, "test.txt"))
}

func TestUntarGzComprehensive(t *testing.T) {
	tempDir := createTempDir(t)

	// Create test file
	testFile := createTestFile(t, tempDir, "test.txt", "content")

	// Create tar.gz archive
	archivePath := filepath.Join(tempDir, "test.tar.gz")
	err := TarGz(testFile, archivePath)
	assert.NoError(t, err)

	// Test decompression
	extractDir := createTestDir(t, tempDir, "extract")
	err = UntarGz(archivePath, extractDir)
	assert.NoError(t, err)

	// Verify extracted file
	assert.FileExists(t, filepath.Join(extractDir, "test.txt"))
}

func TestWalkAndProcessTar(t *testing.T) {
	tempDir := createTempDir(t)

	// Create test files
	createTestFile(t, tempDir, "file1.txt", "content1")
	subDir := createTestDir(t, tempDir, "subdir")
	createTestFile(t, subDir, "file2.txt", "content2")

	// Create tar.gz archive
	archivePath := filepath.Join(tempDir, "test.tar.gz")
	err := TarGz(tempDir, archivePath)
	assert.NoError(t, err)

	// Verify archive was created
	assert.FileExists(t, archivePath)
}

func TestWalkAndProcessZip(t *testing.T) {
	tempDir := createTempDir(t)

	// Create test files
	createTestFile(t, tempDir, "file1.txt", "content1")
	subDir := createTestDir(t, tempDir, "subdir")
	createTestFile(t, subDir, "file2.txt", "content2")

	// Create zip archive
	archivePath := filepath.Join(tempDir, "test.zip")
	err := Zip(tempDir, archivePath)
	assert.NoError(t, err)

	// Verify archive was created
	assert.FileExists(t, archivePath)
}

func TestProcessSingleFileTar(t *testing.T) {
	tempDir := createTempDir(t)

	// Create test file
	testFile := createTestFile(t, tempDir, "test.txt", "content")

	// Create tar.gz archive
	archivePath := filepath.Join(tempDir, "test.tar.gz")
	err := TarGz(testFile, archivePath)
	assert.NoError(t, err)

	// Verify archive was created
	assert.FileExists(t, archivePath)
}

func TestProcessSingleFileZip(t *testing.T) {
	tempDir := createTempDir(t)

	// Create test file
	testFile := createTestFile(t, tempDir, "test.txt", "content")

	// Create zip archive
	archivePath := filepath.Join(tempDir, "test.zip")
	err := Zip(testFile, archivePath)
	assert.NoError(t, err)

	// Verify archive was created
	assert.FileExists(t, archivePath)
}

func TestSafeCopyComprehensive(t *testing.T) {
	tempDir := createTempDir(t)

	// Test with different buffer sizes
	bufferSizes := []int{1024, 4096, 32768}

	for _, bufferSize := range bufferSizes {
		t.Run(fmt.Sprintf("buffer_size_%d", bufferSize), func(t *testing.T) {
			srcFile := createTestFile(t, tempDir, "source.txt", "content")
			dstFile := filepath.Join(tempDir, fmt.Sprintf("dest_%d.txt", bufferSize))

			src, err := os.Open(srcFile)
			assert.NoError(t, err)
			defer src.Close()

			dst, err := os.Create(dstFile)
			assert.NoError(t, err)
			defer dst.Close()

			written, err := safeCopy(context.TODO(), dst, src, 1024, bufferSize)
			assert.NoError(t, err)
			assert.Greater(t, written, int64(0))
		})
	}
}

func TestIsPathSafeComprehensive(t *testing.T) {
	tests := []struct {
		name     string
		target   string
		base     string
		expected bool
	}{
		{"Same directory", "/base/file", "/base", true},
		{"Subdirectory", "/base/sub/file", "/base", true},
		{"Parent directory", "/base/../file", "/base", false},
		{"Absolute path outside", "/other/file", "/base", false},
		{"Empty target", "", "/base", false},
		{"Empty base", "/file", "", false},
		{"Relative path with dots", "../file", "/base", false},
		{"Same directory with trailing slash", "/base/file", "/base/", true},
		{"Nested subdirectories", "/base/dir1/dir2/file", "/base", true},
		{"Parent directory with multiple dots", "/base/../../file", "/base", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isPathSafe(tt.target, tt.base)
			assert.Equal(t, tt.expected, result)
		})
	}
}
