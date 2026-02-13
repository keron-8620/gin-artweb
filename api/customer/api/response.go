package api

import (
	"gin-artweb/api/common"
)

// ApiStandardOut API基础信息
type ApiStandardOut struct {
	// 唯一标识
	ID uint32 `json:"id" example:"1"`

	// 创建时间
	CreatedAt string `json:"created_at" example:"2023-01-01 12:00:00"`

	// 更新时间
	UpdatedAt string `json:"updated_at" example:"2023-01-01 12:00:00"`

	// HTTP路径
	URL string `json:"url" example:"/api/v1/users"`

	// 请求方法
	Method string `json:"method" example:"GET"`

	// 标签
	Label string `json:"label" example:"customer"`

	// 描述
	Descr string `json:"descr" example:"用户管理权限"`
}

// ApiReply 权限响应结构
type ApiReply = common.APIReply[*ApiStandardOut]

// PagApiReply API的分页响应结构
type PagApiReply = common.APIReply[*common.Pag[ApiStandardOut]]
