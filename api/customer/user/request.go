package user

import (
	"gin-artweb/api/common"

	"go.uber.org/zap/zapcore"
)

// CreateUserRequest 用于创建用户的请求结构体
//
// swagger:model CreateUserRequest
type CreateUserRequest struct {
	// 用户名，最大长度50
	// Required: true
	// Max length: 50
	Username string `json:"username" binding:"required,max=50"`

	// 密码，最大长度20
	// Required: true
	// Max length: 20
	Password string `json:"password" binding:"required, max=20"`

	// 是否激活，必填
	IsActive bool `json:"is_active" binding:"required"`

	// 是否是工作人员，必填
	IsStaff bool `json:"is_staff" binding:"required"`

	// 角色ID，必填
	RoleID uint32 `json:"role_id" binding:"required"`
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
	// 用户名，最大长度50
	// Required: true
	// Max length: 50
	Username string `json:"name" binding:"required,max=50"`

	// 是否激活，必填
	IsActive bool `json:"is_active" binding:"required"`

	// 是否是工作人员，必填
	IsStaff bool `json:"is_staff" binding:"required"`

	// 角色ID，必填
	RoleID uint32 `json:"role_id" binding:"required"`
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

	// 用户名称，字符串长度限制
	// Max length: 50
	Username string `form:"name" binding:"omitempty,max=50"`

	// 是否激活
	IsActive *bool `json:"is_active" binding:"omitempty"`

	// 是否是工作人员
	IsStaff *bool `json:"is_staff" binding:"omitempty"`

	// 角色ID
	RoleID uint32 `json:"role_id" binding:"omitempty"`
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
	// 新密码，最大长度20
	// Required: true
	// Max length: 20
	NewPassword string `json:"new_password" binding:"required,max=20"`

	// 确认密码，最大长度20
	// Required: true
	// Max length: 20
	ConfirmPassword string `json:"confirm_password" binding:"required,max=20"`
}

// PatchPasswordRequest 修改密码
//
// swagger:model PatchPasswordRequest
type PatchPasswordRequest struct {
	// 原密码，必填,
	// Required: true
	// Max length: 20
	OldPassword string `json:"old_password" binding:"required"`

	// 新密码，最大长度20
	// Required: true
	// Max length: 20
	NewPassword string `json:"new_password" binding:"required,max=20"`

	// 确认密码，最大长度20
	// Required: true
	// Max length: 20
	ConfirmPassword string `json:"confirm_password" binding:"required,max=20"`
}

type LoginRequest struct {
	// 用户名，必填,
	// Required: true
	// Max length: 50
	Username string `json:"username" binding:"required,max=50"`

	// 密码，最大长度20
	// Required: true
	// Max length: 20
	Password string `json:"password" binding:"required,max=20"`
}
