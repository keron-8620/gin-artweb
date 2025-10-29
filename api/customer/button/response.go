package button

import (
	"gitee.com/keion8620/go-dango-gin/api/customer/menu"
	"gitee.com/keion8620/go-dango-gin/api/customer/permission"
	"gitee.com/keion8620/go-dango-gin/pkg/common"
)

// ButtonOutBase 按钮基础信息
type ButtonOutBase struct {
	// 按钮ID
	Id uint `json:"id" example:"1"`
	// 创建时间
	CreatedAt string `json:"created_at" example:"2023-01-01 12:00:00"`
	// 更新时间
	UpdatedAt string `json:"updated_at" example:"2023-01-01 12:00:00"`
	// 名称
	Name string `json:"name" example:"用户管理"`
	// 排列顺序
	ArrangeOrder uint `json:"arrange_order" example:"1000"`
	// 是否激活
	IsActive bool `json:"is_active" example:"true"`
	// 描述
	Descr string `json:"descr" example:"用户管理"`
}

type ButtonOut struct {
	ButtonOutBase
	Menu        *menu.MenuOutBase               `json:"menu"`
	Permissions []*permission.PermissionOutBase `json:"permissions"`
}

// ButtonBaseReply 按钮响应结构
type ButtonReply = common.APIReply[ButtonOut]

// PagButtonBaseReply 按钮的分页响应结构
type PagButtonBaseReply = common.APIReply[*common.Pag[*ButtonOutBase]]
