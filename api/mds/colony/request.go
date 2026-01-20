package colony

import "gin-artweb/api/common"

// CreateOrUpdateMdsColonyRequest 用于创建mon节点的请求结构体
//
// swagger:model CreateOrUpdateMdsColonyRequest
type CreateOrUpdateMdsColonyRequest struct {
	// 集群号
	ColonyNum string `json:"colony_num" form:"colony_num" binding:"required,max=2"`

	// 解压后名称
	ExtractedName string `json:"extracted_name" form:"extracted_name" binding:"required,max=50"`

	// 程序包ID
	PackageID uint32 `json:"package_id" form:"package_id" binding:"required"`

	// mon节点ID
	MonNodeID uint32 `json:"mon_node_id" form:"mon_node_id" binding:"required"`
}

// ListMdsColonyRequest 用于获取mon节点列表的请求结构体
// 支持分页查询和多种筛选条件
//
// swagger:model ListMdsColonyRequest
type ListMdsColonyRequest struct {
	common.StandardModelQuery

	// 集群号
	ColonyNum string `form:"colony_num"`

	// 解压后名称
	ExtractedName string `form:"extracted_name"`

	// 程序包ID
	PackageID uint32 `form:"package_id"`

	// mon节点ID
	MonNodeID uint32 `form:"mon_node_id"`
}

func (req *ListMdsColonyRequest) Query() (int, int, map[string]any) {
	page, size, query := req.StandardModelQuery.QueryMap(12)
	if req.ColonyNum != "" {
		query["colony_num = ?"] = req.ColonyNum
	}
	if req.ExtractedName != "" {
		query["extracted_name = ?"] = "%" + req.ExtractedName + "%"
	}
	if req.PackageID > 0 {
		query["package_id = ?"] = req.PackageID
	}
	if req.MonNodeID > 0 {
		query["mon_node_id = ?"] = req.MonNodeID
	}
	return page, size, query
}
