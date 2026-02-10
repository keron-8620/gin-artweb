package data

import (
	"time"

	"github.com/patrickmn/go-cache"

	"gin-artweb/internal/infra/customer/biz"
	"gin-artweb/internal/shared/auth"
	"gin-artweb/internal/shared/test"
)

// NewTestPermissionRepo 创建测试用的权限仓库实例
func NewTestPermissionRepo() *permissionRepo {
	db := test.NewTestGormDBWithConfig(nil)
	db.AutoMigrate(&biz.PermissionModel{})
	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	enforcer, _ := auth.NewCasbinEnforcer()
	return &permissionRepo{
		log:      logger,
		gormDB:   db,
		timeouts: dbTimeout,
		enforcer: enforcer,
	}
}

// NewTestMenuRepo 创建测试用的菜单仓库实例
func NewTestMenuRepo() *menuRepo {
	db := test.NewTestGormDBWithConfig(nil)
	db.AutoMigrate(&biz.MenuModel{}, &biz.PermissionModel{})
	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	enforcer, _ := auth.NewCasbinEnforcer()
	return &menuRepo{
		log:      logger,
		gormDB:   db,
		timeouts: dbTimeout,
		enforcer: enforcer,
	}
}

// NewTestButtonRepo 创建测试用的按钮仓库实例
func NewTestButtonRepo() *buttonRepo {
	db := test.NewTestGormDBWithConfig(nil)
	db.AutoMigrate(&biz.ButtonModel{}, &biz.PermissionModel{})
	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	enforcer, _ := auth.NewCasbinEnforcer()
	return &buttonRepo{
		log:      logger,
		gormDB:   db,
		timeouts: dbTimeout,
		enforcer: enforcer,
	}
}

// NewTestRoleRepo 创建测试用的角色仓库实例
func NewTestRoleRepo() *roleRepo {
	db := test.NewTestGormDBWithConfig(nil)
	db.AutoMigrate(&biz.RoleModel{}, &biz.PermissionModel{}, &biz.MenuModel{}, &biz.ButtonModel{})
	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	enforcer, _ := auth.NewCasbinEnforcer()
	return &roleRepo{
		log:      logger,
		gormDB:   db,
		timeouts: dbTimeout,
		enforcer: enforcer,
	}
}

// NewTestUserRepo 创建测试用的用户仓库实例
func NewTestUserRepo() *userRepo {
	db := test.NewTestGormDBWithConfig(nil)
	db.AutoMigrate(&biz.UserModel{})
	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	return &userRepo{
		log:      logger,
		gormDB:   db,
		timeouts: dbTimeout,
	}
}

// NewTestLoginRecordRepo 创建测试用的登录记录仓库实例
func NewTestLoginRecordRepo() *loginRecordRepo {
	db := test.NewTestGormDBWithConfig(nil)
	db.AutoMigrate(&biz.LoginRecordModel{})
	dbTimeout := test.NewTestDBTimeouts()
	logger := test.NewTestZapLogger()
	return &loginRecordRepo{
		log:      logger,
		gormDB:   db,
		timeouts: dbTimeout,
		cache:    cache.New(5*time.Minute, 10*time.Minute),
		maxNum:   5,
		ttl:      5 * time.Minute,
	}
}
