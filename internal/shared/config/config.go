package config

import (
	"log"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
)

// SystemConf 系统配置结构体
type SystemConf struct {
	Server            *ServerConfig      `yaml:"server"`
	Database          *DBConf            `yaml:"database"`
	Log               *LogConfig         `yaml:"log"`
	CORS              *AllowConfig       `yaml:"cors"`
	Security          *SecurityConfig    `yaml:"security"`
	SSH               *SSHConfig         `yaml:"ssh"`
	StorageSyncConfig *StorageSyncConfig `yaml:"storage_sync"`
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

	if conf.Database.Dns != "file::memory:" && !filepath.IsAbs(conf.Database.Dns) {
		conf.Database.Dns = filepath.Join(BaseDir, conf.Database.Dns)
	}

	return conf
}
