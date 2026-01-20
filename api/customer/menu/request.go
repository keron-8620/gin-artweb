package menu

import (
	"go.uber.org/zap/zapcore"

	"gin-artweb/api/common"
)

// CreateMenuRequest 用于创建菜单的请求结构体
//
// swagger:model CreateMenuRequest
type CreateMenuRequest struct {
	// 唯一标识
	ID uint32 `json:"id" form:"id" binding:"required,gt=0"`

	// 前端路由路径
	Path string `json:"path" form:"path" binding:"required,max=100"`

	// 组件路径
	Component string `json:"component" form:"component" binding:"required,max=200"`

	// 名称
	Name string `json:"name" form:"name" binding:"required,max=50"`

	// 菜单元信息
	Meta MetaSchemas `json:"meta" form:"meta" binding:"required"`

	// 排序字段
	ArrangeOrder uint32 `json:"arrange_order" form:"arrange_order" binding:"required"`

	// 是否激活
	IsActive bool `json:"is_active" form:"is_active"`

	// 描述
	Descr string `json:"descr" form:"descr" binding:"omitempty,max=254"`

	// 父级菜单ID
	ParentID *uint32 `json:"parent_id" form:"parent_id" binding:"omitempty"`

	// 权限ID列表
	PermissionIDs []uint32 `json:"permission_ids" form:"permission_ids" binding:"omitempty"`
}

func (req *CreateMenuRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddUint32("id", req.ID)
	enc.AddString("path", req.Path)
	enc.AddString("component", req.Component)
	enc.AddObject("meta", &req.Meta)
	enc.AddString("name", req.Name)
	enc.AddUint32("arrange_order", req.ArrangeOrder)
	enc.AddBool("is_active", req.IsActive)
	enc.AddString("descr", req.Descr)
	if req.ParentID != nil {
		enc.AddUint32("parent_id", *req.ParentID)
	}
	enc.AddArray("permission_ids", zapcore.ArrayMarshalerFunc(func(ae zapcore.ArrayEncoder) error {
		for _, id := range req.PermissionIDs {
			ae.AppendUint32(id)
		}
		return nil
	}))
	return nil
}

// UpdateMenuRequest 用于更新菜单的请求结构体
//
// swagger:model UpdateMenuRequest
type UpdateMenuRequest struct {
	// 前端路由路径
	Path string `json:"path" form:"path" binding:"required,max=100"`

	// 组件路径
	Component string `json:"component" form:"component" binding:"required,max=200"`

	// 名称
	Name string `json:"name" form:"name" binding:"required,max=50"`

	// 菜单元信息
	Meta MetaSchemas `json:"meta" form:"meta" binding:"required"`

	// 排序字段
	ArrangeOrder uint32 `json:"arrange_order" form:"arrange_order" binding:"required"`

	// 是否激活
	IsActive bool `json:"is_active" form:"is_active"`

	// 描述信息
	Descr string `json:"descr" form:"descr" binding:"omitempty,max=254"`

	// 父级菜单ID
	ParentID *uint32 `json:"parent_id" form:"parent_id" binding:"omitempty"`

	// 权限ID列表
	PermissionIDs []uint32 `json:"permission_ids" form:"permission_ids" binding:"omitempty"`
}

func (req *UpdateMenuRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("path", req.Path)
	enc.AddString("component", req.Component)
	enc.AddObject("meta", &req.Meta)
	enc.AddString("name", req.Name)
	enc.AddUint32("arrange_order", req.ArrangeOrder)
	enc.AddBool("is_active", req.IsActive)
	enc.AddString("descr", req.Descr)
	if req.ParentID != nil {
		enc.AddUint32("parent_id", *req.ParentID)
	}
	enc.AddArray("permission_ids", zapcore.ArrayMarshalerFunc(func(ae zapcore.ArrayEncoder) error {
		for _, id := range req.PermissionIDs {
			ae.AppendUint32(id)
		}
		return nil
	}))
	return nil
}

// ListMenuRequest 用于获取菜单列表的请求结构体
// 支持分页查询和多种筛选条件
//
// swagger:model ListMenuRequest
type ListMenuRequest struct {
	common.StandardModelQuery

	// 前端路由路径
	Path string `form:"path" binding:"omitempty,max=100"`

	// 组件路径
	Component string `form:"component" binding:"omitempty,max=200"`

	// 名称
	Name string `form:"name" binding:"omitempty,max=50"`

	// 是否激活
	IsActive *bool `form:"is_active" binding:"omitempty"`

	// 菜单描述
	Descr string `form:"descr" binding:"omitempty,max=254"`

	// 父级菜单ID
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
