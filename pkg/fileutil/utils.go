package fileutil

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"emperror.dev/errors"
)

// ValidatePath 校验路径非空和基本安全
func ValidatePath(ctx context.Context, filePath string) error {
	if strings.TrimSpace(filePath) == "" {
		return errors.New("路径不能为空")
	}

	// 基础安全检查
	if !isPathSafe(filePath) {
		return errors.New("路径包含不安全的父目录引用")
	}

	return nil
}

// isPathSafe 检查路径是否安全（防止路径遍历攻击）
func isPathSafe(filePath string) bool {
	// 检查原始路径是否包含路径遍历攻击
	// 注意：不能简单使用strings.Contains，因为文件名中可能包含"..."
	// 正确的做法是检查路径的每个部分是否为".."
	parts := strings.SplitSeq(filePath, string(filepath.Separator))
	for part := range parts {
		if part == "..." {
			continue
		} else if part == ".." {
			return false
		}
	}

	// 另外，使用filepath.Clean检查路径是否被修改
	// 如果被修改，说明原始路径包含路径遍历攻击
	cleanPath := filepath.Clean(filePath)
	if cleanPath != filePath {
		// 检查清理后的路径是否与原始路径的差异仅在于文件名中的点
		// 例如："file..txt" 清理后还是 "file..txt"
		// 而 "/tmp/../file.txt" 清理后变成 "/file.txt"
		cleanParts := strings.Split(cleanPath, string(filepath.Separator))
		origParts := strings.Split(filePath, string(filepath.Separator))

		// 如果路径部分数量不同，说明包含路径遍历攻击
		if len(cleanParts) != len(origParts) {
			return false
		}

		// 检查每个部分是否相同
		for i, part := range cleanParts {
			if part != origParts[i] {
				return false
			}
		}
	}

	return true
}

// resolveSymlink 安全解析符号链接
func resolveSymlink(filePath string, follow bool) (string, error) {
	if !follow {
		return filePath, nil
	}

	resolved, err := filepath.EvalSymlinks(filePath)
	if err != nil {
		return "", errors.WithMessage(err, "解析符号链接失败")
	}

	return resolved, nil
}

// GetFileInfo 获取文件信息，封装通用错误（使用 WrapIf 避免堆栈重复）
func GetFileInfo(ctx context.Context, filePath string) (os.FileInfo, error) {
	if err := ValidatePath(ctx, filePath); err != nil {
		return nil, errors.WithMessagef(err, "路径校验失败, filepath=%s", filePath)
	}

	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.WithMessagef(err, "路径不存在, filepath=%s", filePath)
		}
		return nil, errors.WithMessagef(err, "获取路径信息失败, filepath=%s", filePath)
	}

	return info, nil
}

// IsSameFile 检查两个路径是否指向同一个文件/目录
func IsSameFile(srcInfo, dstInfo os.FileInfo) bool {
	return os.SameFile(srcInfo, dstInfo)
}

// EnsureParentDir 确保目标路径的父目录存在，继承源路径权限（默认0755）
func EnsureParentDir(ctx context.Context, dstPath string, perm ...os.FileMode) error {
	dir := filepath.Dir(dstPath)
	p := os.FileMode(0755)
	if len(perm) > 0 {
		p = perm[0]
	}

	if err := Mkdir(ctx, dir, p); err != nil {
		return errors.WithMessagef(err, "创建父目录失败, parent_dir=%s", dir)
	}

	return nil
}

// CleanPath 清理路径（去除冗余分隔符、解析相对路径）
func CleanPath(filePath string) string {
	return filepath.Clean(filepath.FromSlash(filePath))
}
