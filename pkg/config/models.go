package config

// type SSLConfig struct {
// 	Enable   bool   `yaml:"enable"`
// 	KeyPath  string `yaml:"key_path"`
// 	CrtPath  string `yaml:"crt_path"`
// 	Password string `yaml:"password"`
// }

// type ServerConfig struct {
// 	Host          string    `yaml:"host"`
// 	Port          int       `yaml:"port"`
// 	SSL           SSLConfig `yaml:"ssl"`
// 	EnableSwagger bool      `yaml:"enable_swagger"`
// }

// // DBConf 数据库配置结构体，用于配置数据库连接参数
// type DBConf struct {
// 	Type            string `yaml:"type" json:"type"`                           // 数据库类型，支持 mysql, postgres, sqlite, sqlserver, opengauss
// 	Dns             string `yaml:"dns" json:"dns"`                             // 数据库连接字符串
// 	MaxIdleConns    int    `yaml:"max_idle_conns" json:"max_idle_conns"`       // 最大空闲连接数
// 	MaxOpenConns    int    `yaml:"max_open_conns" json:"max_open_conns"`       // 最大打开连接数
// 	ConnMaxLifetime int    `yaml:"conn_max_lifetime" json:"conn_max_lifetime"` // 连接最大生命周期(秒)
// 	ReadTimeout     int    `yaml:"read_timeout" json:"read_timeout"`           // 查询单条数据超时
// 	WriteTimeout    int    `yaml:"write_timeout" json:"write_timeout"`         // 写操作超时
// 	ListTimeout     int    `yaml:"list_timeout" json:"list_timeout"`           // 查询列表超时
// }

// // 邮箱配置
// type SMTPConfig struct {
// 	Host     string `yaml:"host"`
// 	Port     int    `yaml:"port"`
// 	User     string `yaml:"user"`
// 	Password string `yaml:"password"`
// 	TLS      bool   `yaml:"tls"`
// }

// // 密码规则配置
// type PasswordRuleConfig struct {
// 	MinLen  int  `yaml:"min_len"`
// 	MaxLen  int  `yaml:"max_len"`
// 	NeedUpp bool `yaml:"need_upp"`
// 	NeedLow bool `yaml:"need_low"`
// 	NeedNum bool `yaml:"need_num"`
// 	NeedSpe bool `yaml:"need_spe"`
// }

// // 安全配置
// type SecurityConfig struct {
// 	RequestTimeout       int    `yaml:"request_timeout"`
// 	CheckTimestamp       bool   `yaml:"check_timestamp"`
// 	Tolerance            int    `yaml:"tolerance"`
// 	FutureTolerance      int    `yaml:"future_tolerance"`
// 	TokenExpireMinutes   int    `yaml:"token_expire_minutes"`
// 	TokenClearMinutes    int    `yaml:"token_clear_minutes"`
// 	LoginFailMaxTimes    int    `yaml:"login_fail_max_times"`
// 	LoginFailLockMinutes int    `yaml:"login_fail_lock_minutes"`
// 	TimeoutShutdown      int    `yaml:"timeout_shutdown"`
// 	MaxUploadFileSize    int    `yaml:"max_upload_file_size"`
// 	PasswordStrength     int    `yaml:"password_strength"`
// 	SecretKey            string `yaml:"secret_key"`
// }

// // AllowConfig 安全配置
// type AllowConfig struct {
// 	AllowOrigins     []string
// 	AllowCredentials bool
// 	AllowMethods     []string
// 	AllowHeaders     []string
// }

// // 限流配置
// type RateLimitConfig struct {
// 	// 每秒请求数
// 	RPS float64 `yaml:"rps"`
	
// 	// 突发请求数
// 	Burst int `yaml:"burst"`
// }

// // LogConfig 日志配置结构体
// type LogConfig struct {
// 	Level      string `mapstructure:"level" json:"level" yaml:"level"`
// 	MaxSize    int    `mapstructure:"max_size" json:"max_size" yaml:"max_size"`
// 	MaxAge     int    `mapstructure:"max_age" json:"max_age" yaml:"max_age"`
// 	MaxBackups int    `mapstructure:"max_backups" json:"max_backups" yaml:"max_backups"`
// 	LocalTime  bool   `mapstructure:"local_time" json:"local_time" yaml:"local_time"`
// 	Compress   bool   `mapstructure:"compress" json:"compress" yaml:"compress"`
// }

// func (c *LogConfig) GetMaxSize() int {
// 	return c.MaxSize
// }

// func (c *LogConfig) GetMaxAge() int {
// 	return c.MaxAge
// }

// func (c *LogConfig) GetMaxBackUps() int {
// 	return c.MaxBackups
// }

// func (c *LogConfig) GetLocalTime() bool {
// 	return c.LocalTime
// }

// func (c *LogConfig) GetCompress() bool {
// 	return c.Compress
// }


