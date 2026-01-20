package node

import "gin-artweb/api/common"

// CreateOrUpdateMonNodeRequest 用于创建mon节点的请求结构体
//
// swagger:model CreateOrUpdateMonNodeRequest
type CreateOrUpdateMonNodeRequest struct {
	// 名称
	Name string `json:"name" form:"name" binding:"required,max=50"`

	// 部署路径
	DeployPath string `json:"deploy_path" form:"deploy_path" binding:"required"`

	// 导出路径
	OutportPath string `json:"outport_path" form:"outport_path" binding:"required"`

	// JAVA_HOME
	JavaHome string `json:"java_home" form:"java_home" bunding:"required"`

	// URL地址
	URL string `json:"url" form:"url" bunding:"required"`

	// 主机ID
	HostID uint32 `json:"host_id" form:"host_id" binding:"required"`
}

// ListMonNodeRequest 用于获取mon节点列表的请求结构体
// 支持分页查询和多种筛选条件
//
// swagger:model ListMonNodeRequest
type ListMonNodeRequest struct {
	common.StandardModelQuery

	// 名称
	Name string `form:"name"`

	// 主机ID
	HostID uint32 `form:"host_id"`
}

func (req *ListMonNodeRequest) Query() (int, int, map[string]any) {
	page, size, query := req.StandardModelQuery.QueryMap(10)
	if req.Name != "" {
		query["name like ?"] = "%" + req.Name + "%"
	}
	if req.HostID > 0 {
		query["host_id = ?"] = req.HostID
	}
	return page, size, query
}
