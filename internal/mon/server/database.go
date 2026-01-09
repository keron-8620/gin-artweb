package server

import (
	"gorm.io/gorm"

	"gin-artweb/internal/mon/biz"
)

func dbAutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&biz.MonNodeModel{},
	)
}
