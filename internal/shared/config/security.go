package config

// TimeoutConfig 超时配置
type TimeoutConfig struct {
	RequestTimeout  int `yaml:"request_timeout"`   // 请求处理超时时间(秒)
	ShutdownTimeout int `yaml:"shutdown_timeout"`  // 服务关闭超时时间(秒)
}

// TimestampConfig 时间戳验证配置
type TimestampConfig struct {
	CheckTimestamp  bool `yaml:"check_timestamp"`   // 是否检查时间戳
	Tolerance       int  `yaml:"tolerance"`         // 时间容忍度(秒)
	FutureTolerance int  `yaml:"future_tolerance"`  // 未来时间容忍度(秒)
}

// TokenConfig Token配置
type TokenConfig struct {
	ExpireMinutes int    `yaml:"expire_minutes"`     // Token过期时间(分钟)
	ClearMinutes  int    `yaml:"clear_minutes"`      // Token清理间隔(分钟)
	SecretKey     string `yaml:"secret_key"`         // 加密密钥
}

// LoginSecurityConfig 登录安全配置
type LoginSecurityConfig struct {
	MaxFailedAttempts int `yaml:"max_failed_attempts"`    // 最大登录失败次数
	LockMinutes       int `yaml:"lock_minutes"`           // 锁定时长(分钟)
}

// UploadConfig 上传配置
type UploadConfig struct {
	MaxFileSize int `yaml:"max_file_size"`  // 最大文件大小(MB)
}

// PasswordConfig 密码配置
type PasswordConfig struct {
	MinLength     int  `yaml:"min_length"`      // 最小长度
	MaxLength     int  `yaml:"max_length"`      // 最大长度
	RequireUpper  bool `yaml:"require_upper"`   // 需要大写字母
	RequireLower  bool `yaml:"require_lower"`   // 需要小写字母
	RequireNumber bool `yaml:"require_number"`  // 需要数字
	RequireSpecial bool `yaml:"require_special"` // 需要特殊字符
	StrengthLevel int  `yaml:"strength_level"`  // 密码强度等级
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	Timeout   TimeoutConfig       `yaml:"timeout"`
	Timestamp TimestampConfig     `yaml:"timestamp"`
	Token     TokenConfig         `yaml:"token"`
	Login     LoginSecurityConfig `yaml:"login"`
	Upload    UploadConfig        `yaml:"upload"`
	Password  PasswordConfig      `yaml:"password"`
}