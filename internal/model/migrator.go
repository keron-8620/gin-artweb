package model

import (
	"gorm.io/gorm"

	"gin-artweb/internal/model/job"
	"gin-artweb/internal/model/mds"
	"gin-artweb/internal/model/mon"
	"gin-artweb/internal/model/oes"
	"gin-artweb/internal/model/resource"
	"gin-artweb/internal/model/sys"
)

func DBAutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		// 系统模型
		&sys.ApiModel{},
		&sys.MenuModel{},
		&sys.ButtonModel{},
		&sys.RoleModel{},
		&sys.UserModel{},
		&sys.LoginRecordModel{},

		// 任务模型
		&job.ScriptModel{},
		&job.ScriptRecordModel{},
		&job.ScheduleModel{},

		// 资源模型
		&resource.HostModel{},
		&resource.PackageModel{},

		// mon模型
		&mon.MonNodeModel{},

		// mds模型
		&mds.MdsColonyModel{},
		&mds.MdsNodeModel{},
		&mds.MdsCronModel{},

		// oes模型
		&oes.OesColonyModel{},
		&oes.OesNodeModel{},
		&oes.OesCronModel{},
	)
}
