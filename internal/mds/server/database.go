package server

import (
	"gorm.io/gorm"

	"gin-artweb/internal/mds/biz"
)

func dbAutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&biz.MdsColonyModel{},
		&biz.MdsNodeModel{},
	)
}
