package user

import (
	"gin-artweb/api/common"
	"gin-artweb/api/customer/role"
)

type UserBaseOut struct {
	// 唯一标识
	ID uint32 `json:"id" example:"1"`

	// 用户名
	Username string `json:"username" example:"judgement"`

	// 是否激活
	IsActive bool `json:"is_active" example:"true"`

	// 是否是工作人员
	IsStaff bool `json:"is_staff" example:"false"`
}

// UserStandardOut用户基础信息
type UserStandardOut struct {
	UserBaseOut

	// 创建时间
	CreatedAt string `json:"created_at" example:"2023-01-01 12:00:00"`

	// 更新时间
	UpdatedAt string `json:"updated_at" example:"2023-01-01 12:00:00"`
}

type UserDetailOut struct {
	UserStandardOut

	// 角色
	Role *role.RoleBaseOut `json:"role"`
}

// UserBaseReply 用户响应结构
type UserReply = common.APIReply[*UserDetailOut]

// PagUserReply 用户的分页响应结构
type PagUserReply = common.APIReply[*common.Pag[UserDetailOut]]

type LoginOut struct {
	// 登录令牌
	Token string `json:"token"`
}

type LoginReply = common.APIReply[LoginOut]
