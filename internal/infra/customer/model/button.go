package model

import (
	"go.uber.org/zap/zapcore"

	"gin-artweb/internal/shared/database"
)

type ButtonModel struct {
	database.StandardModel
	Name     string     `gorm:"column:name;type:varchar(50);not null;uniqueIndex;comment:名称" json:"name"`
	Sort     uint32     `gorm:"column:sort;type:integer;comment:排序" json:"sort"`
	IsActive bool       `gorm:"column:is_active;type:boolean;comment:是否激活" json:"is_active"`
	Descr    string     `gorm:"column:descr;type:varchar(254);comment:描述" json:"descr"`
	MenuID   uint32     `gorm:"column:menu_id;not null;comment:菜单ID" json:"menu_id"`
	Menu     MenuModel  `gorm:"foreignKey:MenuID;references:ID;constraint:OnDelete:CASCADE" json:"menu"`
	Apis     []ApiModel `gorm:"many2many:customer_button_api;joinForeignKey:button_id;joinReferences:api_id;constraint:OnDelete:CASCADE"`
}

func (m *ButtonModel) TableName() string {
	return "customer_button"
}

func (m *ButtonModel) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if m == nil {
		return nil
	}
	if err := m.StandardModel.MarshalLogObject(enc); err != nil {
		return err
	}
	enc.AddString("name", m.Name)
	enc.AddUint32("sort", m.Sort)
	enc.AddBool("is_active", m.IsActive)
	enc.AddString("descr", m.Descr)
	enc.AddUint32("menu_id", m.MenuID)
	enc.AddArray("apis", zapcore.ArrayMarshalerFunc(func(ae zapcore.ArrayEncoder) error {
		for _, api := range m.Apis {
			ae.AppendUint32(api.ID)
		}
		return nil
	}))
	return nil
}
