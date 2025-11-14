package button

import (
	"gin-artweb/api/common"
)

// CreateButtonRequest 用于创建按钮的请求结构体
//
// swagger:model CreateButtonRequest
type CreateButtonRequest struct {
	// 按钮主键，必须大于0
	// Required: true
	// Minimum: 1
	ID uint32 `json:"id" binding:"required,gt=0"`

	// 名称，最大长度50
	// Required: true
	// Max length: 50
	Name string `json:"name" binding:"required,max=50"`

	// 排序字段，必填
	ArrangeOrder uint32 `json:"arrange_order" binding:"required"`

	// 是否激活，必填
	IsActive bool `json:"is_active" binding:"required"`

	// 按钮描述信息，可选，最大长度254
	// Max length: 254
	Descr string `json:"descr" binding:"omitempty,max=254"`

	// 菜单ID，必填
	MenuID uint32 `json:"menu_id" binding:"required"`

	// 关联权限ID列表，可选
	PermissionIDs []uint32 `json:"permission_ids" binding:"omitempty"`
}

// UpdateButtonRequest 用于更新按钮的请求结构体
//
// swagger:model UpdateButtonRequest
type UpdateButtonRequest struct {
	// 名称，最大长度50
	// Required: true
	// Max length: 50
	Name string `json:"name" binding:"required,max=50"`

	// 排序字段，必填
	ArrangeOrder uint32 `json:"arrange_order" binding:"required"`

	// 是否激活，必填
	IsActive bool `json:"is_active" binding:"required"`

	// 按钮描述信息，可选，最大长度254
	// Max length: 254
	Descr string `json:"descr" binding:"omitempty,max=254"`

	// 菜单ID，必填
	MenuID uint32 `json:"menu_id" binding:"required"`

	// 关联权限ID列表，可选
	PermissionIDs []uint32 `json:"permission_ids" binding:"omitempty"`
}

// ListButtonRequest 用于获取按钮列表的请求结构体
// 支持分页查询和多种筛选条件
//
// swagger:model ListButtonRequest
type ListButtonRequest struct {
	common.StandardModelQuery

	// 按钮名称，字符串长度限制
	// Max length: 50
	Name string `form:"name" binding:"omitempty,max=50"`

	// 是否激活筛选
	IsActive *bool `form:"is_active" binding:"omitempty"`

	// 按钮描述，字符串长度限制
	// Max length: 254
	Descr string `form:"descr" binding:"omitempty,max=254"`

	// 菜单ID，必填
	MenuID uint32 `json:"menu_id" binding:"omitempty"`
}

func (req *ListButtonRequest) Query() (int, int, map[string]any) {
	page, size, query := req.StandardModelQuery.QueryMap(10)
	if req.Name != "" {
		query["name like ?"] = "%" + req.Name + "%"
	}
	if req.IsActive != nil {
		query["is_active = ?"] = *req.IsActive
	}
	if req.Descr != "" {
		query["descr like ?"] = "%" + req.Descr + "%"
	}
	if req.MenuID != 0 {
		query["menu_id = ?"] = req.MenuID
	}
	return page, size, query
}
