package node

import (
	"gin-artweb/api/common"
	"gin-artweb/api/resource/host"
)

type MonNodeBaseOut struct {
	// 计划任务ID
	ID uint32 `json:"id" example:"1"`
	// 名称
	Name string `json:"name" example:"test"`
	// 部署路径
	DeployPath string `json:"deploy_path" example:""`
	// 导出路径
	OutportPath string `json:"outport_path" example:""`
	// JAVA_HOME
	JavaHome string `json:"java_home" example:""`
	// URL地址
	URL string `json:"url" example:"http://192.168.11.189:8080"`
}

type MonNodeStandardOut struct {
	MonNodeBaseOut
	// 创建时间
	CreatedAt string `json:"created_at" example:"2023-01-01 12:00:00"`
	// 更新时间
	UpdatedAt string `json:"updated_at" example:"2023-01-01 12:00:00"`
}

type MonNodeDetailOut struct {
	MonNodeStandardOut
	Host *host.HostBaseOut `json:"host"`
}

// MonNodeReply 程序包响应结构
type MonNodeReply = common.APIReply[MonNodeDetailOut]

// PagMonNodeReply 程序包的分页响应结构
type PagMonNodeReply = common.APIReply[*common.Pag[MonNodeDetailOut]]
