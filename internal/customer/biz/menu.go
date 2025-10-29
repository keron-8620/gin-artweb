package biz

import (
	"context"
	"encoding/json"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"gitee.com/keion8620/go-dango-gin/pkg/database"
	"gitee.com/keion8620/go-dango-gin/pkg/errors"
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
	Path         string            `gorm:"column:path;type:varchar(100);not null;uniqueIndex;comment:前端路由" json:"url"`
	Component    string            `gorm:"column:component;type:varchar(200);not null;comment:请求方式" json:"method"`
	Name         string            `gorm:"column:name;type:varchar(50);not null;uniqueIndex;comment:名称" json:"name"`
	Meta         Meta              `gorm:"column:meta;serializer:json;comment:菜单信息" json:"meta"`
	ArrangeOrder uint              `gorm:"column:arrange_order;type:integer;comment:排序" json:"arrange_order"`
	IsActive     bool              `gorm:"column:is_active;type:boolean;comment:是否激活" json:"is_active"`
	Descr        string            `gorm:"column:descr;type:varchar(254);comment:描述" json:"descr"`
	ParentId     *uint             `gorm:"column:parent_id;foreignKey:ParentId;references:Id;constraint:OnDelete:CASCADE;comment:菜单" json:"parent"`
	Parent       *MenuModel        `gorm:"foreignKey:ParentId;constraint:OnDelete:CASCADE"`
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
	enc.AddUint("arrange_order", m.ArrangeOrder)
	enc.AddBool("is_active", m.IsActive)
	enc.AddString("descr", m.Descr)
	if m.ParentId != nil {
		enc.AddUint("parent_id", *m.ParentId)
	}
	return nil
}

type MenuRepo interface {
	CreateModel(context.Context, *MenuModel) error
	UpdateModel(context.Context, map[string]any, map[string]any, ...any) error
	DeleteModel(context.Context, ...any) error
	FindModel(context.Context, []string, ...any) (*MenuModel, error)
	ListModel(context.Context, database.QueryParams) (int64, []MenuModel, error)
	AddGroupPolicy(context.Context, MenuModel) error
	RemoveGroupPolicy(context.Context, MenuModel) error
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

func (uc *MenuUsecase) GetParemtMenu(
	ctx context.Context,
	parentId *uint,
) (*MenuModel, *errors.Error) {
	if parentId == nil || *parentId == 0 {
		return nil, nil
	}
	m, err := uc.menuRepo.FindModel(ctx, nil, *parentId)
	if err != nil {
		rErr := database.NewGormError(err, map[string]any{"parent_id": *parentId})
		uc.log.Error(rErr.Error())
		return nil, rErr
	}
	return m, nil
}

func (uc *MenuUsecase) GetPermissions(
	ctx context.Context,
	permIds []uint,
) ([]PermissionModel, *errors.Error) {
	if len(permIds) == 0 {
		return nil, nil
	}
	qp := database.NewPksQueryParams(permIds)
	_, ms, err := uc.permRepo.ListModel(ctx, qp)
	if err != nil {
		rErr := database.NewGormError(err, nil)
		uc.log.Error(rErr.Error())
		return nil, rErr
	}
	return ms, nil
}

func (uc *MenuUsecase) CreateMenu(
	ctx context.Context,
	permIds []uint,
	m MenuModel,
) (*MenuModel, *errors.Error) {
	menu, err := uc.GetParemtMenu(ctx, m.ParentId)
	if err != nil {
		return nil, err
	}
	if menu != nil {
		m.Parent = menu
	}
	perms, err := uc.GetPermissions(ctx, permIds)
	if err != nil {
		return nil, err
	}
	if len(perms) > 0 {
		m.Permissions = perms
	}
	if err := uc.menuRepo.CreateModel(ctx, &m); err != nil {
		rErr := database.NewGormError(err, nil)
		uc.log.Error(rErr.Error())
		return nil, rErr
	}
	if err := uc.menuRepo.AddGroupPolicy(ctx, m); err != nil {
		rErr := ErrAddGroupPolicy.WithCause(err)
		uc.log.Error(rErr.Error())
		return nil, rErr
	}
	return &m, nil
}

func (uc *MenuUsecase) UpdateMenuById(
	ctx context.Context,
	menuId uint,
	permIds []uint,
	data map[string]any,
) *errors.Error {
	perms, err := uc.GetPermissions(ctx, permIds)
	if err != nil {
		return err
	}
	upmap := make(map[string]any, 1)
	if len(perms) > 0 {
		upmap["Permissions"] = perms
	}
	data["id"] = menuId
	if err := uc.menuRepo.UpdateModel(ctx, data, upmap, "id = ?", menuId); err != nil {
		rErr := database.NewGormError(err, data)
		uc.log.Error(rErr.Error())
		return rErr
	}
	m, rErr := uc.FindMenuById(ctx, []string{"Parent", "Permissions"}, menuId)
	if rErr != nil {
		return rErr
	}
	if err := uc.menuRepo.RemoveGroupPolicy(ctx, *m); err != nil {
		rErr = ErrRemoveGroupPolicy.WithCause(err)
		uc.log.Error(rErr.Error())
		return rErr
	}
	if err := uc.menuRepo.AddGroupPolicy(ctx, *m); err != nil {
		rErr := ErrAddGroupPolicy.WithCause(err)
		uc.log.Error(rErr.Error())
		return rErr
	}
	return nil
}

func (uc *MenuUsecase) DeleteMenuById(
	ctx context.Context,
	menuId uint,
) *errors.Error {
	m, rErr := uc.FindMenuById(ctx, []string{"Parent", "Permissions"}, menuId)
	if rErr != nil {
		return rErr
	}
	if err := uc.menuRepo.DeleteModel(ctx, menuId); err != nil {
		rErr = database.NewGormError(err, map[string]any{"id": menuId})
		uc.log.Error(rErr.Error())
		return rErr
	}
	if err := uc.menuRepo.RemoveGroupPolicy(ctx, *m); err != nil {
		rErr = ErrRemoveGroupPolicy.WithCause(err)
		uc.log.Error(rErr.Error())
		return rErr
	}
	return nil
}

func (uc *MenuUsecase) FindMenuById(
	ctx context.Context,
	preloads []string,
	menuId uint,
) (*MenuModel, *errors.Error) {
	m, err := uc.menuRepo.FindModel(ctx, preloads, menuId)
	if err != nil {
		rErr := database.NewGormError(err, map[string]any{"id": menuId})
		uc.log.Error(rErr.Error())
		return nil, rErr
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
		rErr := database.NewGormError(err, nil)
		uc.log.Error(rErr.Error())
		return 0, nil, rErr
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
			return err
		}
	}
	return nil
}
