package biz

import (
	"context"

	"gin-artweb/pkg/database"

	"go.uber.org/zap/zapcore"
)

type HostModel struct {
	database.StandardModel
	Name     string `gorm:"column:name;type:varchar(50);not null;uniqueIndex;comment:名称" json:"name"`
	Label    string `gorm:"column:label;type:varchar(50);index:idx_member;comment:标签" json:"label"`
	IPAddr   string `gorm:"column:ip_addr;type:varchar(50);comment:IP地址" json:"ip_addr"`
	Port     uint16 `gorm:"column:port;type:smallint;comment:端口" json:"port"`
	Username string `gorm:"column:username;type:varchar(50);comment:用户名" json:"username"`
	PyPath   string `gorm:"column:py_path;type:varchar(254);comment:python路径" json:"py_path"`
	Remark   string `gorm:"column:remark;type:varchar(254);comment:备注" json:"remark"`
}

func (m *HostModel) TableName() string {
	return "resource_host"
}

func (m *HostModel) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if err := m.StandardModel.MarshalLogObject(enc); err != nil {
		return err
	}
	enc.AddString("name", m.Name)
	enc.AddString("label", m.Label)
	enc.AddString("ip_addr", m.IPAddr)
	enc.AddUint16("port", m.Port)
	enc.AddString("username", m.Username)
	enc.AddString("py_path", m.PyPath)
	enc.AddString("remark", m.Remark)
	return nil
}

type HostRepo interface {
	CreateModel(context.Context, *HostModel) error
	UpdateModel(context.Context, map[string]any, ...any) error
	DeleteModel(context.Context, ...any) error
	FindModel(context.Context, []string, ...any) (*HostModel, error)
	ListModel(context.Context, database.QueryParams) (int64, []HostModel, error)
}
