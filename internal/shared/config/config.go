package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/goccy/go-yaml"
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

	// 处理相对路径
	absPath, err := filepath.Abs(exePath)
	if err != nil {
		panic(fmt.Sprintf("转换绝对路径失败: %v", err))
	}

	// 处理符号链接
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

// SystemConf 系统配置结构体
type SystemConf struct {
	Server   *ServerConfig   `yaml:"server"`
	Database *DBConf         `yaml:"database"`
	Log      *LogConfig      `yaml:"log"`
	CORS     *AllowConfig    `yaml:"cors"`
	Security *SecurityConfig `yaml:"security"`
	SSH      *SSHConfig      `yaml:"ssh"`
	Upload   *UploadConfig   `yaml:"upload"`
}

// NewSystemConf 加载系统配置文件
func NewSystemConf(configPath string) *SystemConf {
	// 检查配置文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("FATAL: 配置文件不存在,请检查: %s", configPath)
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("FATAL: 读取配置文件失败: %v", err)
	}

	conf := &SystemConf{}

	// 解析YAML配置
	if err := yaml.Unmarshal(data, conf); err != nil {
		log.Fatalf("FATAL: 配置文件解析失败: %v", err)
	}

	if conf.Database.Type == "sqlite" && conf.Database.Dns == "file::memory:" && !filepath.IsAbs(conf.Database.Dns) {
		conf.Database.Dns = filepath.Join(BaseDir, conf.Database.Dns)
	}

	return conf
}
