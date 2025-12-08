package file

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// WriteJSON 将数据序列化为JSON格式并写入文件
func WriteJSON(filename string, data any, indent uint8) error {
	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return err
	}

	// 创建或截断文件
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// 序列化为JSON并写入文件
	encoder := json.NewEncoder(file)

	// 如果指定了缩进，则使用空格进行格式化
	if indent > 0 {
		indentStr := strings.Repeat(" ", int(indent)) // 使用空格
		encoder.SetIndent("", indentStr)
	}

	return encoder.Encode(data)
}
