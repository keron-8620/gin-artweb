package archive

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewArchiver(t *testing.T) {
	tests := []struct {
		name     string
		format   ArchiveFormat
		expected bool
		error    bool
	}{
		{"ZIP format", FormatZip, true, false},
		{"TAR.GZ format", FormatTarGz, true, false},
		{"Invalid format", "invalid", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			archiver, err := NewArchiver(tt.format)
			if tt.error {
				assert.Error(t, err)
				assert.Nil(t, archiver)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, archiver)
			}
		})
	}
}

func TestArchiverInterface(t *testing.T) {
	formats := []ArchiveFormat{FormatZip, FormatTarGz}
	for _, format := range formats {
		t.Run(string(format), func(t *testing.T) {
			archiver, err := NewArchiver(format)
			assert.NoError(t, err)
			assert.NotNil(t, archiver)

			// Test Compress method exists
			assert.NotPanics(t, func() {
				_ = archiver.Compress
			})

			// Test Decompress method exists
			assert.NotPanics(t, func() {
				_ = archiver.Decompress
			})

			// Test ValidateSingleDir method exists
			assert.NotPanics(t, func() {
				_ = archiver.ValidateSingleDir
			})
		})
	}
}

// Test helper functions
func createTempDir(t *testing.T) string {
	tempDir, err := os.MkdirTemp("", "archive-test-")
	assert.NoError(t, err)
	t.Cleanup(func() {
		os.RemoveAll(tempDir)
	})
	return tempDir
}

func createTestFile(t *testing.T, dir, name, content string) string {
	filePath := filepath.Join(dir, name)
	err := os.WriteFile(filePath, []byte(content), 0644)
	assert.NoError(t, err)
	return filePath
}

func createTestDir(t *testing.T, dir, subdir string) string {
	subDirPath := filepath.Join(dir, subdir)
	err := os.MkdirAll(subDirPath, 0755)
	assert.NoError(t, err)
	return subDirPath
}
