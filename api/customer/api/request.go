package api

import (
	"go.uber.org/zap/zapcore"

	"gin-artweb/api/common"
)

// CreateApiRequest 用于创建权限的请求结构体
//
// swagger:model CreateApiRequest
type CreateApiRequest struct {
	// 唯一标识
	ID uint32 `json:"id" form:"id" binding:"required,gt=0"`

	// URL地址
	URL string `json:"url" form:"url" binding:"required,max=150"`

	// 请求方法
	Method string `json:"method" form:"method" binding:"required,oneof=GET POST PUT DELETE PATCH WS"`

	// 标签
	Label string `json:"label" form:"label" binding:"required,max=50"`

	// 描述信息
	Descr string `json:"descr" form:"descr" binding:"max=254"`
}

func (req *CreateApiRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddUint32("id", req.ID)
	enc.AddString("url", req.URL)
	enc.AddString("method", req.Method)
	enc.AddString("label", req.Label)
	enc.AddString("descr", req.Descr)
	return nil
}

// UpdateApiRequest 用于更新权限的请求结构体
// 包含权限主键、HTTP URL、请求方法和描述信息
//
// swagger:model UpdateApiRequest
type UpdateApiRequest struct {
	// URL地址
	URL string `json:"url" form:"url" binding:"required,max=150"`

	// 请求方法
	Method string `json:"method" form:"method" binding:"required,oneof=GET POST PUT DELETE PATCH WS"`

	// 标签
	Label string `json:"label" form:"label" binding:"required,max=50"`

	// 描述信息
	Descr string `json:"descr" form:"descr" binding:"omitempty,max=254"`
}

func (req *UpdateApiRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("url", req.URL)
	enc.AddString("method", req.Method)
	enc.AddString("label", req.Label)
	enc.AddString("descr", req.Descr)
	return nil
}

// ListApiRequest 用于获取权限列表的请求结构体
// 支持分页查询和多种筛选条件
//
// swagger:model ListApiRequest
type ListApiRequest struct {
	common.StandardModelQuery

	// URL地址
	URL string `form:"url" binding:"omitempty,max=150"`

	// 请求方法
	Method string `form:"method" binding:"omitempty,oneof=GET POST PUT DELETE PATCH WS"`

	// 标签
	Label string `form:"label" binding:"omitempty,max=50"`

	// 描述信息
	Descr string `form:"descr" binding:"omitempty,max=254"`
}

func (req *ListApiRequest) Query() (int, int, map[string]any) {
	page, size, query := req.StandardModelQuery.QueryMap(10)
	if req.URL != "" {
		query["url like ?"] = "%" + req.URL + "%"
	}
	if req.Method != "" {
		query["method = ?"] = req.Method
	}
	if req.Label != "" {
		query["label like ?"] = "%" + req.Label + "%"
	}
	if req.Descr != "" {
		query["descr like ?"] = "%" + req.Descr + "%"
	}
	return page, size, query
}
