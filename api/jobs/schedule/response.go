package schedule

import (
	"gin-artweb/api/common"
	"gin-artweb/api/jobs/script"
)

type ScheduleOutBase struct {
	// 计划任务ID
	ID uint32 `json:"id" example:"1"`
	// 创建时间
	CreatedAt string `json:"created_at" example:"2023-01-01 12:00:00"`
	// 更新时间
	UpdatedAt string `json:"updated_at" example:"2023-01-01 12:00:00"`
	// 名称
	Name string `json:"name" example:"test"`
	// Cron 表达式
	Specification string `json:"specification" example:"0 12 * * 1-5"`
	// 是否启用
	IsEnabled bool `json:"is_enabled" example:"true"`
	// 环境变量 (JSON对象字符串)
	EnvVars map[string]string `json:"env_vars" example:"{}"`
	// 命令行参数
	CommandArgs string `json:"command_args" example:""`
	// 工作目录
	WorkDir string `json:"work_dir" example:""`
	// 超时时间(秒)
	Timeout int `json:"timeout" example:"300"`
	// 用户名
	UserName string `json:"user_name" example:"admin"`
}

type ScheduleOut struct {
	ScheduleOutBase
	// 脚本
	Script *script.ScriptOutBase `json:"script"`
}

// ScheduleReply 程序包响应结构
type ScheduleReply = common.APIReply[ScheduleOut]

// PagScheduleReply 程序包的分页响应结构
type PagScheduleReply = common.APIReply[*common.Pag[ScheduleOut]]
