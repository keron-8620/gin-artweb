package biz

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"gitee.com/keion8620/go-dango-gin/pkg/database"
	"gitee.com/keion8620/go-dango-gin/pkg/errors"
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

type RoleRepo interface {
	CreateModel(context.Context, *RoleModel) error
	UpdateModel(context.Context, map[string]any, map[string]any, ...any) error
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

func (uc *RoleUsecase) GetMenus(
	ctx context.Context,
	menuIds []uint,
) ([]MenuModel, *errors.Error) {
	if len(menuIds) == 0 {
		return nil, nil
	}
	qp := database.NewPksQueryParams(menuIds)
	_, ms, err := uc.menuRepo.ListModel(ctx, qp)
	if err != nil {
		rErr := database.NewGormError(err, nil)
		uc.log.Error(rErr.Error())
		return nil, rErr
	}
	return ms, nil
}

func (uc *RoleUsecase) GetButtons(
	ctx context.Context,
	buttonIds []uint,
) ([]ButtonModel, *errors.Error) {
	if len(buttonIds) == 0 {
		return nil, nil
	}
	qp := database.NewPksQueryParams(buttonIds)
	_, ms, err := uc.buttonRepo.ListModel(ctx, qp)
	if err != nil {
		rErr := database.NewGormError(err, nil)
		uc.log.Error(rErr.Error())
		return nil, rErr
	}
	return ms, nil
}

func (uc *RoleUsecase) CreateRole(
	ctx context.Context,
	permIds []uint,
	menuIds []uint,
	buttonIds []uint,
	m RoleModel,
) (*RoleModel, *errors.Error) {
	perms, err := uc.GetPermissions(ctx, permIds)
	if err != nil {
		return nil, err
	}
	if len(perms) > 0 {
		m.Permissions = perms
	}
	menus, err := uc.GetMenus(ctx, menuIds)
	if err != nil {
		return nil, err
	}
	if len(menus) > 0 {
		m.Menus = menus
	}
	buttons, err := uc.GetButtons(ctx, buttonIds)
	if err != nil {
		return nil, err
	}
	if len(buttons) > 0 {
		m.Buttons = buttons
	}
	if err := uc.roleRepo.CreateModel(ctx, &m); err != nil {
		rErr := database.NewGormError(err, nil)
		uc.log.Error(rErr.Error())
		return nil, rErr
	}
	if err := uc.roleRepo.AddGroupPolicy(ctx, m); err != nil {
		rErr := ErrAddGroupPolicy.WithCause(err)
		uc.log.Error(rErr.Error())
		return nil, rErr
	}
	return &m, nil
}

func (uc *RoleUsecase) UpdateRoleById(
	ctx context.Context,
	roleId uint,
	permIds []uint,
	menuIds []uint,
	buttonIds []uint,
	data map[string]any,
) *errors.Error {
	ummap := make(map[string]any, 3)
	perms, err := uc.GetPermissions(ctx, permIds)
	if err != nil {
		return err
	}
	if len(perms) > 0 {
		ummap["Permissions"] = perms
	}
	menus, err := uc.GetMenus(ctx, menuIds)
	if err != nil {
		return err
	}
	if len(menus) > 0 {
		ummap["Menus"] = menus
	}
	buttons, err := uc.GetButtons(ctx, buttonIds)
	if err != nil {
		return err
	}
	if len(buttons) > 0 {
		ummap["Buttons"] = buttons
	}
	data["id"] = roleId
	if err := uc.roleRepo.UpdateModel(ctx, data, ummap, "id = ?", roleId); err != nil {
		rErr := database.NewGormError(err, data)
		uc.log.Error(rErr.Error())
		return rErr
	}
	m, rErr := uc.FindRoleById(ctx, []string{"Buttons", "Menus", "Permissions"}, roleId)
	if rErr != nil {
		return rErr
	}
	if err := uc.roleRepo.RemoveGroupPolicy(ctx, *m); err != nil {
		rErr := ErrRemoveGroupPolicy.WithCause(err)
		uc.log.Error(rErr.Error())
		return rErr
	}
	if err := uc.roleRepo.AddGroupPolicy(ctx, *m); err != nil {
		rErr := ErrAddGroupPolicy.WithCause(err)
		uc.log.Error(rErr.Error())
		return rErr
	}
	return nil
}

func (uc *RoleUsecase) DeleteRoleById(ctx context.Context, roleId uint) *errors.Error {
	m, rErr := uc.FindRoleById(ctx, []string{"Buttons", "Menus", "Permissions"}, roleId)
	if rErr != nil {
		return rErr
	}
	if err := uc.roleRepo.DeleteModel(ctx, roleId); err != nil {
		rErr := database.NewGormError(err, map[string]any{"id": roleId})
		uc.log.Error(rErr.Error())
		return rErr
	}
	if err := uc.roleRepo.RemoveGroupPolicy(ctx, *m); err != nil {
		rErr := ErrRemoveGroupPolicy.WithCause(err)
		uc.log.Error(rErr.Error())
		return rErr
	}
	return nil
}

func (uc *RoleUsecase) FindRoleById(
	ctx context.Context,
	preloads []string,
	roleId uint,
) (*RoleModel, *errors.Error) {
	m, err := uc.roleRepo.FindModel(ctx, preloads, roleId)
	if err != nil {
		rErr := database.NewGormError(err, map[string]any{"id": roleId})
		uc.log.Error(rErr.Error())
		return nil, rErr
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
		rErr := database.NewGormError(err, nil)
		uc.log.Error(rErr.Error())
		return 0, nil, rErr
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
			return err
		}
	}
	return nil
}
