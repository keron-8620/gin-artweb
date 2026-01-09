package server

import (
	"gorm.io/gorm"

	"gin-artweb/internal/resource/biz"
)

func dbAutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&biz.HostModel{},
		&biz.PackageModel{},
	)
}
