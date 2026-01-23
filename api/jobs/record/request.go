package record

import "gin-artweb/api/common"

// CreateScriptRecordRequest 用于创建计划任务的请求结构体
//
// swagger:model CreateScriptRecordRequest
type CreateScriptRecordRequest struct {
	// 脚本ID
	ScriptID uint32 `json:"script_id" form:"script_id" binding:"required"`

	// 命令行参数
	CommandArgs string `json:"command_args" form:"command_args" binding:"omitempty"`

	// 环境变量 (JSON对象)
	EnvVars string `json:"env_vars" form:"env_vars" binding:"omitempty"`

	// 超时时间(秒)
	Timeout int `json:"timeout" form:"timeout" binding:"required"`

	// 工作目录
	WorkDir string `json:"work_dir" form:"work_dir" binding:"omitempty"`
}

// ListScriptRecordRequest 用于获取计划任务列表的请求结构体
// 支持分页查询和多种筛选条件
//
// swagger:model ListScriptRecordRequest
type ListScriptRecordRequest struct {
	common.StandardModelQuery

	// 筛选计划任务触发类型
	TriggerType string `form:"trigger_type" binding:"omitempty"`

	// 筛选脚本执行的任务状态
	Status int `form:"status" binding:"omitempty"`

	// 按脚本退出码筛选
	ExitCode int `form:"exit_code" binding:"omitempty"`

	// 按脚本ID筛选
	ScriptID uint32 `form:"script_id" binding:"omitempty"`

	// 按用户名筛选
	Username string `form:"username" binding:"omitempty"`
}

func (req *ListScriptRecordRequest) Query() (int, int, map[string]any) {
	page, size, query := req.BaseModelQuery.QueryMap(11)
	if req.TriggerType != "" {
		query["trigger_type = ?"] = req.TriggerType
	}
	if req.Status != 0 {
		query["status = ?"] = req.Status
	}
	if req.ExitCode != 0 {
		query["exit_code = ?"] = req.ExitCode
	}
	if req.ScriptID > 0 {
		query["script_id = ?"] = req.ScriptID
	}
	if req.Username != "" {
		query["username like ?"] = "%" + req.Username + "%"
	}
	return page, size, query
}
