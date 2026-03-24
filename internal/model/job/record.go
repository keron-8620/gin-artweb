package job

import (
	"os"
	"time"

	"go.uber.org/zap/zapcore"

	"gin-artweb/internal/model/common"
	"gin-artweb/internal/shared/database"
)

type ScriptRecordModel struct {
	database.StandardModel
	TriggerType  string      `gorm:"column:trigger_type;type:varchar(20);comment:触发类型(cron/api)" json:"trigger_type"`
	Status       int         `gorm:"column:status;type:tinyint;not null;default:0;comment:执行状态(0-待执行,1-执行中,2-成功,3-失败,4-超时,5-崩溃)" json:"status"`
	ExitCode     int         `gorm:"column:exit_code;comment:退出码" json:"exit_code"`
	EnvVars      string      `gorm:"column:env_vars;type:json;comment:环境变量(JSON对象)" json:"env_vars"`
	CommandArgs  string      `gorm:"column:command_args;type:varchar(254);comment:命令行参数(JSON数组)" json:"command_args"`
	WorkDir      string      `gorm:"column:work_dir;type:varchar(255);comment:工作目录" json:"work_dir"`
	Timeout      int         `gorm:"column:timeout;type:int;not null;default:300;comment:超时时间(秒)" json:"timeout"`
	LogName      string      `gorm:"column:log_name;type:varchar(255);comment:日志文件路径" json:"log_name"`
	ErrorMessage string      `gorm:"column:error_message;type:text;comment:错误信息" json:"error_message"`
	Username     string      `gorm:"column:username;type:varchar(50);comment:用户名" json:"username"`
	ScriptID     uint32      `gorm:"column:script_id;not null;index;comment:脚本ID" json:"script_id"`
	Script       ScriptModel `gorm:"foreignKey:ScriptID;references:ID;constraint:OnDelete:CASCADE" json:"script"`
}

func (m *ScriptRecordModel) TableName() string {
	return "job_script_record"
}

func (m *ScriptRecordModel) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if m == nil {
		return nil
	}
	if err := m.StandardModel.MarshalLogObject(enc); err != nil {
		return err
	}
	enc.AddString("trigger_type", m.TriggerType)
	enc.AddInt("status", m.Status)
	enc.AddInt("exit_code", m.ExitCode)
	enc.AddString("env_vars", m.EnvVars)
	enc.AddString("command_args", m.CommandArgs)
	enc.AddString("work_dir", m.WorkDir)
	enc.AddString("log_path", m.LogName)
	enc.AddString("username", m.Username)
	enc.AddUint32("script_id", m.ScriptID)
	return nil
}

type ExecuteBIZ struct {
	TriggerType string `json:"trigger_type"`
	ScriptID    uint32 `json:"script_id"`
	CommandArgs string `json:"command_args"`
	EnvVars     string `json:"env_vars"`
	Timeout     int    `json:"timeout"`
	WorkDir     string `json:"work_dir"`
	Username    string `json:"username"`
}

func (biz *ExecuteBIZ) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("trigger_type", biz.TriggerType)
	enc.AddUint32("script_id", biz.ScriptID)
	enc.AddString("command_args", biz.CommandArgs)
	enc.AddString("env_vars", biz.EnvVars)
	enc.AddInt("timeout", biz.Timeout)
	enc.AddString("work_dir", biz.WorkDir)
	enc.AddString("username", biz.Username)
	return nil
}

type TaskInfo struct {
	Status   int
	ExitCode int
	ErrMSG   string
	Error    error
	LogFile  *os.File
}

func (t *TaskInfo) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddInt("status", t.Status)
	enc.AddInt("exit_code", t.ExitCode)
	enc.AddString("error_message", t.ErrMSG)
	return nil
}

func (t *TaskInfo) ToUpdateMap() map[string]any {
	return map[string]any{
		"status":        t.Status,
		"exit_code":     t.ExitCode,
		"error_message": t.ErrMSG,
	}
}

// CreateScriptRecordDTO 用于创建计划任务的请求结构体
//
// swagger:model CreateScriptRecordDTO
type CreateScriptRecordDTO struct {
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

// ListScriptRecordDTO 用于获取计划任务列表的请求结构体
// 支持分页查询和多种筛选条件
//
// swagger:model ListScriptRecordDTO
type ListScriptRecordDTO struct {
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

func (t *ListScriptRecordDTO) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if err := t.StandardModelQuery.MarshalLogObject(enc); err != nil {
		return err
	}
	enc.AddString("trigger_type", t.TriggerType)
	enc.AddInt("status", t.Status)
	enc.AddInt("exit_code", t.ExitCode)
	enc.AddUint32("script_id", t.ScriptID)
	enc.AddString("username", t.Username)
	return nil
}

func (dto *ListScriptRecordDTO) ToQueryMap() map[string]any {
	queryMap := dto.BaseModelQuery.ToQueryMap(11)
	if dto.TriggerType != "" {
		queryMap["trigger_type = ?"] = dto.TriggerType
	}
	if dto.Status != 0 {
		queryMap["status = ?"] = dto.Status
	}
	if dto.ExitCode != 0 {
		queryMap["exit_code = ?"] = dto.ExitCode
	}
	if dto.ScriptID > 0 {
		queryMap["script_id = ?"] = dto.ScriptID
	}
	if dto.Username != "" {
		queryMap["username like ?"] = "%" + dto.Username + "%"
	}
	return queryMap
}

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
	Script ScriptStandardOut `json:"script"`
}

// ScriptRecordResp 程序包响应结构
type ScriptRecordResp = common.APIResp[ScriptRecordDetailOut]

// PagScriptRecordResp 程序包的分页响应结构
type PagScriptRecordResp = common.APIResp[*common.Pag[ScriptRecordDetailOut]]

// 实时日志流响应结构
type RealTimeLogResponse struct {
	Line string `json:"line"`
}

func ScriptRecordToStandardOut(
	m ScriptRecordModel,
) *ScriptRecordStandardOut {
	return &ScriptRecordStandardOut{
		ID:           m.ID,
		CreatedAt:    m.CreatedAt.Format(time.DateTime),
		UpdatedAt:    m.UpdatedAt.Format(time.DateTime),
		TriggerType:  m.TriggerType,
		Status:       m.Status,
		ExitCode:     m.ExitCode,
		EnvVars:      m.EnvVars,
		CommandArgs:  m.CommandArgs,
		Timeout:      m.Timeout,
		WorkDir:      m.WorkDir,
		ErrorMessage: m.ErrorMessage,
		Username:     m.Username,
	}
}

func ScriptRecordToDetailOut(
	m ScriptRecordModel,
) *ScriptRecordDetailOut {
	var script *ScriptStandardOut
	if m.Script.ID != 0 {
		script = ScriptModelToStandardOut(m.Script)
	}

	standardOut := ScriptRecordToStandardOut(m)
	result := &ScriptRecordDetailOut{
		ScriptRecordStandardOut: *standardOut,
	}

	if script != nil {
		result.Script = *script
	}

	return result
}

func ListScriptRecordToDetailOut(
	ms []ScriptRecordModel,
) []ScriptRecordDetailOut {
	if len(ms) == 0 {
		return []ScriptRecordDetailOut{}
	}
	mso := make([]ScriptRecordDetailOut, 0, len(ms))
	if len(ms) > 0 {
		for _, m := range ms {
			mo := ScriptRecordToDetailOut(m)
			mso = append(mso, *mo)
		}
	}
	return mso
}

type BizTaskInfo struct {
	// 任务名称
	TaskName string `json:"task_name" example:"mon"`

	// 执行记录ID(0表示非正常执行的任务)
	RecordID uint32 `json:"record_id"`

	// 执行状态(0-待执行,1-执行中,2-成功,3-失败,4-超时,5-崩溃)
	Status int `json:"status" example:"2"`

	// 创建时间
	StartTime string `json:"start_time" example:"2023-01-01 12:00:00"`

	// 更新时间
	EndTime string `json:"end_time" example:"2023-01-01 12:00:00"`

	// 触发类型(cron/api,未执行为空)
	TriggerType string `json:"trigger_type" example:"cron"`
}
