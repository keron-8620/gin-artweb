package permission

import (
	"gitee.com/keion8620/go-dango-gin/pkg/common"
)

// PermissionOutBase 权限基础信息
type PermissionOutBase struct {
	// 权限ID
	Id uint `json:"id" example:"1"`
	// 创建时间
	CreatedAt string `json:"created_at" example:"2023-01-01 12:00:00"`
	// 更新时间
	UpdatedAt string `json:"updated_at" example:"2023-01-01 12:00:00"`
	// HTTP路径
	HttpUrl string `json:"http_url" example:"/api/v1/users"`
	// 请求方法
	Method string `json:"method" example:"GET"`
	// 描述
	Descr string `json:"descr" example:"用户管理权限"`
}

// PermissionReply 权限响应结构
type PermissionReply = common.APIReply[PermissionOutBase]

// PagPermissionReply 权限的分页响应结构
type PagPermissionReply = common.APIReply[*common.Pag[*PermissionOutBase]]
