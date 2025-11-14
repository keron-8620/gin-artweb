package menu

import (
	"gin-artweb/api/common"
)

// CreateMenuRequest 用于创建菜单的请求结构体
//
// swagger:model CreateMenuRequest
type CreateMenuRequest struct {
	// 菜单主键，必须大于0
	// Required: true
	// Minimum: 1
	ID uint32 `json:"id" binding:"required,gt=0"`

	// 前端路由路径，最大长度100
	// Required: true
	// Max length: 100
	Path string `json:"path" binding:"required,max=100"`

	// 组件路径，最大长度200
	// Required: true
	// Max length: 200
	Component string `json:"component" binding:"required,max=200"`

	// 名称，最大长度50
	// Required: true
	// Max length: 50
	Name string `json:"name" binding:"required,max=50"`

	// 菜单元信息，必填
	Meta MetaSchemas `json:"meta" binding:"required"`

	// 排序字段，必填
	ArrangeOrder uint32 `json:"arrange_order" binding:"required"`

	// 是否激活，必填
	IsActive bool `json:"is_active" binding:"required"`

	// 描述，最大长度254
	// Max length: 254
	Descr string `json:"descr" binding:"omitempty,max=254"`

	// 父级菜单ID，可选
	ParentID *uint32 `json:"parent_id" binding:"omitempty"`

	// 关联权限ID列表，可选
	PermissionIDs []uint32 `json:"permission_ids" binding:"omitempty"`
}

// UpdateMenuRequest 用于更新菜单的请求结构体
//
// swagger:model UpdateMenuRequest
type UpdateMenuRequest struct {
	// 前端路由路径，最大长度100
	// Required: true
	// Max length: 100
	Path string `json:"path" binding:"required,max=100"`

	// 组件路径，最大长度200
	// Required: true
	// Max length: 200
	Component string `json:"component" binding:"required,max=200"`

	// 名称，最大长度50
	// Required: true
	// Max length: 50
	Name string `json:"name" binding:"required,max=50"`

	// 菜单元信息，必填
	Meta MetaSchemas `json:"meta" binding:"required"`

	// 排序字段，必填
	ArrangeOrder uint32 `json:"arrange_order" binding:"required"`

	// 是否激活，必填
	IsActive bool `json:"is_active" binding:"required"`

	// 菜单描述信息，可选，最大长度254
	// Max length: 254
	Descr string `json:"descr" binding:"max=254"`

	// 父级菜单ID，可选
	ParentID *uint32 `json:"parent_id" binding:"omitempty"`

	// 关联权限ID列表，可选
	PermissionIDs []uint32 `json:"permission_ids" binding:"omitempty"`
}

// ListMenuRequest 用于获取菜单列表的请求结构体
// 支持分页查询和多种筛选条件
//
// swagger:model ListMenuRequest
type ListMenuRequest struct {
	common.StandardModelQuery

	// 前端路由路径，字符串长度限制
	// Max length: 100
	Path string `form:"path" binding:"omitempty,max=100"`

	// 组件路径，字符串长度限制
	// Max length: 200
	Component string `form:"component" binding:"omitempty,max=200"`

	// 菜单名称，字符串长度限制
	// Max length: 50
	Name string `form:"name" binding:"omitempty,max=50"`

	// 是否激活筛选
	IsActive *bool `form:"is_active" binding:"omitempty"`

	// 菜单描述，字符串长度限制
	// Max length: 254
	Descr string `form:"descr" binding:"omitempty,max=254"`

	// 父级菜单ID筛选
	ParentID *uint32 `form:"parent_id" binding:"omitempty"`
}

func (req *ListMenuRequest) Query() (int, int, map[string]any) {
	page, size, query := req.StandardModelQuery.QueryMap(12)
	if req.Path != "" {
		query["path like ?"] = "%" + req.Path + "%"
	}
	if req.Component != "" {
		query["component like ?"] = "%" + req.Component + "%"
	}
	if req.Name != "" {
		query["name like ?"] = "%" + req.Name + "%"
	}
	if req.IsActive != nil {
		query["is_active = ?"] = *req.IsActive
	}
	if req.Descr != "" {
		query["descr like ?"] = "%" + req.Descr + "%"
	}
	if req.ParentID != nil {
		query["parent_id = ?"] = *req.ParentID
	}
	return page, size, query
}
