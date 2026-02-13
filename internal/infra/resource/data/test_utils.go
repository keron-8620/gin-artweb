package data

import (
	"gin-artweb/internal/infra/resource/model"
	"gin-artweb/internal/shared/test"
)

// NewTestHostRepo 创建测试用的主机仓库实例
func NewTestHostRepo() *HostRepo {
	db := test.NewTestGormDBWithConfig(nil)
	db.AutoMigrate(&model.HostModel{})
	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	return &HostRepo{
		log:      logger,
		gormDB:   db,
		timeouts: dbTimeout,
	}
}

// NewTestPackageRepo 创建测试用的软件包仓库实例
func NewTestPackageRepo() *PackageRepo {
	db := test.NewTestGormDBWithConfig(nil)
	db.AutoMigrate(&model.PackageModel{})
	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	return &PackageRepo{
		log:      logger,
		gormDB:   db,
		timeouts: dbTimeout,
	}
}
