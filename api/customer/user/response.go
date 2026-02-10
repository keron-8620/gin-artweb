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
	AccessToken string `json:"access_token"`
	// 刷新令牌
	RefreshToken string `json:"refresh_token"`
}

type LoginReply = common.APIReply[LoginOut]

// LoginRecordStandardOut登陆记录信息
type LoginRecordStandardOut struct {
	// 唯一标识
	ID uint32 `json:"id" example:"1"`

	// 名称
	Username string `json:"username" example:"judgement"`

	// 登录时间
	LoginAt string `json:"login_at" example:"2023-01-01 12:00:00"`

	// IP地址
	IPAddress string `json:"ip_address" example:"192.168.1.1"`

	// 用户浏览器信息
	UserAgent string `json:"user_agent" example:"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36"`

	// 登录状态
	Status bool `json:"is_active" example:"true"`
}

// PagUserReply 用户的分页响应结构
type PagLoginRecordReply = common.APIReply[*common.Pag[LoginRecordStandardOut]]
