package role

import (
	"gin-artweb/api/common"

	"go.uber.org/zap/zapcore"
)

// CreateOrUpdateRoleRequest 用于创建或更新角色的请求结构体
//
// swagger:model CreateOrUpdateRoleRequest
type CreateOrUpdateRoleRequest struct {
	// 名称
	Name string `json:"name" form:"name" binding:"required,max=50"`

	// 描述信息
	Descr string `json:"descr" form:"descr" binding:"omitempty,max=254"`

	// 权限ID列表
	PermissionIDs []uint32 `json:"permission_ids" form:"permission_ids" binding:"omitempty"`

	// 菜单ID列表
	MenuIDs []uint32 `json:"menu_ids" form:"menu_ids" binding:"omitempty"`

	// 按钮ID列表
	ButtonIDs []uint32 `json:"button_ids" form:"button_ids" binding:"omitempty"`
}

func (req *CreateOrUpdateRoleRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("name", req.Name)
	enc.AddString("descr", req.Descr)
	enc.AddArray("permission_ids", zapcore.ArrayMarshalerFunc(func(ae zapcore.ArrayEncoder) error {
		for _, id := range req.PermissionIDs {
			ae.AppendUint32(id)
		}
		return nil
	}))
	enc.AddArray("menu_ids", zapcore.ArrayMarshalerFunc(func(ae zapcore.ArrayEncoder) error {
		for _, id := range req.MenuIDs {
			ae.AppendUint32(id)
		}
		return nil
	}))
	enc.AddArray("button_ids", zapcore.ArrayMarshalerFunc(func(ae zapcore.ArrayEncoder) error {
		for _, id := range req.ButtonIDs {
			ae.AppendUint32(id)
		}
		return nil
	}))
	return nil
}

// ListRoleRequest 用于获取角色列表的请求结构体
// 支持分页查询和多种筛选条件
//
// swagger:model ListRoleRequest
type ListRoleRequest struct {
	common.StandardModelQuery

	// 名称
	Name string `form:"name" binding:"omitempty,max=50"`

	// 描述信息
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
