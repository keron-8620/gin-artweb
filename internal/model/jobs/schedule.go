package jobs

import (
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap/zapcore"

	"gin-artweb/internal/model/common"
	"gin-artweb/internal/shared/database"
)

type ScheduleModel struct {
	database.StandardModel
	Name          string      `gorm:"column:name;type:varchar(50);not null;uniqueIndex;comment:名称" json:"name"`
	Specification string      `gorm:"column:specification;type:text;comment:条件" json:"specification"`
	IsEnabled     bool        `gorm:"column:is_enabled;type:boolean;comment:是否启用" json:"is_enabled"`
	EnvVars       string      `gorm:"column:env_vars;type:json;comment:环境变量(JSON对象)" json:"env_vars"`
	CommandArgs   string      `gorm:"column:command_args;type:varchar(254);comment:命令行参数" json:"command_args"`
	WorkDir       string      `gorm:"column:work_dir;type:varchar(255);comment:工作目录" json:"work_dir"`
	Timeout       int         `gorm:"column:timeout;type:int;not null;default:300;comment:超时时间(秒)" json:"timeout"`
	IsRetry       bool        `gorm:"column:is_retry;type:boolean;default:false;comment:是否启用重试" json:"is_retry"`
	RetryInterval int         `gorm:"column:retry_interval;type:int;default:60;comment:重试间隔(秒)" json:"retry_interval"`
	MaxRetries    int         `gorm:"column:max_retries;type:int;default:3;comment:最大重试次数" json:"max_retries"`
	Username      string      `gorm:"column:username;type:varchar(50);comment:用户名" json:"username"`
	ScriptID      uint32      `gorm:"column:script_id;not null;index;comment:计划任务ID" json:"script_id"`
	Script        ScriptModel `gorm:"foreignKey:ScriptID;references:ID" json:"script"`
}

func (m *ScheduleModel) TableName() string {
	return "jobs_schedule"
}

func (m *ScheduleModel) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if m == nil {
		return nil
	}
	if err := m.StandardModel.MarshalLogObject(enc); err != nil {
		return err
	}
	enc.AddString("name", m.Name)
	enc.AddString("specification", m.Specification)
	enc.AddBool("is_enabled", m.IsEnabled)
	enc.AddString("env_vars", m.EnvVars)
	enc.AddString("command_args", m.CommandArgs)
	enc.AddString("work_dir", m.WorkDir)
	enc.AddInt("timeout", m.Timeout)
	enc.AddString("username", m.Username)
	enc.AddUint32("script_id", m.ScriptID)
	return nil
}

type ScheduleJobInfo struct {
	EntryID    cron.EntryID `json:"entry_id"`
	ScheduleID uint32       `json:"schedule_id"`
	NextRun    time.Time    `json:"next_run"`
	PrevRun    time.Time    `json:"prev_run"`
}

// CreateScheduleRequest 用于创建计划任务的请求结构体
//
// swagger:model CreateScheduleRequest
type CreateScheduleRequest struct {
	// 计划任务名称
	Name string `json:"name" binding:"required,max=50"`

	// Cron 表达式
	Specification string `json:"specification" binding:"required"`

	// 是否启用
	IsEnabled bool `json:"is_enabled"`

	// 环境变量(JSON对象)
	EnvVars string `json:"env_vars,omitempty"`

	// 命令行参数
	CommandArgs string `json:"command_args,omitempty"`

	// 工作目录
	WorkDir string `json:"work_dir,omitempty"`

	// 超时时间(秒)
	Timeout int `json:"timeout,omitempty"`

	// 是否重试
	IsRetry bool `json:"is_retry"`

	// 重试间隔时间(秒)
	RetryInterval int `json:"retry_interval"`

	// 最大重试次数
	MaxRetries int `json:"max_retries"`

	// 脚本ID
	ScriptID uint32 `json:"script_id" binding:"required"`
}

// UpdateScheduleRequest 用于更新计划任务的请求结构体
//
// swagger:model UpdateScheduleRequest
type UpdateScheduleRequest struct {
	// 计划任务名称
	Name string `json:"name" binding:"required,max=50"`

	// Cron 表达式
	Specification string `json:"specification" binding:"required"`

	// 是否启用
	IsEnabled bool `json:"is_enabled"`

	// 环境变量(JSON对象)
	EnvVars string `json:"env_vars,omitempty"`

	// 命令行参数
	CommandArgs string `json:"command_args,omitempty"`

	// 工作目录
	WorkDir string `json:"work_dir,omitempty"`

	// 超时时间(秒)
	Timeout int `json:"timeout,omitempty"`

	// 是否重试
	IsRetry bool `json:"is_retry"`

	// 重试间隔时间(秒)
	RetryInterval int `json:"retry_interval"`

	// 最大重试次数
	MaxRetries int `json:"max_retries"`

	// 脚本ID
	ScriptID uint32 `json:"script_id" binding:"required"`
}

// ListScheduleRequest 用于获取计划任务列表的请求结构体
// 支持分页查询和多种筛选条件
//
// swagger:model ListScheduleRequest
type ListScheduleRequest struct {
	common.StandardModelQuery

	// 名称
	Name string `form:"name"`

	// 是否启用
	IsEnabled *bool `form:"is_enabled"`

	// 脚本ID
	ScriptID uint32 `form:"script_id"`

	// 用户名
	Username string `form:"username" binding:"omitempty"`
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

type ScheduleStandardOut struct {
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

	// 环境变量(JSON对象)
	EnvVars string `json:"env_vars" example:"{}"`

	// 命令行参数
	CommandArgs string `json:"command_args" example:""`

	// 工作目录
	WorkDir string `json:"work_dir" example:""`

	// 超时时间(秒)
	Timeout int `json:"timeout" example:"300"`

	// 是否重试
	IsRetry bool `json:"is_retry"`

	// 重试间隔时间(秒)
	RetryInterval int `json:"retry_interval"`

	// 最大重试次数
	MaxRetries int `json:"max_retries"`

	// 用户名
	Username string `json:"username" example:"admin"`
}

type ScheduleDetailOut struct {
	ScheduleStandardOut

	// 脚本
	Script *ScriptStandardOut `json:"script"`
}

// ScheduleReply 程序包响应结构
type ScheduleReply = common.APIReply[ScheduleDetailOut]

// PagScheduleReply 程序包的分页响应结构
type PagScheduleReply = common.APIReply[*common.Pag[ScheduleDetailOut]]

func ScheduleToStandardOut(
	m ScheduleModel,
) *ScheduleStandardOut {
	return &ScheduleStandardOut{
		ID:            m.ID,
		CreatedAt:     m.CreatedAt.Format(time.DateTime),
		UpdatedAt:     m.UpdatedAt.Format(time.DateTime),
		Name:          m.Name,
		Specification: m.Specification,
		IsEnabled:     m.IsEnabled,
		EnvVars:       m.EnvVars,
		CommandArgs:   m.CommandArgs,
		WorkDir:       m.WorkDir,
		Timeout:       m.Timeout,
		IsRetry:       m.IsRetry,
		MaxRetries:    m.MaxRetries,
		RetryInterval: m.RetryInterval,
		Username:      m.Username,
	}
}

func ScheduleToDetailOut(
	m ScheduleModel,
) *ScheduleDetailOut {
	var script *ScriptStandardOut
	if m.Script.ID != 0 {
		script = ScriptModelToStandardOut(m.Script)
	}
	return &ScheduleDetailOut{
		ScheduleStandardOut: *ScheduleToStandardOut(m),
		Script:              script,
	}
}

func ListScheduledToDetailOut(
	rms *[]ScheduleModel,
) *[]ScheduleDetailOut {
	if rms == nil {
		return &[]ScheduleDetailOut{}
	}

	ms := *rms
	mso := make([]ScheduleDetailOut, 0, len(ms))
	if len(ms) > 0 {
		for _, m := range ms {
			mo := ScheduleToDetailOut(m)
			mso = append(mso, *mo)
		}
	}
	return &mso
}
