package colony

import "gin-artweb/api/common"

// CreateOrUpdateOesColonyRequest 用于创建mon节点的请求结构体
//
// swagger:model CreateOrUpdateOesColonyRequest
type CreateOrUpdateOesColonyRequest struct {
	// 系统类型
	SystemType string `json:"system_type" form:"system_type" binding:"required,oneof=STK CRD OPT"`

	// 集群号
	ColonyNum string `json:"colony_num" form:"colony_num" binding:"required,max=2"`

	// 解压后名称
	ExtractedName string `json:"extracted_name" form:"extracted_name" binding:"required,max=50"`

	// 是否启用
	IsEnable bool `json:"is_enable" form:"is_enable" binding:"required"`

	// 程序包ID
	PackageID uint32 `json:"package_id" form:"package_id" binding:"required"`

	// xcounter包ID
	XCounterID uint32 `json:"xcounter_id" form:"xcounter_id" binding:"required"`

	// mon节点ID
	MonNodeID uint32 `json:"mon_node_id" form:"mon_node_id" binding:"required"`
}

// ListOesColonyRequest 用于获取mon节点列表的请求结构体
// 支持分页查询和多种筛选条件
//
// swagger:model ListOesColonyRequest
type ListOesColonyRequest struct {
	common.StandardModelQuery

	// 系统类型
	SystemType string `form:"system_type"`

	// 集群号
	ColonyNum string `form:"colony_num"`

	// 解压后名称
	ExtractedName string `form:"extracted_name"`

	// 是否启用
	IsEnable *bool `form:"is_enable"`

	// 程序包ID
	PackageID uint32 `form:"package_id"`

	// xcounter包ID
	XCounterID uint32 `form:"xcounter_id"`

	// mon节点ID
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
	if req.IsEnable != nil {
		query["is_enable = ?"] = *req.IsEnable
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
