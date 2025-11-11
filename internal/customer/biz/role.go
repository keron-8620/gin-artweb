package biz

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"gin-artweb/pkg/database"
	"gin-artweb/pkg/errors"
)

type RoleModel struct {
	database.StandardModel
	Name        string            `gorm:"column:name;type:varchar(50);not null;uniqueIndex;comment:名称" json:"name"`
	Descr       string            `gorm:"column:descr;type:varchar(254);comment:描述" json:"descr"`
	Permissions []PermissionModel `gorm:"many2many:customer_role_permission;joinForeignKey:role_id;joinReferences:permission_id;constraint:OnDelete:CASCADE"`
	Menus       []MenuModel       `gorm:"many2many:customer_role_menu;joinForeignKey:role_id;joinReferences:menu_id;constraint:OnDelete:CASCADE"`
	Buttons     []ButtonModel     `gorm:"many2many:customer_role_button;joinForeignKey:role_id;joinReferences:button_id;constraint:OnDelete:CASCADE"`
}

func (m *RoleModel) TableName() string {
	return "customer_role"
}

func (m *RoleModel) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if err := m.StandardModel.MarshalLogObject(enc); err != nil {
		return err
	}
	enc.AddString("name", m.Name)
	enc.AddString("descr", m.Descr)
	return nil
}

type MenuTreeNode struct {
	MenuModel MenuModel
	Children  []*MenuTreeNode
	Buttons   []ButtonModel
}

type RoleRepo interface {
	CreateModel(context.Context, *RoleModel) error
	UpdateModel(context.Context, map[string]any, []PermissionModel, []MenuModel, []ButtonModel, ...any) error
	DeleteModel(context.Context, ...any) error
	FindModel(context.Context, []string, ...any) (*RoleModel, error)
	ListModel(context.Context, database.QueryParams) (int64, []RoleModel, error)
	RoleModelToSub(RoleModel) string
	AddGroupPolicy(context.Context, RoleModel) error
	RemoveGroupPolicy(context.Context, RoleModel) error
}

type RoleUsecase struct {
	log        *zap.Logger
	permRepo   PermissionRepo
	menuRepo   MenuRepo
	buttonRepo ButtonRepo
	roleRepo   RoleRepo
}

func NewRoleUsecase(
	log *zap.Logger,
	permRepo PermissionRepo,
	menuRepo MenuRepo,
	buttonRepo ButtonRepo,
	roleRepo RoleRepo,
) *RoleUsecase {
	return &RoleUsecase{
		log:        log,
		permRepo:   permRepo,
		menuRepo:   menuRepo,
		buttonRepo: buttonRepo,
		roleRepo:   roleRepo,
	}
}

func (uc *RoleUsecase) GetPermissions(
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

func (uc *RoleUsecase) GetMenus(
	ctx context.Context,
	menuIDs []uint32,
) ([]MenuModel, *errors.Error) {
	if len(menuIDs) == 0 {
		return nil, nil
	}
	qp := database.NewPksQueryParams(menuIDs)
	_, ms, err := uc.menuRepo.ListModel(ctx, qp)
	if err != nil {
		return nil, database.NewGormError(err, nil)
	}
	return ms, nil
}

func (uc *RoleUsecase) GetButtons(
	ctx context.Context,
	buttonIDs []uint32,
) ([]ButtonModel, *errors.Error) {
	if len(buttonIDs) == 0 {
		return nil, nil
	}
	qp := database.NewPksQueryParams(buttonIDs)
	_, ms, err := uc.buttonRepo.ListModel(ctx, qp)
	if err != nil {
		return nil, database.NewGormError(err, nil)
	}
	return ms, nil
}

func (uc *RoleUsecase) CreateRole(
	ctx context.Context,
	permIDs []uint32,
	menuIDs []uint32,
	buttonIDs []uint32,
	m RoleModel,
) (*RoleModel, *errors.Error) {
	perms, err := uc.GetPermissions(ctx, permIDs)
	if err != nil {
		return nil, err
	}
	if len(perms) > 0 {
		m.Permissions = perms
	}
	menus, err := uc.GetMenus(ctx, menuIDs)
	if err != nil {
		return nil, err
	}
	if len(menus) > 0 {
		m.Menus = menus
	}
	buttons, err := uc.GetButtons(ctx, buttonIDs)
	if err != nil {
		return nil, err
	}
	if len(buttons) > 0 {
		m.Buttons = buttons
	}
	if err := uc.roleRepo.CreateModel(ctx, &m); err != nil {
		return nil, database.NewGormError(err, nil)
	}
	if err := uc.roleRepo.AddGroupPolicy(ctx, m); err != nil {
		return nil, ErrAddGroupPolicy.WithCause(err)
	}
	return &m, nil
}

func (uc *RoleUsecase) UpdateRoleByID(
	ctx context.Context,
	roleID uint32,
	permIDs []uint32,
	menuIDs []uint32,
	buttonIDs []uint32,
	data map[string]any,
) *errors.Error {
	perms, err := uc.GetPermissions(ctx, permIDs)
	if err != nil {
		return err
	}
	menus, err := uc.GetMenus(ctx, menuIDs)
	if err != nil {
		return err
	}
	buttons, err := uc.GetButtons(ctx, buttonIDs)
	if err != nil {
		return err
	}
	data["id"] = roleID
	if err := uc.roleRepo.UpdateModel(ctx, data, perms, menus, buttons, "id = ?", roleID); err != nil {
		return database.NewGormError(err, data)
	}
	m, rErr := uc.FindRoleByID(ctx, []string{"Permissions", "Menus", "Buttons"}, roleID)
	if rErr != nil {
		return rErr
	}
	if err := uc.roleRepo.RemoveGroupPolicy(ctx, *m); err != nil {
		return ErrRemoveGroupPolicy.WithCause(err)
	}
	if err := uc.roleRepo.AddGroupPolicy(ctx, *m); err != nil {
		return ErrAddGroupPolicy.WithCause(err)
	}
	return nil
}

func (uc *RoleUsecase) DeleteRoleByID(
	ctx context.Context,
	roleID uint32,
) *errors.Error {
	m, rErr := uc.FindRoleByID(ctx, []string{"Permissions", "Menus", "Buttons"}, roleID)
	if rErr != nil {
		return rErr
	}
	if err := uc.roleRepo.DeleteModel(ctx, roleID); err != nil {
		return database.NewGormError(err, map[string]any{"id": roleID})
	}
	if err := uc.roleRepo.RemoveGroupPolicy(ctx, *m); err != nil {
		return ErrRemoveGroupPolicy.WithCause(err)
	}
	return nil
}

func (uc *RoleUsecase) FindRoleByID(
	ctx context.Context,
	preloads []string,
	roleID uint32,
) (*RoleModel, *errors.Error) {
	m, err := uc.roleRepo.FindModel(ctx, preloads, roleID)
	if err != nil {
		return nil, database.NewGormError(err, map[string]any{"id": roleID})
	}
	return m, nil
}

func (uc *RoleUsecase) ListRole(
	ctx context.Context,
	page, size int,
	query map[string]any,
	orderBy []string,
	isCount bool,
	preloads []string,
) (int64, []RoleModel, *errors.Error) {
	qp := database.QueryParams{
		Preloads: preloads,
		Query:    query,
		OrderBy:  orderBy,
		Limit:    max(size, 0),
		Offset:   max(page-1, 0),
		IsCount:  isCount,
	}
	count, ms, err := uc.roleRepo.ListModel(ctx, qp)
	if err != nil {
		return 0, nil, database.NewGormError(err, nil)
	}
	return count, ms, nil
}

func (uc *RoleUsecase) LoadRolePolicy(ctx context.Context) error {
	_, rms, err := uc.ListRole(ctx, 0, 0, nil, nil, false, nil)
	if err != nil {
		return err
	}
	for _, rm := range rms {
		if err := uc.roleRepo.AddGroupPolicy(ctx, rm); err != nil {
			return ErrAddGroupPolicy.WithCause(err)
		}
	}
	return nil
}

func (uc *RoleUsecase) GetRoleMenuTree(
	ctx context.Context,
	roleID uint32,
) ([]*MenuTreeNode, *errors.Error) {
	m, rErr := uc.FindRoleByID(ctx, []string{"Menus", "Buttons"}, roleID)
	if rErr != nil {
		return nil, rErr
	}
	roleMenuMap := make(map[uint32]MenuModel)
	for _, menu := range m.Menus {
		roleMenuMap[menu.ID] = menu
	}
	roleButtonMap := make(map[uint32]ButtonModel)
	for _, button := range m.Buttons {
		roleButtonMap[button.ID] = button
	}
	var result []*MenuTreeNode
	for _, menu := range m.Menus {
		if menu.ParentID == nil {
			mt, err := uc.buildMenuTree(menu, roleMenuMap, roleButtonMap)
			if err != nil {
				return nil, err
			}
			result = append(result, mt)
		}
	}
	return result, nil
}

func (uc *RoleUsecase) buildMenuTree(
	m MenuModel,
	mp map[uint32]MenuModel,
	bp map[uint32]ButtonModel,
) (*MenuTreeNode, *errors.Error) {
	var children []MenuModel
	for _, menu := range mp {
		if menu.ParentID != nil && *menu.ParentID == m.ID {
			children = append(children, menu)
		}
	}
	var childTrees []*MenuTreeNode
	for _, child := range children {
		childTree, err := uc.buildMenuTree(child, mp, bp)
		if err != nil {
			return nil, err
		}
		childTrees = append(childTrees, childTree)
	}
	var buttons []ButtonModel
	for _, button := range bp {
		if button.MenuID == m.ID {
			buttons = append(buttons, button)
		}
	}
	return &MenuTreeNode{
		MenuModel: m,
		Children:  childTrees,
		Buttons:   buttons,
	}, nil
}
