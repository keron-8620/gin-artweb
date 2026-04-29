package archive

import (
	"archive/tar"
	"archive/zip"
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorTypes(t *testing.T) {
	// Test error types are defined
	errors := []error{
		ErrFileSizeExceeded,
		ErrFileCountExceeded,
		ErrInvalidPath,
		ErrEmptyArchive,
		ErrMultipleTopLevelEntries,
		ErrSingleEntryNotDir,
	}

	for _, err := range errors {
		assert.NotNil(t, err)
		assert.NotEmpty(t, err.Error())
	}
}

func TestApplyOptions(t *testing.T) {
	tests := []struct {
		name     string
		options  []ArchiveOption
		expected ArchiveOptions
	}{
		{
			name:     "Default options",
			options:  []ArchiveOption{},
			expected: DefaultArchiveOptions,
		},
		{
			name:    "Custom max file size",
			options: []ArchiveOption{WithMaxFileSize(1024 * 1024)},
			expected: ArchiveOptions{
				Context:             DefaultArchiveOptions.Context,
				MaxFileSize:         1024 * 1024,
				MaxFiles:            DefaultArchiveOptions.MaxFiles,
				ExcludePatterns:     DefaultArchiveOptions.ExcludePatterns,
				IncludeOnly:         DefaultArchiveOptions.IncludeOnly,
				FollowSymlinks:      DefaultArchiveOptions.FollowSymlinks,
				BufferSize:          DefaultArchiveOptions.BufferSize,
				MaxExtractPathDepth: DefaultArchiveOptions.MaxExtractPathDepth,
				AllowedExtensions:   DefaultArchiveOptions.AllowedExtensions,
			},
		},
		{
			name: "Multiple options",
			options: []ArchiveOption{
				WithMaxFiles(100),
				WithBufferSize(64 * 1024),
				WithFollowSymlinks(true),
			},
			expected: ArchiveOptions{
				Context:             DefaultArchiveOptions.Context,
				MaxFileSize:         DefaultArchiveOptions.MaxFileSize,
				MaxFiles:            100,
				ExcludePatterns:     DefaultArchiveOptions.ExcludePatterns,
				IncludeOnly:         DefaultArchiveOptions.IncludeOnly,
				FollowSymlinks:      true,
				BufferSize:          64 * 1024,
				MaxExtractPathDepth: DefaultArchiveOptions.MaxExtractPathDepth,
				AllowedExtensions:   DefaultArchiveOptions.AllowedExtensions,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := applyOptions(tt.options...)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSafeCopy(t *testing.T) {
	tempDir := createTempDir(t)

	// Create test file
	srcFile := createTestFile(t, tempDir, "source.txt", "test content")
	dstFile := filepath.Join(tempDir, "destination.txt")

	// Read source file
	src, err := os.Open(srcFile)
	assert.NoError(t, err)
	defer src.Close()

	// Create destination file
	dst, err := os.Create(dstFile)
	assert.NoError(t, err)
	defer dst.Close()

	// Test safeCopy
	written, err := safeCopy(context.TODO(), dst, src, 1024, 32*1024)
	assert.NoError(t, err)
	assert.Greater(t, written, int64(0))

	// Verify destination file was written
	assert.FileExists(t, dstFile)
}

func TestCleanArchiveName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Normal path", "dir/file.txt", "dir/file.txt"},
		{"Path with ./ prefix", "./dir/file.txt", "dir/file.txt"},
		{"Path with ../", "../dir/file.txt", "../dir/file.txt"}, // This should be caught by isPathSafe
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanArchiveName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractTopLevelName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Single file", "file.txt", "file.txt"},
		{"File in directory", "dir/file.txt", "dir"},
		{"Nested directories", "dir1/dir2/file.txt", "dir1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractTopLevelName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCleanZipName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Normal path", "dir/file.txt", "dir/file.txt"},
		{"Path with ./ prefix", "./dir/file.txt", "dir/file.txt"},
		{"Path with trailing slash", "dir/", "dir"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanZipName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractZipTopLevelName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Single file", "file.txt", "file.txt"},
		{"File in directory", "dir/file.txt", "dir"},
		{"Nested directories", "dir1/dir2/file.txt", "dir1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractZipTopLevelName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestZipFileHandler(t *testing.T) {
	tempDir := createTempDir(t)
	createTestFile(t, tempDir, "test.txt", "content")

	// Test that zipFileHandler type is correctly defined
	var handler zipFileHandler = func(path string, info os.FileInfo, header *zip.FileHeader, file *os.File, opts ArchiveOptions) error {
		return nil
	}
	assert.NotNil(t, handler)
}

func TestFileHandler(t *testing.T) {
	// Test that fileHandler type is correctly defined
	var handler fileHandler = func(path string, info os.FileInfo, header *tar.Header, reader io.Reader, opts ArchiveOptions) error {
		return nil
	}
	assert.NotNil(t, handler)
}

func TestWithExcludePatterns(t *testing.T) {
	option := WithExcludePatterns([]string{"*.txt", "*.log"})
	options := DefaultArchiveOptions
	option(&options)
	assert.Len(t, options.ExcludePatterns, 2)
	assert.Equal(t, "*.txt", options.ExcludePatterns[0])
	assert.Equal(t, "*.log", options.ExcludePatterns[1])
}

func TestWithIncludeOnly(t *testing.T) {
	option := WithIncludeOnly([]string{"*.go", "*.md"})
	options := DefaultArchiveOptions
	option(&options)
	assert.Len(t, options.IncludeOnly, 2)
	assert.Equal(t, "*.go", options.IncludeOnly[0])
	assert.Equal(t, "*.md", options.IncludeOnly[1])
}

func TestWithMax解压PathDepth(t *testing.T) {
	option := WithMax解压PathDepth(50)
	options := DefaultArchiveOptions
	option(&options)
	assert.Equal(t, 50, options.MaxExtractPathDepth)
}

func TestWithAllowedExtensions(t *testing.T) {
	option := WithAllowedExtensions([]string{".jpg", ".png"})
	options := DefaultArchiveOptions
	option(&options)
	assert.Len(t, options.AllowedExtensions, 2)
	assert.Equal(t, ".jpg", options.AllowedExtensions[0])
	assert.Equal(t, ".png", options.AllowedExtensions[1])
}
