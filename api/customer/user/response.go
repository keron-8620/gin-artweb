package user

import (
	"gin-artweb/api/customer/role"
	"gin-artweb/pkg/common"
)

// UserOutBase用户基础信息
type UserOutBase struct {
	// 用户ID
	ID uint32 `json:"id" example:"1"`
	// 创建时间
	CreatedAt string `json:"created_at" example:"2023-01-01 12:00:00"`
	// 更新时间
	UpdatedAt string `json:"updated_at" example:"2023-01-01 12:00:00"`
	// 名称
	Username string `json:"username" example:"judgement"`
	// 是否激活
	IsActive bool `json:"is_active" example:"true"`
	// 是否是工作人员
	IsStaff bool `json:"is_staff" example:"false"`
}

type UserOut struct {
	UserOutBase
	Role *role.RoleOutBase `json:"role"`
}

// UserBaseReply 用户响应结构
type UserReply = common.APIReply[UserOut]

// PagUserReply 用户的分页响应结构
type PagUserReply = common.APIReply[*common.Pag[*UserOut]]

type LoginOut struct {
	Token string `json:"token"`
}

type LoginReply = common.APIReply[LoginOut]
