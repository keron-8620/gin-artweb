package model

import (
	"go.uber.org/zap/zapcore"

	"gin-artweb/internal/shared/database"
)

type HostModel struct {
	database.StandardModel
	Name    string `gorm:"column:name;type:varchar(50);not null;uniqueIndex;comment:名称" json:"name"`
	Label   string `gorm:"column:label;type:varchar(50);index:idx_host_label;comment:标签" json:"label"`
	SSHIP   string `gorm:"column:ssh_ip;type:varchar(108);uniqueIndex:idx_host_ip_port_user;comment:IP地址" json:"ssh_ip"`
	SSHPort uint16 `gorm:"column:ssh_port;type:smallint;uniqueIndex:idx_host_ip_port_user;comment:端口" json:"ssh_port"`
	SSHUser string `gorm:"column:ssh_user;type:varchar(50);uniqueIndex:idx_host_ip_port_user;comment:用户名" json:"ssh_user"`
	PyPath  string `gorm:"column:py_path;type:varchar(254);comment:python路径" json:"py_path"`
	Remark  string `gorm:"column:remark;type:varchar(254);comment:备注" json:"remark"`
}

func (m *HostModel) TableName() string {
	return "resource_host"
}

func (m *HostModel) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if m == nil {
		return nil
	}
	if err := m.StandardModel.MarshalLogObject(enc); err != nil {
		return err
	}
	enc.AddString("name", m.Name)
	enc.AddString("label", m.Label)
	enc.AddString("ssh_ip", m.SSHIP)
	enc.AddUint16("ssh_port", m.SSHPort)
	enc.AddString("ssh_user", m.SSHUser)
	enc.AddString("py_path", m.PyPath)
	enc.AddString("remark", m.Remark)
	return nil
}
