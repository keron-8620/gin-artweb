package colony

import (
	"gin-artweb/api/common"
	mon "gin-artweb/api/mon/node"
	"gin-artweb/api/resource/pkg"
)

type OesColonyBaseOut struct {
	// ID
	ID uint32 `json:"id" example:"1"`

	// 系统类型
	SystemType string `json:"system_type" example:"STK"`

	// 集群号
	ColonyNum string `json:"colony_num" example:"01"`

	// 解压后名称
	ExtractedName string `json:"extracted_name" example:"oes"`
}

type OesColonyStandardOut struct {
	OesColonyBaseOut

	// 创建时间
	CreatedAt string `json:"created_at" example:"2023-01-01 12:00:00"`

	// 更新时间
	UpdatedAt string `json:"updated_at" example:"2023-01-01 12:00:00"`
}

type OesColonyDetailOut struct {
	OesColonyStandardOut

	// oes程序包
	Package *pkg.PackageStandardOut `json:"package"`

	// xcounter程序包
	XCounter *pkg.PackageStandardOut `json:"xcounter"`

	// mon节点
	MonNode *mon.MonNodeBaseOut `json:"mon_node"`
}

// OesColonyReply 程序包响应结构
type OesColonyReply = common.APIReply[OesColonyDetailOut]

// PagOesColonyReply 程序包的分页响应结构
type PagOesColonyReply = common.APIReply[*common.Pag[OesColonyDetailOut]]

// oes 任务状态
type OesColonyTaskInfo struct {
	// 集群号
	ColonyNum string `json:"colony_num" example:"01"`

	// 任务状态
	Tasks []common.TaskInfo `json:"tasks"`
}

// ListOesTasksInfoReply 多个oes集群的任务状态响应结构
type ListOesTasksInfoReply = common.APIReply[[]OesColonyTaskInfo]
