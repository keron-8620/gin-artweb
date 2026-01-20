package record

import (
	"gin-artweb/api/common"
	"gin-artweb/api/jobs/script"
)

type ScriptRecordStandardOut struct {
	// 脚本执行记录ID
	ID uint32 `json:"id" example:"1"`

	// 创建时间
	CreatedAt string `json:"created_at" example:"2023-01-01 12:00:00"`

	// 更新时间
	UpdatedAt string `json:"updated_at" example:"2023-01-01 12:00:00"`

	// 触发类型(cron/api)
	TriggerType string `json:"trigger_type" example:"cron"`

	// 执行状态(0-待执行,1-执行中,2-成功,3-失败,4-超时,5-崩溃)
	Status int `json:"status" example:"2"`

	// 退出码
	ExitCode int `json:"exit_code" example:"0"`

	// 环境变量(JSON对象)
	EnvVars string `json:"env_vars,omitempty" example:"{\"ENV\":\"production\"}"`

	// 命令行参数
	CommandArgs string `json:"command_args,omitempty" example:"[\"--verbose\"]"`

	// 工作目录
	WorkDir string `json:"work_dir,omitempty" example:"/home/user/work"`

	// 超时时间(秒)
	Timeout int `json:"timeout" example:"300"`

	// 错误信息
	ErrorMessage string `json:"error_message,omitempty" example:""`

	// 用户名
	Username string `json:"username" example:"admin"`
}

type ScriptRecordDetailOut struct {
	ScriptRecordStandardOut

	// 脚本信息
	Script script.ScriptStandardOut `json:"script"`
}

// ScriptRecordReply 程序包响应结构
type ScriptRecordReply = common.APIReply[ScriptRecordDetailOut]

// PagScriptRecordReply 程序包的分页响应结构
type PagScriptRecordReply = common.APIReply[*common.Pag[ScriptRecordDetailOut]]
