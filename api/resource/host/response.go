package host

import (
	"gin-artweb/api/common"
)

type HostBaseOut struct {
	// 主机ID
	ID uint32 `json:"id" example:"1"`

	// 名称
	Name string `json:"name" example:"artweb主机"`

	// 标签
	Label string `json:"label" example:"artweb"`

	// IP地址
	SSHIP string `json:"ssh_ip" example:"192.168.1.1"`

	// 端口
	SSHPort uint16 `json:"ssh_port" example:"22"`

	// 用户名
	SSHUser string `json:"ssh_user" example:"root"`

	// Python路径
	PyPath string `json:"py_path" example:"/usr/bin/python3"`

	// 备注
	Remark string `json:"remark" example:"测试"`
}

// HostStandardOut 主机基础信息
type HostStandardOut struct {
	HostBaseOut

	// 创建时间
	CreatedAt string `json:"created_at" example:"2023-01-01 12:00:00"`

	// 更新时间
	UpdatedAt string `json:"updated_at" example:"2023-01-01 12:00:00"`
}

// HostReply 主机响应结构
type HostReply = common.APIReply[HostStandardOut]

// PagHostReply 主机的分页响应结构
type PagHostReply = common.APIReply[*common.Pag[HostStandardOut]]
