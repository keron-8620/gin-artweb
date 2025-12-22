package colony

import "gin-artweb/api/common"

// CreateOrUpdateOesColonyRequest 用于创建mon节点的请求结构体
//
// swagger:model CreateOrUpdateOesColonyRequest
type CreateOrUpdateOesColonyRequest struct {
	// 系统类型
	// required: true
	// example: "stk"
	SystemType string `json:"system_type" form:"system_type" binding:"required,oneof=STK CRD OPT"`

	// 集群号
	// required: true
	// example: "01"
	ColonyNum string `json:"colony_num" form:"colony_num" binding:"required,max=2"`

	// 解压后名称
	// required: true
	// example: "oes"
	ExtractedName string `json:"extracted_name" form:"extracted_name" binding:"required,max=50"`

	// 程序包ID
	// required: true
	// example: 1
	PackageID uint32 `json:"package_id" form:"package_id" binding:"required"`

	// xcounter包ID
	// required: true
	// example: 1
	XCounterID uint32 `json:"xcounter_id" form:"xcounter_id" binding:"required"`

	// mon节点ID
	// required: true
	// example: 1
	MonNodeID uint32 `json:"mon_node_id" form:"mon_node_id" binding:"required"`
}

// ListOesColonyRequest 用于获取mon节点列表的请求结构体
// 支持分页查询和多种筛选条件
//
// swagger:model ListOesColonyRequest
type ListOesColonyRequest struct {
	common.StandardModelQuery

	// 系统类型
	// required: true
	// example: "stk"
	SystemType string `form:"system_type"`

	// 集群号
	// required: false
	// example: "01"
	ColonyNum string `form:"colony_num"`

	// 解压后名称
	// required: false
	// example: "oes"
	ExtractedName string `form:"extracted_name"`

	// 程序包ID
	// required: false
	// example: 1
	PackageID uint32 `form:"package_id"`

	// xcounter包ID
	// required: false
	// example: 1
	XCounterID uint32 `form:"xcounter_id"`

	// mon节点ID
	// required: false
	// example: 1
	MonNodeID uint32 `form:"mon_node_id"`
}

func (req *ListOesColonyRequest) Query() (int, int, map[string]any) {
	page, size, query := req.StandardModelQuery.QueryMap(14)
	if req.SystemType != "" {
		query["system_type = ?"] = req.SystemType
	}
	if req.ColonyNum != "" {
		query["colony_num = ?"] = req.ColonyNum
	}
	if req.ExtractedName != "" {
		query["extracted_name = ?"] = "%" + req.ExtractedName + "%"
	}
	if req.PackageID > 0 {
		query["package_id = ?"] = req.PackageID
	}
	if req.XCounterID > 0 {
		query["xcounter_id"] = req.XCounterID
	}
	if req.MonNodeID > 0 {
		query["mon_node_id = ?"] = req.MonNodeID
	}
	return page, size, query
}
