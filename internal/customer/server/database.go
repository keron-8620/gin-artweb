package server

import (
	"gorm.io/gorm"

	"gin-artweb/internal/customer/biz"
)

func dbAutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&biz.PermissionModel{},
		&biz.MenuModel{},
		&biz.ButtonModel{},
		&biz.RoleModel{},
		&biz.UserModel{},
		&biz.LoginRecordModel{},
	)
}
