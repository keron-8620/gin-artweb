package archive

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type zipFileHandler func(path string, info os.FileInfo, header *zip.FileHeader, file *os.File, opts ArchiveOptions) error

func Zip(src, dst string, opts ...ArchiveOption) error {
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
	zipWriter := zip.NewWriter(f)
	defer zipWriter.Close()
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to stat source %s: %w", src, err)
	}
	if srcInfo.IsDir() {
		return walkAndProcessZip(src, options, func(path string, info os.FileInfo, header *zip.FileHeader, file *os.File, opts ArchiveOptions) error {
			relPath, err := filepath.Rel(filepath.Dir(src), path)
			if err != nil {
				return fmt.Errorf("failed to calculate relative path for %s: %w", path, err)
			}
			if relPath == "." {
				return nil
			}
			header.Name = filepath.ToSlash(relPath)
			if info.IsDir() {
				header.Name += "/"
			}
			w, err := zipWriter.CreateHeader(header)
			if err != nil {
				return fmt.Errorf("failed to create zip header for %s: %w", path, err)
			}
			if !info.IsDir() && info.Mode().IsRegular() && file != nil {
				if _, err := safeCopy(opts.Context, w, file, opts.MaxFileSize, opts.BufferSize); err != nil {
					return fmt.Errorf("failed to copy content for %s: %w", path, err)
				}
			}
			return nil
		})
	}
	return processSingleFileZip(src, srcInfo, zipWriter, options)
}

func walkAndProcessZip(src string, options ArchiveOptions, handler zipFileHandler) error {
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
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return fmt.Errorf("failed to create zip header for %s: %w", path, err)
		}
		var file *os.File
		if !info.IsDir() && info.Mode().IsRegular() {
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
			f, err := os.Open(path) // #nosec G304,G122 - symlink checked above
			if err != nil {
				return fmt.Errorf("failed to open file %s: %w", path, err)
			}
			file = f
			defer f.Close()
		}
		return handler(path, info, header, file, options)
	})
}

func processSingleFileZip(src string, srcInfo os.FileInfo, zipWriter *zip.Writer, options ArchiveOptions) error {
	fileCount := 1
	if options.MaxFiles > 0 && fileCount > options.MaxFiles {
		return fmt.Errorf("%w: %d", ErrFileCountExceeded, options.MaxFiles)
	}
	if options.MaxFileSize > 0 && srcInfo.Size() > options.MaxFileSize {
		return fmt.Errorf("%w: %s size %d > limit %d", ErrFileSizeExceeded, src, srcInfo.Size(), options.MaxFileSize)
	}
	header, err := zip.FileInfoHeader(srcInfo)
	if err != nil {
		return fmt.Errorf("failed to create zip header for %s: %w", src, err)
	}
	header.Name = filepath.Base(src)
	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return fmt.Errorf("failed to create zip header for %s: %w", src, err)
	}
	file, err := os.Open(src) // #nosec G304 - src validated by caller
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", src, err)
	}
	defer file.Close()
	if _, err := safeCopy(options.Context, writer, file, options.MaxFileSize, options.BufferSize); err != nil {
		return fmt.Errorf("failed to copy content for %s: %w", src, err)
	}
	return nil
}

func Unzip(src, dst string, opts ...ArchiveOption) error {
	options := applyOptions(opts...)
	if err := checkContext(options.Context); err != nil {
		return fmt.Errorf("context check failed: %w", err)
	}
	reader, err := zip.OpenReader(src)
	if err != nil {
		return fmt.Errorf("failed to open zip file %s: %w", src, err)
	}
	defer reader.Close()
	if err := os.MkdirAll(dst, 0750); err != nil {
		return fmt.Errorf("failed to create destination directory %s: %w", dst, err)
	}
	fileCount := 0
	for _, file := range reader.File {
		if err := checkContext(options.Context); err != nil {
			return fmt.Errorf("context check failed during unzip: %w", err)
		}
		fileCount++
		if options.MaxFiles > 0 && fileCount > options.MaxFiles {
			return fmt.Errorf("%w: %d", ErrFileCountExceeded, options.MaxFiles)
		}
		cleanName := filepath.Clean(filepath.FromSlash(file.Name))
		parts := strings.SplitSeq(cleanName, string(filepath.Separator))
		for part := range parts {
			if part == ".." {
				return fmt.Errorf("%w: %s", ErrInvalidPath, file.Name)
			}
		}
		target := filepath.Join(dst, cleanName)
		if !isPathSafe(target, dst) {
			return fmt.Errorf("%w: %s", ErrInvalidPath, target)
		}
		if err := processZipEntry(target, file, options); err != nil {
			return err
		}
	}
	return nil
}

func processZipEntry(target string, file *zip.File, options ArchiveOptions) error {
	fileinfo := file.FileInfo()
	if fileinfo.IsDir() {
		if err := os.MkdirAll(target, fileinfo.Mode()); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", target, err)
		}
		return nil
	}
	if options.MaxFileSize > 0 && fileinfo.Size() > options.MaxFileSize {
		return fmt.Errorf("%w: %s size %d > limit %d", ErrFileSizeExceeded, file.Name, fileinfo.Size(), options.MaxFileSize)
	}
	parentDir := filepath.Dir(target)
	if err := os.MkdirAll(parentDir, 0750); err != nil {
		return fmt.Errorf("failed to create parent directory %s: %w", parentDir, err)
	}
	targetFile, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR|os.O_TRUNC, fileinfo.Mode()) // #nosec G304 - target validated by isPathSafe
	if err != nil {
		return fmt.Errorf("failed to create target file %s: %w", target, err)
	}
	defer targetFile.Close()
	srcFile, err := file.Open()
	if err != nil {
		return fmt.Errorf("failed to open zip entry %s: %w", file.Name, err)
	}
	defer srcFile.Close()
	if _, err := safeCopy(options.Context, targetFile, srcFile, options.MaxFileSize, options.BufferSize); err != nil {
		return fmt.Errorf("failed to copy content for %s: %w", file.Name, err)
	}
	return nil
}

func ValidateSingleDirZip(src string, opts ...ArchiveOption) (string, error) {
	options := applyOptions(opts...)
	if err := checkContext(options.Context); err != nil {
		return "", fmt.Errorf("context check failed: %w", err)
	}
	reader, err := zip.OpenReader(src)
	if err != nil {
		return "", fmt.Errorf("failed to open zip file %s: %w", src, err)
	}
	defer reader.Close()
	return validateZipSingleDir(reader.File, options)
}

func validateZipSingleDir(files []*zip.File, options ArchiveOptions) (string, error) {
	topLevelEntries := make(map[string]bool)
	var firstDirName string
	for _, file := range files {
		if err := checkContext(options.Context); err != nil {
			return "", fmt.Errorf("context check failed during validation: %w", err)
		}
		name := cleanZipName(file.Name)
		topLevelName := extractZipTopLevelName(name)
		if topLevelName == "" {
			continue
		}
		topLevelEntries[topLevelName] = true
		if firstDirName == "" && file.FileInfo().IsDir() {
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

func cleanZipName(name string) string {
	name = filepath.Clean(name)
	name = strings.TrimPrefix(name, "./")
	name = strings.TrimSuffix(name, "/")
	return name
}

func extractZipTopLevelName(name string) string {
	if strings.Contains(name, "/") {
		parts := strings.Split(name, "/")
		return parts[0]
	}
	return name
}
