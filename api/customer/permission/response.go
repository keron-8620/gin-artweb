package permission

import (
	"gin-artweb/pkg/common"
)

// PermissionOutBase 权限基础信息
type PermissionOutBase struct {
	// 权限ID
	Id uint32 `json:"id" example:"1"`
	// 创建时间
	CreatedAt string `json:"created_at" example:"2023-01-01 12:00:00"`
	// 更新时间
	UpdatedAt string `json:"updated_at" example:"2023-01-01 12:00:00"`
	// HTTP路径
	Url string `json:"Url" example:"/api/v1/users"`
	// 请求方法
	Method string `json:"method" example:"GET"`
	// 标签
	Label string `json:"label" example:"customer"`
	// 描述
	Descr string `json:"descr" example:"用户管理权限"`
}

// PermissionReply 权限响应结构
type PermissionReply = common.APIReply[PermissionOutBase]

// PagPermissionReply 权限的分页响应结构
type PagPermissionReply = common.APIReply[*common.Pag[*PermissionOutBase]]
