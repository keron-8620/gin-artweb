package model

import (
	"go.uber.org/zap/zapcore"

	"gin-artweb/internal/shared/database"
)

type UserModel struct {
	database.StandardModel
	Username string    `gorm:"column:username;type:varchar(50);not null;uniqueIndex;comment:用户名" json:"username"`
	Password string    `gorm:"column:password;type:varchar(150);not null;comment:密码" json:"password"`
	IsActive bool      `gorm:"column:is_active;type:boolean;comment:是否激活" json:"is_active"`
	IsStaff  bool      `gorm:"column:is_staff;type:boolean;comment:是否是工作人员" json:"is_staff"`
	RoleID   uint32    `gorm:"column:role_id;not null;comment:角色ID" json:"role_id"`
	Role     RoleModel `gorm:"foreignKey:RoleID;references:ID;constraint:OnDelete:CASCADE" json:"role"`
}

func (m *UserModel) TableName() string {
	return "customer_user"
}

func (m *UserModel) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if m == nil {
		return nil
	}
	if err := m.StandardModel.MarshalLogObject(enc); err != nil {
		return err
	}
	enc.AddString("username", m.Username)
	enc.AddBool("is_active", m.IsActive)
	enc.AddBool("is_staff", m.IsStaff)
	enc.AddUint32("role_id", m.RoleID)
	return nil
}
