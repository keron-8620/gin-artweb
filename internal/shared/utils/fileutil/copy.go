package fileutil

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// CopyFile copies a single file from src to dst.
// It preserves the file permissions and modification time.
func CopyFile(src, dst string) error {
	// Validate input paths
	if src == "" {
		return fmt.Errorf("source path cannot be empty")
	}
	if dst == "" {
		return fmt.Errorf("destination path cannot be empty")
	}

	// Check if source file exists
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to get source file info: %w", err)
	}

	// Check if source is actually a file
	if srcInfo.IsDir() {
		return fmt.Errorf("source is a directory, not a file")
	}

	// Handle case where destination is an existing directory
	if dstInfo, err := os.Stat(dst); err == nil {
		if os.SameFile(srcInfo, dstInfo) {
			return nil // Same file, nothing to do
		}
		if dstInfo.IsDir() {
			dst = filepath.Join(dst, filepath.Base(src))
		}
	} else if os.IsNotExist(err) {
		// Destination doesn't exist, check if parent directory exists
		if dirInfo, err := os.Stat(filepath.Dir(dst)); err == nil && dirInfo.IsDir() {
			// Parent directory exists, this is fine
		} else if os.IsNotExist(err) {
			// Need to create parent directory
		} else {
			return fmt.Errorf("failed to check parent directory: %w", err)
		}
	} else {
		return fmt.Errorf("failed to check destination: %w", err)
	}

	// Ensure destination directory exists
	dstDir := filepath.Dir(dst)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	// Create destination file
	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	// Copy file content
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	// Sync to disk
	if err := dstFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync destination file: %w", err)
	}

	// Preserve modification time
	if err := os.Chtimes(dst, time.Now(), srcInfo.ModTime()); err != nil {
		return fmt.Errorf("failed to preserve modification time: %w", err)
	}

	return nil
}

// CopyDir recursively copies a directory tree from src to dst.
// It preserves file permissions, modification times and directory structure.
func CopyDir(src, dst string) error {
	// Validate input paths
	if src == "" {
		return fmt.Errorf("source path cannot be empty")
	}
	if dst == "" {
		return fmt.Errorf("destination path cannot be empty")
	}

	// Get source directory info
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to get source directory info: %w", err)
	}

	// Check if source is actually a directory
	if !srcInfo.IsDir() {
		return fmt.Errorf("source is a file, not a directory")
	}

	// Handle case where destination is an existing directory
	if dstInfo, err := os.Stat(dst); err == nil {
		if os.SameFile(srcInfo, dstInfo) {
			return nil // Same directory, nothing to do
		}
		if dstInfo.IsDir() {
			dst = filepath.Join(dst, filepath.Base(src))
		}
	} else if os.IsNotExist(err) {
		// Destination doesn't exist, which is fine
	} else {
		return fmt.Errorf("failed to check destination: %w", err)
	}

	// Create destination directory with same permissions
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Read directory entries
	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("failed to read source directory: %w", err)
	}

	// Process each entry
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		// Get file info
		entryInfo, err := entry.Info()
		if err != nil {
			return fmt.Errorf("failed to get entry info for %s: %w", entry.Name(), err)
		}

		if entryInfo.IsDir() {
			// Recursively copy subdirectory
			if err := CopyDir(srcPath, dstPath); err != nil {
				return fmt.Errorf("failed to copy directory %s: %w", entry.Name(), err)
			}
		} else {
			// Copy regular file
			if err := CopyFile(srcPath, dstPath); err != nil {
				return fmt.Errorf("failed to copy file %s: %w", entry.Name(), err)
			}
		}
	}

	// Preserve directory modification time
	if err := os.Chtimes(dst, time.Now(), srcInfo.ModTime()); err != nil {
		return fmt.Errorf("failed to preserve directory modification time: %w", err)
	}

	return nil
}

// Copy copies either a file or directory from src to dst.
// It automatically detects the type and uses appropriate function.
func Copy(src, dst string) error {
	// Validate input paths
	if src == "" {
		return fmt.Errorf("source path cannot be empty")
	}
	if dst == "" {
		return fmt.Errorf("destination path cannot be empty")
	}

	// Get source info
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to get source info: %w", err)
	}

	// Choose appropriate copy function based on source type
	if srcInfo.IsDir() {
		return CopyDir(src, dst)
	}
	return CopyFile(src, dst)
}