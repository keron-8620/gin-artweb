package config

import "fmt"

// ValidateCriticalConfig 验证关键配置项
func ValidateCriticalConfig(conf *SystemConf) error {
	// 验证端口范围
	if conf.Server.Port <= 0 || conf.Server.Port > 65535 {
		return fmt.Errorf("HTTP端口 %d 无效,必须在1-65535之间", conf.Server.Port)
	}

	// if conf.SMTP.Port <= 0 || conf.SMTP.Port > 65535 {
	// 	return fmt.Errorf("SMTP端口 %d 无效,必须在1-65535之间", conf.SMTP.Port)
	// }

	// 验证必需的配置项
	if conf.Database.Type == "" {
		return fmt.Errorf("数据库类型不能为空")
	}

	if conf.Database.Dns == "" {
		return fmt.Errorf("数据库连接字符串不能为空")
	}

	return nil
}