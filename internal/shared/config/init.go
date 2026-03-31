package config

import (
	"time"

	"github.com/casbin/casbin/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type SystemInit struct {
	Conf      *SystemConf
	DB        *gorm.DB
	DBTimeout *DBTimeout
	Enforcer  *casbin.Enforcer
	Crontab   *cron.Cron
	JwtConf   *JWTConfig
	Loggers   *Loggers
}

// DBTimeout 数据库操作超时参数
type DBTimeout struct {
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	ListTimeout  time.Duration
}

type Loggers struct {
	Server  *zap.Logger
	Handler *zap.Logger
	Service *zap.Logger
	Repo    *zap.Logger
}

type JWTConfig struct {
	Issuer                 string            // 令牌签发者
	AccessTokenExpiration  time.Duration     // 访问令牌过期时间
	RefreshTokenExpiration time.Duration     // 刷新令牌过期时间
	AccessSecret           []byte            // 访问令牌密钥
	RefreshSecret          []byte            // 刷新令牌密钥
	AccessMethod           jwt.SigningMethod // 访问令牌签名方法
	RefreshMethod          jwt.SigningMethod // 刷新令牌签名方法
}
