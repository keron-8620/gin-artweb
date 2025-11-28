package script

import (
	"errors"
	"mime/multipart"

	"gin-artweb/api/common"

	"go.uber.org/zap/zapcore"
)

type UploadScriptRequest struct {
	// 上传的程序包文件
	File *multipart.FileHeader `form:"file" binding:"required"`

	// 脚本描述，字符串长度限制
	// Max length: 254
	Descr string `form:"descr" binding:"omitempty,max=254"`

	// 项目，最大长度50
	// Required: true
	// Max length: 50
	Project string `form:"project" binding:"required"`

	// 标签，最大长度50
	// Required: true
	// Max length: 50
	Label string `form:"label" binding:"required"`

	// 语言，最大长度50
	// Required: true
	// Max length: 50
	Language string `form:"language" binding:"required"`

	// 状态，必填
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

	// 名称，最大长度50
	// Max length: 50
	Name string `json:"name" binding:"omitempty,max=50"`

	// 脚本描述，字符串长度限制
	// Max length: 254
	Descr string `form:"descr" binding:"omitempty,max=254"`

	// 项目，最大长度50
	// Max length: 50
	Project string `form:"project" binding:"omitempty"`

	// 标签，最大长度50
	// Max length: 50
	Label string `form:"label" binding:"omitempty"`

	// 语言，最大长度50
	// Max length: 50
	Language string `form:"language" binding:"omitempty"`

	// 状态
	Status *bool `form:"status"`

	// 是否是内置脚本
	IsBuiltin *bool `json:"is_builtin" binding:"omitempty"`

	// 最后修改的用户
	UserID uint32 `json:"user_id" binding:"omitempty"`
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
