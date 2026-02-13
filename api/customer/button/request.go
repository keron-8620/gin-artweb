package button

import (
	"gin-artweb/api/common"

	"go.uber.org/zap/zapcore"
)

// CreateButtonRequest 用于创建按钮的请求结构体
//
// swagger:model CreateButtonRequest
type CreateButtonRequest struct {
	// 唯一标识
	ID uint32 `json:"id" form:"id" binding:"required,gt=0"`

	// 名称
	Name string `json:"name" form:"name" binding:"required,max=50"`

	// 排序字段
	Sort uint32 `json:"sort" form:"sort" binding:"required"`

	// 是否激活
	IsActive bool `json:"is_active" form:"is_active" binding:"required"`

	// 描述信息
	Descr string `json:"descr" form:"descr" binding:"omitempty,max=254"`

	// 菜单ID
	MenuID uint32 `json:"menu_id" form:"menu_id" binding:"required"`

	// API ID列表
	ApiIDs []uint32 `json:"api_ids" form:"api_ids" binding:"omitempty"`
}

func (req *CreateButtonRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddUint32("id", req.ID)
	enc.AddString("name", req.Name)
	enc.AddUint32("sort", req.Sort)
	enc.AddBool("is_active", req.IsActive)
	enc.AddString("descr", req.Descr)
	enc.AddUint32("menu_id", req.MenuID)
	enc.AddArray("api_ids", zapcore.ArrayMarshalerFunc(func(ae zapcore.ArrayEncoder) error {
		for _, id := range req.ApiIDs {
			ae.AppendUint32(id)
		}
		return nil
	}))
	return nil
}

// UpdateButtonRequest 用于更新按钮的请求结构体
//
// swagger:model UpdateButtonRequest
type UpdateButtonRequest struct {
	// 名称
	Name string `json:"name" form:"name" binding:"required,max=50"`

	// 排序字段
	Sort uint32 `json:"sort" form:"sort" binding:"omitempty"`

	// 是否激活
	IsActive bool `json:"is_active" form:"is_active"`

	// 描述信息
	Descr string `json:"descr" form:"descr" binding:"omitempty,max=254"`

	// 菜单ID
	MenuID uint32 `json:"menu_id" form:"menu_id" binding:"required"`

	// API ID列表
	ApiIDs []uint32 `json:"api_ids" form:"api_ids" binding:"omitempty"`
}

func (req *UpdateButtonRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("name", req.Name)
	enc.AddUint32("sort", req.Sort)
	enc.AddBool("is_active", req.IsActive)
	enc.AddString("descr", req.Descr)
	enc.AddUint32("menu_id", req.MenuID)
	enc.AddArray("api_ids", zapcore.ArrayMarshalerFunc(func(ae zapcore.ArrayEncoder) error {
		for _, id := range req.ApiIDs {
			ae.AppendUint32(id)
		}
		return nil
	}))
	return nil
}

// ListButtonRequest 用于获取按钮列表的请求结构体
// 支持分页查询和多种筛选条件
//
// swagger:model ListButtonRequest
type ListButtonRequest struct {
	common.StandardModelQuery

	// 按钮名称
	Name string `form:"name" binding:"omitempty,max=50"`

	// 是否激活
	IsActive *bool `form:"is_active" binding:"omitempty"`

	// 描述信息
	Descr string `form:"descr" binding:"omitempty,max=254"`

	// 菜单ID
	MenuID uint32 `form:"menu_id" binding:"omitempty"`
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
