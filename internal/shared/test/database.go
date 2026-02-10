package test

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewTestGormDBWithConfig(config *gorm.Config) *gorm.DB {
	if config == nil {
		config = &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		}
	}

	db, err := gorm.Open(sqlite.Open("file::memory:"), config)
	if err != nil {
		panic(err)
	}
	return db
}

func CloseTestGormDB(db *gorm.DB) error {
	if db == nil {
		return nil
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
