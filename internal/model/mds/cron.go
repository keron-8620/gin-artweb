package mds

import (
	"go.uber.org/zap/zapcore"

	"gin-artweb/internal/model/job"
	"gin-artweb/internal/shared/database"
)

type MdsCronModel struct {
	database.StandardModel
	MdsColonyID uint32            `gorm:"column:mds_colony_id;not null;comment:mds集群ID" json:"mds_colony_id"`
	MdsColony   MdsColonyModel    `gorm:"foreignKey:MdsColonyID;references:ID;constraint:OnDelete:CASCADE" json:"mds_colony"`
	ScheduleID  uint32            `gorm:"column:schedule_id;not null;comment:计划任务ID" json:"schedule_id"`
	Schedule    job.ScheduleModel `gorm:"foreignKey:ScheduleID;references:ID;constraint:OnDelete:CASCADE" json:"schedule"`
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
	enc.AddUint32("mds_colony_id", m.MdsColonyID)
	enc.AddUint32("schedule_id", m.ScheduleID)
	return nil
}
