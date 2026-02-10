package test

import (
	"time"

	"gin-artweb/internal/shared/config"
)

// NewTestSecurityConfig 创建用于测试环境的安全配置
// 该配置针对测试场景进行了优化，关闭了一些生产环境严格验证，
// 同时设置了合理的超时时间和安全参数，以提高测试效率和准确性。
func NewTestSecurityConfig() *config.SecurityConfig {
	return &config.SecurityConfig{
		Timeout: config.TimeoutConfig{
			RequestTimeout:  10, // 请求超时时间为10秒，加快测试反馈速度
			ShutdownTimeout: 5,  // 关闭超时时间为5秒，加速测试结束
		},
		Timestamp: config.TimestampConfig{
			CheckTimestamp: false, // 测试环境关闭时间戳验证，避免因时间差异导致测试失败
		},
		Token: config.TokenConfig{
			AccessMinutes:  10,      // Token 10分钟后过期，便于测试过期逻辑
			RefreshMinutes: 10,      // 10分钟后清理过期Token
			AccessMethod:   "HS256", // 使用HS256算法签名访问令牌
			RefreshMethod:  "HS256", // 使用HS256算法签名刷新令牌
		},
		Login: config.LoginSecurityConfig{
			MaxFailedAttempts: 5,  // 最多允许5次登录失败尝试
			LockMinutes:       30, // 登录失败锁定30分钟
		},
		Upload: config.UploadConfig{
			MaxPkgSize: 500, // 最大上传文件500M大小(MB)
		},
		Password: config.PasswordConfig{
			StrengthLevel: 3, // 中高等密码强度要求，可测试各种密码强度规则
		},
	}
}

func NewTestDBTimeouts() *config.DBTimeout {
	return &config.DBTimeout{
		ListTimeout:  10 * time.Second, // 批量查询超时时间为10秒
		ReadTimeout:  5 * time.Second,  // 单个查询超时为5秒
		WriteTimeout: 5 * time.Second,  // 单个写入超时为5秒
	}
}
