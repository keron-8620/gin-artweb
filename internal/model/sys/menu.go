package sys

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
	Apis      []ApiModel  `gorm:"many2many:sys_menu_api;joinForeignKey:menu_id;joinReferences:api_id;constraint:OnDelete:CASCADE"`
}

func (m *MenuModel) TableName() string {
	return "sys_menu"
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
	if err := enc.AddObject("meta", &m.Meta); err != nil {
		return err
	}
	enc.AddString("name", m.Name)
	enc.AddUint32("sort", m.Sort)
	enc.AddBool("is_active", m.IsActive)
	enc.AddString("descr", m.Descr)
	if m.ParentID != nil {
		enc.AddUint32("parent_id", *m.ParentID)
	}
	if err := enc.AddArray("apis", zapcore.ArrayMarshalerFunc(func(ae zapcore.ArrayEncoder) error {
		for _, api := range m.Apis {
			ae.AppendUint32(api.ID)
		}
		return nil
	})); err != nil {
		return err
	}
	return nil
}

func (m *MenuModel) ToUpdateMap() map[string]any {
	return map[string]any{
		"path":      m.Path,
		"component": m.Component,
		"name":      m.Name,
		"meta":      m.Meta.Json(),
		"sort":      m.Sort,
		"is_active": m.IsActive,
		"descr":     m.Descr,
		"parent_id": m.ParentID,
	}
}

func ListMenuModelToUint32s(ms []MenuModel) []uint32 {
	if len(ms) == 0 {
		return []uint32{}
	}

	ids := make([]uint32, len(ms))
	for i, m := range ms {
		ids[i] = m.ID
	}
	return ids
}

// CreateMenuDTO 用于创建菜单的请求结构体
//
// swagger:model CreateMenuDTO
type CreateMenuDTO struct {
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

func (dto *CreateMenuDTO) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddUint32("id", dto.ID)
	enc.AddString("path", dto.Path)
	enc.AddString("component", dto.Component)
	if err := enc.AddObject("meta", &dto.Meta); err != nil {
		return err
	}
	enc.AddString("name", dto.Name)
	enc.AddUint32("sort", dto.Sort)
	enc.AddBool("is_active", dto.IsActive)
	enc.AddString("descr", dto.Descr)
	if dto.ParentID != nil {
		enc.AddUint32("parent_id", *dto.ParentID)
	}
	if err := enc.AddArray("api_ids", zapcore.ArrayMarshalerFunc(func(ae zapcore.ArrayEncoder) error {
		for _, id := range dto.ApiIDs {
			ae.AppendUint32(id)
		}
		return nil
	})); err != nil {
		return err
	}
	return nil
}

func (dto *CreateMenuDTO) ToModel() MenuModel {
	m := MenuModel{
		StandardModel: database.StandardModel{
			BaseModel: database.BaseModel{ID: dto.ID},
		},
		Path:      dto.Path,
		Component: dto.Component,
		Name:      dto.Name,
		Meta: MetaSchemas{
			Icon:  dto.Meta.Icon,
			Title: dto.Meta.Title,
		},
		Sort:     dto.Sort,
		IsActive: dto.IsActive,
		Descr:    dto.Descr,
		ParentID: dto.ParentID,
	}
	if m.ParentID != nil && *m.ParentID == 0 {
		m.ParentID = nil
	}
	return m
}

// UpdateMenuDTO 用于更新菜单的请求结构体
//
// swagger:model UpdateMenuDTO
type UpdateMenuDTO struct {
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

func (dto *UpdateMenuDTO) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("path", dto.Path)
	enc.AddString("component", dto.Component)
	if err := enc.AddObject("meta", &dto.Meta); err != nil {
		return err
	}
	enc.AddString("name", dto.Name)
	enc.AddUint32("sort", dto.Sort)
	enc.AddBool("is_active", dto.IsActive)
	enc.AddString("descr", dto.Descr)
	if dto.ParentID != nil {
		enc.AddUint32("parent_id", *dto.ParentID)
	}
	if err := enc.AddArray("api_ids", zapcore.ArrayMarshalerFunc(func(ae zapcore.ArrayEncoder) error {
		for _, id := range dto.ApiIDs {
			ae.AppendUint32(id)
		}
		return nil
	})); err != nil {
		return err
	}
	return nil
}

func (dto *UpdateMenuDTO) ToUpdateMap() map[string]any {
	data := map[string]any{
		"path":      dto.Path,
		"component": dto.Component,
		"name":      dto.Name,
		"meta":      dto.Meta.Json(),
		"sort":      dto.Sort,
		"is_active": dto.IsActive,
		"descr":     dto.Descr,
	}
	if dto.ParentID != nil && *dto.ParentID != 0 {
		data["parent_id"] = dto.ParentID
	}
	return data
}

// ListMenuDTO 用于获取菜单列表的请求结构体
// 支持分页查询和多种筛选条件
//
// swagger:model ListMenuDTO
type ListMenuDTO struct {
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

func (dto *ListMenuDTO) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if dto == nil {
		return nil
	}
	if err := dto.StandardModelQuery.MarshalLogObject(enc); err != nil {
		return err
	}
	enc.AddString("path", dto.Path)
	enc.AddString("component", dto.Component)
	enc.AddString("name", dto.Name)
	if dto.IsActive != nil {
		enc.AddBool("is_active", *dto.IsActive)
	}
	enc.AddString("descr", dto.Descr)
	if dto.ParentID != nil {
		enc.AddUint32("parent_id", *dto.ParentID)
	}
	return nil
}

func (dto *ListMenuDTO) ToQueryMap() map[string]any {
	queryMap := dto.StandardModelQuery.ToQueryMap(12)
	if dto.Path != "" {
		queryMap["path like ?"] = "%" + dto.Path + "%"
	}
	if dto.Component != "" {
		queryMap["component like ?"] = "%" + dto.Component + "%"
	}
	if dto.Name != "" {
		queryMap["name like ?"] = "%" + dto.Name + "%"
	}
	if dto.IsActive != nil {
		queryMap["is_active = ?"] = *dto.IsActive
	}
	if dto.Descr != "" {
		queryMap["descr like ?"] = "%" + dto.Descr + "%"
	}
	if dto.ParentID != nil {
		queryMap["parent_id = ?"] = *dto.ParentID
	}
	return queryMap
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

// MenuResp 菜单响应结构
type MenuResp = common.APIResp[*MenuDetailOut]

// PagMenuResp 菜单的分页响应结构
type PagMenuResp = common.APIResp[*common.Pag[MenuStandardOut]]

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
	ms []MenuModel,
) []MenuStandardOut {
	if len(ms) == 0 {
		return []MenuStandardOut{}
	}
	mso := make([]MenuStandardOut, 0, len(ms))
	for _, m := range ms {
		mo := MenuModelToStandardOut(m)
		mso = append(mso, *mo)
	}
	return mso
}
