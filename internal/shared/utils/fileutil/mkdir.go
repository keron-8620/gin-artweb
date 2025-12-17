package fileutil

import (
	"fmt"
	"os"
	"path/filepath"
)

// Mkdir creates a directory named path with the specified permission bits.
// If the directory already exists, it does nothing and returns nil.
func Mkdir(path string, perm os.FileMode) error {
	// Validate input path
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	// Check if directory already exists
	if info, err := os.Stat(path); err == nil {
		if info.IsDir() {
			// Directory already exists
			return nil
		}
		// Path exists but is not a directory
		return fmt.Errorf("path exists but is not a directory: %s", path)
	} else if !os.IsNotExist(err) {
		// Other stat error
		return fmt.Errorf("failed to check path: %w", err)
	}

	// Create the directory
	if err := os.Mkdir(path, perm); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	return nil
}

// MkdirAll creates a directory named path along with any necessary parents,
// with the specified permission bits.
// If the directory already exists, it does nothing and returns nil.
func MkdirAll(path string, perm os.FileMode) error {
	// Validate input path
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	// Clean the path
	path = filepath.Clean(path)

	// Check if directory already exists
	if info, err := os.Stat(path); err == nil {
		if info.IsDir() {
			// Directory already exists
			return nil
		}
		// Path exists but is not a directory
		return fmt.Errorf("path exists but is not a directory: %s", path)
	} else if !os.IsNotExist(err) {
		// Other stat error
		return fmt.Errorf("failed to check path: %w", err)
	}

	// Create the directory and any necessary parents
	if err := os.MkdirAll(path, perm); err != nil {
		return fmt.Errorf("failed to create directory path: %w", err)
	}

	return nil
}

// MkdirWithParents creates a directory with specified permissions along with its parent directories.
// This is an alias for MkdirAll for convenience.
func MkdirWithParents(path string, perm os.FileMode) error {
	return MkdirAll(path, perm)
}

// MkdirTemp creates a new temporary directory in the directory dir with a name beginning with prefix
// and returns the path of the new directory.
// If dir is the empty string, TempDir uses the default directory for temporary files (see os.TempDir).
// Multiple programs calling MkdirTemp simultaneously will not choose the same directory.
func MkdirTemp(dir, prefix string) (string, error) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp(dir, prefix)
	if err != nil {
		return "", fmt.Errorf("failed to create temporary directory: %w", err)
	}

	return tempDir, nil
}

// EnsureDir ensures that the directory containing the given file path exists.
// This is useful when you want to ensure the parent directory of a file exists before creating the file.
func EnsureDir(filePath string) error {
	// Validate input path
	if filePath == "" {
		return fmt.Errorf("file path cannot be empty")
	}

	// Get the directory part of the file path
	dir := filepath.Dir(filePath)

	// Create directory with default permissions if it doesn't exist
	return MkdirAll(dir, 0755)
}