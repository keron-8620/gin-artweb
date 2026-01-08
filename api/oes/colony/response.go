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
	SystemType string `json:"system_type" example:"stk"`
	// 集群号
	ColonyNum string `json:"colony_num" example:"test"`
	// 解压后名称
	ExtractedName string `json:"extracted_name" example:""`
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
	Package  *pkg.PackageStandardOut `json:"package"`
	XCounter *pkg.PackageStandardOut `json:"xcounter"`
	MonNode  *mon.MonNodeBaseOut     `json:"mon_node"`
}

// OesColonyReply 程序包响应结构
type OesColonyReply = common.APIReply[OesColonyDetailOut]

// PagOesColonyReply 程序包的分页响应结构
type PagOesColonyReply = common.APIReply[*common.Pag[OesColonyStandardOut]]

// 现货的任务状态
type StkTaskStatus struct {
	Mon     string `json:"mon"`
	Counter string `json:"counter"`
	Bse     string `json:"bse"`
	Sse     string `json:"sse"`
	Szse    string `json:"szse"`
	Csdc    string `json:"csdc"`
}

// ListStkTaskStatusReply 多个oes现货集群的任务状态响应结构
type ListStkTaskStatusReply = common.APIReply[map[string]StkTaskStatus]

// 两融的任务状态
type CrdTaskStatus struct {
	Mon      string `json:"mon"`
	Counter  string `json:"counter"`
	Sse      string `json:"sse"`
	Szse     string `json:"szse"`
	SseLate  string `json:"sse_late"`
	SzseLate string `json:"szse_late"`
	Csdc     string `json:"csdc"`
}

// ListCrdTaskStatusReply 多个oes两融集群的任务状态响应结构
type ListCrdTaskStatusReply = common.APIReply[map[string]CrdTaskStatus]

// 期权任务状态
type OptTaskStatus struct {
	Mon     string `json:"mon"`
	Counter string `json:"counter"`
	Sse     string `json:"sse"`
	Szse    string `json:"szse"`
}

// ListOptTaskStatusReply 多个oes期权集群的任务状态响应结构
type ListOptTaskStatusReply = common.APIReply[map[string]OptTaskStatus]
