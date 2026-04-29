package common

import (
	"bufio"
	"os"
	"strconv"
	"strings"

	"emperror.dev/errors"
)

// readUint32FromFile 从指定文件读取单个数字并转换为uint32
func ReadUint32FromFile(filePath string) (uint32, error) {
	if _, err := os.Stat(filePath); err != nil {
		if !os.IsNotExist(err) {
			return 0, errors.WrapIfWithDetails(err, "获取文件状态失败", "filepath", filePath)
		}
		return 0, nil
	}
	// 读取文件内容
	file, err := os.Open(filePath) // #nosec G304
	if err != nil {
		return 0, errors.WrapIfWithDetails(err, "打开文件失败", "filepath", filePath)
	}
	defer file.Close() // 确保文件句柄关闭

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		if err = scanner.Err(); err != nil {
			return 0, errors.WrapIfWithDetails(err, "文件读取失败", "filepath", filePath)
		}
		return 0, errors.WrapIfWithDetails(errors.New("文件为空"), "文件为空", "filepath", filePath)
	}
	content := scanner.Text()

	// 去除空白字符（防止文件有换行/空格）
	numberStr := strings.TrimSpace(content)

	// 将字符串转换为uint64（先转uint64避免溢出判断），再转为uint32
	numberUint64, err := strconv.ParseUint(numberStr, 10, 32)
	if err != nil {
		return 0, errors.WrapIfWithDetails(err, "转换为uint32失败", "filepath", filePath)
	}

	// 转为uint32
	return uint32(numberUint64), nil
}
