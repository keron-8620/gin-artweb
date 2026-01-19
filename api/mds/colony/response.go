package colony

import (
	"gin-artweb/api/common"
	mon "gin-artweb/api/mon/node"
	"gin-artweb/api/resource/pkg"
)

type MdsColonyBaseOut struct {
	// ID
	ID uint32 `json:"id" example:"1"`
	// 集群号
	ColonyNum string `json:"colony_num" example:"test"`
	// 解压后名称
	ExtractedName string `json:"extracted_name" example:""`
}

type MdsColonyStandardOut struct {
	MdsColonyBaseOut
	// 创建时间
	CreatedAt string `json:"created_at" example:"2023-01-01 12:00:00"`
	// 更新时间
	UpdatedAt string `json:"updated_at" example:"2023-01-01 12:00:00"`
}

type MdsColonyDetailOut struct {
	MdsColonyStandardOut
	Package *pkg.PackageStandardOut `json:"package"`
	MonNode *mon.MonNodeBaseOut     `json:"mon_node"`
}

// MdsColonyReply mds集群配置的响应结构
type MdsColonyReply = common.APIReply[MdsColonyDetailOut]

// PagMdsColonyReply mds集群配置的分页响应结构
type PagMdsColonyReply = common.APIReply[*common.Pag[MdsColonyDetailOut]]

// mds 任务状态
type MdsTaskStatus struct {
	Mon  string `json:"mon"`
	Bse  string `json:"bse"`
	Sse  string `json:"sse"`
	Szse string `json:"szse"`
}

// ListMdsTaskStatusReply 多个mds集群的任务状态响应结构
type ListMdsTaskStatusReply = common.APIReply[map[string]MdsTaskStatus]
