package config

// SSLConfig SSL配置
type SSLConfig struct {
	Enable   bool   `yaml:"enable"`
	KeyPath  string `yaml:"key_path"`
	CrtPath  string `yaml:"crt_path"`
	Password string `yaml:"password"`
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	// 每秒请求数
	RPS float64 `yaml:"rps"`

	// 突发请求数
	Burst int `yaml:"burst"`
}

// TimeoutConfig 超时配置
type TimeoutConfig struct {
	Request  int `yaml:"request"`  // 请求处理超时时间(秒)
	Shutdown int `yaml:"shutdown"` // 服务关闭超时时间(秒)
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host    string          `yaml:"host"`
	Port    int             `yaml:"port"`
	SSL     SSLConfig       `yaml:"ssl"`
	Rate    RateLimitConfig `yaml:"rate"`
	Timeout TimeoutConfig   `yaml:"timeout"`
	Swagger bool            `yaml:"swagger"`
}
