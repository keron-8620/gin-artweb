package fileutil

import (
	"fmt"
	"os"
	"path/filepath"
)

type FileInfo struct {
	Name     string      `json:"name"`
	Size     int64       `json:"size"`
	Mode     os.FileMode `json:"mode"`
	ModTime  string      `json:"mod_time"`
	IsDir    bool        `json:"is_dir"`
	Children []*FileInfo `json:"children,omitempty"`
	Path     string      `json:"path"`
}

func ListFileInfo(path string) (*FileInfo, error) {
	// 检查路径是否为空
	if path == "" {
		return nil, os.ErrInvalid
	}

	// 获取文件信息以确认路径存在
	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	root := &FileInfo{
		Name:    fileInfo.Name(),
		Size:    fileInfo.Size(),
		Mode:    fileInfo.Mode(),
		ModTime: fileInfo.ModTime().String(),
		IsDir:   fileInfo.IsDir(),
		Path:    path,
	}

	if root.IsDir {
		entries, err := os.ReadDir(path)
		if err != nil {
			return nil, err
		}

		for _, entry := range entries {
			childPath := filepath.Join(path, entry.Name())
			childNode, err := ListFileInfo(childPath)
			if err != nil {
				return nil, fmt.Errorf("获取子目录信息失败: %s, 错误: %v", childPath, err)
			}
			root.Children = append(root.Children, childNode)
		}
	}

	return root, nil
}
