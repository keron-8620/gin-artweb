package test

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewTestGormDBWithConfig(config *gorm.Config) *gorm.DB {
	return openTestGormDB("file::memory:?cache=shared", config)
}

// NewNamedTestGormDBWithConfig 创建一个具名的内存数据库。
// 同名连接共享数据，不同名称完全隔离，避免测试包之间互相污染。
func NewNamedTestGormDBWithConfig(name string, config *gorm.Config) *gorm.DB {
	if config == nil {
		config = &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		}
	}

	name = strings.NewReplacer("/", "_", "\\", "_", " ", "_").Replace(name)
	if name == "" {
		name = uuid.NewString()
	}
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", name)
	return openTestGormDB(dsn, config)
}

func openTestGormDB(dsn string, config *gorm.Config) *gorm.DB {
	if config == nil {
		config = &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		}
	}

	db, err := gorm.Open(sqlite.Open(dsn), config)
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
