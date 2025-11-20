package config

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	// 每秒请求数
	RPS float64 `yaml:"rps"`
	
	// 突发请求数
	Burst int `yaml:"burst"`
}