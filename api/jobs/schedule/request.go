package schedule

import "gin-artweb/api/common"

// CreateScheduleRequest 用于创建计划任务的请求结构体
//
// swagger:model CreateScheduleRequest
type CreateScheduleRequest struct {
	// 计划任务名称
	// required: true
	// example: "每日备份"
	Name string `json:"name" binding:"required,max=50"`

	// Cron 表达式
	// required: true
	// example: "0 2 * * *"
	Specification string `json:"specification" binding:"required"`

	// 是否启用
	// required: false
	// default: true
	IsEnabled bool `json:"is_enabled"`

	// 环境变量 (JSON对象字符串)
	// required: false
	// example: "{\"ENV\":\"production\"}"
	EnvVars string `json:"env_vars,omitempty"`

	// 命令行参数
	// required: false
	// example: "[\"--verbose\", \"--output=/tmp\"]"
	CommandArgs string `json:"command_args,omitempty"`

	// 工作目录
	// required: false
	// example: "/home/user/work"
	WorkDir string `json:"work_dir,omitempty"`

	// 超时时间(秒)
	// required: false
	// default: 300
	// minimum: 1
	Timeout int `json:"timeout,omitempty"`

	// 是否重试
	// required: false
	// default: false
	IsRetry bool `json:"is_retry"`

	// 重试间隔时间(秒)
	// required: false
	// default: 300
	// minimum: 1
	RetryInterval int `json:"retry_interval"`

	// 最大重试次数
	// required: false
	// default: 3
	MaxRetries int `json:"max_retries"`

	// 脚本ID
	// required: true
	// example: 1
	ScriptID uint32 `json:"script_id" binding:"required"`
}

// UpdateScheduleRequest 用于更新计划任务的请求结构体
//
// swagger:model UpdateScheduleRequest
type UpdateScheduleRequest struct {
	// 计划任务名称
	// required: true
	// example: "每日备份"
	Name string `json:"name" binding:"required,max=50"`

	// Cron 表达式
	// required: true
	// example: "0 2 * * *"
	Specification string `json:"specification" binding:"required"`

	// 是否启用
	// required: false
	// default: true
	IsEnabled bool `json:"is_enabled"`

	// 环境变量 (JSON对象字符串)
	// required: false
	// example: "{\"ENV\":\"production\"}"
	EnvVars string `json:"env_vars,omitempty"`

	// 命令行参数
	// required: false
	// example: "[\"--verbose\", \"--output=/tmp\"]"
	CommandArgs string `json:"command_args,omitempty"`

	// 工作目录
	// required: false
	// example: "/home/user/work"
	WorkDir string `json:"work_dir,omitempty"`

	// 超时时间(秒)
	// required: false
	// default: 300
	// minimum: 1
	Timeout int `json:"timeout,omitempty"`

	// 是否重试
	// required: false
	// default: false
	IsRetry bool `json:"is_retry"`

	// 重试间隔时间(秒)
	// required: false
	// default: 300
	// minimum: 1
	RetryInterval int `json:"retry_interval"`

	// 最大重试次数
	// required: false
	// default: 3
	MaxRetries int `json:"max_retries"`

	// 脚本ID
	// required: true
	// example: 1
	ScriptID uint32 `json:"script_id" binding:"required"`
}

// ListScheduleRequest 用于获取计划任务列表的请求结构体
// 支持分页查询和多种筛选条件
//
// swagger:model ListScheduleRequest
type ListScheduleRequest struct {
	common.StandardModelQuery

	// 按名称搜索
	// required: false
	Name string `form:"name"`

	// 按启用状态筛选
	// required: false
	IsEnabled *bool `form:"is_enabled"`

	// 按脚本ID筛选
	// required: false
	ScriptID uint32 `form:"script_id"`

	// 按用户名筛选
	// required: false
	Username string `json:"username" binding:"omitempty"`
}

func (req *ListScheduleRequest) Query() (int, int, map[string]any) {
	page, size, query := req.BaseModelQuery.QueryMap(14)
	if req.Name != "" {
		query["name like ?"] = "%" + req.Name + "%"
	}
	if req.IsEnabled != nil {
		query["is_enabled = ?"] = *req.IsEnabled
	}
	if req.ScriptID > 0 {
		query["script_id = ?"] = req.ScriptID
	}
	if req.Username != "" {
		query["username like ?"] = "%" + req.Username + "%"
	}
	return page, size, query
}
