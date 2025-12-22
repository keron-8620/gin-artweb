package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

type PathConf struct {
	BaseDir     string
	ConfigDir   string
	HtmlDir     string
	LogsDir     string
	StorageDir  string
	ResourceDir string
}

func getBaseDir() string {
	exePath, err := os.Executable()
	if err != nil {
		panic(fmt.Sprintf("获取执行路径失败: %v", err))
	}

	// 处理符号链接和相对路径
	absPath, err := filepath.Abs(exePath)
	if err != nil {
		panic(fmt.Sprintf("转换绝对路径失败: %v", err))
	}

	resolvedPath, err := filepath.EvalSymlinks(absPath)
	if err != nil {
		panic(fmt.Sprintf("解析符号链接失败: %v", err))
	}

	// Windows平台特殊处理
	if runtime.GOOS == "windows" {
		resolvedPath = filepath.ToSlash(resolvedPath)
	}

	binDir := filepath.Dir(resolvedPath)
	return filepath.Dir(binDir)
}

var (
	BaseDir     = getBaseDir()
	ConfigDir   = filepath.Join(BaseDir, "config")
	LogDir      = filepath.Join(BaseDir, "logs")
	StorageDir  = filepath.Join(BaseDir, "storage")
	ResourceDir = filepath.Join(BaseDir, "resource")
)
