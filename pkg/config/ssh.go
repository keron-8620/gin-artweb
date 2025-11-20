package config


type SSHConfig struct {
    // 私钥路径
    PrivateKeyPath string `mapstructure:"private_key_path" json:"private_key_path"`
    
    // 公钥路径
    PublicKeyPath string `mapstructure:"public_key_path" json:"public_key_path"`
    
    // 连接超时时间（秒）
    ConnectTimeout int `mapstructure:"connect_timeout" json:"connect_timeout"`
}
