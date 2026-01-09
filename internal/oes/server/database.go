package server

import (
	"gorm.io/gorm"

	"gin-artweb/internal/oes/biz"
)

func dbAutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&biz.OesColonyModel{},
		&biz.OesNodeModel{},
	)
}
