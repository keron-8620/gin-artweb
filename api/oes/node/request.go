package node

import "gin-artweb/api/common"

// CreateOrUpdateOesNodeRequest 用于创建oes节点的请求结构体
//
// swagger:model CreateOrUpdateOesNodeRequest
type CreateOrUpdateOesNodeRequest struct {
	// 节点角色
	// required: true
	// example: "01"
	NodeRole string `json:"node_role" form:"node_role" binding:"required,oneof=master follow arbiter"`

	// 是否启用
	// required: true
	// example: true
	IsEnable bool `json:"is_enable" form:"is_enable"`

	// oes集群ID
	// required: true
	// example: 1
	OesColonyID uint32 `json:"oes_colony_id" form:"oes_colony_id" binding:"required"`

	// 主机ID
	// required: true
	// example: 1
	HostID uint32 `json:"host_id" form:"host_id" binding:"required"`
}

// ListOesNodeRequest 用于获取oes节点列表的请求结构体
// 支持分页查询和多种筛选条件
//
// swagger:model ListOesNodeRequest
type ListOesNodeRequest struct {
	common.StandardModelQuery

	// 节点角色
	// example: "master"
	NodeRole string `form:"node_role"`

	// 是否启用
	// required: false
	// example: true
	IsEnable *bool `form:"is_enable"`

	// oes集群ID
	// required: false
	// example: 1
	OesColonyID uint32 `form:"oes_colony_id"`

	// 主机ID
	// required: false
	// example: 1
	HostID uint32 `form:"host_id"`
}

func (req *ListOesNodeRequest) Query() (int, int, map[string]any) {
	page, size, query := req.StandardModelQuery.QueryMap(12)
	if req.NodeRole != "" {
		query["NodeRole = ?"] = req.NodeRole
	}
	if req.IsEnable != nil {
		query["is_enable = ?"] = *req.IsEnable
	}
	if req.OesColonyID > 0 {
		query["oes_colony_id = ?"] = req.OesColonyID
	}
	if req.HostID > 0 {
		query["host_id = ?"] = req.HostID
	}
	return page, size, query
}
