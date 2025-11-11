package biz

import (
	"context"
	"encoding/json"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"gin-artweb/pkg/database"
	"gin-artweb/pkg/errors"
)

type Meta struct {
	Title string `json:"title"`
	Icon  string `json:"icon"`
}

func (m *Meta) Json() string {
	metaBytes, err := json.Marshal(m)
	if err != nil {
		return "{}"
	}
	return string(metaBytes)
}

type MenuModel struct {
	database.StandardModel
	Path         string            `gorm:"column:path;type:varchar(100);not null;uniqueIndex;comment:前端路由" json:"path"`
	Component    string            `gorm:"column:component;type:varchar(200);not null;comment:前端组件" json:"component"`
	Name         string            `gorm:"column:name;type:varchar(50);not null;uniqueIndex;comment:名称" json:"name"`
	Meta         Meta              `gorm:"column:meta;serializer:json;comment:菜单信息" json:"meta"`
	ArrangeOrder uint32            `gorm:"column:arrange_order;type:integer;comment:排序" json:"arrange_order"`
	IsActive     bool              `gorm:"column:is_active;type:boolean;comment:是否激活" json:"is_active"`
	Descr        string            `gorm:"column:descr;type:varchar(254);comment:描述" json:"descr"`
	ParentID     *uint32           `gorm:"column:parent_id;comment:父菜单ID" json:"parent_id"`
	Parent       *MenuModel        `gorm:"foreignKey:ParentID;references:ID;constraint:OnDelete:CASCADE" json:"parent"`
	Permissions  []PermissionModel `gorm:"many2many:customer_menu_permission;joinForeignKey:menu_id;joinReferences:permission_id;constraint:OnDelete:CASCADE"`
}

func (m *MenuModel) TableName() string {
	return "customer_menu"
}

func (m *MenuModel) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if err := m.StandardModel.MarshalLogObject(enc); err != nil {
		return err
	}
	enc.AddString("path", m.Path)
	enc.AddString("component", m.Component)
	enc.AddString("meta", m.Meta.Json())
	enc.AddString("name", m.Name)
	enc.AddUint32("arrange_order", m.ArrangeOrder)
	enc.AddBool("is_active", m.IsActive)
	enc.AddString("descr", m.Descr)
	if m.ParentID != nil {
		enc.AddUint32("parent_id", *m.ParentID)
	}
	return nil
}

type MenuRepo interface {
	CreateModel(context.Context, *MenuModel) error
	UpdateModel(context.Context, map[string]any, []PermissionModel, ...any) error
	DeleteModel(context.Context, ...any) error
	FindModel(context.Context, []string, ...any) (*MenuModel, error)
	ListModel(context.Context, database.QueryParams) (int64, []MenuModel, error)
	AddGroupPolicy(context.Context, MenuModel) error
	RemoveGroupPolicy(context.Context, MenuModel, bool) error
}

type MenuUsecase struct {
	log      *zap.Logger
	permRepo PermissionRepo
	menuRepo MenuRepo
}

func NewMenuUsecase(
	log *zap.Logger,
	permRepo PermissionRepo,
	menuRepo MenuRepo,
) *MenuUsecase {
	return &MenuUsecase{
		log:      log,
		permRepo: permRepo,
		menuRepo: menuRepo,
	}
}

func (uc *MenuUsecase) GetParentMenu(
	ctx context.Context,
	parentID *uint32,
) (*MenuModel, *errors.Error) {
	if parentID == nil || *parentID == 0 {
		return nil, nil
	}
	m, err := uc.menuRepo.FindModel(ctx, nil, *parentID)
	if err != nil {
		return nil, database.NewGormError(err, map[string]any{"parent_id": *parentID})
	}
	return m, nil
}

func (uc *MenuUsecase) GetPermissions(
	ctx context.Context,
	permIDs []uint32,
) ([]PermissionModel, *errors.Error) {
	if len(permIDs) == 0 {
		return nil, nil
	}
	qp := database.NewPksQueryParams(permIDs)
	_, ms, err := uc.permRepo.ListModel(ctx, qp)
	if err != nil {
		return nil, database.NewGormError(err, nil)
	}
	return ms, nil
}

func (uc *MenuUsecase) CreateMenu(
	ctx context.Context,
	permIDs []uint32,
	m MenuModel,
) (*MenuModel, *errors.Error) {
	menu, err := uc.GetParentMenu(ctx, m.ParentID)
	if err != nil {
		return nil, err
	}
	if menu != nil {
		m.Parent = menu
	}
	perms, err := uc.GetPermissions(ctx, permIDs)
	if err != nil {
		return nil, err
	}
	if len(perms) > 0 {
		m.Permissions = perms
	}
	if err := uc.menuRepo.CreateModel(ctx, &m); err != nil {
		return nil, database.NewGormError(err, nil)
	}
	if err := uc.menuRepo.AddGroupPolicy(ctx, m); err != nil {
		return nil, ErrAddGroupPolicy.WithCause(err)
	}
	return &m, nil
}

func (uc *MenuUsecase) UpdateMenuByID(
	ctx context.Context,
	menuID uint32,
	permIDs []uint32,
	data map[string]any,
) *errors.Error {
	perms, err := uc.GetPermissions(ctx, permIDs)
	if err != nil {
		return err
	}
	data["id"] = menuID
	if err := uc.menuRepo.UpdateModel(ctx, data, perms, "id = ?", menuID); err != nil {
		return database.NewGormError(err, data)
	}
	m, rErr := uc.FindMenuByID(ctx, []string{"Parent", "Permissions"}, menuID)
	if rErr != nil {
		return rErr
	}
	if err := uc.menuRepo.RemoveGroupPolicy(ctx, *m, false); err != nil {
		return ErrRemoveGroupPolicy.WithCause(err)
	}
	if err := uc.menuRepo.AddGroupPolicy(ctx, *m); err != nil {
		return ErrAddGroupPolicy.WithCause(err)
	}
	return nil
}

func (uc *MenuUsecase) DeleteMenuByID(
	ctx context.Context,
	menuID uint32,
) *errors.Error {
	m, rErr := uc.FindMenuByID(ctx, []string{"Parent", "Permissions"}, menuID)
	if rErr != nil {
		return rErr
	}
	if err := uc.menuRepo.DeleteModel(ctx, menuID); err != nil {
		return database.NewGormError(err, map[string]any{"id": menuID})
	}
	if err := uc.menuRepo.RemoveGroupPolicy(ctx, *m, true); err != nil {
		return ErrRemoveGroupPolicy.WithCause(err)
	}
	return nil
}

func (uc *MenuUsecase) FindMenuByID(
	ctx context.Context,
	preloads []string,
	menuID uint32,
) (*MenuModel, *errors.Error) {
	m, err := uc.menuRepo.FindModel(ctx, preloads, menuID)
	if err != nil {
		return nil, database.NewGormError(err, map[string]any{"id": menuID})
	}
	return m, nil
}

func (uc *MenuUsecase) ListMenu(
	ctx context.Context,
	page, size int,
	query map[string]any,
	orderBy []string,
	isCount bool,
	preloads []string,
) (int64, []MenuModel, *errors.Error) {
	qp := database.QueryParams{
		Preloads: preloads,
		Query:    query,
		OrderBy:  orderBy,
		Limit:    max(size, 0),
		Offset:   max(page-1, 0),
		IsCount:  isCount,
	}
	count, ms, err := uc.menuRepo.ListModel(ctx, qp)
	if err != nil {
		return 0, nil, database.NewGormError(err, nil)
	}
	return count, ms, nil
}

func (uc *MenuUsecase) LoadMenuPolicy(ctx context.Context) error {
	_, mms, err := uc.ListMenu(ctx, 0, 0, nil, nil, false, nil)
	if err != nil {
		return err
	}
	for _, mm := range mms {
		if err := uc.menuRepo.AddGroupPolicy(ctx, mm); err != nil {
			return ErrAddGroupPolicy.WithCause(err)
		}
	}
	return nil
}
