package mds

import (
	"go.uber.org/zap/zapcore"

	"gin-artweb/internal/shared/database"

	"gin-artweb/internal/model/jobs"
)

type MdsCronModel struct {
	database.StandardModel
	TaskName    string             `gorm:"column:task_name;type:varchar(50);not null;uniqueIndex:idx_task_colony;comment:任务名称" json:"task_name"`
	MdsColonyID uint32             `gorm:"column:mds_colony_id;not null;uniqueIndex:idx_task_colony;comment:mds集群ID" json:"mds_colony_id"`
	MdsColony   MdsColonyModel     `gorm:"foreignKey:MdsColonyID;references:ID;constraint:OnDelete:CASCADE" json:"mds_colony"`
	ScheduleID  uint32             `gorm:"column:schedule_id;not null;uniqueIndex;comment:计划ID" json:"schedule_id"`
	Schedule    jobs.ScheduleModel `gorm:"foreignKey:ScheduleID;references:ID;constraint:OnDelete:CASCADE" json:"schedule"`
}

func (m *MdsCronModel) TableName() string {
	return "mds_cron"
}

func (m *MdsCronModel) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if m == nil {
		return nil
	}
	if err := m.StandardModel.MarshalLogObject(enc); err != nil {
		return err
	}
	enc.AddString("task_name", m.TaskName)
	enc.AddUint32("mds_colony_id", m.MdsColonyID)
	enc.AddUint32("schedule_id", m.ScheduleID)
	return nil
}
