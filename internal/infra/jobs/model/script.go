package model

import (
	"go.uber.org/zap/zapcore"

	"gin-artweb/internal/shared/database"
)

type ScriptModel struct {
	database.StandardModel
	Name      string `gorm:"column:name;type:varchar(50);not null;index:idx_script_project_label_name;comment:名称" json:"name"`
	Descr     string `gorm:"column:descr;type:varchar(254);comment:描述" json:"descr"`
	Project   string `gorm:"column:project;type:varchar(50);index:idx_script_project_label_name;comment:项目" json:"project"`
	Label     string `gorm:"column:label;type:varchar(50);index:idx_script_project_label_name;;comment:标签" json:"label"`
	Language  string `gorm:"column:language;type:varchar(50);comment:脚本语言" json:"language"`
	Status    bool   `gorm:"column:status;type:boolean;comment:是否启用" json:"status"`
	IsBuiltin bool   `gorm:"column:is_builtin;type:boolean;comment:是否是内置脚本" json:"is_builtin"`
	Username  string `gorm:"column:username;type:varchar(50);comment:用户名" json:"username"`
}

func (m *ScriptModel) TableName() string {
	return "jobs_script"
}

func (m *ScriptModel) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if m == nil {
		return nil
	}
	if err := m.StandardModel.MarshalLogObject(enc); err != nil {
		return err
	}
	enc.AddString("name", m.Name)
	enc.AddString("descr", m.Descr)
	enc.AddString("project", m.Project)
	enc.AddString("label", m.Label)
	enc.AddString("language", m.Language)
	enc.AddBool("status", m.Status)
	enc.AddBool("is_builtin", m.IsBuiltin)
	enc.AddString("username", m.Username)
	return nil
}
