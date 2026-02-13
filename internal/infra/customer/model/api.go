package model

import (
	"go.uber.org/zap/zapcore"

	"gin-artweb/internal/shared/database"
)

type ApiModel struct {
	database.StandardModel
	URL    string `gorm:"column:url;type:varchar(150);not null;uniqueIndex:idx_api_url_method;comment:HTTP的URL地址" json:"url"`
	Method string `gorm:"column:method;type:varchar(10);not null;uniqueIndex:idx_api_url_method;comment:请求方法" json:"method"`
	Label  string `gorm:"column:label;type:varchar(50);not null;index:label;comment:标签" json:"label"`
	Descr  string `gorm:"column:descr;type:varchar(254);comment:描述" json:"descr"`
}

func (m *ApiModel) TableName() string {
	return "customer_api"
}

func (m *ApiModel) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if m == nil {
		return nil
	}
	if err := m.StandardModel.MarshalLogObject(enc); err != nil {
		return err
	}
	enc.AddString("url", m.URL)
	enc.AddString("method", m.Method)
	enc.AddString("label", m.Label)
	enc.AddString("descr", m.Descr)
	return nil
}
