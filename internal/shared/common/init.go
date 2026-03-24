package common

import (
	"github.com/casbin/casbin/v2"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"gin-artweb/internal/shared/auth"
	"gin-artweb/internal/shared/config"
)

type Loggers struct {
	Server  *zap.Logger
	Handler *zap.Logger
	Service *zap.Logger
	Data    *zap.Logger
}

type Initialize struct {
	Conf      *config.SystemConf
	DB        *gorm.DB
	DBTimeout *config.DBTimeout
	Enforcer  *casbin.Enforcer
	Crontab   *cron.Cron
	JwtConf   *auth.JWTConfig
	Loggers   *Loggers
}
