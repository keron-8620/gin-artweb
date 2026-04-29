package oes

import (
	"go.uber.org/zap/zapcore"

	"gin-artweb/internal/model/job"
	"gin-artweb/internal/shared/database"
)

type OesCronModel struct {
	database.BaseModel
	OesColonyID uint32            `gorm:"column:oes_colony_id;not null;comment:oes集群ID" json:"oes_colony_id"`
	OesColony   OesColonyModel    `gorm:"foreignKey:OesColonyID;references:ID;constraint:OnDelete:CASCADE" json:"oes_colony"`
	ScheduleID  uint32            `gorm:"column:schedule_id;not null;comment:计划任务ID" json:"schedule_id"`
	Schedule    job.ScheduleModel `gorm:"foreignKey:ScheduleID;references:ID;constraint:OnDelete:CASCADE" json:"schedule"`
}

func (m *OesCronModel) TableName() string {
	return "oes_cron"
}

func (m *OesCronModel) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if m == nil {
		return nil
	}
	if err := m.BaseModel.MarshalLogObject(enc); err != nil {
		return err
	}
	enc.AddUint32("oes_colony_id", m.OesColonyID)
	enc.AddUint32("schedule_id", m.ScheduleID)
	return nil
}

func ListOesCronModelToUint32s(ms []OesCronModel) []uint32 {
	if len(ms) == 0 {
		return []uint32{}
	}

	ids := make([]uint32, len(ms))
	for i, m := range ms {
		ids[i] = m.ID
	}
	return ids
}
