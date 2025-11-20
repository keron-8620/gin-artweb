package server

import (
	"go.uber.org/zap"
	"gorm.io/gorm"

	"gin-artweb/internal/resource/biz"
)

func dbAutoMigrate(db *gorm.DB, logger *zap.Logger) error {
	if err := db.AutoMigrate(
		&biz.HostModel{},
		&biz.PackageModel{},
	); err != nil {
		logger.Error(
			"数据库自动迁移resource模型失败", 
			zap.Error(err),
		)
		return err
	}
	return nil
}
