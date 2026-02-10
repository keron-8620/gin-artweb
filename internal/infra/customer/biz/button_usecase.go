package biz

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
)

const (
	ButtonTableName = "customer_button"
	ButtonIDKey     = "button_id"
	ButtonIDsKey    = "button_ids"
)

type ButtonModel struct {
	database.StandardModel
	Name         string            `gorm:"column:name;type:varchar(50);not null;uniqueIndex;comment:名称" json:"name"`
	ArrangeOrder uint32            `gorm:"column:arrange_order;type:integer;comment:排序" json:"arrange_order"`
	IsActive     bool              `gorm:"column:is_active;type:boolean;comment:是否激活" json:"is_active"`
	Descr        string            `gorm:"column:descr;type:varchar(254);comment:描述" json:"descr"`
	MenuID       uint32            `gorm:"column:menu_id;not null;comment:菜单ID" json:"menu_id"`
	Menu         MenuModel         `gorm:"foreignKey:MenuID;references:ID;constraint:OnDelete:CASCADE" json:"menu"`
	Permissions  []PermissionModel `gorm:"many2many:customer_button_permission;joinForeignKey:button_id;joinReferences:permission_id;constraint:OnDelete:CASCADE"`
}

func (m *ButtonModel) TableName() string {
	return ButtonTableName
}

func (m *ButtonModel) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if m == nil {
		return nil
	}
	if err := m.StandardModel.MarshalLogObject(enc); err != nil {
		return err
	}
	enc.AddString("name", m.Name)
	enc.AddUint32("arrange_order", m.ArrangeOrder)
	enc.AddBool("is_active", m.IsActive)
	enc.AddString("descr", m.Descr)
	enc.AddUint32("menu_id", m.MenuID)
	enc.AddArray(PermissionIDsKey, zapcore.ArrayMarshalerFunc(func(ae zapcore.ArrayEncoder) error {
		for _, perm := range m.Permissions {
			ae.AppendUint32(perm.ID)
		}
		return nil
	}))
	return nil
}

func ListButtonModelToUint32s(bms *[]ButtonModel) []uint32 {
	if bms == nil {
		return []uint32{}
	}
	ms := *bms
	if len(ms) == 0 {
		return []uint32{}
	}

	ids := make([]uint32, len(ms))
	for i, m := range ms {
		ids[i] = m.ID
	}
	return ids
}

type ButtonUsecase struct {
	log        *zap.Logger
	permRepo   PermissionRepo
	menuRepo   MenuRepo
	buttonRepo ButtonRepo
}

type ButtonRepo interface {
	CreateModel(context.Context, *ButtonModel, *[]PermissionModel) error
	UpdateModel(context.Context, map[string]any, *[]PermissionModel, ...any) error
	DeleteModel(context.Context, ...any) error
	GetModel(context.Context, []string, ...any) (*ButtonModel, error)
	ListModel(context.Context, database.QueryParams) (int64, *[]ButtonModel, error)
	AddGroupPolicy(context.Context, *ButtonModel) error
	RemoveGroupPolicy(context.Context, *ButtonModel, bool) error
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
	menuID uint32,
) (*MenuModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	uc.log.Info(
		"开始查询按钮关联的菜单",
		zap.Uint32(MenuIDKey, menuID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, err := uc.menuRepo.GetModel(ctx, nil, menuID)
	if err != nil {
		uc.log.Error(
			"查询按钮关联的菜单失败",
			zap.Error(err),
			zap.Uint32(MenuIDKey, menuID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, map[string]any{"menu_id": menuID})
	}

	uc.log.Info(
		"查询按钮关联的菜单成功",
		zap.Uint32(MenuIDKey, menuID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return m, nil
}

func (uc *ButtonUsecase) GetPermissions(
	ctx context.Context,
	permIDs []uint32,
) (*[]PermissionModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	if len(permIDs) == 0 {
		return &[]PermissionModel{}, nil
	}

	uc.log.Info(
		"开始查询按钮关联的权限列表",
		zap.Uint32s(PermissionIDsKey, permIDs),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	qp := database.QueryParams{
		Query: map[string]any{"id in ?": permIDs},
	}
	_, ms, err := uc.permRepo.ListModel(ctx, qp)
	if err != nil {
		uc.log.Error(
			"查询按钮关联的权限列表失败",
			zap.Error(err),
			zap.Uint32s(PermissionIDsKey, permIDs),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, nil)
	}

	uc.log.Info(
		"查询按钮关联的权限列表成功",
		zap.Uint32s(PermissionIDsKey, permIDs),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return ms, nil
}

func (uc *ButtonUsecase) CreateButton(
	ctx context.Context,
	permIDs []uint32,
	m ButtonModel,
) (*ButtonModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	uc.log.Info(
		"开始创建按钮",
		zap.Uint32s(PermissionIDsKey, permIDs),
		zap.Object(database.ModelKey, &m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	menu, err := uc.GetMenu(ctx, m.MenuID)
	if err != nil {
		return nil, err
	}
	m.Menu = *menu

	perms, err := uc.GetPermissions(ctx, permIDs)
	if err != nil {
		return nil, err
	}

	if err := uc.buttonRepo.CreateModel(ctx, &m, perms); err != nil {
		uc.log.Error(
			"创建按钮失败",
			zap.Error(err),
			zap.Object(database.ModelKey, &m),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, nil)
	}

	if len(*perms) > 0 {
		m.Permissions = *perms
	}

	if err := uc.buttonRepo.AddGroupPolicy(ctx, &m); err != nil {
		uc.log.Error(
			"添加按钮组策略失败",
			zap.Error(err),
			zap.Object(database.ModelKey, &m),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.FromError(err)
	}

	uc.log.Info(
		"创建按钮成功",
		zap.Object(database.ModelKey, &m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return &m, nil
}

func (uc *ButtonUsecase) UpdateButtonByID(
	ctx context.Context,
	buttonID uint32,
	permIDs []uint32,
	data map[string]any,
) (*ButtonModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	uc.log.Info(
		"开始更新按钮",
		zap.Uint32(ButtonIDKey, buttonID),
		zap.Uint32s(PermissionIDsKey, permIDs),
		zap.Any(database.UpdateDataKey, data),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	perms, err := uc.GetPermissions(ctx, permIDs)
	if err != nil {
		return nil, err
	}

	data["id"] = buttonID
	if err := uc.buttonRepo.UpdateModel(ctx, data, perms, "id = ?", buttonID); err != nil {
		uc.log.Error(
			"更新按钮失败",
			zap.Error(err),
			zap.Uint32(ButtonIDKey, buttonID),
			zap.Any(database.UpdateDataKey, data),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, data)
	}

	m, rErr := uc.FindButtonByID(ctx, []string{"Menu", "Permissions"}, buttonID)
	if rErr != nil {
		return nil, rErr
	}

	if err := uc.buttonRepo.RemoveGroupPolicy(ctx, m, false); err != nil {
		uc.log.Error(
			"移除旧按钮组策略失败",
			zap.Error(err),
			zap.Uint32(ButtonIDKey, buttonID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.FromError(err)
	}

	if err := uc.buttonRepo.AddGroupPolicy(ctx, m); err != nil {
		uc.log.Error(
			"添加新按钮组策略失败",
			zap.Error(err),
			zap.Uint32(ButtonIDKey, buttonID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.FromError(err)
	}

	uc.log.Info(
		"更新按钮成功",
		zap.Uint32(ButtonIDKey, buttonID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return m, nil
}

func (uc *ButtonUsecase) DeleteButtonByID(
	ctx context.Context,
	buttonID uint32,
) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	uc.log.Info(
		"开始删除按钮",
		zap.Uint32(ButtonIDKey, buttonID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, rErr := uc.FindButtonByID(ctx, []string{"Menu", "Permissions"}, buttonID)
	if rErr != nil {
		return rErr
	}

	if err := uc.buttonRepo.DeleteModel(ctx, buttonID); err != nil {
		uc.log.Error(
			"删除按钮失败",
			zap.Error(err),
			zap.Uint32(ButtonIDKey, buttonID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.NewGormError(err, map[string]any{"id": buttonID})
	}

	if err := uc.buttonRepo.RemoveGroupPolicy(ctx, m, true); err != nil {
		uc.log.Error(
			"移除按钮组策略失败",
			zap.Error(err),
			zap.Uint32(ButtonIDKey, buttonID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.FromError(err)
	}

	uc.log.Info(
		"删除按钮成功",
		zap.Uint32(ButtonIDKey, buttonID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return nil
}

func (uc *ButtonUsecase) FindButtonByID(
	ctx context.Context,
	preloads []string,
	buttonID uint32,
) (*ButtonModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	uc.log.Info(
		"开始查询按钮",
		zap.Strings(database.PreloadKey, preloads),
		zap.Uint32(ButtonIDKey, buttonID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, err := uc.buttonRepo.GetModel(ctx, preloads, buttonID)
	if err != nil {
		uc.log.Error(
			"查询按钮失败",
			zap.Error(err),
			zap.Uint32(ButtonIDKey, buttonID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, map[string]any{"id": buttonID})
	}

	uc.log.Info(
		"查询按钮成功",
		zap.Uint32(ButtonIDKey, buttonID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return m, nil
}

func (uc *ButtonUsecase) ListButton(
	ctx context.Context,
	qp database.QueryParams,
) (int64, *[]ButtonModel, *errors.Error) {
	if ctx.Err() != nil {
		return 0, nil, errors.FromError(ctx.Err())
	}

	uc.log.Info(
		"开始查询按钮列表",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	count, ms, err := uc.buttonRepo.ListModel(ctx, qp)
	if err != nil {
		uc.log.Error(
			"查询按钮列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}

	uc.log.Info(
		"查询按钮列表成功",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return count, ms, nil
}

func (uc *ButtonUsecase) LoadButtonPolicy(ctx context.Context) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	uc.log.Info(
		"开始加载按钮策略",
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	qp := database.QueryParams{
		Preloads: []string{"Permissions"},
		Columns:  []string{"id", "menu_id"},
	}

	_, bms, err := uc.ListButton(ctx, qp)
	if err != nil {
		uc.log.Error(
			"加载按钮策略时查询按钮列表失败",
			zap.Error(err),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return err
	}

	var policyCount int
	if bms != nil {
		ms := *bms
		policyCount = len(ms)
		for i := range ms {
			if err := uc.buttonRepo.AddGroupPolicy(ctx, &ms[i]); err != nil {
				uc.log.Error(
					"加载按钮策略失败",
					zap.Error(err),
					zap.Uint32(MenuIDKey, ms[i].ID),
					zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
				)
				return errors.FromError(err)
			}
		}
	}

	uc.log.Info(
		"加载按钮策略成功",
		zap.Int("policy_count", policyCount),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return nil
}
