package config

import (
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
)

// SystemConf 系统配置结构体
type SystemConf struct {
	Server   *ServerConfig    `yaml:"server"`
	Database *DBConf          `yaml:"database"`
	Log      *LogConfig       `yaml:"log"`
	CORS     *AllowConfig     `yaml:"cors"`
	SMTP     *SMTPConfig      `yaml:"smtp"`
	Security *SecurityConfig  `yaml:"security"`
	Rate     *RateLimitConfig `yaml:"rate"`
	SSH      *SSHConfig       `yaml:"ssh"`
}

// NewSystemConf 加载系统配置文件
func NewSystemConf(configPath string) *SystemConf {
	// 检查配置文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic(fmt.Sprintf("FATAL: 配置文件不存在,请检查: %s", configPath))
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		panic(fmt.Sprintf("FATAL: 读取配置文件失败: %v", err))
	}

	conf := &SystemConf{}

	// 解析YAML配置
	if err := yaml.Unmarshal(data, conf); err != nil {
		panic(fmt.Sprintf("FATAL: 配置文件解析失败: %v", err))
	}

	// 验证关键配置
	if err := ValidateCriticalConfig(conf); err != nil {
		panic(fmt.Sprintf("FATAL: 关键配置验证失败: %v", err))
	}

	return conf
}
