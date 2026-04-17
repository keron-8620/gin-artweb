package sys

import (
	"time"

	"go.uber.org/zap/zapcore"

	"gin-artweb/internal/model/common"
	"gin-artweb/internal/shared/database"
)

type ButtonModel struct {
	database.StandardModel
	Name     string     `gorm:"column:name;type:varchar(50);not null;uniqueIndex;comment:名称" json:"name"`
	Sort     uint32     `gorm:"column:sort;type:integer;comment:排序" json:"sort"`
	IsActive bool       `gorm:"column:is_active;type:boolean;comment:是否激活" json:"is_active"`
	Descr    string     `gorm:"column:descr;type:varchar(254);comment:描述" json:"descr"`
	MenuID   uint32     `gorm:"column:menu_id;not null;comment:菜单ID" json:"menu_id"`
	Menu     MenuModel  `gorm:"foreignKey:MenuID;references:ID;constraint:OnDelete:CASCADE" json:"menu"`
	Apis     []ApiModel `gorm:"many2many:sys_button_api;joinForeignKey:button_id;joinReferences:api_id;constraint:OnDelete:CASCADE"`
}

func (m *ButtonModel) TableName() string {
	return "sys_button"
}

func (m *ButtonModel) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if m == nil {
		return nil
	}
	if err := m.StandardModel.MarshalLogObject(enc); err != nil {
		return err
	}
	enc.AddString("name", m.Name)
	enc.AddUint32("sort", m.Sort)
	enc.AddBool("is_active", m.IsActive)
	enc.AddString("descr", m.Descr)
	enc.AddUint32("menu_id", m.MenuID)
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

func (m *ButtonModel) ToUpdateMap() map[string]any {
	return map[string]any{
		"name":      m.Name,
		"sort":      m.Sort,
		"is_active": m.IsActive,
		"descr":     m.Descr,
		"menu_id":   m.MenuID,
	}
}

func ListButtonModelToUint32s(ms []ButtonModel) []uint32 {
	if len(ms) == 0 {
		return []uint32{}
	}

	ids := make([]uint32, len(ms))
	for i, m := range ms {
		ids[i] = m.ID
	}
	return ids
}

// CreateButtonDTO 用于创建按钮的请求结构体
//
// swagger:model CreateButtonDTO
type CreateButtonDTO struct {
	// 唯一标识
	ID uint32 `json:"id" form:"id" binding:"required,gt=0"`

	// 名称
	Name string `json:"name" form:"name" binding:"required,max=50"`

	// 排序字段
	Sort uint32 `json:"sort" form:"sort" binding:"required"`

	// 是否激活
	IsActive bool `json:"is_active" form:"is_active" binding:"required"`

	// 描述信息
	Descr string `json:"descr" form:"descr" binding:"omitempty,max=254"`

	// 菜单ID
	MenuID uint32 `json:"menu_id" form:"menu_id" binding:"required"`

	// API ID列表
	ApiIDs []uint32 `json:"api_ids" form:"api_ids" binding:"omitempty"`
}

func (dto *CreateButtonDTO) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddUint32("id", dto.ID)
	enc.AddString("name", dto.Name)
	enc.AddUint32("sort", dto.Sort)
	enc.AddBool("is_active", dto.IsActive)
	enc.AddString("descr", dto.Descr)
	enc.AddUint32("menu_id", dto.MenuID)
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

func (dto *CreateButtonDTO) ToModel() ButtonModel {
	return ButtonModel{
		StandardModel: database.StandardModel{
			BaseModel: database.BaseModel{ID: dto.ID},
		},
		Name:     dto.Name,
		Sort:     dto.Sort,
		IsActive: dto.IsActive,
		Descr:    dto.Descr,
		MenuID:   dto.MenuID,
	}
}

// UpdateButtonDTO 用于更新按钮的请求结构体
//
// swagger:model UpdateButtonDTO
type UpdateButtonDTO struct {
	// 名称
	Name string `json:"name" form:"name" binding:"required,max=50"`

	// 排序字段
	Sort uint32 `json:"sort" form:"sort" binding:"omitempty"`

	// 是否激活
	IsActive bool `json:"is_active" form:"is_active"`

	// 描述信息
	Descr string `json:"descr" form:"descr" binding:"omitempty,max=254"`

	// 菜单ID
	MenuID uint32 `json:"menu_id" form:"menu_id" binding:"required"`

	// API ID列表
	ApiIDs []uint32 `json:"api_ids" form:"api_ids" binding:"omitempty"`
}

func (dto *UpdateButtonDTO) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("name", dto.Name)
	enc.AddUint32("sort", dto.Sort)
	enc.AddBool("is_active", dto.IsActive)
	enc.AddString("descr", dto.Descr)
	enc.AddUint32("menu_id", dto.MenuID)
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

func (dto *UpdateButtonDTO) ToUpdateMap() map[string]any {
	return map[string]any{
		"name":      dto.Name,
		"sort":      dto.Sort,
		"is_active": dto.IsActive,
		"descr":     dto.Descr,
		"menu_id":   dto.MenuID,
	}
}

// ListButtonDTO 用于获取按钮列表的请求结构体
// 支持分页查询和多种筛选条件
//
// swagger:model ListButtonDTO
type ListButtonDTO struct {
	common.StandardModelQuery

	// 按钮名称
	Name string `form:"name" binding:"omitempty,max=50"`

	// 是否激活
	IsActive *bool `form:"is_active" binding:"omitempty"`

	// 描述信息
	Descr string `form:"descr" binding:"omitempty,max=254"`

	// 菜单ID
	MenuID uint32 `form:"menu_id" binding:"omitempty"`
}

func (dto *ListButtonDTO) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if dto == nil {
		return nil
	}
	if err := dto.StandardModelQuery.MarshalLogObject(enc); err != nil {
		return err
	}
	enc.AddString("name", dto.Name)
	if dto.IsActive != nil {
		enc.AddBool("is_active", *dto.IsActive)
	}
	enc.AddString("descr", dto.Descr)
	enc.AddUint32("menu_id", dto.MenuID)
	return nil
}

func (dto *ListButtonDTO) ToQueryMap() map[string]any {
	queryMap := dto.StandardModelQuery.ToQueryMap(10)
	if dto.Name != "" {
		queryMap["name like ?"] = "%" + dto.Name + "%"
	}
	if dto.IsActive != nil {
		queryMap["is_active = ?"] = *dto.IsActive
	}
	if dto.Descr != "" {
		queryMap["descr like ?"] = "%" + dto.Descr + "%"
	}
	if dto.MenuID != 0 {
		queryMap["menu_id = ?"] = dto.MenuID
	}
	return queryMap
}

type ButtonBaseOut struct {
	// 唯一标识
	ID uint32 `json:"id" example:"1"`

	// 名称
	Name string `json:"name" example:"用户管理"`

	// 排序字段
	Sort uint32 `json:"sort" example:"1000"`

	// 是否激活
	IsActive bool `json:"is_active" example:"true"`

	// 描述
	Descr string `json:"descr" example:"用户管理"`
}

// ButtonStandardOut 按钮标准信息
type ButtonStandardOut struct {
	ButtonBaseOut

	// 创建时间
	CreatedAt string `json:"created_at" example:"2023-01-01 12:00:00"`

	// 更新时间
	UpdatedAt string `json:"updated_at" example:"2023-01-01 12:00:00"`
}

// ButtonDetailOut 按钮详情信息
type ButtonDetailOut struct {
	ButtonStandardOut

	// 菜单
	Menu *MenuStandardOut `json:"menu"`

	// API ID列表
	ApiIDs []uint32 `json:"api_ids"`
}

// ButtonBaseResp 按钮响应结构
type ButtonResp = common.APIResp[*ButtonDetailOut]

// PagButtonResp 按钮的分页响应结构
type PagButtonResp = common.APIResp[*common.Pag[ButtonStandardOut]]

func ButtonModelToBaseOut(
	m ButtonModel,
) *ButtonBaseOut {
	return &ButtonBaseOut{
		ID:       m.ID,
		Name:     m.Name,
		Sort:     m.Sort,
		IsActive: m.IsActive,
		Descr:    m.Descr,
	}
}

func ButtonModelToStandardOut(
	m ButtonModel,
) *ButtonStandardOut {
	return &ButtonStandardOut{
		ButtonBaseOut: *ButtonModelToBaseOut(m),
		CreatedAt:     m.CreatedAt.Format(time.DateTime),
		UpdatedAt:     m.UpdatedAt.Format(time.DateTime),
	}
}

func ButtonModelToDetailOut(
	m ButtonModel,
) *ButtonDetailOut {
	var menu *MenuStandardOut
	if m.Menu.ID != 0 {
		menu = MenuModelToStandardOut(m.Menu)
	}
	var ApiIDs = []uint32{}
	if len(m.Apis) > 0 {
		ApiIDs = make([]uint32, len(m.Apis))
		for i, p := range m.Apis {
			ApiIDs[i] = p.ID
		}
	}
	return &ButtonDetailOut{
		ButtonStandardOut: *ButtonModelToStandardOut(m),
		Menu:              menu,
		ApiIDs:            ApiIDs,
	}
}

func ListButtonModelToStandardOut(
	ms []ButtonModel,
) []ButtonStandardOut {
	if len(ms) == 0 {
		return []ButtonStandardOut{}
	}
	mso := make([]ButtonStandardOut, 0, len(ms))
	if len(ms) > 0 {
		for _, m := range ms {
			mo := ButtonModelToStandardOut(m)
			mso = append(mso, *mo)
		}
	}
	return mso
}
