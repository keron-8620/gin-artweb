package model

import (
	"go.uber.org/zap/zapcore"

	"gin-artweb/internal/shared/database"
)

type ScriptRecordModel struct {
	database.StandardModel
	TriggerType  string      `gorm:"column:trigger_type;type:varchar(20);comment:触发类型(cron/api)" json:"trigger_type"`
	Status       int         `gorm:"column:status;type:tinyint;not null;default:0;comment:执行状态(0-待执行,1-执行中,2-成功,3-失败,4-超时,5-崩溃)" json:"status"`
	ExitCode     int         `gorm:"column:exit_code;comment:退出码" json:"exit_code"`
	EnvVars      string      `gorm:"column:env_vars;type:json;comment:环境变量(JSON对象)" json:"env_vars"`
	CommandArgs  string      `gorm:"column:command_args;type:varchar(254);comment:命令行参数(JSON数组)" json:"command_args"`
	WorkDir      string      `gorm:"column:work_dir;type:varchar(255);comment:工作目录" json:"work_dir"`
	Timeout      int         `gorm:"column:timeout;type:int;not null;default:300;comment:超时时间(秒)" json:"timeout"`
	LogName      string      `gorm:"column:log_name;type:varchar(255);comment:日志文件路径" json:"log_name"`
	ErrorMessage string      `gorm:"column:error_message;type:text;comment:错误信息" json:"error_message"`
	Username     string      `gorm:"column:username;type:varchar(50);comment:用户名" json:"username"`
	ScriptID     uint32      `gorm:"column:script_id;not null;index;comment:脚本ID" json:"script_id"`
	Script       ScriptModel `gorm:"foreignKey:ScriptID;references:ID;constraint:OnDelete:CASCADE" json:"script"`
}

func (m *ScriptRecordModel) TableName() string {
	return "jobs_script_record"
}

func (m *ScriptRecordModel) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if m == nil {
		return nil
	}
	if err := m.StandardModel.MarshalLogObject(enc); err != nil {
		return err
	}
	enc.AddString("trigger_type", m.TriggerType)
	enc.AddInt("status", m.Status)
	enc.AddInt("exit_code", m.ExitCode)
	enc.AddString("env_vars", m.EnvVars)
	enc.AddString("command_args", m.CommandArgs)
	enc.AddString("work_dir", m.WorkDir)
	enc.AddString("log_path", m.LogName)
	enc.AddString("username", m.Username)
	enc.AddUint32("script_id", m.ScriptID)
	return nil
}
