package role

import (
	"gin-artweb/api/common"
	"gin-artweb/api/customer/button"
	"gin-artweb/api/customer/menu"
	"gin-artweb/api/customer/permission"
)

// RoleOutBase 角色基础信息
type RoleOutBase struct {
	// 角色ID
	ID uint32 `json:"id" example:"1"`
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

// RoleMenuPerm 角色菜单权限
type RoleMenuPerm struct {
	menu.MenuOutBase
	// 子菜单
	Children []RoleMenuPerm `json:"children"`
	// 按钮
	Buttons []button.ButtonOutBase `json:"buttons"`
}

// RolePermTreeReply 角色响应结构
type RoleMenuTreeReply = common.APIReply[[]*RoleMenuPerm]
