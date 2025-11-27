package record

import (
	"time"

	"gin-artweb/api/common"
)

// ListLoginRecordRequest 用于获取用户登陆记录的请求结构体
// 支持分页查询和多种筛选条件
//
// swagger:model ListUserRequest
type ListLoginRecordRequest struct {
	common.BaseModelQuery

	// 用户名称，字符串长度限制
	// Max length: 50
	Username string `form:"name" binding:"omitempty,max=50"`

	// IP 地址，字符串长度限制
	// Max length: 108
	IPAddress string `form:"ip_address" binding:"omitempty,max=108"`

	// 登陆状态
	Status *bool `form:"status" binding:"omitempty"`

	// 登陆时间之前的记录 (RFC3339格式)
	// example: 2023-01-01T00:00:00Z
	BeforeLoginAt string `form:"before_login_at" binding:"omitempty"`

	// 登陆时间之后的记录 (RFC3339格式)
	// example: 2023-01-01T00:00:00Z
	AfterLoginAt string `form:"after_login_at" binding:"omitempty"`
}

func (req *ListLoginRecordRequest) Query() (int, int, map[string]any) {
	page, size, query := req.BaseModelQuery.QueryMap(9)
	if req.Username != "" {
		query["username = ?"] = req.Username
	}
	if req.IPAddress != "" {
		query["ip_address = ?"] = req.IPAddress
	}
	if req.Status != nil {
		query["status = ?"] = &req.Status
	}
	if req.BeforeLoginAt != "" {
		bft, err := time.Parse(time.RFC3339, req.BeforeLoginAt)
		if err == nil {
			query["login_at < ?"] = bft
		}
	}
	if req.AfterLoginAt != "" {
		act, err := time.Parse(time.RFC3339, req.AfterLoginAt)
		if err == nil {
			query["login_at > ?"] = act
		}
	}
	return page, size, query
}
