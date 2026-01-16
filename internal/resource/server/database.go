package server

import (
	"gorm.io/gorm"

	"gin-artweb/internal/resource/biz"
)

func DBAutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&biz.HostModel{},
		&biz.PackageModel{},
	)
}
