package role

import (
	"gin-artweb/api/common"
	"gin-artweb/api/customer/button"
	"gin-artweb/api/customer/menu"
)

type RoleBaseOut struct {
	// 唯一标识
	ID uint32 `json:"id" example:"1"`

	// 名称
	Name string `json:"name" example:"用户管理"`

	// 描述
	Descr string `json:"descr" example:"用户管理"`
}

// RoleStandardOut 角色基础信息
type RoleStandardOut struct {
	RoleBaseOut

	// 创建时间
	CreatedAt string `json:"created_at" example:"2023-01-01 12:00:00"`

	// 更新时间
	UpdatedAt string `json:"updated_at" example:"2023-01-01 12:00:00"`
}

type RoleDetailOut struct {
	RoleStandardOut

	// APIID列表
	ApiIDs []uint32 `json:"api_ids"`

	// 菜单ID列表
	MenuIDs []uint32 `json:"menu_ids"`

	// 按钮ID列表
	ButtonIDs []uint32 `json:"button_ids"`
}

// RoleBaseReply 角色响应结构
type RoleReply = common.APIReply[*RoleDetailOut]

// PagRoleReply 角色的分页响应结构
type PagRoleReply = common.APIReply[*common.Pag[RoleStandardOut]]

// RoleMenuPerm 角色菜单权限
type RoleMenuPerm struct {
	menu.MenuBaseOut

	// 子菜单
	Children []RoleMenuPerm `json:"children"`

	// 按钮
	Buttons []button.ButtonBaseOut `json:"buttons"`
}

// RoleMenuTreeReply 角色响应结构
type RoleMenuTreeReply = common.APIReply[*[]RoleMenuPerm]
