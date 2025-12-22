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
	Package *pkg.PackageStandardOut `json:"package"`
	XCounter *pkg.PackageStandardOut `json:"xcounter"`
	MonNode *mon.MonNodeBaseOut     `json:"mon_node"`
}

// OesColonyReply 程序包响应结构
type OesColonyReply = common.APIReply[OesColonyDetailOut]

// PagOesColonyReply 程序包的分页响应结构
type PagOesColonyReply = common.APIReply[*common.Pag[OesColonyStandardOut]]
