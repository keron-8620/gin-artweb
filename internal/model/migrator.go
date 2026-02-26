package model

import (
	"gorm.io/gorm"

	"gin-artweb/internal/model/customer"
	"gin-artweb/internal/model/jobs"
	"gin-artweb/internal/model/mds"
	"gin-artweb/internal/model/mon"
	"gin-artweb/internal/model/oes"
	"gin-artweb/internal/model/resource"
)

func DBAutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		// 客户模型
		&customer.ApiModel{},
		&customer.MenuModel{},
		&customer.ButtonModel{},
		&customer.RoleModel{},
		&customer.UserModel{},
		&customer.LoginRecordModel{},

		// 任务模型
		&jobs.ScriptModel{},
		&jobs.ScriptRecordModel{},
		&jobs.ScheduleModel{},

		// 资源模型
		&resource.HostModel{},
		&resource.PackageModel{},

		// mon模型
		&mon.MonNodeModel{},

		// mds模型
		&mds.MdsColonyModel{},
		&mds.MdsNodeModel{},

		// oes模型
		&oes.OesColonyModel{},
		&oes.OesNodeModel{},
	)
}
