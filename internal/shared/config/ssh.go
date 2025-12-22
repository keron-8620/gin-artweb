package config

type SSHConfig struct {
	// 私钥路径
	Private string `mapstructure:"private" json:"private"`

	// 连接超时时间（秒）
	Timeout int `mapstructure:"timeout" json:"timeout"`
}
