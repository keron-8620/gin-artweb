package model

import (
	"go.uber.org/zap/zapcore"

	"gin-artweb/internal/shared/database"
)

type RoleModel struct {
	database.StandardModel
	Name    string        `gorm:"column:name;type:varchar(50);not null;uniqueIndex;comment:名称" json:"name"`
	Descr   string        `gorm:"column:descr;type:varchar(254);comment:描述" json:"descr"`
	Apis    []ApiModel    `gorm:"many2many:customer_role_api;joinForeignKey:role_id;joinReferences:api_id;constraint:OnDelete:CASCADE"`
	Menus   []MenuModel   `gorm:"many2many:customer_role_menu;joinForeignKey:role_id;joinReferences:menu_id;constraint:OnDelete:CASCADE"`
	Buttons []ButtonModel `gorm:"many2many:customer_role_button;joinForeignKey:role_id;joinReferences:button_id;constraint:OnDelete:CASCADE"`
}

func (m *RoleModel) TableName() string {
	return "customer_role"
}

func (m *RoleModel) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if m == nil {
		return nil
	}
	if err := m.StandardModel.MarshalLogObject(enc); err != nil {
		return err
	}
	enc.AddString("name", m.Name)
	enc.AddString("descr", m.Descr)
	enc.AddArray("apis", zapcore.ArrayMarshalerFunc(func(ae zapcore.ArrayEncoder) error {
		for _, api := range m.Apis {
			ae.AppendUint32(api.ID)
		}
		return nil
	}))
	enc.AddArray("menus", zapcore.ArrayMarshalerFunc(func(ae zapcore.ArrayEncoder) error {
		for _, menu := range m.Menus {
			ae.AppendUint32(menu.ID)
		}
		return nil
	}))
	enc.AddArray("buttons", zapcore.ArrayMarshalerFunc(func(ae zapcore.ArrayEncoder) error {
		for _, button := range m.Buttons {
			ae.AppendUint32(button.ID)
		}
		return nil
	}))
	return nil
}
