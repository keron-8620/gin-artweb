package script

import (
	"gin-artweb/api/common"
)

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

// ScriptReply 程序包响应结构
type ScriptReply = common.APIReply[ScriptStandardOut]

// PagScriptReply程序包的分页响应结构
type PagScriptReply = common.APIReply[*common.Pag[ScriptStandardOut]]
