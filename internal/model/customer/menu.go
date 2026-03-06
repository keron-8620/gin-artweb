package customer

import (
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap/zapcore"

	"gin-artweb/internal/model/common"
	"gin-artweb/internal/shared/database"
)

type MetaSchemas struct {
	// 标题
	Title string `json:"title" example:"用户管理"`
	// 图标
	Icon string `json:"icon" example:"icon"`
}

func (m *MetaSchemas) Json() string {
	jd, _ := json.Marshal(m)
	return string(jd)
}

func (m *MetaSchemas) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("title", m.Title)
	enc.AddString("icon", m.Icon)
	return nil
}

func NewMetaSchemas(ms string) (*MetaSchemas, error) {
	if ms == "" {
		return nil, fmt.Errorf("meta is empty")
	}
	var meta MetaSchemas
	err := json.Unmarshal([]byte(ms), &meta)
	if err != nil {
		return nil, fmt.Errorf("解析 MetaSchemas 失败: %w", err)
	}
	return &meta, nil
}

type MenuModel struct {
	database.StandardModel
	Path      string      `gorm:"column:path;type:varchar(100);not null;uniqueIndex;comment:前端路由" json:"path"`
	Component string      `gorm:"column:component;type:varchar(200);not null;comment:前端组件" json:"component"`
	Name      string      `gorm:"column:name;type:varchar(50);not null;uniqueIndex;comment:名称" json:"name"`
	Meta      MetaSchemas `gorm:"column:meta;serializer:json;comment:菜单信息" json:"meta"`
	Sort      uint32      `gorm:"column:sort;type:integer;comment:排序" json:"sort"`
	IsActive  bool        `gorm:"column:is_active;type:boolean;comment:是否激活" json:"is_active"`
	Descr     string      `gorm:"column:descr;type:varchar(254);comment:描述" json:"descr"`
	ParentID  *uint32     `gorm:"column:parent_id;comment:父菜单ID" json:"parent_id"`
	Parent    *MenuModel  `gorm:"foreignKey:ParentID;references:ID;constraint:OnDelete:CASCADE" json:"parent"`
	Apis      []ApiModel  `gorm:"many2many:customer_menu_api;joinForeignKey:menu_id;joinReferences:api_id;constraint:OnDelete:CASCADE"`
}

func (m *MenuModel) TableName() string {
	return "customer_menu"
}

func (m *MenuModel) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if m == nil {
		return nil
	}
	if err := m.StandardModel.MarshalLogObject(enc); err != nil {
		return err
	}
	enc.AddString("path", m.Path)
	enc.AddString("component", m.Component)
	enc.AddObject("meta", &m.Meta)
	enc.AddString("name", m.Name)
	enc.AddUint32("sort", m.Sort)
	enc.AddBool("is_active", m.IsActive)
	enc.AddString("descr", m.Descr)
	if m.ParentID != nil {
		enc.AddUint32("parent_id", *m.ParentID)
	}
	enc.AddArray("apis", zapcore.ArrayMarshalerFunc(func(ae zapcore.ArrayEncoder) error {
		for _, api := range m.Apis {
			ae.AppendUint32(api.ID)
		}
		return nil
	}))
	return nil
}

// CreateMenuRequest 用于创建菜单的请求结构体
//
// swagger:model CreateMenuRequest
type CreateMenuRequest struct {
	// 唯一标识
	ID uint32 `json:"id" form:"id" binding:"required,gt=0"`

	// 前端路由路径
	Path string `json:"path" form:"path" binding:"required,max=100"`

	// 组件路径
	Component string `json:"component" form:"component" binding:"required,max=200"`

	// 名称
	Name string `json:"name" form:"name" binding:"required,max=50"`

	// 菜单元信息
	Meta MetaSchemas `json:"meta" form:"meta" binding:"required"`

	// 排序字段
	Sort uint32 `json:"sort" form:"sort" binding:"required"`

	// 是否激活
	IsActive bool `json:"is_active" form:"is_active"`

	// 描述
	Descr string `json:"descr" form:"descr" binding:"omitempty,max=254"`

	// 父级菜单ID
	ParentID *uint32 `json:"parent_id" form:"parent_id" binding:"omitempty"`

	// 权限ID列表
	ApiIDs []uint32 `json:"api_ids" form:"api_ids" binding:"omitempty"`
}

func (req *CreateMenuRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddUint32("id", req.ID)
	enc.AddString("path", req.Path)
	enc.AddString("component", req.Component)
	enc.AddObject("meta", &req.Meta)
	enc.AddString("name", req.Name)
	enc.AddUint32("sort", req.Sort)
	enc.AddBool("is_active", req.IsActive)
	enc.AddString("descr", req.Descr)
	if req.ParentID != nil {
		enc.AddUint32("parent_id", *req.ParentID)
	}
	enc.AddArray("api_ids", zapcore.ArrayMarshalerFunc(func(ae zapcore.ArrayEncoder) error {
		for _, id := range req.ApiIDs {
			ae.AppendUint32(id)
		}
		return nil
	}))
	return nil
}

// UpdateMenuRequest 用于更新菜单的请求结构体
//
// swagger:model UpdateMenuRequest
type UpdateMenuRequest struct {
	// 前端路由路径
	Path string `json:"path" form:"path" binding:"required,max=100"`

	// 组件路径
	Component string `json:"component" form:"component" binding:"required,max=200"`

	// 名称
	Name string `json:"name" form:"name" binding:"required,max=50"`

	// 菜单元信息
	Meta MetaSchemas `json:"meta" form:"meta" binding:"required"`

	// 排序字段
	Sort uint32 `json:"sort" form:"sort" binding:"required"`

	// 是否激活
	IsActive bool `json:"is_active" form:"is_active"`

	// 描述信息
	Descr string `json:"descr" form:"descr" binding:"omitempty,max=254"`

	// 父级菜单ID
	ParentID *uint32 `json:"parent_id" form:"parent_id" binding:"omitempty"`

	// API ID列表
	ApiIDs []uint32 `json:"api_ids" form:"api_ids" binding:"omitempty"`
}

func (req *UpdateMenuRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("path", req.Path)
	enc.AddString("component", req.Component)
	enc.AddObject("meta", &req.Meta)
	enc.AddString("name", req.Name)
	enc.AddUint32("sort", req.Sort)
	enc.AddBool("is_active", req.IsActive)
	enc.AddString("descr", req.Descr)
	if req.ParentID != nil {
		enc.AddUint32("parent_id", *req.ParentID)
	}
	enc.AddArray("api_ids", zapcore.ArrayMarshalerFunc(func(ae zapcore.ArrayEncoder) error {
		for _, id := range req.ApiIDs {
			ae.AppendUint32(id)
		}
		return nil
	}))
	return nil
}

// ListMenuRequest 用于获取菜单列表的请求结构体
// 支持分页查询和多种筛选条件
//
// swagger:model ListMenuRequest
type ListMenuRequest struct {
	common.StandardModelQuery

	// 前端路由路径
	Path string `form:"path" binding:"omitempty,max=100"`

	// 组件路径
	Component string `form:"component" binding:"omitempty,max=200"`

	// 名称
	Name string `form:"name" binding:"omitempty,max=50"`

	// 是否激活
	IsActive *bool `form:"is_active" binding:"omitempty"`

	// 菜单描述
	Descr string `form:"descr" binding:"omitempty,max=254"`

	// 父级菜单ID
	ParentID *uint32 `form:"parent_id" binding:"omitempty"`
}

func (req *ListMenuRequest) Query() (int, int, map[string]any) {
	page, size, query := req.StandardModelQuery.QueryMap(12)
	if req.Path != "" {
		query["path like ?"] = "%" + req.Path + "%"
	}
	if req.Component != "" {
		query["component like ?"] = "%" + req.Component + "%"
	}
	if req.Name != "" {
		query["name like ?"] = "%" + req.Name + "%"
	}
	if req.IsActive != nil {
		query["is_active = ?"] = *req.IsActive
	}
	if req.Descr != "" {
		query["descr like ?"] = "%" + req.Descr + "%"
	}
	if req.ParentID != nil {
		query["parent_id = ?"] = *req.ParentID
	}
	return page, size, query
}

// MenuStandardOut 菜单基础输出结构体
type MenuBaseOut struct {
	// 唯一标识
	ID uint32 `json:"id" example:"1"`

	// 前端路由
	Path string `json:"path" example:"/api/v1/users"`

	// 组件路径
	Component string `json:"component" example:"GET"`

	// 名称
	Name string `json:"name" example:"用户管理"`

	//菜单信息
	Meta MetaSchemas `json:"meta"`

	// 排序字段
	Sort uint32 `json:"sort" example:"1000"`

	// 是否激活
	IsActive bool `json:"is_active" example:"true"`

	// 描述
	Descr string `json:"descr" example:"用户管理"`
}

// MenuStandardOut 菜单标准输出结构体
type MenuStandardOut struct {
	MenuBaseOut

	// 创建时间
	CreatedAt string `json:"created_at" example:"2023-01-01 12:00:00"`

	// 更新时间
	UpdatedAt string `json:"updated_at" example:"2023-01-01 12:00:00"`
}

// MenuDetailOut 菜单详情输出结构体
type MenuDetailOut struct {
	MenuStandardOut

	// 父级菜单
	Parent *MenuStandardOut `json:"parent"`

	// API ID列表
	ApiIDs []uint32 `json:"api_ids"`
}

// MenuReply 菜单响应结构
type MenuReply = common.APIReply[*MenuDetailOut]

// PagMenuReply 菜单的分页响应结构
type PagMenuReply = common.APIReply[*common.Pag[MenuStandardOut]]

func MenuModelToBaseOut(
	m MenuModel,
) *MenuBaseOut {
	return &MenuBaseOut{
		ID:        m.ID,
		Path:      m.Path,
		Component: m.Component,
		Name:      m.Name,
		Meta:      m.Meta,
		Sort:      m.Sort,
		IsActive:  m.IsActive,
		Descr:     m.Descr,
	}
}

func MenuModelToStandardOut(
	m MenuModel,
) *MenuStandardOut {
	return &MenuStandardOut{
		MenuBaseOut: *MenuModelToBaseOut(m),
		CreatedAt:   m.CreatedAt.Format(time.DateTime),
		UpdatedAt:   m.UpdatedAt.Format(time.DateTime),
	}
}

func MenuModelToDetailOut(
	m MenuModel,
) *MenuDetailOut {
	var parent *MenuStandardOut
	if m.Parent != nil {
		parent = MenuModelToStandardOut(*m.Parent)
	}
	var apiIDs = []uint32{}
	if len(m.Apis) > 0 {
		apiIDs = make([]uint32, len(m.Apis))
		for i, p := range m.Apis {
			apiIDs[i] = p.ID
		}
	}
	return &MenuDetailOut{
		MenuStandardOut: *MenuModelToStandardOut(m),
		Parent:          parent,
		ApiIDs:          apiIDs,
	}
}

func ListMenuModelToStandardOut(
	mms *[]MenuModel,
) *[]MenuStandardOut {
	if mms == nil {
		return &[]MenuStandardOut{}
	}
	ms := *mms
	mso := make([]MenuStandardOut, 0, len(ms))
	for _, m := range ms {
		mo := MenuModelToStandardOut(m)
		mso = append(mso, *mo)
	}
	return &mso
}
