package host

import (
	"gin-artweb/api/common"
)

// HostOutBase 主机基础信息
type HostOutBase struct {
	// 主机ID
	ID uint32 `json:"id" example:"1"`
	// 创建时间
	CreatedAt string `json:"created_at" example:"2023-01-01 12:00:00"`
	// 更新时间
	UpdatedAt string `json:"updated_at" example:"2023-01-01 12:00:00"`
	// 名称
	Name string `json:"name" example:"artweb主机"`
	// 标签
	Label string `json:"label" example:"artweb"`
	// IP地址
	IPAddr string `json:"ip_addr" example:"192.168.1.1"`
	// 端口
	Port uint16 `json:"port" example:"22"`
	// 用户名
	Username string `json:"username" example:"root"`
	// Python路径
	PyPath string `json:"py_path" example:"/usr/bin/python3"`
	// 备注
	Remark string `json:"remark" example:"测试"`
}

// HostReply 主机响应结构
type HostReply = common.APIReply[HostOutBase]

// PagHostReply 主机的分页响应结构
type PagHostReply = common.APIReply[*common.Pag[HostOutBase]]
