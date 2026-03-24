package sys

import (
	"time"

	"go.uber.org/zap/zapcore"

	"gin-artweb/internal/model/common"
	"gin-artweb/internal/shared/database"
)

type UserModel struct {
	database.StandardModel
	Username string    `gorm:"column:username;type:varchar(50);not null;uniqueIndex;comment:用户名" json:"username"`
	Password string    `gorm:"column:password;type:varchar(150);not null;comment:密码" json:"password"`
	IsActive bool      `gorm:"column:is_active;type:boolean;comment:是否激活" json:"is_active"`
	IsStaff  bool      `gorm:"column:is_staff;type:boolean;comment:是否是工作人员" json:"is_staff"`
	RoleID   uint32    `gorm:"column:role_id;not null;comment:角色ID" json:"role_id"`
	Role     RoleModel `gorm:"foreignKey:RoleID;references:ID;constraint:OnDelete:CASCADE" json:"role"`
}

func (m *UserModel) TableName() string {
	return "sys_user"
}

func (m *UserModel) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if m == nil {
		return nil
	}
	if err := m.StandardModel.MarshalLogObject(enc); err != nil {
		return err
	}
	enc.AddString("username", m.Username)
	enc.AddBool("is_active", m.IsActive)
	enc.AddBool("is_staff", m.IsStaff)
	enc.AddUint32("role_id", m.RoleID)
	return nil
}

// CreateUserDTO 用于创建用户的请求结构体
//
// swagger:model CreateUserDTO
type CreateUserDTO struct {
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

func (dto *CreateUserDTO) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("username", dto.Username)
	enc.AddBool("is_active", dto.IsActive)
	enc.AddBool("is_staff", dto.IsStaff)
	enc.AddUint32("role_id", dto.RoleID)
	return nil
}

// UpdateUserDTO 用于更新用户的请求结构体
// 包含用户路径、组件、名称等信息
//
// swagger:model UpdateUserDTO
type UpdateUserDTO struct {
	// 用户名
	Username string `json:"username" form:"username" binding:"required,max=50"`

	// 是否激活
	IsActive bool `json:"is_active" form:"is_active"`

	// 是否是工作人员
	IsStaff bool `json:"is_staff" form:"is_staff"`

	// 角色ID
	RoleID uint32 `json:"role_id" form:"role_id" binding:"required"`
}

func (dto *UpdateUserDTO) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("username", dto.Username)
	enc.AddBool("is_active", dto.IsActive)
	enc.AddBool("is_staff", dto.IsStaff)
	enc.AddUint32("role_id", dto.RoleID)
	return nil
}

func (dto *UpdateUserDTO) ToUpdateMap() map[string]any {
	return map[string]any{
		"username":  dto.Username,
		"is_active": dto.IsActive,
		"is_staff":  dto.IsStaff,
		"role_id":   dto.RoleID,
	}
}

// ListUserDTO 用于获取用户列表的请求结构体
// 支持分页查询和多种筛选条件
//
// swagger:model ListUserDTO
type ListUserDTO struct {
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

func (dto *ListUserDTO) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if dto == nil {
		return nil
	}
	if err := dto.StandardModelQuery.MarshalLogObject(enc); err != nil {
		return err
	}
	enc.AddString("username", dto.Username)
	if dto.IsActive != nil {
		enc.AddBool("is_active", *dto.IsActive)
	}
	if dto.IsStaff != nil {
		enc.AddBool("is_staff", *dto.IsStaff)
	}
	enc.AddUint32("role_id", dto.RoleID)
	return nil
}

func (dto *ListUserDTO) ToQueryMap() map[string]any {
	queryMap := dto.StandardModelQuery.ToQueryMap(10)
	if dto.Username != "" {
		queryMap["username = ?"] = dto.Username
	}
	if dto.IsActive != nil {
		queryMap["is_active = ?"] = *dto.IsActive
	}
	if dto.IsStaff != nil {
		queryMap["is_staff = ?"] = *dto.IsStaff
	}
	if dto.RoleID != 0 {
		queryMap["role_id = ?"] = dto.RoleID
	}
	return queryMap
}

// ResetPasswordDTO 重置用户的密码
//
// swagger:model ResetPasswordDTO
type ResetPasswordDTO struct {
	// 新密码
	NewPassword string `json:"new_password" form:"new_password" binding:"required,max=20"`

	// 确认密码
	ConfirmPassword string `json:"confirm_password" form:"confirm_password" binding:"required,eqfield=NewPassword"`
}

// PatchPasswordDTO 修改密码
//
// swagger:model PatchPasswordDTO
type PatchPasswordDTO struct {
	// 原密码
	OldPassword string `json:"old_password" form:"old_password" binding:"required"`

	// 新密码
	NewPassword string `json:"new_password" form:"new_password" binding:"required,max=20"`

	// 确认密码
	ConfirmPassword string `json:"confirm_password" form:"confirm_password" binding:"required,eqfield=NewPassword"`
}

type LoginDTO struct {
	// 用户名
	Username string `json:"username" form:"username" binding:"required,max=50"`

	// 密码
	Password string `json:"password" form:"password" binding:"required,max=20"`
}

type RequestContext struct {
	// IP地址
	IP string `json:"ip"`

	// 客户端用户代理（浏览器/设备信息）
	UserAgent string `json:"user_agent"`
}

type RefreshTokenDTO struct {
	// 刷新令牌
	RefreshToken string `json:"refresh_token" form:"refresh_token" binding:"required"`
}

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
	Role *RoleBaseOut `json:"role"`
}

// UserBaseResp 用户响应结构
type UserResp = common.APIResp[*UserDetailOut]

// PagUserResp 用户的分页响应结构
type PagUserResp = common.APIResp[*common.Pag[UserDetailOut]]

type LoginOut struct {
	// 登录令牌
	AccessToken string `json:"access_token"`
	// 刷新令牌
	RefreshToken string `json:"refresh_token"`
}

type LoginResp = common.APIResp[LoginOut]

func UserModelToBaseOut(
	m UserModel,
) *UserBaseOut {
	return &UserBaseOut{
		ID:       m.ID,
		Username: m.Username,
		IsActive: m.IsActive,
		IsStaff:  m.IsStaff,
	}
}

func UserModelToStandardOut(
	m UserModel,
) *UserStandardOut {
	return &UserStandardOut{
		UserBaseOut: *UserModelToBaseOut(m),
		CreatedAt:   m.CreatedAt.Format(time.DateTime),
		UpdatedAt:   m.UpdatedAt.Format(time.DateTime),
	}
}

func UserModelToDetailOut(
	m UserModel,
) *UserDetailOut {
	var role *RoleBaseOut
	if m.Role.ID != 0 {
		role = RoleModelToBaseOut(m.Role)
	}
	return &UserDetailOut{
		UserStandardOut: *UserModelToStandardOut(m),
		Role:            role,
	}
}

func ListUserModelToDetailOut(
	ums []UserModel,
) []UserDetailOut {
	if len(ums) == 0 {
		return []UserDetailOut{}
	}
	mso := make([]UserDetailOut, 0, len(ums))
	if len(ums) > 0 {
		for _, m := range ums {
			mo := UserModelToDetailOut(m)
			mso = append(mso, *mo)
		}
	}
	return mso
}

func LoginRecordModelToStandardOut(
	m LoginRecordModel,
) *LoginRecordStandardOut {
	return &LoginRecordStandardOut{
		ID:        m.ID,
		Username:  m.Username,
		LoginAt:   m.LoginAt.Format(time.DateTime),
		Status:    m.Status,
		IPAddress: m.IPAddress,
		UserAgent: m.UserAgent,
	}
}

func ListLoginRecordModelToStandardOut(
	ms []LoginRecordModel,
) []LoginRecordStandardOut {
	if len(ms) == 0 {
		return []LoginRecordStandardOut{}
	}
	mso := make([]LoginRecordStandardOut, 0, len(ms))
	if len(ms) > 0 {
		for _, m := range ms {
			mo := LoginRecordModelToStandardOut(m)
			mso = append(mso, *mo)
		}
	}
	return mso
}
