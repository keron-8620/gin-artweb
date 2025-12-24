package common

import (
	"github.com/casbin/casbin/v2"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"

	"gin-artweb/internal/shared/config"
)

type Initialize struct {
	Conf      *config.SystemConf
	DB        *gorm.DB
	DBTimeout *config.DBTimeout
	Enforcer  *casbin.Enforcer
	Crontab   *cron.Cron
}
