package server

import (
	"gorm.io/gorm"

	"gin-artweb/internal/business/mon/biz"
)

func DBAutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&biz.MonNodeModel{},
	)
}
