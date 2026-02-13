package model

import (
	"go.uber.org/zap/zapcore"

	"gin-artweb/internal/shared/database"
)

type ScheduleModel struct {
	database.StandardModel
	Name          string      `gorm:"column:name;type:varchar(50);not null;uniqueIndex;comment:名称" json:"name"`
	Specification string      `gorm:"column:specification;type:text;comment:条件" json:"specification"`
	IsEnabled     bool        `gorm:"column:is_enabled;type:boolean;comment:是否启用" json:"is_enabled"`
	EnvVars       string      `gorm:"column:env_vars;type:json;comment:环境变量(JSON对象)" json:"env_vars"`
	CommandArgs   string      `gorm:"column:command_args;type:varchar(254);comment:命令行参数" json:"command_args"`
	WorkDir       string      `gorm:"column:work_dir;type:varchar(255);comment:工作目录" json:"work_dir"`
	Timeout       int         `gorm:"column:timeout;type:int;not null;default:300;comment:超时时间(秒)" json:"timeout"`
	IsRetry       bool        `gorm:"column:is_retry;type:boolean;default:false;comment:是否启用重试" json:"is_retry"`
	RetryInterval int         `gorm:"column:retry_interval;type:int;default:60;comment:重试间隔(秒)" json:"retry_interval"`
	MaxRetries    int         `gorm:"column:max_retries;type:int;default:3;comment:最大重试次数" json:"max_retries"`
	Username      string      `gorm:"column:username;type:varchar(50);comment:用户名" json:"username"`
	ScriptID      uint32      `gorm:"column:script_id;not null;index;comment:计划任务ID" json:"script_id"`
	Script        ScriptModel `gorm:"foreignKey:ScriptID;references:ID" json:"script"`
}

func (m *ScheduleModel) TableName() string {
	return "jobs_schedule"
}

func (m *ScheduleModel) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if m == nil {
		return nil
	}
	if err := m.StandardModel.MarshalLogObject(enc); err != nil {
		return err
	}
	enc.AddString("name", m.Name)
	enc.AddString("specification", m.Specification)
	enc.AddBool("is_enabled", m.IsEnabled)
	enc.AddString("env_vars", m.EnvVars)
	enc.AddString("command_args", m.CommandArgs)
	enc.AddString("work_dir", m.WorkDir)
	enc.AddInt("timeout", m.Timeout)
	enc.AddString("username", m.Username)
	enc.AddUint32("script_id", m.ScriptID)
	return nil
}
