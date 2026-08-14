package archive

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrFileSizeExceeded        = errors.New("file size exceeds limit")
	ErrFileCountExceeded       = errors.New("file count exceeds limit")
	ErrInvalidPath             = errors.New("invalid path")
	ErrEmptyArchive            = errors.New("archive is empty")
	ErrMultipleTopLevelEntries = errors.New("archive contains multiple top-level entries")
	ErrSingleEntryNotDir       = errors.New("single entry is not a directory")
)

type fileHandler func(path string, info os.FileInfo, header *tar.Header, reader io.Reader, opts ArchiveOptions) error

func TarGz(src, dst string, opts ...ArchiveOption) error {
	options := applyOptions(opts...)
	if err := checkContext(options.Context); err != nil {
		return fmt.Errorf("context check failed: %w", err)
	}
	if filepath.Base(dst) != dst && !filepath.IsAbs(dst) {
		return fmt.Errorf("%w: %s", ErrInvalidPath, dst)
	}
	dstFile, err := os.Create(dst) // #nosec G304 - path validated above
	if err != nil {
		return fmt.Errorf("failed to create destination file %s: %w", dst, err)
	}
	if err := dstFile.Close(); err != nil {
		return fmt.Errorf("failed to close destination file: %w", err)
	}
	f, err := os.OpenFile(dst, os.O_WRONLY, 0600) // #nosec G304 - path validated above
	if err != nil {
		return fmt.Errorf("failed to open destination file: %w", err)
	}
	defer f.Close()
	gzWriter := gzip.NewWriter(f)
	defer gzWriter.Close()
	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to stat source %s: %w", src, err)
	}
	if srcInfo.IsDir() {
		return walkAndProcessTar(src, options, func(path string, info os.FileInfo, header *tar.Header, reader io.Reader, opts ArchiveOptions) error {
			relPath, err := filepath.Rel(filepath.Dir(src), path)
			if err != nil {
				return fmt.Errorf("failed to calculate relative path for %s: %w", path, err)
			}
			if relPath == "." {
				return nil
			}
			header.Name = relPath
			if err := tarWriter.WriteHeader(header); err != nil {
				return fmt.Errorf("failed to write tar header for %s: %w", path, err)
			}
			if info.Mode().IsRegular() && reader != nil {
				if _, err := safeCopy(opts.Context, tarWriter, reader, opts.MaxFileSize, opts.BufferSize); err != nil {
					return fmt.Errorf("failed to copy content for %s: %w", path, err)
				}
			}
			return nil
		})
	}
	return processSingleFileTar(src, srcInfo, tarWriter, options)
}

func walkAndProcessTar(src string, options ArchiveOptions, handler fileHandler) error {
	fileCount := 0
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if err := checkContext(options.Context); err != nil {
			return fmt.Errorf("context check failed during walk: %w", err)
		}
		relPath, err := filepath.Rel(filepath.Dir(src), path)
		if err != nil {
			return fmt.Errorf("failed to calculate relative path for %s: %w", path, err)
		}
		if relPath == "." {
			return nil
		}
		fileCount++
		if options.MaxFiles > 0 && fileCount > options.MaxFiles {
			return fmt.Errorf("%w: %d", ErrFileCountExceeded, options.MaxFiles)
		}
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return fmt.Errorf("failed to create tar header for %s: %w", path, err)
		}
		var reader io.Reader
		if info.Mode().IsRegular() {
			if options.MaxFileSize > 0 && info.Size() > options.MaxFileSize {
				return fmt.Errorf("%w: %s size %d > limit %d", ErrFileSizeExceeded, path, info.Size(), options.MaxFileSize)
			}
			lstat, err := os.Lstat(path)
			if err != nil {
				return fmt.Errorf("failed to lstat file %s: %w", path, err)
			}
			if lstat.Mode()&os.ModeSymlink != 0 {
				return fmt.Errorf("symlink not allowed: %s", path)
			}
			file, err := os.Open(path) // #nosec G304,G122 - symlink checked above
			if err != nil {
				return fmt.Errorf("failed to open file %s: %w", path, err)
			}
			defer file.Close()
			reader = file
		}
		return handler(path, info, header, reader, options)
	})
}

func processSingleFileTar(src string, srcInfo os.FileInfo, tarWriter *tar.Writer, options ArchiveOptions) error {
	fileCount := 1
	if options.MaxFiles > 0 && fileCount > options.MaxFiles {
		return fmt.Errorf("%w: %d", ErrFileCountExceeded, options.MaxFiles)
	}
	if options.MaxFileSize > 0 && srcInfo.Size() > options.MaxFileSize {
		return fmt.Errorf("%w: %s size %d > limit %d", ErrFileSizeExceeded, src, srcInfo.Size(), options.MaxFileSize)
	}
	header, err := tar.FileInfoHeader(srcInfo, "")
	if err != nil {
		return fmt.Errorf("failed to create tar header for %s: %w", src, err)
	}
	header.Name = filepath.Base(src)
	if err := tarWriter.WriteHeader(header); err != nil {
		return fmt.Errorf("failed to write tar header for %s: %w", src, err)
	}
	file, err := os.Open(src) // #nosec G304 - src validated by caller
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", src, err)
	}
	defer file.Close()
	if _, err := safeCopy(options.Context, tarWriter, file, options.MaxFileSize, options.BufferSize); err != nil {
		return fmt.Errorf("failed to copy content for %s: %w", src, err)
	}
	return nil
}

func UntarGz(src, dst string, opts ...ArchiveOption) error {
	options := applyOptions(opts...)
	if err := checkContext(options.Context); err != nil {
		return fmt.Errorf("context check failed: %w", err)
	}
	srcFile, err := os.Open(src) // #nosec G304 - src validated by caller
	if err != nil {
		return fmt.Errorf("failed to open source file %s: %w", src, err)
	}
	defer srcFile.Close()
	gzReader, err := gzip.NewReader(srcFile)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzReader.Close()
	tarReader := tar.NewReader(gzReader)
	if err := os.MkdirAll(dst, 0750); err != nil {
		return fmt.Errorf("failed to create destination directory %s: %w", dst, err)
	}

	fileCount := 0
	for {
		if err := checkContext(options.Context); err != nil {
			return fmt.Errorf("context check failed during untar: %w", err)
		}
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}
		fileCount++
		if options.MaxFiles > 0 && fileCount > options.MaxFiles {
			return fmt.Errorf("%w: %d", ErrFileCountExceeded, options.MaxFiles)
		}
		cleanName := filepath.Clean(header.Name)
		parts := strings.SplitSeq(cleanName, string(filepath.Separator))
		for part := range parts {
			if part == ".." {
				return fmt.Errorf("%w: %s", ErrInvalidPath, header.Name)
			}
		}
		target := filepath.Join(dst, cleanName)
		if !isPathSafe(target, dst) {
			return fmt.Errorf("%w: %s", ErrInvalidPath, target)
		}
		if err := processTarEntry(target, dst, header, tarReader, options); err != nil {
			return err
		}
	}
	return nil
}

func processTarEntry(target, extractRoot string, header *tar.Header, tarReader *tar.Reader, options ArchiveOptions) error {
	switch header.Typeflag {
	case tar.TypeDir:
		mode := os.FileMode(header.Mode & 0755)
		if options.ExtractDirMode != 0 {
			mode = options.ExtractDirMode
		}
		if err := ensureExtractDir(target, mode, options.ExtractDirMode != 0); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", target, err)
		}
	case tar.TypeReg:
		if options.MaxFileSize > 0 && header.Size > options.MaxFileSize {
			return fmt.Errorf("%w: %s size %d > limit %d", ErrFileSizeExceeded, header.Name, header.Size, options.MaxFileSize)
		}
		parentDir := filepath.Dir(target)
		parentMode := os.FileMode(0750)
		if options.ExtractDirMode != 0 {
			parentMode = options.ExtractDirMode
		}
		if err := ensureExtractDir(parentDir, parentMode, options.ExtractDirMode != 0); err != nil {
			return fmt.Errorf("failed to create parent directory %s: %w", parentDir, err)
		}

		file, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, header.FileInfo().Mode()) // #nosec G304 - target validated by isPathSafe
		if err != nil {
			return fmt.Errorf("failed to create target file %s: %w", target, err)
		}
		defer file.Close()
		if _, err := safeCopy(options.Context, file, tarReader, options.MaxFileSize, options.BufferSize); err != nil {
			return fmt.Errorf("failed to copy content for %s: %w", header.Name, err)
		}
	case tar.TypeSymlink:
		if filepath.IsAbs(header.Linkname) {
			return fmt.Errorf("absolute symlink not allowed: %s -> %s", header.Name, header.Linkname)
		}
		cleanLinkName := filepath.Clean(header.Linkname)
		linkTarget := filepath.Join(filepath.Dir(target), cleanLinkName) // #nosec G305 - isPathSafe check below
		// A relative link may use .. to reference a sibling inside the archive.
		// Validate it against the extraction root, not the link's parent directory.
		if !isPathSafe(linkTarget, extractRoot) {
			return fmt.Errorf("symlink points outside directory: %s -> %s", header.Name, header.Linkname)
		}
		if !options.FollowSymlinks {
			if err := os.Symlink(header.Linkname, target); err != nil {
				return fmt.Errorf("failed to create symlink %s -> %s: %w", target, header.Linkname, err)
			}
		} else {
			return fmt.Errorf("following symlinks not allowed: %s -> %s", header.Name, header.Linkname)
		}
	}
	return nil
}

func ensureExtractDir(path string, mode os.FileMode, override bool) error {
	if err := os.MkdirAll(path, mode); err != nil {
		return err
	}
	if override {
		return os.Chmod(path, mode)
	}
	return nil
}

func ValidateSingleDirTarGz(src string, opts ...ArchiveOption) (string, error) {
	options := applyOptions(opts...)
	if err := checkContext(options.Context); err != nil {
		return "", fmt.Errorf("context check failed: %w", err)
	}
	srcFile, err := os.Open(src) // #nosec G304 - src validated by caller
	if err != nil {
		return "", fmt.Errorf("failed to open source file %s: %w", src, err)
	}
	defer srcFile.Close()
	gzReader, err := gzip.NewReader(srcFile)
	if err != nil {
		return "", fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzReader.Close()
	tarReader := tar.NewReader(gzReader)
	return validateTarSingleDir(tarReader, options)
}

func validateTarSingleDir(tarReader *tar.Reader, options ArchiveOptions) (string, error) {
	topLevelEntries := make(map[string]bool)
	var firstDirName string
	for {
		if err := checkContext(options.Context); err != nil {
			return "", fmt.Errorf("context check failed during validation: %w", err)
		}
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("failed to read tar header: %w", err)
		}
		name := cleanArchiveName(header.Name)
		topLevelName := extractTopLevelName(name)
		if topLevelName == "" {
			continue
		}
		topLevelEntries[topLevelName] = true
		if firstDirName == "" && header.Typeflag == tar.TypeDir {
			firstDirName = topLevelName
		}
		if len(topLevelEntries) > 1 {
			keys := make([]string, 0, len(topLevelEntries))
			for k := range topLevelEntries {
				keys = append(keys, k)
			}
			return "", fmt.Errorf("%w: %v", ErrMultipleTopLevelEntries, keys)
		}
	}
	if len(topLevelEntries) == 0 {
		return "", ErrEmptyArchive
	}
	if len(topLevelEntries) > 1 {
		keys := make([]string, 0, len(topLevelEntries))
		for k := range topLevelEntries {
			keys = append(keys, k)
		}
		return "", fmt.Errorf("%w: %v", ErrMultipleTopLevelEntries, keys)
	}
	if firstDirName == "" {
		return "", ErrSingleEntryNotDir
	}
	return firstDirName, nil
}

func cleanArchiveName(name string) string {
	name = filepath.Clean(name)
	name = strings.TrimPrefix(name, "./")
	return name
}

func extractTopLevelName(name string) string {
	if strings.Contains(name, "/") {
		parts := strings.Split(name, "/")
		return parts[0]
	}
	return name
}
