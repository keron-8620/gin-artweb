package fileutil

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"emperror.dev/errors"
)

// FileInfo 文件/目录的结构化信息（JSON 友好）
type FileInfo struct {
	Name     string      `json:"name"`
	Size     int64       `json:"size"`
	ModTime  string      `json:"mod_time"` // 标准化时间格式（RFC3339）
	IsDir    bool        `json:"is_dir"`
	Children []*FileInfo `json:"children,omitempty"`
}

// ListFileInfo 递归获取路径的文件/目录信息
// 示例:
//
//	info, _ := ListFileInfo(context.Background(), "/tmp/test")
//	// 返回 /tmp/test 的所有层级信息
func ListFileInfo(ctx context.Context, filePath string) (*FileInfo, error) {
	if err := ValidatePath(ctx, filePath); err != nil {
		return nil, errors.WithMessage(err, "路径校验失败")
	}

	// 单次 os.Stat 获取所有信息，避免重复调用
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, errors.WithMessagef(err, "获取路径基础信息失败, filepath=%s", filePath)
	}

	// 构建根节点（标准化时间格式）
	root := &FileInfo{
		Name:    info.Name(),
		Size:    info.Size(),
		ModTime: info.ModTime().Format(time.RFC3339), // 统一时间格式
		IsDir:   info.IsDir(),
	}

	// 递归处理子目录（仅目录需要）
	if root.IsDir {
		entries, err := os.ReadDir(filePath)
		if err != nil {
			return nil, errors.WithMessagef(err, "读取目录条目失败, dirpath=%s", filePath)
		}

		for _, entry := range entries {
			childPath := filepath.Join(filePath, entry.Name())
			// 复用 entry.Info()，减少系统调用
			childInfo, err := entry.Info()
			if err != nil {
				return nil, errors.WithMessagef(err, "获取子项详细信息失败, child_path=%s", childPath)
			}

			childNode := &FileInfo{
				Name:    childInfo.Name(),
				Size:    childInfo.Size(),
				ModTime: childInfo.ModTime().Format(time.RFC3339),
				IsDir:   childInfo.IsDir(),
			}

			// 递归处理子目录
			if childNode.IsDir {
				grandChild, err := ListFileInfo(ctx, childPath)
				if err != nil {
					return nil, errors.WithMessagef(err, "递归处理子目录失败, child_path=%s", childPath)
				}
				childNode.Children = grandChild.Children
			}

			root.Children = append(root.Children, childNode)
		}
	}

	return root, nil
}
