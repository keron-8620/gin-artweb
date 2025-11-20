package server

import (
	"go.uber.org/zap"
	"gorm.io/gorm"

	"gin-artweb/internal/customer/biz"
)

func dbAutoMigrate(db *gorm.DB, logger *zap.Logger) error {
	if err := db.AutoMigrate(
		&biz.PermissionModel{},
		&biz.MenuModel{},
		&biz.ButtonModel{},
		&biz.RoleModel{},
		&biz.UserModel{},
		&biz.LoginRecordModel{},
	); err != nil {
		logger.Error(
			"数据库自动迁移customer模型失败", 
			zap.Error(err),
		)
		return err
	}
	return nil
}
