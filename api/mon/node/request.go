package node

import "gin-artweb/api/common"

// CreateMonNodeRequest 用于创建mon节点的请求结构体
//
// swagger:model CreateMonNodeRequest
type CreateMonNodeRequest struct {
	// 名称
	// required: true
	// example: "mon上海节点"
	Name string `json:"name" binding:"required,max=50"`

	// 部署路径
	// required: true
	// example: "/home/monuser/mon"
	DeployPath string `json:"deploy_path" binding:"required"`

	// 导出路径
	// required: true
	// example: "/mnt/quant360/import/mon"
	OutportPath string `json:"outport_path" binding:"required"`

	// JAVA_HOME
	// required: true
	// example: "/home/monuser/jdk-11.0.1"
	JavaHome string `json:"java_home" bunding:"required"`

	// URL地址
	// required: true
	// example: "http://192.168.11.189:8080/mon"
	URL string `json:"url" bunding:"required"`

	// 主机ID
	// required: true
	// example: 1
	HostID uint32 `json:"host_id" binding:"required"`
}

// UpdateMonNodeRequest 用于更新mon节点的请求结构体
//
// swagger:model UpdateMonNodeRequest
type UpdateMonNodeRequest struct {
	// 名称
	// required: true
	// example: "mon上海节点"
	Name string `json:"name" binding:"required,max=50"`

	// 部署路径
	// required: true
	// example: "/home/monuser/mon"
	DeployPath string `json:"deploy_path" binding:"required"`

	// 导出路径
	// required: true
	// example: "/mnt/quant360/import/mon"
	OutportPath string `json:"outport_path" binding:"required"`

	// JAVA_HOME
	// required: true
	// example: "/home/monuser/jdk-11.0.1"
	JavaHome string `json:"java_home" bunding:"required"`

	// URL地址
	// required: true
	// example: "http://192.168.11.189:8080/mon"
	URL string `json:"url" bunding:"required"`

	// 主机ID
	// required: true
	// example: 1
	HostID uint32 `json:"host_id" binding:"required"`
}

// ListMonNodeRequest 用于获取mon节点列表的请求结构体
// 支持分页查询和多种筛选条件
//
// swagger:model ListMonNodeRequest
type ListMonNodeRequest struct {
	common.StandardModelQuery

	// 按名称搜索
	// required: false
	Name string `form:"name"`

	// 按主机ID筛选
	// required: false
	HostID uint32 `form:"host_id"`
}

func (req *ListMonNodeRequest) Query() (int, int, map[string]any) {
	page, size, query := req.BaseModelQuery.QueryMap(10)
	if req.Name != "" {
		query["name like ?"] = "%" + req.Name + "%"
	}
	if req.HostID > 0 {
		query["host_id = ?"] = req.HostID
	}
	return page, size, query
}
