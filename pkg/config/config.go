package config

import (
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
)

// SystemConf 系统配置结构体
type SystemConf struct {
	Server   *ServerConfig       `yaml:"server"`
	Database *DBConf             `yaml:"database"`
	Log      *LogConfig          `yaml:"log"`
	CORS     *AllowConfig        `yaml:"cors"`
	SMTP     *SMTPConfig         `yaml:"smtp"`
	Security *SecurityConfig     `yaml:"security"`
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
	if err := validateCriticalConfig(conf); err != nil {
		panic(fmt.Sprintf("FATAL: 关键配置验证失败: %v", err))
	}

	return conf
}

// validateCriticalConfig 验证关键配置项
func validateCriticalConfig(conf *SystemConf) error {
	// 验证端口范围
	if conf.Server.Port <= 0 || conf.Server.Port > 65535 {
		return fmt.Errorf("HTTP端口 %d 无效,必须在1-65535之间", conf.Server.Port)
	}

	if conf.SMTP.Port <= 0 || conf.SMTP.Port > 65535 {
		return fmt.Errorf("SMTP端口 %d 无效,必须在1-65535之间", conf.SMTP.Port)
	}

	// 验证必需的配置项
	if conf.Database.Type == "" {
		return fmt.Errorf("数据库类型不能为空")
	}

	if conf.Database.Dns == "" {
		return fmt.Errorf("数据库连接字符串不能为空")
	}

	return nil
}
