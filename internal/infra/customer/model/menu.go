package model

import (
	"go.uber.org/zap/zapcore"

	"gin-artweb/internal/shared/database"
)

type Meta struct {
	Title string `json:"title"`
	Icon  string `json:"icon"`
}

func (m *Meta) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("title", m.Title)
	enc.AddString("icon", m.Icon)
	return nil
}

type MenuModel struct {
	database.StandardModel
	Path      string     `gorm:"column:path;type:varchar(100);not null;uniqueIndex;comment:前端路由" json:"path"`
	Component string     `gorm:"column:component;type:varchar(200);not null;comment:前端组件" json:"component"`
	Name      string     `gorm:"column:name;type:varchar(50);not null;uniqueIndex;comment:名称" json:"name"`
	Meta      Meta       `gorm:"column:meta;serializer:json;comment:菜单信息" json:"meta"`
	Sort      uint32     `gorm:"column:sort;type:integer;comment:排序" json:"sort"`
	IsActive  bool       `gorm:"column:is_active;type:boolean;comment:是否激活" json:"is_active"`
	Descr     string     `gorm:"column:descr;type:varchar(254);comment:描述" json:"descr"`
	ParentID  *uint32    `gorm:"column:parent_id;comment:父菜单ID" json:"parent_id"`
	Parent    *MenuModel `gorm:"foreignKey:ParentID;references:ID;constraint:OnDelete:CASCADE" json:"parent"`
	Apis      []ApiModel `gorm:"many2many:customer_menu_api;joinForeignKey:menu_id;joinReferences:api_id;constraint:OnDelete:CASCADE"`
}

func (m *MenuModel) TableName() string {
	return "customer_menu"
}

func (m *MenuModel) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if m == nil {
		return nil
	}
	if err := m.StandardModel.MarshalLogObject(enc); err != nil {
		return err
	}
	enc.AddString("path", m.Path)
	enc.AddString("component", m.Component)
	enc.AddObject("meta", &m.Meta)
	enc.AddString("name", m.Name)
	enc.AddUint32("sort", m.Sort)
	enc.AddBool("is_active", m.IsActive)
	enc.AddString("descr", m.Descr)
	if m.ParentID != nil {
		enc.AddUint32("parent_id", *m.ParentID)
	}
	enc.AddArray("apis", zapcore.ArrayMarshalerFunc(func(ae zapcore.ArrayEncoder) error {
		for _, api := range m.Apis {
			ae.AppendUint32(api.ID)
		}
		return nil
	}))
	return nil
}
