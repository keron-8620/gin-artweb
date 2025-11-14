package permission

import "gin-artweb/api/common"

// CreatePermissionRequest 用于创建权限的请求结构体
//
// swagger:model CreatePermissionRequest
type CreatePermissionRequest struct {
	// 权限主键，必须大于0
	// Required: true
	// Minimum: 1
	ID uint32 `json:"id" binding:"required,gt=0"`

	// 权限对应的HTTP URL，最大长度150
	// Required: true
	// Max length: 150
	URL string `json:"url" binding:"required,max=150"`

	// HTTP请求方法，枚举值验证
	// Required: true
	// Enum: GET,POST,PUT,DELETE,PATCH,WS
	Method string `json:"method" binding:"required,oneof=GET POST PUT DELETE PATCH WS"`

	// 权限描述信息，最大长度50
	// Required: true
	// Max length: 50
	Label string `json:"label" binding:"required,max=50"`

	// 权限描述信息，可选，最大长度254
	// Max length: 254
	Descr string `json:"descr" binding:"max=254"`
}

// UpdatePermissionRequest 用于更新权限的请求结构体
// 包含权限主键、HTTP URL、请求方法和描述信息
//
// swagger:model UpdatePermissionRequest
type UpdatePermissionRequest struct {
	// 权限对应的HTTP URL，最大长度150
	// Required: true
	// Max length: 150
	URL string `json:"url" binding:"required,max=150"`

	// HTTP请求方法，枚举值验证
	// Required: true
	// Enum: GET,POST,PUT,DELETE,PATCH,WS
	Method string `json:"method" binding:"required,oneof=GET POST PUT DELETE PATCH WS"`

	// 权限描述信息，最大长度50
	// Required: true
	// Max length: 50
	Label string `json:"label" binding:"required,max=50"`

	// 权限描述信息，可选，最大长度254
	// Max length: 254
	Descr string `json:"descr" binding:"omitempty,max=254"`
}

// ListPermissionRequest 用于获取权限列表的请求结构体
// 支持分页查询和多种筛选条件
//
// swagger:model ListPermissionRequest
type ListPermissionRequest struct {
	common.StandardModelQuery

	// 权限对应的HTTP URL，字符串长度限制
	// Max length: 150
	URL string `form:"url" binding:"omitempty,max=150"`

	// HTTP请求方法，枚举值验证
	// Enum: GET,POST,PUT,DELETE,PATCH,WS
	Method string `form:"method" binding:"omitempty,oneof=GET POST PUT DELETE PATCH WS"`

	// 权限描述信息，最大长度50
	// Max length: 50
	Label string `json:"label" binding:"omitempty,max=50"`

	// 描述信息，字符串长度限制
	// Max length: 254
	Descr string `form:"descr" binding:"omitempty,max=254"`
}

func (req *ListPermissionRequest) Query() (int, int, map[string]any) {
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
