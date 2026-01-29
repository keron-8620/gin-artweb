package server

import (
	"gorm.io/gorm"

	"gin-artweb/internal/business/mds/biz"
)

func DBAutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&biz.MdsColonyModel{},
		&biz.MdsNodeModel{},
	)
}
