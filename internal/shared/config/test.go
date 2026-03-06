package config

import (
	"time"
)

// NewTestSecurityConfig 创建用于测试环境的安全配置
// 该配置针对测试场景进行了优化，关闭了一些生产环境严格验证，
// 同时设置了合理的超时时间和安全参数，以提高测试效率和准确性。
func NewTestSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		Timestamp: TimestampConfig{
			CheckTimestamp: false, // 测试环境关闭时间戳验证，避免因时间差异导致测试失败
		},
		Token: TokenConfig{
			AccessMinutes:  10,      // Token 10分钟后过期，便于测试过期逻辑
			RefreshMinutes: 10,      // 10分钟后清理过期Token
			AccessMethod:   "HS256", // 使用HS256算法签名访问令牌
			RefreshMethod:  "HS256", // 使用HS256算法签名刷新令牌
		},
		Login: LoginSecurityConfig{
			MaxFailedAttempts: 5,  // 最多允许5次登录失败尝试
			LockMinutes:       30, // 登录失败锁定30分钟
		},
		Password: PasswordConfig{
			StrengthLevel: 3, // 中高等密码强度要求，可测试各种密码强度规则
		},
	}
}

func NewTestUploadConfig() *UploadConfig {
	return &UploadConfig{
		MaxPkgSize:    500, // 最大上传程序包大小500M(MB)
		MaxScriptSize: 1,   // 脚本最大上传大小1M(MB)
		MaxConfSize:   1,   // 配置文件最大上传大小1M(MB)
	}
}

func NewTestDBTimeouts() *DBTimeout {
	return &DBTimeout{
		ListTimeout:  10 * time.Second, // 批量查询超时时间为10秒
		ReadTimeout:  5 * time.Second,  // 单个查询超时为5秒
		WriteTimeout: 5 * time.Second,  // 单个写入超时为5秒
	}
}
