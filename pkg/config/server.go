package config

// SSLConfig SSL配置
type SSLConfig struct {
	Enable   bool   `yaml:"enable"`
	KeyPath  string `yaml:"key_path"`
	CrtPath  string `yaml:"crt_path"`
	Password string `yaml:"password"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host          string    `yaml:"host"`
	Port          int       `yaml:"port"`
	SSL           SSLConfig `yaml:"ssl"`
	EnableSwagger bool      `yaml:"enable_swagger"`
}