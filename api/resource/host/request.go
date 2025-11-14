package host

import "gin-artweb/api/common"

// CreateHosrRequest 用于创建主机的请求结构体
//
// swagger:model CreateHosrRequest
type CreateHosrRequest struct {
	// 名称，最大长度50
	// Required: true
	// Max length: 50
	Name string `json:"name" binding:"required,max=50"`

	// 标签，最大长度50
	// Required: true
	// Max length: 50
	Label string `json:"label" binding:"required,max=50"`

	// ip地址，最大长度108
	// Required: true
	// Max length: 108
	IPAddr string `json:"ip_addr" binding:"required,max=108"`

	// 端口，必填
	// Required: true
	Port uint16

	// 用户名，最大长度50
	// Required: true
	// Max length: 50
	Username string `json:"username" binding:"required,max=50"`

	// 密码，最大长度150
	// Required: true
	// Max length: 150
	Password string `json:"password" binding:"required,max=150"`

	// python路径，最大长度254
	// Required: true
	// Max length: 254
	PyPath string `json:"py_path" binding:"omitempty,max=254"`

	// 备注，最大长度254
	// Max length: 254
	Remark string `json:"remark" binding:"max=254"`
}

// UpdateHostRequest 用于更新主机的请求结构体
// 包含主机主键、HTTP URL、请求方法和描述信息
//
// swagger:model UpdateHostRequest
type UpdateHostRequest struct {
	// 名称，最大长度50
	// Required: true
	// Max length: 50
	Name string `json:"name" binding:"required,max=50"`

	// 标签，最大长度50
	// Required: true
	// Max length: 50
	Label string `json:"label" binding:"required,max=50"`

	// ip地址，最大长度108
	// Required: true
	// Max length: 108
	IPAddr string `json:"ip_addr" binding:"required,max=108"`

	// 端口，必填
	// Required: true
	Port uint16

	// 用户名，最大长度50
	// Required: true
	// Max length: 50
	Username string `json:"username" binding:"required,max=50"`

	// 密码，最大长度150
	// Required: true
	// Max length: 150
	Password string `json:"password" binding:"required,max=150"`

	// python路径，最大长度254
	// Required: true
	// Max length: 254
	PyPath string `json:"py_path" binding:"omitempty,max=254"`

	// 备注，最大长度254
	// Max length: 254
	Remark string `json:"remark" binding:"max=254"`
}

// ListHostRequest 用于获取主机列表的请求结构体
// 支持分页查询和多种筛选条件
//
// swagger:model ListHostRequest
type ListHostRequest struct {
	common.StandardModelQuery

	// 名称，最大长度50
	// Required: true
	// Max length: 50
	Name string `json:"name" binding:"required,max=50"`

	// 标签，最大长度50
	// Required: true
	// Max length: 50
	Label string `json:"label" binding:"required,max=50"`

	// ip地址，最大长度108
	// Required: true
	// Max length: 108
	IPAddr string `json:"ip_addr" binding:"required,max=108"`

	// 端口，必填
	// Required: true
	Port *uint16

	// 用户名，最大长度50
	// Required: true
	// Max length: 50
	Username string `json:"username" binding:"required,max=50"`

	// python路径，最大长度254
	// Required: true
	// Max length: 254
	PyPath string `json:"py_path" binding:"omitempty,max=254"`

	// 备注，最大长度254
	// Max length: 254
	Remark string `json:"remark" binding:"max=254"`
}

func (req *ListHostRequest) Query() (int, int, map[string]any) {
	page, size, query := req.StandardModelQuery.QueryMap(13)
	if req.Name != "" {
		query["name like ?"] = "%" + req.Name + "%"
	}
	if req.Label != "" {
		query["label = ?"] = req.Label
	}
	if req.IPAddr != "" {
		query["ip_addr = ?"] = req.IPAddr
	}
	if req.Port != nil {
		query["port = ?"] = *req.Port
	}
	if req.Username != "" {
		query["username list ?"] = "%" + req.Username + "%"
	}
	if req.PyPath != "" {
		query["py_path lisk ?"] = "%" + req.PyPath + "%"
	}
	if req.Remark != "" {
		query["remark like ?"] = "%" + req.Remark + "%"
	}
	return page, size, query
}
