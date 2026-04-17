package job

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
	CreateType    int8        `gorm:"column:create_type;not null;index;comment:创建类型, 0:手动, 1:业务" json:"create_type"`
	Username      string      `gorm:"column:username;type:varchar(50);comment:用户名" json:"username"`
	ScriptID      uint32      `gorm:"column:script_id;not null;index;comment:计划任务ID" json:"script_id"`
	Script        ScriptModel `gorm:"foreignKey:ScriptID;references:ID" json:"script"`
}

func (m *ScheduleModel) TableName() string {
	return "job_schedule"
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
	enc.AddBool("is_retry", m.IsRetry)
	enc.AddInt("retry_interval", m.RetryInterval)
	enc.AddInt("max_retries", m.MaxRetries)
	enc.AddInt8("create_type", m.CreateType)
	enc.AddString("username", m.Username)
	enc.AddUint32("script_id", m.ScriptID)
	return nil
}

func (m *ScheduleModel) ToUpdateMap() map[string]any {
	return map[string]any{
		"name":           m.Name,
		"specification":  m.Specification,
		"is_enabled":     m.IsEnabled,
		"env_vars":       m.EnvVars,
		"command_args":   m.CommandArgs,
		"work_dir":       m.WorkDir,
		"timeout":        m.Timeout,
		"is_retry":       m.IsRetry,
		"retry_interval": m.RetryInterval,
		"max_retries":    m.MaxRetries,
		"create_type":    m.CreateType,
		"username":       m.Username,
		"script_id":      m.ScriptID,
	}
}

func ListScheduleModelToUint32s(ms []ScheduleModel) []uint32 {
	if len(ms) == 0 {
		return []uint32{}
	}

	ids := make([]uint32, len(ms))
	for i, m := range ms {
		ids[i] = m.ID
	}
	return ids
}

type ScheduleJobInfo struct {
	EntryID    cron.EntryID `json:"entry_id"`
	ScheduleID uint32       `json:"schedule_id"`
	NextRun    time.Time    `json:"next_run"`
	PrevRun    time.Time    `json:"prev_run"`
}

func (info ScheduleJobInfo) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddInt("entry_id", int(info.EntryID))
	enc.AddUint32("schedule_id", info.ScheduleID)
	enc.AddTime("next_run", info.NextRun)
	enc.AddTime("prev_run", info.PrevRun)
	return nil
}

// ScheduleUpsertDTO 用于创建计划任务的请求结构体
//
// swagger:model ScheduleUpsertDTO
type ScheduleUpsertDTO struct {
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

	// 创建类型
	CreateType int8 `json:"create_type" example:"0"`

	// 脚本ID
	ScriptID uint32 `json:"script_id" binding:"required"`
}

func (dto *ScheduleUpsertDTO) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if dto == nil {
		return nil
	}
	enc.AddString("name", dto.Name)
	enc.AddString("specification", dto.Specification)
	enc.AddBool("is_enabled", dto.IsEnabled)
	enc.AddString("env_vars", dto.EnvVars)
	enc.AddString("command_args", dto.CommandArgs)
	enc.AddString("work_dir", dto.WorkDir)
	enc.AddInt("timeout", dto.Timeout)
	enc.AddBool("is_retry", dto.IsRetry)
	enc.AddInt("retry_interval", dto.RetryInterval)
	enc.AddInt("max_retries", dto.MaxRetries)
	enc.AddInt8("create_type", dto.CreateType)
	enc.AddUint32("script_id", dto.ScriptID)
	return nil
}

func (dto *ScheduleUpsertDTO) ToModel(username string) ScheduleModel {
	return ScheduleModel{
		Name:          dto.Name,
		Specification: dto.Specification,
		IsEnabled:     dto.IsEnabled,
		EnvVars:       dto.EnvVars,
		CommandArgs:   dto.CommandArgs,
		WorkDir:       dto.WorkDir,
		Timeout:       dto.Timeout,
		IsRetry:       dto.IsRetry,
		RetryInterval: dto.RetryInterval,
		MaxRetries:    dto.MaxRetries,
		CreateType:    dto.CreateType,
		Username:      username,
		ScriptID:      dto.ScriptID,
	}
}

func (dto *ScheduleUpsertDTO) ToUpdateMap(username string) map[string]any {
	return map[string]any{
		"name":           dto.Name,
		"specification":  dto.Specification,
		"is_enabled":     dto.IsEnabled,
		"env_vars":       dto.EnvVars,
		"command_args":   dto.CommandArgs,
		"work_dir":       dto.WorkDir,
		"timeout":        dto.Timeout,
		"is_retry":       dto.IsRetry,
		"retry_interval": dto.RetryInterval,
		"max_retries":    dto.MaxRetries,
		"create_type":    dto.CreateType,
		"username":       username,
		"script_id":      dto.ScriptID,
	}
}

// ListScheduleDTO 用于获取计划任务列表的请求结构体
// 支持分页查询和多种筛选条件
//
// swagger:model ListScheduleDTO
type ListScheduleDTO struct {
	common.StandardModelQuery

	// 名称
	Name string `form:"name"`

	// 是否启用
	IsEnabled *bool `form:"is_enabled"`

	// 创建类型
	CreateType *int8 `form:"create_type" binding:"omitempty"`

	// 用户名
	Username string `form:"username" binding:"omitempty"`

	// 脚本ID
	ScriptID uint32 `form:"script_id"`
}

func (dto *ListScheduleDTO) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if dto == nil {
		return nil
	}
	if err := dto.StandardModelQuery.MarshalLogObject(enc); err != nil {
		return err
	}
	enc.AddString("name", dto.Name)
	if dto.IsEnabled != nil {
		enc.AddBool("is_enabled", *dto.IsEnabled)
	}
	if dto.CreateType != nil {
		enc.AddInt8("create_type", *dto.CreateType)
	}
	enc.AddString("username", dto.Username)
	enc.AddUint32("script_id", dto.ScriptID)
	return nil
}

func (dto *ListScheduleDTO) ToQueryMap() map[string]any {
	queryMap := dto.BaseModelQuery.ToQueryMap(14)
	if dto.Name != "" {
		queryMap["name like ?"] = "%" + dto.Name + "%"
	}
	if dto.IsEnabled != nil {
		queryMap["is_enabled = ?"] = *dto.IsEnabled
	}
	if dto.CreateType != nil {
		queryMap["create_type = ?"] = *dto.CreateType
	}
	if dto.Username != "" {
		queryMap["username like ?"] = "%" + dto.Username + "%"
	}
	if dto.ScriptID > 0 {
		queryMap["script_id = ?"] = dto.ScriptID
	}
	return queryMap
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

	// 创建类型
	CreateType int8 `json:"create_type" example:"1"`

	// 用户名
	Username string `json:"username" example:"admin"`
}

type ScheduleDetailOut struct {
	ScheduleStandardOut

	// 脚本
	Script *ScriptStandardOut `json:"script"`
}

// ScheduleResp 程序包响应结构
type ScheduleResp = common.APIResp[ScheduleDetailOut]

// PagScheduleResp 程序包的分页响应结构
type PagScheduleResp = common.APIResp[*common.Pag[ScheduleDetailOut]]

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
		CreateType:    m.CreateType,
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
	ms []ScheduleModel,
) []ScheduleDetailOut {
	if len(ms) == 0 {
		return []ScheduleDetailOut{}
	}
	mso := make([]ScheduleDetailOut, 0, len(ms))
	if len(ms) > 0 {
		for _, m := range ms {
			mo := ScheduleToDetailOut(m)
			mso = append(mso, *mo)
		}
	}
	return mso
}
