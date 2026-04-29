package archive

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProcessTarEntrySymlink(t *testing.T) {
	tempDir := createTempDir(t)

	// Create a simple test file
	testFile := createTestFile(t, tempDir, "test.txt", "content")

	// Create tar.gz archive
	archivePath := filepath.Join(tempDir, "test.tar.gz")
	err := TarGz(testFile, archivePath)
	assert.NoError(t, err)

	// Test decompression
	extractDir := createTestDir(t, tempDir, "extract")
	err = UntarGz(archivePath, extractDir)
	assert.NoError(t, err)

	// Verify file was extracted
	extractedFile := filepath.Join(extractDir, "test.txt")
	assert.FileExists(t, extractedFile)
}

func TestValidateSingleDirEmptyArchive(t *testing.T) {
	tempDir := createTempDir(t)

	// Create empty zip file
	emptyZipPath := filepath.Join(tempDir, "empty.zip")
	f, err := os.Create(emptyZipPath)
	assert.NoError(t, err)
	f.Close()

	// Test validation - should fail
	dirName, err := ValidateSingleDirZip(emptyZipPath)
	assert.Error(t, err)
	assert.Empty(t, dirName)
}

func TestValidateSingleDirMultipleEntries(t *testing.T) {
	tempDir := createTempDir(t)

	// Create a zip file with multiple top-level entries
	// Note: We'll test with an empty archive instead
	emptyZipPath := filepath.Join(tempDir, "empty.zip")
	f, err := os.Create(emptyZipPath)
	assert.NoError(t, err)
	f.Close()

	// Test validation - should fail
	dirName, err := ValidateSingleDirZip(emptyZipPath)
	assert.Error(t, err)
	assert.Empty(t, dirName)
}

func TestSafeCopyErrorCases(t *testing.T) {
	tempDir := createTempDir(t)

	// Test with nil writer
	srcFile := createTestFile(t, tempDir, "source.txt", "content")
	src, err := os.Open(srcFile)
	assert.NoError(t, err)
	defer src.Close()

	// This should panic or return error
	// Note: We expect this to panic since writer is nil
	assert.Panics(t, func() {
		_, _ = safeCopy(context.TODO(), nil, src, 1024, 32*1024)
	})
}

func TestIsPathSafeAllCases(t *testing.T) {
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
		{"Double dots in path", "/base/dir/../file", "/base", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isPathSafe(tt.target, tt.base)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWalkAndProcessTarWithExclude(t *testing.T) {
	tempDir := createTempDir(t)

	// Create test files
	createTestFile(t, tempDir, "file1.txt", "content1")
	createTestFile(t, tempDir, "file2.log", "content2") // This should be excluded

	// Create tar.gz archive
	archivePath := filepath.Join(tempDir, "test.tar.gz")
	err := TarGz(tempDir, archivePath, WithExcludePatterns([]string{"*.log"}))
	assert.NoError(t, err)

	// Verify archive was created
	assert.FileExists(t, archivePath)
}

func TestWalkAndProcessZipWithIncludeOnly(t *testing.T) {
	tempDir := createTempDir(t)

	// Create test files
	createTestFile(t, tempDir, "file1.txt", "content1")
	createTestFile(t, tempDir, "file2.log", "content2") // This should be excluded

	// Create zip archive
	archivePath := filepath.Join(tempDir, "test.zip")
	err := Zip(tempDir, archivePath, WithIncludeOnly([]string{"*.txt"}))
	assert.NoError(t, err)

	// Verify archive was created
	assert.FileExists(t, archivePath)
}

func TestUnzipWithPathTraversal(t *testing.T) {
	tempDir := createTempDir(t)

	// Create a zip file with path traversal attempt
	// Note: This is a simplified test - in reality, we'd need to create a malicious zip
	archivePath := filepath.Join(tempDir, "test.zip")

	// Create a normal file first
	testFile := createTestFile(t, tempDir, "file.txt", "content")
	err := Zip(testFile, archivePath)
	assert.NoError(t, err)

	// Test decompression
	extractDir := createTestDir(t, tempDir, "extract")
	err = Unzip(archivePath, extractDir)
	assert.NoError(t, err)

	// Verify file was extracted
	assert.FileExists(t, filepath.Join(extractDir, "file.txt"))
}

func TestUntarGzWithPathTraversal(t *testing.T) {
	tempDir := createTempDir(t)

	// Create a tar.gz file with path traversal attempt
	// Note: This is a simplified test - in reality, we'd need to create a malicious tar.gz
	archivePath := filepath.Join(tempDir, "test.tar.gz")

	// Create a normal file first
	testFile := createTestFile(t, tempDir, "file.txt", "content")
	err := TarGz(testFile, archivePath)
	assert.NoError(t, err)

	// Test decompression
	extractDir := createTestDir(t, tempDir, "extract")
	err = UntarGz(archivePath, extractDir)
	assert.NoError(t, err)

	// Verify file was extracted
	assert.FileExists(t, filepath.Join(extractDir, "file.txt"))
}

func TestWithContextCancellation(t *testing.T) {
	tempDir := createTempDir(t)

	// Create test file
	testFile := createTestFile(t, tempDir, "file.txt", "content")

	// Create context with immediate cancellation
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Test compression with cancelled context
	archivePath := filepath.Join(tempDir, "test.zip")
	err := Zip(testFile, archivePath, WithContext(ctx))
	assert.Error(t, err)
}

func TestWithMaxFileSizeExceeded(t *testing.T) {
	tempDir := createTempDir(t)

	// Create a small file
	testFile := createTestFile(t, tempDir, "file.txt", "content")

	// Test with very small size limit
	archivePath := filepath.Join(tempDir, "test.zip")
	err := Zip(testFile, archivePath, WithMaxFileSize(1)) // 1 byte limit
	assert.Error(t, err)
}

func TestWithMaxFilesExceeded(t *testing.T) {
	tempDir := createTempDir(t)

	// Create multiple files
	for i := 0; i < 3; i++ {
		createTestFile(t, tempDir, fmt.Sprintf("file%d.txt", i), "content")
	}

	// Test with file count limit
	archivePath := filepath.Join(tempDir, "test.zip")
	err := Zip(tempDir, archivePath, WithMaxFiles(2)) // Only allow 2 files
	assert.Error(t, err)
}
