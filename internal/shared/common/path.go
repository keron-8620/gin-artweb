package common

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"emperror.dev/errors"

	"gin-artweb/internal/shared/config"
)

func GetHostVarsExportPath(pk uint32) string {
	filename := fmt.Sprintf("host_%d.yaml", pk)
	return filepath.Join(config.StorageDir, "host_vars", filename)
}

func GetPackageStoragePath(filename string) string {
	return filepath.Join(config.StorageDir, "packages", filename)
}

func GetScriptStoragePath(project, label, name string, isBuiltin bool) string {
	if isBuiltin {
		return filepath.Join(config.ResourceDir, project, "script", label, name)
	}
	return filepath.Join(config.StorageDir, "script", project, label, name)
}

func GetScriptLogStoragePath(data, logname string) string {
	return filepath.Join(config.StorageDir, "logs", data, logname)
}

func GetMonNodeExportPath(pk uint32) string {
	return filepath.Join(config.StorageDir, "mon", "config", fmt.Sprintf("%d", pk), "mon.yaml")
}

func GetMdsColonyBinDir(colonyNum string) string {
	return filepath.Join(config.StorageDir, "mds", "bin", colonyNum)
}

func GetMdsColonyConfigDir(colonyNum string) string {
	return filepath.Join(config.StorageDir, "mds", "config", colonyNum)
}

func GetOesColonyBinDir(colonyNum string) string {
	return filepath.Join(config.StorageDir, "oes", "bin", colonyNum)
}

func GetOesColonyConfigDir(colonyNum string) string {
	return filepath.Join(config.StorageDir, "oes", "config", colonyNum)
}

// readUint32FromFile 从指定文件读取单个数字并转换为uint32
func ReadUint32FromFile(filePath string) (uint32, error) {
	if _, err := os.Stat(filePath); err != nil {
		if !os.IsNotExist(err) {
			return 0, errors.WrapIfWithDetails(err, "获取文件状态失败", "filepath", filePath)
		}
		return 0, nil
	}
	// 读取文件内容
	file, err := os.Open(filePath)
	if err != nil {
		return 0, errors.WrapIfWithDetails(err, "打开文件失败", "filepath", filePath)
	}
	defer file.Close() // 确保文件句柄关闭

	// 读取第一行内容（因为文件只有一个数字）
	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		return 0, errors.WrapIfWithDetails(err, "文件为空或读取失败", "filepath", filePath)
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
