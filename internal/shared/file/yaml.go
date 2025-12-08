package file

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// WriteYAML 将给定的数据写入指定路径的 YAML 文件
func WriteYAML(filename string, data any) error {
	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return err
	}

	// 序列化为 YAML 字节流
	out, err := yaml.Marshal(data)
	if err != nil {
		return err
	}

	// 写入文件（会覆盖已有内容）
	return os.WriteFile(filename, out, 0644)
}
