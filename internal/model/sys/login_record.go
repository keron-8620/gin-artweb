package sys

import (
	"time"

	"go.uber.org/zap/zapcore"

	"gin-artweb/internal/model/common"
	"gin-artweb/internal/shared/database"
)

type LoginRecordModel struct {
	database.BaseModel
	Username  string    `gorm:"column:username;type:varchar(50);comment:用户名" json:"username"`
	LoginAt   time.Time `gorm:"column:login_at;autoCreateTime;comment:登录时间" json:"login_at"`
	IPAddress string    `gorm:"column:ip_address;type:varchar(108);comment:ip地址" json:"ip_address"`
	UserAgent string    `gorm:"column:user_agent;type:varchar(254);comment:客户端信息" json:"user_agent"`
	Status    bool      `gorm:"column:status;type:boolean;comment:是否登录成功" json:"status"`
}

func (m *LoginRecordModel) TableName() string {
	return "sys_login_record"
}

func (m *LoginRecordModel) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if m == nil {
		return nil
	}
	if err := m.BaseModel.MarshalLogObject(enc); err != nil {
		return err
	}
	enc.AddString("username", m.Username)
	enc.AddTime("login_at", m.LoginAt)
	enc.AddString("ip_address", m.IPAddress)
	enc.AddString("user_agent", m.UserAgent)
	enc.AddBool("status", m.Status)
	return nil
}

func ListLoginRecordModelToUint32s(ms []LoginRecordModel) []uint32 {
	if len(ms) == 0 {
		return []uint32{}
	}

	ids := make([]uint32, len(ms))
	for i, m := range ms {
		ids[i] = m.ID
	}
	return ids
}

// ListLoginRecordDTO 用于获取用户登陆记录的请求结构体
// 支持分页查询和多种筛选条件
//
// swagger:model ListLoginRecordDTO
type ListLoginRecordDTO struct {
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

func (dto *ListLoginRecordDTO) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if dto == nil {
		return nil
	}
	if err := dto.BaseModelQuery.MarshalLogObject(enc); err != nil {
		return err
	}
	enc.AddString("username", dto.Username)
	enc.AddString("ip_address", dto.IPAddress)
	if dto.Status != nil {
		enc.AddBool("status", *dto.Status)
	}
	enc.AddString("before_login_at", dto.BeforeLoginAt)
	enc.AddString("after_login_at", dto.AfterLoginAt)
	return nil
}

func (dto *ListLoginRecordDTO) ToQueryMap() map[string]any {
	queryMap := dto.BaseModelQuery.ToQueryMap(9)
	if dto.Username != "" {
		queryMap["username = ?"] = dto.Username
	}
	if dto.IPAddress != "" {
		queryMap["ip_address = ?"] = dto.IPAddress
	}
	if dto.Status != nil {
		queryMap["status = ?"] = dto.Status
	}
	if dto.BeforeLoginAt != "" {
		bft, err := time.Parse(time.RFC3339, dto.BeforeLoginAt)
		if err == nil {
			queryMap["login_at < ?"] = bft
		}
	}
	if dto.AfterLoginAt != "" {
		act, err := time.Parse(time.RFC3339, dto.AfterLoginAt)
		if err == nil {
			queryMap["login_at > ?"] = act
		}
	}
	return queryMap
}

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

// PagLoginRecordResp 用户登陆记录的分页响应结构
type PagLoginRecordResp = common.APIResp[*common.Pag[LoginRecordStandardOut]]
