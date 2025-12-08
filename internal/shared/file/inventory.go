package file

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func WriteInventoryFile(filename string, hosts []uint32) error {
	// 确保目录存在
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// 构建文件内容
	var content strings.Builder
	for _, host := range hosts {
		content.WriteString(fmt.Sprintf("db_host_%d\n", host))
	}

	// 写入文件（会覆盖已有内容）
	return os.WriteFile(filename, []byte(content.String()), 0644)
}
