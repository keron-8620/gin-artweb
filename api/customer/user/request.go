package user

import (
	"time"

	"go.uber.org/zap/zapcore"

	"gin-artweb/api/common"
)

// CreateUserRequest 用于创建用户的请求结构体
//
// swagger:model CreateUserRequest
type CreateUserRequest struct {
	// 用户名
	Username string `json:"username" form:"username" binding:"required,max=50"`

	// 密码
	Password string `json:"password" form:"password" binding:"required,max=20"`

	// 是否激活
	IsActive bool `json:"is_active" form:"is_active"`

	// 是否是工作人员
	IsStaff bool `json:"is_staff" form:"is_staff"`

	// 角色ID
	RoleID uint32 `json:"role_id" form:"role_id" binding:"required"`
}

func (req *CreateUserRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("username", req.Username)
	enc.AddBool("is_active", req.IsActive)
	enc.AddBool("is_staff", req.IsStaff)
	enc.AddUint32("role_id", req.RoleID)
	return nil
}

// UpdateUserRequest 用于更新用户的请求结构体
// 包含用户路径、组件、名称等信息
//
// swagger:model UpdateUserRequest
type UpdateUserRequest struct {
	// 用户名
	Username string `json:"username" form:"username" binding:"required,max=50"`

	// 是否激活
	IsActive bool `json:"is_active" form:"is_active"`

	// 是否是工作人员
	IsStaff bool `json:"is_staff" form:"is_staff"`

	// 角色ID
	RoleID uint32 `json:"role_id" form:"role_id" binding:"required"`
}

func (req *UpdateUserRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("username", req.Username)
	enc.AddBool("is_active", req.IsActive)
	enc.AddBool("is_staff", req.IsStaff)
	enc.AddUint32("role_id", req.RoleID)
	return nil
}

// ListUserRequest 用于获取用户列表的请求结构体
// 支持分页查询和多种筛选条件
//
// swagger:model ListUserRequest
type ListUserRequest struct {
	common.StandardModelQuery

	// 用户名
	Username string `form:"username" binding:"omitempty,max=50"`

	// 是否激活
	IsActive *bool `form:"is_active" binding:"omitempty"`

	// 是否是工作人员
	IsStaff *bool `form:"is_staff" binding:"omitempty"`

	// 角色ID
	RoleID uint32 `form:"role_id" binding:"omitempty"`
}

func (req *ListUserRequest) Query() (int, int, map[string]any) {
	page, size, query := req.StandardModelQuery.QueryMap(10)
	if req.Username != "" {
		query["username = ?"] = req.Username
	}
	if req.IsActive != nil {
		query["is_active = ?"] = *req.IsActive
	}
	if req.IsStaff != nil {
		query["is_staff = ?"] = *req.IsStaff
	}
	if req.RoleID != 0 {
		query["role_id = ?"] = req.RoleID
	}
	return page, size, query
}

// ResetPasswordRequest 重置用户的密码
//
// swagger:model ResetPasswordRequest
type ResetPasswordRequest struct {
	// 新密码
	NewPassword string `json:"new_password" form:"new_password" binding:"required,max=20"`

	// 确认密码
	ConfirmPassword string `json:"confirm_password" form:"confirm_password" binding:"required,eqfield=NewPassword"`
}

// PatchPasswordRequest 修改密码
//
// swagger:model PatchPasswordRequest
type PatchPasswordRequest struct {
	// 原密码
	OldPassword string `json:"old_password" form:"old_password" binding:"required"`

	// 新密码
	NewPassword string `json:"new_password" form:"new_password" binding:"required,max=20"`

	// 确认密码
	ConfirmPassword string `json:"confirm_password" form:"confirm_password" binding:"required,eqfield=NewPassword"`
}

type LoginRequest struct {
	// 用户名
	Username string `json:"username" form:"username" binding:"required,max=50"`

	// 密码
	Password string `json:"password" form:"password" binding:"required,max=20"`
}

type RefreshTokenRequest struct {
	// 刷新令牌
	RefreshToken string `json:"refresh_token" form:"refresh_token" binding:"required"`
}

// ListLoginRecordRequest 用于获取用户登陆记录的请求结构体
// 支持分页查询和多种筛选条件
//
// swagger:model ListUserRequest
type ListLoginRecordRequest struct {
	common.BaseModelQuery

	// 用户名
	Username string `form:"name" binding:"omitempty,max=50"`

	// IP 地址
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
