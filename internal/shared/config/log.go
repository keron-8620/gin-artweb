package config

// LogConfig 日志配置结构体
type LogConfig struct {
	Level      string `mapstructure:"level" json:"level" yaml:"level"`
	MaxSize    int    `mapstructure:"max_size" json:"max_size" yaml:"max_size"`
	MaxAge     int    `mapstructure:"max_age" json:"max_age" yaml:"max_age"`
	MaxBackups int    `mapstructure:"max_backups" json:"max_backups" yaml:"max_backups"`
	LocalTime  bool   `mapstructure:"local_time" json:"local_time" yaml:"local_time"`
	Compress   bool   `mapstructure:"compress" json:"compress" yaml:"compress"`
}
