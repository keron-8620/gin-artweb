package script

import (
	"gin-artweb/api/common"
	"gin-artweb/api/customer/user"
)

// ScriptOutBase 程序包基础信息
type ScriptOutBase struct {
	// 脚本ID
	ID uint32 `json:"id" example:"1"`
	// 创建时间
	CreatedAt string `json:"created_at" example:"2023-01-01 12:00:00"`
	// 更新时间
	UpdatedAt string `json:"updated_at" example:"2023-01-01 12:00:00"`
	// 名称
	Name string `json:"name" example:"test.sh"`
	// 描述
	Descr string `json:"descr" example:"这是一个测试脚本"`
	// 项目
	Project string `json:"project" example:"artweb"`
	// 标签
	Label string `json:"label" example:"cmd"`
	// 语言
	Language string `json:"language" example:"bash"`
	// 状态
	Status bool `json:"status" example:"true"`
	// 是否是内置脚本
	IsBuiltin bool `json:"is_builtin" example:"true"`
}

type ScriptOut struct {
	ScriptOutBase
	// 用户ID
	User *user.UserOutBase `json:"user"`
}

// ScriptReply 程序包响应结构
type ScriptReply = common.APIReply[ScriptOut]

// PagScriptReply程序包的分页响应结构
type PagScriptReply = common.APIReply[*common.Pag[ScriptOutBase]]
