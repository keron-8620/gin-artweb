package biz

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"gin-artweb/pkg/database"
	"gin-artweb/pkg/errors"
)

type ButtonModel struct {
	database.StandardModel
	Name         string            `gorm:"column:name;type:varchar(50);not null;uniqueIndex;comment:名称" json:"name"`
	ArrangeOrder uint32            `gorm:"column:arrange_order;type:integer;comment:排序" json:"arrange_order"`
	IsActive     bool              `gorm:"column:is_active;type:boolean;comment:是否激活" json:"is_active"`
	Descr        string            `gorm:"column:descr;type:varchar(254);comment:描述" json:"descr"`
	MenuId       uint32            `gorm:"column:menu_id;foreignKey:MenuId;references:Id;not null;constraint:OnDelete:CASCADE;comment:菜单" json:"menu"`
	Menu         MenuModel         `gorm:"foreignKey:MenuId;constraint:OnDelete:CASCADE"`
	Permissions  []PermissionModel `gorm:"many2many:customer_button_permission;joinForeignKey:button_id;joinReferences:permission_id;constraint:OnDelete:CASCADE"`
}

func (m *ButtonModel) TableName() string {
	return "customer_button"
}

func (m *ButtonModel) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if err := m.StandardModel.MarshalLogObject(enc); err != nil {
		return err
	}
	enc.AddString("name", m.Name)
	enc.AddUint32("arrange_order", m.ArrangeOrder)
	enc.AddBool("is_active", m.IsActive)
	enc.AddString("descr", m.Descr)
	enc.AddUint32("menu_id", m.MenuId)
	return nil
}

type ButtonUsecase struct {
	log        *zap.Logger
	permRepo   PermissionRepo
	menuRepo   MenuRepo
	buttonRepo ButtonRepo
}

type ButtonRepo interface {
	CreateModel(context.Context, *ButtonModel) error
	UpdateModel(context.Context, map[string]any, map[string]any, ...any) error
	DeleteModel(context.Context, ...any) error
	FindModel(context.Context, []string, ...any) (*ButtonModel, error)
	ListModel(context.Context, database.QueryParams) (int64, []ButtonModel, error)
	AddGroupPolicy(context.Context, ButtonModel) error
	RemoveGroupPolicy(context.Context, ButtonModel, bool) error
}

func NewButtonUsecase(
	log *zap.Logger,
	permRepo PermissionRepo,
	menuRepo MenuRepo,
	buttonRepo ButtonRepo,
) *ButtonUsecase {
	return &ButtonUsecase{
		log:        log,
		permRepo:   permRepo,
		menuRepo:   menuRepo,
		buttonRepo: buttonRepo,
	}
}

func (uc *ButtonUsecase) GetMenu(
	ctx context.Context,
	menuId uint32,
) (*MenuModel, *errors.Error) {
	m, err := uc.menuRepo.FindModel(ctx, nil, menuId)
	if err != nil {
		return nil, database.NewGormError(err, map[string]any{"menu_id": menuId})
	}
	return m, nil
}

func (uc *ButtonUsecase) GetPermissions(
	ctx context.Context,
	permIds []uint32,
) ([]PermissionModel, *errors.Error) {
	if len(permIds) == 0 {
		return nil, nil
	}
	qp := database.NewPksQueryParams(permIds)
	_, ms, err := uc.permRepo.ListModel(ctx, qp)
	if err != nil {
		return nil, database.NewGormError(err, nil)
	}
	return ms, nil
}

func (uc *ButtonUsecase) CreateButton(
	ctx context.Context,
	permIds []uint32,
	m ButtonModel,
) (*ButtonModel, *errors.Error) {
	menu, err := uc.GetMenu(ctx, m.MenuId)
	if err != nil {
		return nil, err
	}
	m.Menu = *menu
	perms, err := uc.GetPermissions(ctx, permIds)
	if err != nil {
		return nil, err
	}
	if len(perms) > 0 {
		m.Permissions = perms
	}
	if err := uc.buttonRepo.CreateModel(ctx, &m); err != nil {
		return nil, database.NewGormError(err, nil)
	}
	if err := uc.buttonRepo.AddGroupPolicy(ctx, m); err != nil {
		return nil, ErrRemoveGroupPolicy.WithCause(err)
	}
	return &m, nil
}

func (uc *ButtonUsecase) UpdateButtonById(
	ctx context.Context,
	buttonId uint32,
	permIds []uint32,
	data map[string]any,
) *errors.Error {
	ummap := make(map[string]any, 1)
	perms, err := uc.GetPermissions(ctx, permIds)
	if err != nil {
		return err
	}
	if len(perms) > 0 {
		ummap["Permissions"] = perms
	}
	data["id"] = buttonId
	if err := uc.buttonRepo.UpdateModel(ctx, data, ummap, "id = ?", buttonId); err != nil {
		return database.NewGormError(err, data)
	}
	m, rErr := uc.FindButtonById(ctx, []string{"Menu", "Permissions"}, buttonId)
	if rErr != nil {
		return rErr
	}
	if err := uc.buttonRepo.RemoveGroupPolicy(ctx, *m, false); err != nil {
		return ErrRemoveGroupPolicy.WithCause(err)
	}
	if err := uc.buttonRepo.AddGroupPolicy(ctx, *m); err != nil {
		return ErrAddGroupPolicy.WithCause(err)
	}
	return nil
}

func (uc *ButtonUsecase) DeleteButtonById(
	ctx context.Context,
	buttonId uint32,
) *errors.Error {
	m, rErr := uc.FindButtonById(ctx, []string{"Menu", "Permissions"}, buttonId)
	if rErr != nil {
		return rErr
	}
	if err := uc.buttonRepo.DeleteModel(ctx, buttonId); err != nil {
		return database.NewGormError(err, map[string]any{"id": buttonId})
	}
	if err := uc.buttonRepo.RemoveGroupPolicy(ctx, *m, true); err != nil {
		return ErrRemoveGroupPolicy.WithCause(err)
	}
	return nil
}

func (uc *ButtonUsecase) FindButtonById(
	ctx context.Context,
	preloads []string,
	buttonId uint32,
) (*ButtonModel, *errors.Error) {
	m, err := uc.buttonRepo.FindModel(ctx, preloads, buttonId)
	if err != nil {
		return nil, database.NewGormError(err, map[string]any{"id": buttonId})
	}
	return m, nil
}

func (uc *ButtonUsecase) ListButton(
	ctx context.Context,
	page, size int,
	query map[string]any,
	orderBy []string,
	isCount bool,
	preloads []string,
) (int64, []ButtonModel, *errors.Error) {
	qp := database.QueryParams{
		Preloads: preloads,
		Query:    query,
		OrderBy:  orderBy,
		Limit:    max(size, 0),
		Offset:   max(page-1, 0),
		IsCount:  isCount,
	}
	count, ms, err := uc.buttonRepo.ListModel(ctx, qp)
	if err != nil {
		return 0, nil, database.NewGormError(err, nil)
	}
	return count, ms, nil
}

func (uc *ButtonUsecase) LoadButtonPolicy(ctx context.Context) error {
	_, bms, err := uc.ListButton(ctx, 0, 0, nil, nil, false, nil)
	if err != nil {
		return err
	}
	for _, bm := range bms {
		if err := uc.buttonRepo.AddGroupPolicy(ctx, bm); err != nil {
			return ErrAddGroupPolicy.WithCause(err)
		}
	}
	return nil
}
