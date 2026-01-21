package node

import (
	"gin-artweb/api/common"
	"gin-artweb/api/mds/colony"
	"gin-artweb/api/resource/host"
)

type MdsNodeBaseOut struct {
	// ID
	ID uint32 `json:"id" example:"1"`

	// 节点角色
	NodeRole string `json:"node_role" example:"master"`

	// 是否启用
	IsEnable bool `json:"is_enable"`
}

type MdsNodeStandardOut struct {
	MdsNodeBaseOut

	// 创建时间
	CreatedAt string `json:"created_at" example:"2023-01-01 12:00:00"`

	// 更新时间
	UpdatedAt string `json:"updated_at" example:"2023-01-01 12:00:00"`
}

type MdsNodeDetailOut struct {
	MdsNodeStandardOut

	// mds集群
	MdsColony *colony.MdsColonyBaseOut `json:"mds_colony"`

	// 主机
	Host *host.HostBaseOut `json:"host"`
}

// MdsNodeReply mds节点配置的响应结构
type MdsNodeReply = common.APIReply[MdsNodeDetailOut]

// PagMdsNodeReply mds节点配置的分页响应结构
type PagMdsNodeReply = common.APIReply[*common.Pag[MdsNodeDetailOut]]
