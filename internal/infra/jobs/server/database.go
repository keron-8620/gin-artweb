package server

import (
	"gorm.io/gorm"

	"gin-artweb/internal/infra/jobs/biz"
)

func DBAutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&biz.ScriptModel{},
		&biz.ScriptRecordModel{},
		&biz.ScheduleModel{},
	)
}
