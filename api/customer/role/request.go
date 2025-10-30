package role

import (
	"gin-artweb/pkg/database"
)

// CreateRoleRequest 用于创建角色的请求结构体
//
// swagger:model CreateRoleRequest
type CreateRoleRequest struct {
	// 名称，必填，最大长度50
	// Required: true
	// Max length: 50
	Name string `json:"name" binding:"required,max=50"`

	// 描述，可选，最大长度254
	// Max length: 254
	Descr string `json:"descr" binding:"omitempty,max=254"`

	// 关联权限ID列表，可选
	PermissionIds []uint32 `json:"permission_ids" binding:"omitempty"`

	// 关联菜单ID列表，可选
	MenuIds []uint32 `json:"menu_ids" binding:"omitempty"`

	// 关联按钮ID列表，可选
	ButtonIds []uint32 `json:"button_ids" binding:"omitempty"`
}

// UpdateRoleRequest 用于更新角色的请求结构体
//
// swagger:model UpdateRoleRequest
type UpdateRoleRequest struct {
	// 名称，必填，最大长度50
	// Required: true
	// Max length: 50
	Name string `json:"name" binding:"required,max=50"`

	// 角色描述信息，可选，最大长度254
	// Max length: 254
	Descr string `json:"descr" binding:"omitempty,max=254"`

	// 关联权限ID列表，可选
	PermissionIds []uint32 `json:"permission_ids" binding:"omitempty"`

	// 关联菜单ID列表，可选
	MenuIds []uint32 `json:"menu_ids" binding:"omitempty"`

	// 关联按钮ID列表，可选
	ButtonIds []uint32 `json:"button_ids" binding:"omitempty"`
}

// ListRoleRequest 用于获取角色列表的请求结构体
// 支持分页查询和多种筛选条件
//
// swagger:model ListRoleRequest
type ListRoleRequest struct {
	database.StandardModelQuery

	// 角色名称，字符串长度限制
	// Max length: 50
	Name string `form:"name" binding:"omitempty,max=50"`

	// 角色描述，字符串长度限制
	// Max length: 254
	Descr string `form:"descr" binding:"omitempty,max=254"`
}

func (req *ListRoleRequest) Query() (int, int, map[string]any) {
	page, size, query := req.StandardModelQuery.QueryMap(8)
	if req.Name != "" {
		query["name like ?"] = "%" + req.Name + "%"
	}
	if req.Descr != "" {
		query["descr like ?"] = "%" + req.Descr + "%"
	}
	return page, size, query
}
