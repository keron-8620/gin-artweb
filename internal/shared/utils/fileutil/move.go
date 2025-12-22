package fileutil

import (
	"fmt"
	"os"
	"path/filepath"
)

// Move 将文件或目录从源路径移动到目标路径。
// 如果目标路径是已存在的目录，则源文件/目录将被移动到该目录中。
// 如果目标路径是文件或不存在的路径，则源文件/目录将被重命名为目标路径。
func Move(src, dst string) error {
	// 验证输入路径
	if src == "" {
		return fmt.Errorf("源路径不能为空")
	}
	if dst == "" {
		return fmt.Errorf("目标路径不能为空")
	}

	// 检查源是否存在
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("获取源信息失败: %w", err)
	}

	// 检查源和目标是否相同
	if dstInfo, err := os.Stat(dst); err == nil {
		if os.SameFile(srcInfo, dstInfo) {
			// 源和目标相同，无需操作
			return nil
		}
		
		// 如果目标是目录，则将源移动到该目录中
		if dstInfo.IsDir() {
			dst = filepath.Join(dst, filepath.Base(src))
			
			// 路径调整后再次检查是否相同
			if dstInfo, err := os.Stat(dst); err == nil {
				if os.SameFile(srcInfo, dstInfo) {
					return nil
				}
			}
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("检查目标路径失败: %w", err)
	}

	// 确保目标的父目录存在
	dstDir := filepath.Dir(dst)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("创建目标父目录失败: %w", err)
	}

	// 执行移动操作
	if err := os.Rename(src, dst); err != nil {
		return fmt.Errorf("将 %s 移动到 %s 失败: %w", src, dst, err)
	}

	return nil
}
