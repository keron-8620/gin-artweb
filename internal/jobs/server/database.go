package server

import (
	"go.uber.org/zap"
	"gorm.io/gorm"

	"gin-artweb/internal/jobs/biz"
)

func dbAutoMigrate(db *gorm.DB, logger *zap.Logger) error {
	if err := db.AutoMigrate(
		&biz.ScriptModel{},
		&biz.ScriptRecordModel{},
		&biz.ScheduleModel{},
	); err != nil {
		logger.Error(
			"数据库自动迁移jobs模型失败",
			zap.Error(err),
		)
		return err
	}
	return nil
}
