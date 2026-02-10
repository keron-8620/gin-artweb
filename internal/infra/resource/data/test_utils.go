package data

import (
	"gin-artweb/internal/infra/resource/biz"
	"gin-artweb/internal/shared/test"
)

// NewTestHostRepo 创建测试用的主机仓库实例
func NewTestHostRepo() *hostRepo {
	db := test.NewTestGormDBWithConfig(nil)
	db.AutoMigrate(&biz.HostModel{})
	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	return &hostRepo{
		log:      logger,
		gormDB:   db,
		timeouts: dbTimeout,
	}
}

// NewTestPackageRepo 创建测试用的软件包仓库实例
func NewTestPackageRepo() *packageRepo {
	db := test.NewTestGormDBWithConfig(nil)
	db.AutoMigrate(&biz.PackageModel{})
	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	return &packageRepo{
		log:      logger,
		gormDB:   db,
		timeouts: dbTimeout,
	}
}
