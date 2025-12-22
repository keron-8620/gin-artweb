package node

import (
	"gin-artweb/api/common"
	"gin-artweb/api/oes/colony"
	"gin-artweb/api/resource/host"
)

type OesNodeBaseOut struct {
	// ID
	ID uint32 `json:"id" example:"1"`
	// 节点角色
	NodeRole string `json:"node_role" example:"master"`
	// 是否启用
	IsEnable bool `json:"is_enable"`
}

type OesNodeStandardOut struct {
	OesNodeBaseOut
	// 创建时间
	CreatedAt string `json:"created_at" example:"2023-01-01 12:00:00"`
	// 更新时间
	UpdatedAt string `json:"updated_at" example:"2023-01-01 12:00:00"`
}

type OesNodeDetailOut struct {
	OesNodeStandardOut
	OesColony *colony.OesColonyBaseOut `json:"oes_colony"`
	Host      *host.HostBaseOut        `json:"host"`
}

// MonNodeReply 程序包响应结构
type OesNodeReply = common.APIReply[OesNodeDetailOut]

// PagMonNodeReply 程序包的分页响应结构
type PagOesNodeReply = common.APIReply[*common.Pag[OesNodeStandardOut]]
