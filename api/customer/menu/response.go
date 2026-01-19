package menu

import (
	"gin-artweb/api/common"
)

// MenuStandardOut 菜单基础输出结构体
type MenuBaseOut struct {
	// 菜单ID
	ID uint32 `json:"id" example:"1"`
	// 前端路由
	Path string `json:"path" example:"/api/v1/users"`
	// 组件路径
	Component string `json:"component" example:"GET"`
	// 名称
	Name string `json:"name" example:"用户管理"`
	//菜单信息
	Meta MetaSchemas `json:"meta"`
	// 排列顺序
	ArrangeOrder uint32 `json:"arrange_order" example:"1000"`
	// 是否激活
	IsActive bool `json:"is_active" example:"true"`
	// 描述
	Descr string `json:"descr" example:"用户管理"`
}

// MenuStandardOut 菜单标准输出结构体
type MenuStandardOut struct {
	MenuBaseOut
	// 创建时间
	CreatedAt string `json:"created_at" example:"2023-01-01 12:00:00"`
	// 更新时间
	UpdatedAt string `json:"updated_at" example:"2023-01-01 12:00:00"`
}

// MenuDetailOut 菜单详情输出结构体
type MenuDetailOut struct {
	MenuStandardOut
	Parent        *MenuStandardOut `json:"parent"`
	PermissionIDs []uint32         `json:"permission_ids"`
}

// MenuReply 菜单响应结构
type MenuReply = common.APIReply[*MenuDetailOut]

// PagMenuReply 菜单的分页响应结构
type PagMenuReply = common.APIReply[*common.Pag[MenuStandardOut]]
