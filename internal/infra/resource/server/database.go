package server

import (
	"gorm.io/gorm"

	"gin-artweb/internal/infra/resource/model"
)

func DBAutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.HostModel{},
		&model.PackageModel{},
	)
}
