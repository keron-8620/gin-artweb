package config

import "time"

// SystemConf 系统配置结构体
type SystemConf struct {
	Server   *ServerConfig   `yaml:"server"`
	Database *DBConf         `yaml:"database"`
	Log      *LogConfig      `yaml:"log"`
	CORS     *AllowConfig    `yaml:"cors"`
	Security *SecurityConfig `yaml:"security"`
	SSH      *SSHConfig      `yaml:"ssh"`
	Upload   *UploadConfig   `yaml:"upload"`
	API      *APIConfig      `yaml:"api"`
}

type APIConfig struct {
	TotalTimeout     time.Duration `yaml:"total_timeout"`
	BizCreateTimeout time.Duration `yaml:"biz_create_timeout"`
	BizUpdateTimeout time.Duration `yaml:"biz_update_timeout"`
	BizQueryTimeout  time.Duration `yaml:"biz_query_timeout"`
	BizDeleteTimeout time.Duration `yaml:"biz_delete_timeout"`
}

// AllowConfig CORS配置
type AllowConfig struct {
	AllowOrigins     []string `yaml:"allow_origins"`
	AllowCredentials bool     `yaml:"allow_credentials"`
	AllowMethods     []string `yaml:"allow_methods"`
	AllowHeaders     []string `yaml:"allow_headers"`
	ExposeHeaders    []string `yaml:"expose_headers"`
}

// DBConf 数据库配置结构体，用于配置数据库连接参数
type DBConf struct {
	Type            string        `yaml:"type" json:"type"`                             // 数据库类型，支持 mysql, postgres, sqlite, sqlserver, opengauss
	Dns             string        `yaml:"dns" json:"dns"`                               // 数据库连接字符串
	MaxIdleConns    int           `yaml:"max_idle_conns" json:"max_idle_conns"`         // 最大空闲连接数
	MaxOpenConns    int           `yaml:"max_open_conns" json:"max_open_conns"`         // 最大打开连接数
	LogSQL          bool          `yaml:"log_sql" json:"log_sql"`                       // 是否打印SQL
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" json:"conn_max_lifetime"`   // 连接最大生命周期(秒)
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time" json:"conn_max_idle_time"` // 连接最大空闲时间(秒)
	ReadTimeout     time.Duration `yaml:"read_timeout" json:"read_timeout"`             // 查询单条数据超时
	WriteTimeout    time.Duration `yaml:"write_timeout" json:"write_timeout"`           // 写操作超时
	ListTimeout     time.Duration `yaml:"list_timeout" json:"list_timeout"`             // 查询列表超时
	ReadSlow        time.Duration `yaml:"read_slow" json:"read_slow"`                   // 查询单条数据慢查询阈值
	WriteSlow       time.Duration `yaml:"write_slow" json:"write_slow"`                 // 写入慢查询阈值
	ListSlow        time.Duration `yaml:"list_slow" json:"list_slow"`                   // 查询列表慢查询阈值
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

type HostGuardConfig struct {
	Enable       bool     `yaml:"enable"`        // 是否启用host请求头防护
	TrustedHosts []string `yaml:"trusted_hosts"` // 受信任的host列表
}

// TimestampConfig 时间戳验证配置
type TimestampConfig struct {
	CheckTimestamp  bool `yaml:"check_timestamp"`  // 是否检查时间戳
	Tolerance       int  `yaml:"tolerance"`        // 时间容忍度(毫秒)
	FutureTolerance int  `yaml:"future_tolerance"` // 未来时间容忍度(毫秒)
}

// TokenConfig Token配置
type TokenConfig struct {
	AccessMinutes  int    `yaml:"access_minutes"`  // Token过期时间(分钟)
	RefreshMinutes int    `yaml:"refresh_minutes"` // 刷新令牌过期时间(分钟)
	AccessMethod   string `yaml:"access_method"`   // 访问令牌签名方法
	RefreshMethod  string `yaml:"refresh_method"`  // 刷新令牌签名方法
}

// LoginSecurityConfig 登录安全配置
type LoginSecurityConfig struct {
	MaxFailedAttempts int `yaml:"max_failed_attempts"` // 最大登录失败次数
	LockMinutes       int `yaml:"lock_minutes"`        // 锁定时长(分钟)
}

// PasswordConfig 密码配置
type PasswordConfig struct {
	StrengthLevel int `yaml:"strength_level"` // 密码强度等级
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	HostGuard HostGuardConfig     `yaml:"host_guard"` // host请求头配置
	Timestamp TimestampConfig     `yaml:"timestamp"`  // 时间戳验证配置
	Token     TokenConfig         `yaml:"token"`      // Token配置
	Login     LoginSecurityConfig `yaml:"login"`      // 登录安全配置
	Password  PasswordConfig      `yaml:"password"`   // 密码配置
}

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

type SSHConfig struct {
	// 私钥路径
	Private string `mapstructure:"private" json:"private"`

	// 连接超时时间（秒）
	Timeout int `mapstructure:"timeout" json:"timeout"`
}

// UploadConfig 上传配置
type UploadConfig struct {
	MaxPkgSize    int `yaml:"max_pkg_size"`    // 最大上传程序包大小(MB)
	MaxScriptSize int `yaml:"max_script_size"` // 脚本最大上传大小(MB)
	MaxConfSize   int `yaml:"max_conf_size"`   // 配置文件最大上传大小(MB)
}
