package oes

import (
	"go.uber.org/zap/zapcore"

	"gin-artweb/internal/shared/database"

	"gin-artweb/internal/model/jobs"
)

type OesCronModel struct {
	database.StandardModel
	TaskName    string             `gorm:"column:task_name;type:varchar(50);not null;uniqueIndex:idx_task_colony_num;comment:任务名称" json:"task_name"`
	OesColonyID uint32             `gorm:"column:oes_colony_id;not null;comment:oes集群ID" json:"oes_colony_id"`
	OesColony   OesColonyModel     `gorm:"foreignKey:OesColonyID;references:ID;constraint:OnDelete:CASCADE" json:"oes_colony"`
	ScheduleID  uint32             `gorm:"column:schedule_id;not null;uniqueIndex;comment:计划ID" json:"schedule_id"`
	Schedule    jobs.ScheduleModel `gorm:"foreignKey:ScheduleID;references:ID;constraint:OnDelete:CASCADE;constraint:OnUpdate:CASCADE" json:"schedule"`
}

func (m *OesCronModel) TableName() string {
	return "oes_cron"
}

func (m *OesCronModel) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if m == nil {
		return nil
	}
	if err := m.StandardModel.MarshalLogObject(enc); err != nil {
		return err
	}
	enc.AddString("task_name", m.TaskName)
	enc.AddUint32("oes_colony_id", m.OesColonyID)
	enc.AddUint32("schedule_id", m.ScheduleID)
	return nil
}
