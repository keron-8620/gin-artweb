// Package database 提供数据库连接和配置功能
// 支持多种数据库类型包括MySQL、PostgreSQL、SQLite、SQLServer和OpenGauss
package database

import (
	"log"
	"time"

	"emperror.dev/errors"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/database/driver/opengauss"
)

// NewGormConfig 创建 gorm 配置
func NewGormConfig(dbLog *log.Logger) *gorm.Config {
	var gc gorm.Config
	gc.SkipDefaultTransaction = false
	gc.FullSaveAssociations = false
	gc.TranslateError = true
	if dbLog != nil {
		gc.Logger = logger.New(dbLog, logger.Config{
			SlowThreshold:             time.Second,
			Colorful:                  false,
			IgnoreRecordNotFoundError: false,
			ParameterizedQueries:      false,
			LogLevel:                  logger.Info,
		})
	}
	return &gc
}

// InitGormDB 初始化GORM数据库连接
// c: 数据库配置信息
// gc: GORM配置信息
// 返回GORM数据库实例和可能的错误
func NewGormDB(c *config.DBConf, gc *gorm.Config) (*gorm.DB, error) {
	var (
		db      *gorm.DB // GORM数据库实例
		openErr error    // 数据库打开错误
	)

	// 根据数据库类型选择相应的驱动并建立连接
	switch c.Type {
	case "mysql":
		db, openErr = gorm.Open(mysql.Open(c.Dns), gc)
	case "postgres":
		db, openErr = gorm.Open(postgres.Open(c.Dns), gc)
	case "sqlite":
		db, openErr = gorm.Open(sqlite.Open(c.Dns), gc)
	case "sqlserver":
		db, openErr = gorm.Open(sqlserver.Open(c.Dns), gc)
	case "opengauss":
		db, openErr = gorm.Open(opengauss.Open(c.Dns), gc)
	default:
		// 不支持的数据库驱动类型
		return nil, errors.NewWithDetails("不支持的数据库驱动类型: %s", c.Type)
	}

	// 如果数据库连接打开失败，返回错误
	if openErr != nil {
		return nil, errors.Wrap(openErr, "数据库连接打开失败")
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, errors.Wrap(err, "获取数据库连接失败")
	}
	// 设置连接池参数
	sqlDB.SetMaxIdleConns(c.MaxIdleConns)                                    // 最大空闲连接数
	sqlDB.SetMaxOpenConns(c.MaxOpenConns)                                    // 最大打开连接数
	sqlDB.SetConnMaxLifetime(time.Duration(c.ConnMaxLifetime) * time.Second) // 连接最大生命周期
	sqlDB.SetConnMaxIdleTime(time.Duration(c.ConnMaxIdleTime) * time.Second) // 空闲连接最大存活时间
	return db, nil
}

// CloseGormDB 关闭GORM数据库连接
// db: GORM数据库实例
// 返回关闭操作可能产生的错误
func CloseGormDB(db *gorm.DB) error {
	// 检查数据库实例是否有效
	if db == nil {
		return errors.New("数据库实例为空")
	}

	// 获取底层数据库连接
	sqlDB, err := db.DB()
	if err != nil {
		return errors.Wrap(err, "获取数据库连接失败")
	}

	// 关闭数据库连接
	return errors.Wrap(sqlDB.Close(), "数据库连接关闭失败")
}
