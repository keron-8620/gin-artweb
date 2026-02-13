package model

import (
	"time"

	"go.uber.org/zap/zapcore"

	"gin-artweb/internal/shared/database"
)

type LoginRecordModel struct {
	database.BaseModel
	Username  string    `gorm:"column:username;type:varchar(50);comment:用户名" json:"username"`
	LoginAt   time.Time `gorm:"column:login_at;autoCreateTime;comment:登录时间" json:"login_at"`
	IPAddress string    `gorm:"column:ip_address;type:varchar(108);comment:ip地址" json:"ip_address"`
	UserAgent string    `gorm:"column:user_agent;type:varchar(254);comment:客户端信息" json:"user_agent"`
	Status    bool      `gorm:"column:status;type:boolean;comment:是否登录成功" json:"status"`
}

func (m *LoginRecordModel) TableName() string {
	return "customer_login_record"
}

func (m *LoginRecordModel) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if m == nil {
		return nil
	}
	if err := m.BaseModel.MarshalLogObject(enc); err != nil {
		return err
	}
	enc.AddString("username", m.Username)
	enc.AddTime("login_at", m.LoginAt)
	enc.AddString("ip_address", m.IPAddress)
	enc.AddString("user_agent", m.UserAgent)
	enc.AddBool("status", m.Status)
	return nil
}
