package mds

import (
	"go.uber.org/zap/zapcore"

	"gin-artweb/internal/model/job"
	"gin-artweb/internal/shared/database"
)

type MdsCronModel struct {
	database.BaseModel
	MdsColonyID uint32            `gorm:"column:mds_colony_id;not null;uniqueIndex:uk_mds_cron_colony_schedule;comment:mds集群ID"`
	MdsColony   MdsColonyModel    `gorm:"foreignKey:MdsColonyID;references:ID;constraint:OnDelete:CASCADE" json:"mds_colony"`
	ScheduleID  uint32            `gorm:"column:schedule_id;not null;uniqueIndex:uk_mds_cron_colony_schedule;comment:计划任务ID"`
	Schedule    job.ScheduleModel `gorm:"foreignKey:ScheduleID;references:ID;constraint:OnDelete:CASCADE" json:"schedule"`
}

func (m *MdsCronModel) TableName() string {
	return "mds_cron"
}

func (m *MdsCronModel) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if m == nil {
		return nil
	}
	if err := m.BaseModel.MarshalLogObject(enc); err != nil {
		return err
	}
	enc.AddUint32("mds_colony_id", m.MdsColonyID)
	enc.AddUint32("schedule_id", m.ScheduleID)
	return nil
}

func ListMdsCronModelToUint32s(ms []MdsCronModel) []uint32 {
	if len(ms) == 0 {
		return []uint32{}
	}

	ids := make([]uint32, len(ms))
	for i, m := range ms {
		ids[i] = m.ID
	}
	return ids
}
