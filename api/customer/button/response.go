package button

import (
	"gin-artweb/api/common"
	"gin-artweb/api/customer/menu"
)

type ButtonBaseOut struct {
	// 按钮ID
	ID uint32 `json:"id" example:"1"`
	// 名称
	Name string `json:"name" example:"用户管理"`
	// 排列顺序
	ArrangeOrder uint32 `json:"arrange_order" example:"1000"`
	// 是否激活
	IsActive bool `json:"is_active" example:"true"`
	// 描述
	Descr string `json:"descr" example:"用户管理"`
}

// ButtonStandardOut 按钮标准信息
type ButtonStandardOut struct {
	ButtonBaseOut
	// 创建时间
	CreatedAt string `json:"created_at" example:"2023-01-01 12:00:00"`
	// 更新时间
	UpdatedAt string `json:"updated_at" example:"2023-01-01 12:00:00"`
}

// ButtonDetailOut 按钮详情信息
type ButtonDetailOut struct {
	ButtonStandardOut
	Menu        *menu.MenuStandardOut `json:"menu"`
	Permissions []uint32              `json:"permissions"`
}

// ButtonBaseReply 按钮响应结构
type ButtonReply = common.APIReply[*ButtonDetailOut]

// PagButtonReply 按钮的分页响应结构
type PagButtonReply = common.APIReply[*common.Pag[ButtonStandardOut]]
