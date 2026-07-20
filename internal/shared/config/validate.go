package config

import (
	"fmt"
	"strings"
	"time"
)

var supportedDatabaseTypes = map[string]struct{}{
	"mysql": {}, "postgres": {}, "sqlite": {}, "sqlserver": {}, "opengauss": {},
}

var supportedJWTMethods = map[string]struct{}{
	"HS256": {}, "HS384": {}, "HS512": {},
}

// Validate 在系统启动前集中校验配置，避免运行过程中出现空指针或危险默认值。
func (c *SystemConf) Validate() error {
	if c == nil {
		return fmt.Errorf("系统配置不能为空")
	}
	if c.Server == nil || c.Database == nil || c.Log == nil || c.CORS == nil ||
		c.Security == nil || c.SSH == nil || c.Upload == nil || c.API == nil {
		return fmt.Errorf("配置必须包含 server、database、log、cors、security、ssh、upload 和 api")
	}

	if strings.TrimSpace(c.Server.Host) == "" {
		return fmt.Errorf("server.host 不能为空")
	}
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("server.port 必须在 1-65535 之间")
	}
	if c.Server.Rate.RPS <= 0 || c.Server.Rate.Burst <= 0 {
		return fmt.Errorf("server.rate.rps 和 server.rate.burst 必须大于 0")
	}
	if c.Server.Timeout.Request <= 0 || c.Server.Timeout.Shutdown <= 0 {
		return fmt.Errorf("server.timeout.request 和 server.timeout.shutdown 必须大于 0")
	}
	if c.Server.SSL.Enable && (strings.TrimSpace(c.Server.SSL.CrtPath) == "" || strings.TrimSpace(c.Server.SSL.KeyPath) == "") {
		return fmt.Errorf("启用 SSL 时必须配置证书和私钥路径")
	}

	if _, ok := supportedDatabaseTypes[c.Database.Type]; !ok {
		return fmt.Errorf("不支持的 database.type: %s", c.Database.Type)
	}
	if strings.TrimSpace(c.Database.Dsn) == "" {
		return fmt.Errorf("database.dsn 不能为空")
	}
	if c.Database.MaxIdleConns < 0 || c.Database.MaxOpenConns <= 0 || c.Database.MaxIdleConns > c.Database.MaxOpenConns {
		return fmt.Errorf("数据库连接池配置无效")
	}
	if c.Database.ReadTimeout <= 0 || c.Database.WriteTimeout <= 0 || c.Database.ListTimeout <= 0 {
		return fmt.Errorf("数据库操作超时必须大于 0")
	}
	if c.Database.ReadSlow <= 0 || c.Database.WriteSlow <= 0 || c.Database.ListSlow <= 0 {
		return fmt.Errorf("数据库慢查询阈值必须大于 0")
	}

	if strings.TrimSpace(c.Log.Level) == "" || c.Log.MaxSize <= 0 || c.Log.MaxAge < 0 || c.Log.MaxBackups < 0 {
		return fmt.Errorf("日志配置无效")
	}
	if len(c.CORS.AllowOrigins) == 0 || len(c.CORS.AllowMethods) == 0 || len(c.CORS.AllowHeaders) == 0 {
		return fmt.Errorf("CORS 的 origins、methods 和 headers 不能为空")
	}
	if c.CORS.AllowCredentials && containsString(c.CORS.AllowOrigins, "*") {
		return fmt.Errorf("允许凭据时 CORS 不得使用通配来源")
	}

	if c.Security.HostGuard.Enable && len(c.Security.HostGuard.TrustedHosts) == 0 {
		return fmt.Errorf("启用 Host Guard 时 trusted_hosts 不能为空")
	}
	if c.Security.Token.AccessDuration <= 0 || c.Security.Token.RefreshDuration <= 0 ||
		c.Security.Token.RefreshDuration <= c.Security.Token.AccessDuration {
		return fmt.Errorf("Token 有效期配置无效，refresh_duration 必须大于 access_duration")
	}
	if _, ok := supportedJWTMethods[c.Security.Token.AccessMethod]; !ok {
		return fmt.Errorf("不支持的 access token 签名算法: %s", c.Security.Token.AccessMethod)
	}
	if _, ok := supportedJWTMethods[c.Security.Token.RefreshMethod]; !ok {
		return fmt.Errorf("不支持的 refresh token 签名算法: %s", c.Security.Token.RefreshMethod)
	}
	if c.Security.Login.MaxFailedAttempts <= 0 || c.Security.Login.LockDuration <= 0 {
		return fmt.Errorf("登录锁定配置无效")
	}
	if c.Security.Password.StrengthLevel < 0 || c.Security.Password.StrengthLevel > 4 {
		return fmt.Errorf("密码强度等级必须在 0-4 之间")
	}
	if c.Security.Timestamp.CheckTimestamp && (c.Security.Timestamp.Tolerance <= 0 || c.Security.Timestamp.FutureTolerance <= 0) {
		return fmt.Errorf("启用时间戳校验时容忍时间必须大于 0")
	}

	if c.SSH.Timeout <= 0 {
		return fmt.Errorf("ssh.timeout 必须大于 0")
	}
	if c.Upload.MaxPkgSize <= 0 || c.Upload.MaxScriptSize <= 0 || c.Upload.MaxConfSize <= 0 {
		return fmt.Errorf("上传大小限制必须大于 0")
	}
	if !validAPITimeouts(c.API) {
		return fmt.Errorf("API 超时配置必须大于 0，且业务超时不能超过 total_timeout")
	}
	return nil
}

func validAPITimeouts(c *APIConfig) bool {
	if c.TotalTimeout <= 0 {
		return false
	}
	for _, timeout := range []time.Duration{
		c.BizCreateTimeout, c.BizUpdateTimeout, c.BizQueryTimeout, c.BizDeleteTimeout,
	} {
		if timeout <= 0 || timeout > c.TotalTimeout {
			return false
		}
	}
	return true
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
