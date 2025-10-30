package menu

import (
	"gin-artweb/api/customer/permission"
	"gin-artweb/pkg/common"
)

// MenuOutBase 菜单基础信息
type MenuOutBase struct {
	// 菜单ID
	Id uint32 `json:"id" example:"1"`
	// 创建时间
	CreatedAt string `json:"created_at" example:"2023-01-01 12:00:00"`
	// 更新时间
	UpdatedAt string `json:"updated_at" example:"2023-01-01 12:00:00"`
	// 前端路由
	Path string `json:"path" example:"/api/v1/users"`
	// 组件路径
	Component string `json:"xomponent" example:"GET"`
	// 名称
	Name string `json:"name" example:"用户管理"`
	//菜单信息
	Meta MetaSchemas `json:"meta"`
	// 标签
	Label string `json:"label" example:"customer"`
	// 排列顺序
	ArrangeOrder uint32 `json:"arrange_order" example:"1000"`
	// 是否激活
	IsActive bool `json:"is_active" example:"true"`
	// 描述
	Descr string `json:"descr" example:"用户管理"`
}

type MenuOut struct {
	MenuOutBase
	Parent      *MenuOutBase                    `json:"parent,omitempty"`
	Permissions []*permission.PermissionOutBase `json:"permissions"`
}

// MenuBaseReply 菜单响应结构
type MenuReply = common.APIReply[MenuOut]

// PagMenuBaseReply 菜单的分页响应结构
type PagMenuBaseReply = common.APIReply[*common.Pag[*MenuOutBase]]
