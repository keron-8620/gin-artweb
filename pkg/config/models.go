package config

type SSLConfig struct {
	Enable   bool   `yaml:"enable"`
	KeyPath  string `yaml:"key_path"`
	CrtPath  string `yaml:"crt_path"`
	Password string `yaml:"password"`
}

type ServerConfig struct {
	Host          string    `yaml:"host"`
	Port          int       `yaml:"port"`
	SSL           SSLConfig `yaml:"ssl"`
	EnableSwagger bool      `yaml:"enable_swagger"`
}

type SMTPConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	TLS      bool   `yaml:"tls"`
}

type PasswordRuleConfig struct {
	MinLen  int  `yaml:"min_len"`
	MaxLen  int  `yaml:"max_len"`
	NeedUpp bool `yaml:"need_upp"`
	NeedLow bool `yaml:"need_low"`
	NeedNum bool `yaml:"need_num"`
	NeedSpe bool `yaml:"need_spe"`
}

type SecurityConfig struct {
	CheckTimestamp       bool   `yaml:"check_timestamp"`
	TimestampRange       int    `yaml:"timestamp_range"`
	TokenExpireMinutes   int    `yaml:"token_expire_minutes"`
	TokenClearMinutes    int    `yaml:"token_clear_minutes"`
	LoginFailMaxTimes    int    `yaml:"login_fail_max_times"`
	LoginFailLockMinutes int    `yaml:"login_fail_lock_minutes"`
	TimeoutShutdown      int    `yaml:"timeout_shutdown"`
	MaxUploadFileSize    int    `yaml:"max_upload_file_size"`
	PasswordStrength     int    `yaml:"password_strength"`
	SecretKey            string `yaml:"secret_key"`
}

// AllowConfig 安全配置
type AllowConfig struct {
	AllowOrigins     []string
	AllowCredentials bool
	AllowMethods     []string
	AllowHeaders     []string
}

// LogConfig 日志配置结构体
type LogConfig struct {
	Level      string `mapstructure:"level" json:"level" yaml:"level"`
	MaxSize    int    `mapstructure:"max_size" json:"max_size" yaml:"max_size"`
	MaxAge     int    `mapstructure:"max_age" json:"max_age" yaml:"max_age"`
	MaxBackups int    `mapstructure:"max_backups" json:"max_backups" yaml:"max_backups"`
	LocalTime  bool   `mapstructure:"local_time" json:"local_time" yaml:"local_time"`
	Compress   bool   `mapstructure:"compress" json:"compress" yaml:"compress"`
}

func (c *LogConfig) GetMaxSize() int {
	return c.MaxSize
}

func (c *LogConfig) GetMaxAge() int {
	return c.MaxAge
}

func (c *LogConfig) GetMaxBackUps() int {
	return c.MaxBackups
}

func (c *LogConfig) GetLocalTime() bool {
	return c.LocalTime
}

func (c *LogConfig) GetCompress() bool {
	return c.Compress
}

// DBConf 数据库配置结构体，用于配置数据库连接参数
type DBConf struct {
	Type            string `yaml:"type" json:"type"`                       // 数据库类型，支持 mysql, postgres, sqlite, sqlserver, opengauss
	Dns             string `yaml:"dns" json:"dns"`                         // 数据库连接字符串
	MaxIdleConns    int    `yaml:"maxIdleConns" json:"maxIdleConns"`       // 最大空闲连接数
	MaxOpenConns    int    `yaml:"maxOpenConns" json:"maxOpenConns"`       // 最大打开连接数
	ConnMaxLifetime int    `yaml:"connMaxLifetime" json:"connMaxLifetime"` // 连接最大生命周期(秒)
}
