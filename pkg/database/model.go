package database

import (
	"time"

	"go.uber.org/zap/zapcore"
)

type BaseModel struct {
	Id uint32 `gorm:"column:id;primary_key;AUTO_INCREMENT;comment:编号" json:"id"`
}

func (m *BaseModel) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddUint32("id", m.Id)
	return nil
}

type StandardModel struct {
	BaseModel

	CreatedAt time.Time `gorm:"column:created_at;comment:创建时间" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;comment:修改时间" json:"updated_at"`
}

func (m *StandardModel) CreateSetTime() {
	now := time.Now()
	m.CreatedAt = now
	m.UpdatedAt = now
}

func (m *StandardModel) UpdateSetTime() {
	m.UpdatedAt = time.Now()
}

func (m *StandardModel) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	m.BaseModel.MarshalLogObject(enc)
	enc.AddTime("created_at", m.CreatedAt)
	enc.AddTime("updated_at", m.UpdatedAt)
	return nil
}
