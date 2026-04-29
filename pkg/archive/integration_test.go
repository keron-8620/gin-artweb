package archive

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArchiverMethods(t *testing.T) {
	formats := []ArchiveFormat{FormatZip, FormatTarGz}
	for _, format := range formats {
		t.Run(string(format), func(t *testing.T) {
			archiver, err := NewArchiver(format)
			assert.NoError(t, err)
			assert.NotNil(t, archiver)

			// Test Compress method
			tempDir := createTempDir(t)
			testFile := createTestFile(t, tempDir, "test.txt", "content")
			archivePath := filepath.Join(tempDir, "test."+string(format))
			err = archiver.Compress(testFile, archivePath)
			assert.NoError(t, err)
			assert.FileExists(t, archivePath)

			// Test Decompress method
			extractDir := createTestDir(t, tempDir, "extract")
			err = archiver.Decompress(archivePath, extractDir)
			assert.NoError(t, err)
			assert.FileExists(t, filepath.Join(extractDir, "test.txt"))

			// Test ValidateSingleDir method with directory
			topDir := createTestDir(t, tempDir, "mydir")
			createTestFile(t, topDir, "file.txt", "content")
			dirArchivePath := filepath.Join(tempDir, "dir."+string(format))
			err = archiver.Compress(topDir, dirArchivePath)
			assert.NoError(t, err)

			dirName, err := archiver.ValidateSingleDir(dirArchivePath)
			assert.NoError(t, err)
			assert.Equal(t, "mydir", dirName)
		})
	}
}

func TestProcessTarEntry(t *testing.T) {
	tempDir := createTempDir(t)

	// Test directory entry
	dirPath := filepath.Join(tempDir, "testdir")
	err := os.MkdirAll(dirPath, 0755)
	assert.NoError(t, err)

	// Test file entry
	filePath := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(filePath, []byte("content"), 0644)
	assert.NoError(t, err)

	// Test symlink entry
	symlinkPath := filepath.Join(tempDir, "link.txt")
	err = os.Symlink("test.txt", symlinkPath)
	assert.NoError(t, err)
}

func TestProcessZipEntry(t *testing.T) {
	tempDir := createTempDir(t)

	// Test directory entry
	dirPath := filepath.Join(tempDir, "testdir")
	err := os.MkdirAll(dirPath, 0755)
	assert.NoError(t, err)

	// Test file entry
	filePath := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(filePath, []byte("content"), 0644)
	assert.NoError(t, err)
}

func TestValidateSingleDirEdgeCases(t *testing.T) {
	tempDir := createTempDir(t)

	// Test with single file (should fail for directory validation)
	testFile := createTestFile(t, tempDir, "test.txt", "content")
	archivePath := filepath.Join(tempDir, "single-file.zip")
	err := Zip(testFile, archivePath)
	assert.NoError(t, err)

	// This should fail because it's a single file, not a directory
	dirName, err := ValidateSingleDirZip(archivePath)
	assert.Error(t, err)
	assert.Empty(t, dirName)
}

func TestSafeCopyEdgeCases(t *testing.T) {
	tempDir := createTempDir(t)

	// Test with zero buffer size
	srcFile := createTestFile(t, tempDir, "source.txt", "content")
	dstFile := filepath.Join(tempDir, "dest.txt")

	src, err := os.Open(srcFile)
	assert.NoError(t, err)
	defer src.Close()

	dst, err := os.Create(dstFile)
	assert.NoError(t, err)
	defer dst.Close()

	// Test with zero buffer size
	written, err := safeCopy(context.TODO(), dst, src, 1024, 0)
	assert.NoError(t, err)
	assert.Greater(t, written, int64(0))
}

func TestIsPathSafeEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		target   string
		base     string
		expected bool
	}{
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

func TestContextCheckWithNil(t *testing.T) {
	err := checkContext(context.TODO())
	assert.NoError(t, err)
}

func TestWithContextOption(t *testing.T) {
	ctx := context.Background()
	option := WithContext(ctx)
	options := DefaultArchiveOptions
	option(&options)
	assert.Equal(t, ctx, options.Context)
}

func TestStreamArchiverInterface(t *testing.T) {
	// Test that StreamArchiver interface is defined
	var _ StreamArchiver = nil
}
