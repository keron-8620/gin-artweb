package fileutil

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Remove removes the named file or directory.
// If the path is a directory, it will be removed only if it's empty.
// If the path does not exist, Remove returns nil (no error).
func Remove(path string) error {
	// Validate input path
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	// Check if path exists
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			// Path doesn't exist, nothing to remove
			return nil
		}
		return fmt.Errorf("failed to check path: %w", err)
	}

	// Remove the file or empty directory
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to remove %s: %w", path, err)
	}

	return nil
}

// RemoveAll removes path and any children it contains.
// It removes everything it can but returns the first error it encounters.
// If the path does not exist, RemoveAll returns nil (no error).
func RemoveAll(path string) error {
	// Validate input path
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	// Check if path exists
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			// Path doesn't exist, nothing to remove
			return nil
		}
		return fmt.Errorf("failed to check path: %w", err)
	}

	// Remove path and all its contents
	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("failed to remove all %s: %w", path, err)
	}

	return nil
}

// RemoveIfExists removes the named file or directory if it exists.
// It's an alias for Remove for semantic clarity.
func RemoveIfExists(path string) error {
	return Remove(path)
}

// SafeRemoveAll removes path and any children it contains,
// but includes safety checks to prevent accidental deletion of important paths.
func SafeRemoveAll(path string) error {
	// Validate input path
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	// Resolve to absolute path for safety checks
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	// Safety checks to prevent accidental deletion of system paths
	unsafePaths := []string{
		"/",
		"/usr",
		"/usr/local",
		"/etc",
		"/var",
		"/lib",
		"/lib64",
		"/bin",
		"/sbin",
		"/boot",
		"/dev",
		"/proc",
		"/sys",
	}

	for _, unsafePath := range unsafePaths {
		if strings.HasPrefix(absPath, unsafePath) && absPath == unsafePath {
			return fmt.Errorf("refusing to remove protected system path: %s", absPath)
		}
	}

	// Also prevent removal of current working directory or parent paths
	cwd, err := os.Getwd()
	if err == nil {
		if absPath == cwd || strings.HasPrefix(cwd, absPath+"/") {
			return fmt.Errorf("refusing to remove current working directory or parent path: %s", absPath)
		}
	}

	// Check if path exists
	if _, err := os.Stat(absPath); err != nil {
		if os.IsNotExist(err) {
			// Path doesn't exist, nothing to remove
			return nil
		}
		return fmt.Errorf("failed to check path: %w", err)
	}

	// Remove path and all its contents
	if err := os.RemoveAll(absPath); err != nil {
		return fmt.Errorf("failed to safely remove all %s: %w", absPath, err)
	}

	return nil
}

// RemoveEmptyDir removes a directory only if it's empty.
// Returns an error if the path is not a directory or if it contains files.
func RemoveEmptyDir(path string) error {
	// Validate input path
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	// Check if path exists
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Path doesn't exist, nothing to remove
			return nil
		}
		return fmt.Errorf("failed to check path: %w", err)
	}

	// Check if path is a directory
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", path)
	}

	// Try to remove (this will fail if directory is not empty)
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to remove directory (may not be empty): %w", err)
	}

	return nil
}