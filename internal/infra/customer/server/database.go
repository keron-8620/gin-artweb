package server

import (
	"gorm.io/gorm"

	"gin-artweb/internal/infra/customer/model"
)

func DBAutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.ApiModel{},
		&model.MenuModel{},
		&model.ButtonModel{},
		&model.RoleModel{},
		&model.UserModel{},
		&model.LoginRecordModel{},
	)
}
