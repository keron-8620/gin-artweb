package user

import (
	"gin-artweb/api/common"

	"go.uber.org/zap/zapcore"
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
	Username string `form:"name" binding:"omitempty,max=50"`

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
