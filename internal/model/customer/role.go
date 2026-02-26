package customer

import (
	"time"

	"go.uber.org/zap/zapcore"

	"gin-artweb/internal/model/common"
	"gin-artweb/internal/shared/database"
)

type RoleModel struct {
	database.StandardModel
	Name    string        `gorm:"column:name;type:varchar(50);not null;uniqueIndex;comment:名称" json:"name"`
	Descr   string        `gorm:"column:descr;type:varchar(254);comment:描述" json:"descr"`
	Apis    []ApiModel    `gorm:"many2many:customer_role_api;joinForeignKey:role_id;joinReferences:api_id;constraint:OnDelete:CASCADE"`
	Menus   []MenuModel   `gorm:"many2many:customer_role_menu;joinForeignKey:role_id;joinReferences:menu_id;constraint:OnDelete:CASCADE"`
	Buttons []ButtonModel `gorm:"many2many:customer_role_button;joinForeignKey:role_id;joinReferences:button_id;constraint:OnDelete:CASCADE"`
}

func (m *RoleModel) TableName() string {
	return "customer_role"
}

func (m *RoleModel) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if m == nil {
		return nil
	}
	if err := m.StandardModel.MarshalLogObject(enc); err != nil {
		return err
	}
	enc.AddString("name", m.Name)
	enc.AddString("descr", m.Descr)
	enc.AddArray("apis", zapcore.ArrayMarshalerFunc(func(ae zapcore.ArrayEncoder) error {
		for _, api := range m.Apis {
			ae.AppendUint32(api.ID)
		}
		return nil
	}))
	enc.AddArray("menus", zapcore.ArrayMarshalerFunc(func(ae zapcore.ArrayEncoder) error {
		for _, menu := range m.Menus {
			ae.AppendUint32(menu.ID)
		}
		return nil
	}))
	enc.AddArray("buttons", zapcore.ArrayMarshalerFunc(func(ae zapcore.ArrayEncoder) error {
		for _, button := range m.Buttons {
			ae.AppendUint32(button.ID)
		}
		return nil
	}))
	return nil
}

// type MenuTreeNode struct {
// 	Menu     MenuModel
// 	Children []*MenuTreeNode
// 	Buttons  []ButtonModel
// }

// CreateOrUpdateRoleRequest 用于创建或更新角色的请求结构体
//
// swagger:model CreateOrUpdateRoleRequest
type CreateOrUpdateRoleRequest struct {
	// 名称
	Name string `json:"name" form:"name" binding:"required,max=50"`

	// 描述信息
	Descr string `json:"descr" form:"descr" binding:"omitempty,max=254"`

	// APIID列表
	ApiIDs []uint32 `json:"api_ids" form:"api_ids" binding:"omitempty"`

	// 菜单ID列表
	MenuIDs []uint32 `json:"menu_ids" form:"menu_ids" binding:"omitempty"`

	// 按钮ID列表
	ButtonIDs []uint32 `json:"button_ids" form:"button_ids" binding:"omitempty"`
}

func (req *CreateOrUpdateRoleRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("name", req.Name)
	enc.AddString("descr", req.Descr)
	enc.AddArray("api_ids", zapcore.ArrayMarshalerFunc(func(ae zapcore.ArrayEncoder) error {
		for _, id := range req.ApiIDs {
			ae.AppendUint32(id)
		}
		return nil
	}))
	enc.AddArray("menu_ids", zapcore.ArrayMarshalerFunc(func(ae zapcore.ArrayEncoder) error {
		for _, id := range req.MenuIDs {
			ae.AppendUint32(id)
		}
		return nil
	}))
	enc.AddArray("button_ids", zapcore.ArrayMarshalerFunc(func(ae zapcore.ArrayEncoder) error {
		for _, id := range req.ButtonIDs {
			ae.AppendUint32(id)
		}
		return nil
	}))
	return nil
}

// ListRoleRequest 用于获取角色列表的请求结构体
// 支持分页查询和多种筛选条件
//
// swagger:model ListRoleRequest
type ListRoleRequest struct {
	common.StandardModelQuery

	// 名称
	Name string `form:"name" binding:"omitempty,max=50"`

	// 描述信息
	Descr string `form:"descr" binding:"omitempty,max=254"`
}

func (req *ListRoleRequest) Query() (int, int, map[string]any) {
	page, size, query := req.StandardModelQuery.QueryMap(8)
	if req.Name != "" {
		query["name like ?"] = "%" + req.Name + "%"
	}
	if req.Descr != "" {
		query["descr like ?"] = "%" + req.Descr + "%"
	}
	return page, size, query
}

type RoleBaseOut struct {
	// 唯一标识
	ID uint32 `json:"id" example:"1"`

	// 名称
	Name string `json:"name" example:"用户管理"`

	// 描述
	Descr string `json:"descr" example:"用户管理"`
}

// RoleStandardOut 角色基础信息
type RoleStandardOut struct {
	RoleBaseOut

	// 创建时间
	CreatedAt string `json:"created_at" example:"2023-01-01 12:00:00"`

	// 更新时间
	UpdatedAt string `json:"updated_at" example:"2023-01-01 12:00:00"`
}

type RoleDetailOut struct {
	RoleStandardOut

	// APIID列表
	ApiIDs []uint32 `json:"api_ids"`

	// 菜单ID列表
	MenuIDs []uint32 `json:"menu_ids"`

	// 按钮ID列表
	ButtonIDs []uint32 `json:"button_ids"`
}

// RoleBaseReply 角色响应结构
type RoleReply = common.APIReply[*RoleDetailOut]

// PagRoleReply 角色的分页响应结构
type PagRoleReply = common.APIReply[*common.Pag[RoleStandardOut]]

// MenuTreeNode 菜单树结点
type MenuTreeNode struct {
	MenuBaseOut

	// 子菜单
	Children []MenuTreeNode `json:"children"`

	// 按钮
	Buttons []ButtonBaseOut `json:"buttons"`
}

// RoleMenuTreeReply 角色响应结构
type RoleMenuTreeReply = common.APIReply[*[]MenuTreeNode]

func RoleModelToBaseOut(
	m RoleModel,
) *RoleBaseOut {
	return &RoleBaseOut{
		ID:    m.ID,
		Name:  m.Name,
		Descr: m.Descr,
	}
}

func RoleModelToStandardOut(
	m RoleModel,
) *RoleStandardOut {
	return &RoleStandardOut{
		RoleBaseOut: *RoleModelToBaseOut(m),
		CreatedAt:   m.CreatedAt.Format(time.DateTime),
		UpdatedAt:   m.UpdatedAt.Format(time.DateTime),
	}
}

func RoleModelToDetailOut(
	m RoleModel,
) *RoleDetailOut {
	var apiIDs = []uint32{}
	if len(m.Apis) > 0 {
		apiIDs = make([]uint32, len(m.Apis))
		for i, p := range m.Apis {
			apiIDs[i] = p.ID
		}
	}

	var menuIDs = []uint32{}
	if len(m.Menus) > 0 {
		menuIDs = make([]uint32, len(m.Menus))
		for i, p := range m.Menus {
			menuIDs[i] = p.ID
		}
	}

	var buttonIDs = []uint32{}
	if len(m.Buttons) > 0 {
		buttonIDs = make([]uint32, len(m.Buttons))
		for i, p := range m.Buttons {
			buttonIDs[i] = p.ID
		}
	}
	return &RoleDetailOut{
		RoleStandardOut: *RoleModelToStandardOut(m),
		ApiIDs:          apiIDs,
		MenuIDs:         menuIDs,
		ButtonIDs:       buttonIDs,
	}
}

func ListRoleModelToStandardOut(
	rms *[]RoleModel,
) *[]RoleStandardOut {
	if rms == nil {
		return &[]RoleStandardOut{}
	}
	ms := *rms
	mso := make([]RoleStandardOut, 0, len(ms))
	if len(ms) > 0 {
		for _, m := range ms {
			mo := RoleModelToStandardOut(m)
			mso = append(mso, *mo)
		}
	}
	return &mso
}
