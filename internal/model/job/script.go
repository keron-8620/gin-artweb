package job

import (
	"errors"
	"io"
	"mime/multipart"
	"time"

	"go.uber.org/zap/zapcore"

	"gin-artweb/internal/model/common"
	"gin-artweb/internal/shared/database"
)

type ScriptModel struct {
	database.StandardModel
	Name      string `gorm:"column:name;type:varchar(50);not null;index:idx_script_project_label_name;comment:名称" json:"name"`
	Descr     string `gorm:"column:descr;type:varchar(254);comment:描述" json:"descr"`
	ParamDesc string `gorm:"column:param_desc;type:varchar(254);comment:参数描述" json:"param_desc"`
	Project   string `gorm:"column:project;type:varchar(50);index:idx_script_project_label_name;comment:项目" json:"project"`
	Label     string `gorm:"column:label;type:varchar(50);index:idx_script_project_label_name;;comment:标签" json:"label"`
	Language  string `gorm:"column:language;type:varchar(50);comment:脚本语言" json:"language"`
	Status    bool   `gorm:"column:status;type:boolean;comment:是否启用" json:"status"`
	IsBuiltin bool   `gorm:"column:is_builtin;type:boolean;comment:是否是内置脚本" json:"is_builtin"`
	Username  string `gorm:"column:username;type:varchar(50);comment:用户名" json:"username"`
}

func (m *ScriptModel) TableName() string {
	return "job_script"
}

func (m *ScriptModel) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if m == nil {
		return nil
	}
	if err := m.StandardModel.MarshalLogObject(enc); err != nil {
		return err
	}
	enc.AddString("name", m.Name)
	enc.AddString("descr", m.Descr)
	enc.AddString("project", m.Project)
	enc.AddString("label", m.Label)
	enc.AddString("language", m.Language)
	enc.AddBool("status", m.Status)
	enc.AddBool("is_builtin", m.IsBuiltin)
	enc.AddString("username", m.Username)
	return nil
}

func ListScriptModelToUint32s(ms []ScriptModel) []uint32 {
	if len(ms) == 0 {
		return []uint32{}
	}

	ids := make([]uint32, len(ms))
	for i, m := range ms {
		ids[i] = m.ID
	}
	return ids
}

type UploadScriptDTO struct {
	// 上传的程序包文件
	File *multipart.FileHeader `form:"file" binding:"required"`

	// 描述信息
	Descr string `form:"descr" binding:"omitempty,max=254"`

	// 参数描述
	ParamDesc string `form:"param_desc" binding:"omitempty,max=254"`

	// 项目
	Project string `form:"project" binding:"required"`

	// 标签
	Label string `form:"label" binding:"required"`

	// 脚本语言
	Language string `form:"language" binding:"required"`

	// 状态
	Status bool `form:"status"`

	// 是否是内置脚本
	IsBuiltin bool `form:"is_builtin"`
}

func (dto *UploadScriptDTO) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if dto.File == nil {
		return errors.New("文件不能为空")
	}
	enc.AddString("name", dto.File.Filename)
	enc.AddString("descr", dto.Descr)
	enc.AddString("project", dto.Project)
	enc.AddString("label", dto.Label)
	enc.AddString("language", dto.Language)
	enc.AddBool("status", dto.Status)
	enc.AddBool("is_builtin", dto.IsBuiltin)
	return nil
}

type ScriptUpsertDTO struct {
	Filename  string    // 文件名
	File      io.Reader // 文件内容
	Descr     string    // 描述信息
	ParamDesc string    // 参数描述
	Project   string    // 项目
	Label     string    // 标签
	Language  string    // 脚本语言
	Status    bool      // 状态
	IsBuiltin bool      // 是否是内置脚本
}

func (dto *ScriptUpsertDTO) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("name", dto.Filename)
	enc.AddString("descr", dto.Descr)
	enc.AddString("project", dto.Project)
	enc.AddString("label", dto.Label)
	enc.AddString("language", dto.Language)
	enc.AddBool("status", dto.Status)
	enc.AddBool("is_builtin", dto.IsBuiltin)
	return nil
}

func (dto *ScriptUpsertDTO) ToModel(username string) ScriptModel {
	return ScriptModel{
		Name:      dto.Filename,
		Descr:     dto.Descr,
		ParamDesc: dto.ParamDesc,
		Project:   dto.Project,
		Label:     dto.Label,
		Language:  dto.Language,
		Status:    dto.Status,
		IsBuiltin: dto.IsBuiltin,
		Username:  username,
	}
}

func (dto *ScriptUpsertDTO) ToUpdateMap(username string) map[string]any {
	return map[string]any{
		"name":       dto.Filename,
		"descr":      dto.Descr,
		"param_desc": dto.ParamDesc,
		"project":    dto.Project,
		"label":      dto.Label,
		"language":   dto.Language,
		"status":     dto.Status,
		"is_builtin": dto.IsBuiltin,
		"username":   username,
	}
}

type ListScriptDTO struct {
	common.StandardModelQuery

	// 名称
	Name string `form:"name" binding:"omitempty,max=50"`

	// 描述信息
	Descr string `form:"descr" binding:"omitempty,max=254"`

	// 项目
	Project string `form:"project" binding:"omitempty"`

	// 标签
	Label string `form:"label" binding:"omitempty"`

	// 脚本语言
	Language string `form:"language" binding:"omitempty"`

	// 状态
	Status *bool `form:"status"`

	// 是否是内置脚本
	IsBuiltin *bool `form:"is_builtin"`

	// 最后修改的用户
	UserID uint32 `form:"user_id" binding:"omitempty"`
}

func (dto *ListScriptDTO) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if dto == nil {
		return nil
	}
	if err := dto.StandardModelQuery.MarshalLogObject(enc); err != nil {
		return err
	}
	enc.AddString("name", dto.Name)
	enc.AddString("descr", dto.Descr)
	enc.AddString("project", dto.Project)
	enc.AddString("label", dto.Label)
	enc.AddString("language", dto.Language)
	if dto.Status != nil {
		enc.AddBool("status", *dto.Status)
	}
	if dto.IsBuiltin != nil {
		enc.AddBool("is_builtin", *dto.IsBuiltin)
	}
	enc.AddUint32("user_id", dto.UserID)
	return nil
}

func (dto *ListScriptDTO) ToQueryMap() map[string]any {
	queryMap := dto.BaseModelQuery.ToQueryMap(14)
	if dto.Name != "" {
		queryMap["name like ?"] = "%" + dto.Name + "%"
	}
	if dto.Descr != "" {
		queryMap["descr like ?"] = "%" + dto.Descr + "%"
	}
	if dto.Project != "" {
		queryMap["project = ?"] = dto.Project
	}
	if dto.Label != "" {
		queryMap["label = ?"] = dto.Label
	}
	if dto.Language != "" {
		queryMap["language = ?"] = dto.Language
	}
	if dto.Status != nil {
		queryMap["status = ?"] = *dto.Status
	}
	if dto.IsBuiltin != nil {
		queryMap["is_builtin = ?"] = *dto.IsBuiltin
	}
	if dto.UserID != 0 {
		queryMap["user_id = ?"] = dto.UserID
	}
	return queryMap
}

// ScriptStandardOut 程序包基础信息
type ScriptStandardOut struct {
	// 唯一标识
	ID uint32 `json:"id" example:"1"`

	// 创建时间
	CreatedAt string `json:"created_at" example:"2023-01-01 12:00:00"`

	// 更新时间
	UpdatedAt string `json:"updated_at" example:"2023-01-01 12:00:00"`

	// 名称
	Name string `json:"name" example:"test.sh"`

	// 描述信息
	Descr string `json:"descr" example:"这是一个测试脚本"`

	// 参数描述
	ParamDesc string `json:"param_desc" example:"--param1=value1 --param2=value2"`

	// 项目
	Project string `json:"project" example:"artweb"`

	// 标签
	Label string `json:"label" example:"cmd"`

	// 脚本语言
	Language string `json:"language" example:"bash"`

	// 状态
	Status bool `json:"status" example:"true"`

	// 是否是内置脚本
	IsBuiltin bool `json:"is_builtin" example:"true"`

	// 用户名
	Username string `json:"username" example:"admin"`
}

// ScriptResp 脚本响应结构
type ScriptResp = common.APIResp[ScriptStandardOut]

// PagScriptResp 脚本分页响应结构
type PagScriptResp = common.APIResp[*common.Pag[ScriptStandardOut]]

// ListProjectResp 项目列表响应结构
type ListProjectResp = common.APIResp[[]string]

// ListLableResp 标签列表响应结构
type ListLableResp = common.APIResp[[]string]

func ScriptModelToStandardOut(
	m ScriptModel,
) *ScriptStandardOut {
	return &ScriptStandardOut{
		ID:        m.ID,
		CreatedAt: m.CreatedAt.Format(time.DateTime),
		UpdatedAt: m.UpdatedAt.Format(time.DateTime),
		Name:      m.Name,
		Descr:     m.Descr,
		ParamDesc: m.ParamDesc,
		Project:   m.Project,
		Label:     m.Label,
		Language:  m.Language,
		Status:    m.Status,
		IsBuiltin: m.IsBuiltin,
		Username:  m.Username,
	}
}

func ListScriptModelToOutBase(
	ms []ScriptModel,
) []ScriptStandardOut {
	if len(ms) == 0 {
		return []ScriptStandardOut{}
	}
	mso := make([]ScriptStandardOut, 0, len(ms))
	if len(ms) > 0 {
		for _, m := range ms {
			mo := ScriptModelToStandardOut(m)
			mso = append(mso, *mo)
		}
	}
	return mso
}
