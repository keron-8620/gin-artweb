package role

import (
	"gitee.com/keion8620/go-dango-gin/api/customer/button"
	"gitee.com/keion8620/go-dango-gin/api/customer/menu"
	"gitee.com/keion8620/go-dango-gin/api/customer/permission"
	"gitee.com/keion8620/go-dango-gin/pkg/common"
)

// RoleOutBase 角色基础信息
type RoleOutBase struct {
	// 角色ID
	Id uint `json:"id" example:"1"`
	// 创建时间
	CreatedAt string `json:"created_at" example:"2023-01-01 12:00:00"`
	// 更新时间
	UpdatedAt string `json:"updated_at" example:"2023-01-01 12:00:00"`
	// 名称
	Name string `json:"name" example:"用户管理"`
	// 描述
	Descr string `json:"descr" example:"用户管理"`
}

type RoleOut struct {
	RoleOutBase
	Permissions []*permission.PermissionOutBase `json:"permissions"`
	Menus       []*menu.MenuOutBase             `json:"menus"`
	Buttons     []*button.ButtonOutBase         `json:"buttons"`
}

// RoleBaseReply 角色响应结构
type RoleReply = common.APIReply[RoleOut]

// PagRoleBaseReply 角色的分页响应结构
type PagRoleBaseReply = common.APIReply[*common.Pag[*RoleOutBase]]
