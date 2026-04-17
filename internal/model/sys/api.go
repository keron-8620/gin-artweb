package sys

import (
	"time"

	"go.uber.org/zap/zapcore"

	"gin-artweb/internal/model/common"
	"gin-artweb/internal/shared/database"
)

type ApiModel struct {
	database.StandardModel
	URL    string `gorm:"column:url;type:varchar(150);not null;uniqueIndex:idx_api_url_method;comment:HTTP的URL地址" json:"url"`
	Method string `gorm:"column:method;type:varchar(10);not null;uniqueIndex:idx_api_url_method;comment:请求方法" json:"method"`
	Label  string `gorm:"column:label;type:varchar(50);not null;index:label;comment:标签" json:"label"`
	Descr  string `gorm:"column:descr;type:varchar(254);comment:描述" json:"descr"`
}

func (m *ApiModel) TableName() string {
	return "sys_api"
}

func (m *ApiModel) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if m == nil {
		return nil
	}
	if err := m.StandardModel.MarshalLogObject(enc); err != nil {
		return err
	}
	enc.AddString("url", m.URL)
	enc.AddString("method", m.Method)
	enc.AddString("label", m.Label)
	enc.AddString("descr", m.Descr)
	return nil
}

func (m *ApiModel) ToUpdateMap() map[string]any {
	return map[string]any{
		"url":    m.URL,
		"method": m.Method,
		"label":  m.Label,
		"descr":  m.Descr,
	}
}

func ListApiModelToUint32s(ms []ApiModel) []uint32 {
	if len(ms) == 0 {
		return []uint32{}
	}

	ids := make([]uint32, len(ms))
	for i, m := range ms {
		ids[i] = m.ID
	}
	return ids
}

// CreateApiDTO 用于创建权限的请求结构体
//
// swagger:model CreateApiDTO
type CreateApiDTO struct {
	// 唯一标识
	ID uint32 `json:"id" form:"id" binding:"required,gt=0"`

	// URL地址
	URL string `json:"url" form:"url" binding:"required,max=150"`

	// 请求方法
	Method string `json:"method" form:"method" binding:"required,oneof=GET POST PUT DELETE PATCH WS"`

	// 标签
	Label string `json:"label" form:"label" binding:"required,max=50"`

	// 描述信息
	Descr string `json:"descr" form:"descr" binding:"max=254"`
}

func (dto *CreateApiDTO) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddUint32("id", dto.ID)
	enc.AddString("url", dto.URL)
	enc.AddString("method", dto.Method)
	enc.AddString("label", dto.Label)
	enc.AddString("descr", dto.Descr)
	return nil
}

func (dto *CreateApiDTO) ToModel() ApiModel {
	return ApiModel{
		StandardModel: database.StandardModel{
			BaseModel: database.BaseModel{ID: dto.ID},
		},
		URL:    dto.URL,
		Method: dto.Method,
		Label:  dto.Label,
		Descr:  dto.Descr,
	}
}

// UpdateApiDTO 用于更新权限的请求结构体
// 包含权限主键、HTTP URL、请求方法和描述信息
//
// swagger:model UpdateApiDTO
type UpdateApiDTO struct {
	// URL地址
	URL string `json:"url" form:"url" binding:"required,max=150"`

	// 请求方法
	Method string `json:"method" form:"method" binding:"required,oneof=GET POST PUT DELETE PATCH WS"`

	// 标签
	Label string `json:"label" form:"label" binding:"required,max=50"`

	// 描述信息
	Descr string `json:"descr" form:"descr" binding:"omitempty,max=254"`
}

func (dto *UpdateApiDTO) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("url", dto.URL)
	enc.AddString("method", dto.Method)
	enc.AddString("label", dto.Label)
	enc.AddString("descr", dto.Descr)
	return nil
}

func (dto *UpdateApiDTO) ToUpdateMap() map[string]any {
	return map[string]any{
		"url":    dto.URL,
		"method": dto.Method,
		"label":  dto.Label,
		"descr":  dto.Descr,
	}
}

// ListApiDTO 用于获取权限列表的请求结构体
// 支持分页查询和多种筛选条件
//
// swagger:model ListApiDTO
type ListApiDTO struct {
	common.StandardModelQuery

	// URL地址
	URL string `form:"url" binding:"omitempty,max=150"`

	// 请求方法
	Method string `form:"method" binding:"omitempty,oneof=GET POST PUT DELETE PATCH WS"`

	// 标签
	Label string `form:"label" binding:"omitempty,max=50"`

	// 描述信息
	Descr string `form:"descr" binding:"omitempty,max=254"`
}

func (dto *ListApiDTO) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if dto == nil {
		return nil
	}
	if err := dto.StandardModelQuery.MarshalLogObject(enc); err != nil {
		return err
	}
	enc.AddString("url", dto.URL)
	enc.AddString("method", dto.Method)
	enc.AddString("label", dto.Label)
	enc.AddString("descr", dto.Descr)
	return nil
}

func (dto *ListApiDTO) ToQueryMap() map[string]any {
	queryMap := dto.StandardModelQuery.ToQueryMap(10)
	if dto.URL != "" {
		queryMap["url like ?"] = "%" + dto.URL + "%"
	}
	if dto.Method != "" {
		queryMap["method = ?"] = dto.Method
	}
	if dto.Label != "" {
		queryMap["label like ?"] = "%" + dto.Label + "%"
	}
	if dto.Descr != "" {
		queryMap["descr like ?"] = "%" + dto.Descr + "%"
	}
	return queryMap
}

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

// ApiResp 权限响应结构
type ApiResp = common.APIResp[*ApiStandardOut]

// PagApiResp API的分页响应结构
type PagApiResp = common.APIResp[*common.Pag[ApiStandardOut]]

func ApiModelToStandardOut(m ApiModel) *ApiStandardOut {
	return &ApiStandardOut{
		ID:        m.ID,
		CreatedAt: m.CreatedAt.Format(time.DateTime),
		UpdatedAt: m.UpdatedAt.Format(time.DateTime),
		URL:       m.URL,
		Method:    m.Method,
		Label:     m.Label,
		Descr:     m.Descr,
	}
}

func ListApiModelToStandardOut(ms []ApiModel) []ApiStandardOut {
	if len(ms) == 0 {
		return []ApiStandardOut{}
	}
	mso := make([]ApiStandardOut, 0, len(ms))
	for _, m := range ms {
		mo := ApiModelToStandardOut(m)
		mso = append(mso, *mo)
	}
	return mso
}
