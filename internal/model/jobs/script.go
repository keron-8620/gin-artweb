package jobs

import (
	"errors"
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
	Project   string `gorm:"column:project;type:varchar(50);index:idx_script_project_label_name;comment:项目" json:"project"`
	Label     string `gorm:"column:label;type:varchar(50);index:idx_script_project_label_name;;comment:标签" json:"label"`
	Language  string `gorm:"column:language;type:varchar(50);comment:脚本语言" json:"language"`
	Status    bool   `gorm:"column:status;type:boolean;comment:是否启用" json:"status"`
	IsBuiltin bool   `gorm:"column:is_builtin;type:boolean;comment:是否是内置脚本" json:"is_builtin"`
	Username  string `gorm:"column:username;type:varchar(50);comment:用户名" json:"username"`
}

func (m *ScriptModel) TableName() string {
	return "jobs_script"
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

type UploadScriptRequest struct {
	// 上传的程序包文件
	File *multipart.FileHeader `form:"file" binding:"required"`

	// 描述信息
	Descr string `form:"descr" binding:"omitempty,max=254"`

	// 项目
	Project string `form:"project" binding:"required"`

	// 标签
	Label string `form:"label" binding:"required"`

	// 脚本语言
	Language string `form:"language" binding:"required"`

	// 状态
	Status bool `form:"status"`
}

func (req *UploadScriptRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if req.File == nil {
		return errors.New("文件不能为空")
	}
	enc.AddString("name", req.File.Filename)
	enc.AddString("descr", req.Descr)
	enc.AddString("project", req.Project)
	enc.AddString("label", req.Label)
	enc.AddString("language", req.Language)
	enc.AddBool("status", req.Status)
	return nil
}

type ListScriptRequest struct {
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
	IsBuiltin *bool `form:"is_builtin" binding:"omitempty"`

	// 最后修改的用户
	UserID uint32 `form:"user_id" binding:"omitempty"`
}

func (req *ListScriptRequest) Query() (int, int, map[string]any) {
	page, size, query := req.BaseModelQuery.QueryMap(14)
	if req.Name != "" {
		query["name like ?"] = "%" + req.Name + "%"
	}
	if req.Descr != "" {
		query["descr like ?"] = "%" + req.Descr + "%"
	}
	if req.Project != "" {
		query["project = ?"] = req.Project
	}
	if req.Label != "" {
		query["label = ?"] = req.Label
	}
	if req.Language != "" {
		query["language = ?"] = req.Language
	}
	if req.Status != nil {
		query["status = ?"] = *req.Status
	}
	if req.IsBuiltin != nil {
		query["is_builtin = ?"] = *req.IsBuiltin
	}
	if req.UserID != 0 {
		query["user_id = ?"] = req.UserID
	}
	return page, size, query
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

// ScriptReply 脚本响应结构
type ScriptReply = common.APIReply[ScriptStandardOut]

// PagScriptReply 脚本分页响应结构
type PagScriptReply = common.APIReply[*common.Pag[ScriptStandardOut]]

// ListProjectReply 项目列表响应结构
type ListProjectReply = common.APIReply[[]string]

// ListLableReply 标签列表响应结构
type ListLableReply = common.APIReply[[]string]

func ScriptModelToStandardOut(
	m ScriptModel,
) *ScriptStandardOut {
	return &ScriptStandardOut{
		ID:        m.ID,
		CreatedAt: m.CreatedAt.Format(time.DateTime),
		UpdatedAt: m.UpdatedAt.Format(time.DateTime),
		Name:      m.Name,
		Descr:     m.Descr,
		Project:   m.Project,
		Label:     m.Label,
		Language:  m.Language,
		Status:    m.Status,
		IsBuiltin: m.IsBuiltin,
		Username:  m.Username,
	}
}

func ListScriptModelToOutBase(
	pms *[]ScriptModel,
) *[]ScriptStandardOut {
	if pms == nil {
		return &[]ScriptStandardOut{}
	}

	ms := *pms
	mso := make([]ScriptStandardOut, 0, len(ms))
	if len(ms) > 0 {
		for _, m := range ms {
			mo := ScriptModelToStandardOut(m)
			mso = append(mso, *mo)
		}
	}
	return &mso
}
