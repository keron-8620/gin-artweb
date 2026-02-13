package server

import (
	"gorm.io/gorm"

	"gin-artweb/internal/infra/jobs/model"
)

func DBAutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.ScriptModel{},
		&model.ScriptRecordModel{},
		&model.ScheduleModel{},
	)
}
