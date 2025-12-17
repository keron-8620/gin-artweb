package fileutil

import (
	"fmt"
	"os"
	"path/filepath"
)

// Move moves a file or directory from src to dst.
// If dst is an existing directory, src will be moved into that directory.
// If dst is a file or non-existent path, src will be renamed to dst.
func Move(src, dst string) error {
	// Validate input paths
	if src == "" {
		return fmt.Errorf("source path cannot be empty")
	}
	if dst == "" {
		return fmt.Errorf("destination path cannot be empty")
	}

	// Check if source exists
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to get source info: %w", err)
	}

	// Check if source and destination are the same
	if dstInfo, err := os.Stat(dst); err == nil {
		if os.SameFile(srcInfo, dstInfo) {
			// Source and destination are the same, nothing to do
			return nil
		}
		
		// If destination is a directory, move src into that directory
		if dstInfo.IsDir() {
			dst = filepath.Join(dst, filepath.Base(src))
			
			// Check again if they're the same after path adjustment
			if dstInfo, err := os.Stat(dst); err == nil {
				if os.SameFile(srcInfo, dstInfo) {
					return nil
				}
			}
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to check destination: %w", err)
	}

	// Ensure parent directory of destination exists
	dstDir := filepath.Dir(dst)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination parent directory: %w", err)
	}

	// Perform the move operation
	if err := os.Rename(src, dst); err != nil {
		return fmt.Errorf("failed to move %s to %s: %w", src, dst, err)
	}

	return nil
}

// MoveFile moves a file from src to dst.
// It behaves the same as Move but provides semantic clarity when working specifically with files.
func MoveFile(src, dst string) error {
	return Move(src, dst)
}

// MoveDir moves a directory from src to dst.
// It behaves the same as Move but provides semantic clarity when working specifically with directories.
func MoveDir(src, dst string) error {
	return Move(src, dst)
}