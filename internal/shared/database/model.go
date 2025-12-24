package database

import (
	"time"

	"go.uber.org/zap/zapcore"
	"gorm.io/gorm"
)

type BaseModel struct {
	ID uint32 `gorm:"column:id;primary_key;AUTO_INCREMENT;comment:编号" json:"id"`
}

func (m *BaseModel) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddUint32("id", m.ID)
	return nil
}

type StandardModel struct {
	BaseModel

	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime;comment:修改时间" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index;comment:删除时间" json:"deleted_at,omitempty"`
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
	if err := m.BaseModel.MarshalLogObject(enc); err != nil {
		return err
	}
	enc.AddTime("created_at", m.CreatedAt)
	enc.AddTime("updated_at", m.UpdatedAt)
	if m.DeletedAt.Valid {
		enc.AddTime("deleted_at", m.DeletedAt.Time)
	}
	return nil
}
