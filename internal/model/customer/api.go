package customer

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
	return "customer_api"
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

// CreateApiRequest 用于创建权限的请求结构体
//
// swagger:model CreateApiRequest
type CreateApiRequest struct {
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

func (req *CreateApiRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddUint32("id", req.ID)
	enc.AddString("url", req.URL)
	enc.AddString("method", req.Method)
	enc.AddString("label", req.Label)
	enc.AddString("descr", req.Descr)
	return nil
}

// UpdateApiRequest 用于更新权限的请求结构体
// 包含权限主键、HTTP URL、请求方法和描述信息
//
// swagger:model UpdateApiRequest
type UpdateApiRequest struct {
	// URL地址
	URL string `json:"url" form:"url" binding:"required,max=150"`

	// 请求方法
	Method string `json:"method" form:"method" binding:"required,oneof=GET POST PUT DELETE PATCH WS"`

	// 标签
	Label string `json:"label" form:"label" binding:"required,max=50"`

	// 描述信息
	Descr string `json:"descr" form:"descr" binding:"omitempty,max=254"`
}

func (req *UpdateApiRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("url", req.URL)
	enc.AddString("method", req.Method)
	enc.AddString("label", req.Label)
	enc.AddString("descr", req.Descr)
	return nil
}

// ListApiRequest 用于获取权限列表的请求结构体
// 支持分页查询和多种筛选条件
//
// swagger:model ListApiRequest
type ListApiRequest struct {
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

func (req *ListApiRequest) Query() (int, int, map[string]any) {
	page, size, query := req.StandardModelQuery.QueryMap(10)
	if req.URL != "" {
		query["url like ?"] = "%" + req.URL + "%"
	}
	if req.Method != "" {
		query["method = ?"] = req.Method
	}
	if req.Label != "" {
		query["label like ?"] = "%" + req.Label + "%"
	}
	if req.Descr != "" {
		query["descr like ?"] = "%" + req.Descr + "%"
	}
	return page, size, query
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

// ApiReply 权限响应结构
type ApiReply = common.APIReply[*ApiStandardOut]

// PagApiReply API的分页响应结构
type PagApiReply = common.APIReply[*common.Pag[ApiStandardOut]]

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

func ListApiModelToStandardOut(pms *[]ApiModel) *[]ApiStandardOut {
	if pms == nil {
		return &[]ApiStandardOut{}
	}

	ms := *pms
	mso := make([]ApiStandardOut, 0, len(ms))
	for _, m := range ms {
		mo := ApiModelToStandardOut(m)
		mso = append(mso, *mo)
	}
	return &mso
}
