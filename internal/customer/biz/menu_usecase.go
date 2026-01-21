package biz

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
	"gin-artweb/pkg/ctxutil"
)

const (
	MenuTableName = "customer_menu"
	MenuIDKey     = "menu_id"
	MenuIDsKey    = "menu_ids"
)

type Meta struct {
	Title string `json:"title"`
	Icon  string `json:"icon"`
}

func (m *Meta) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("title", m.Title)
	enc.AddString("icon", m.Icon)
	return nil
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
	return MenuTableName
}

func (m *MenuModel) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if m == nil {
		return database.GormModelIsNil(MenuTableName)
	}
	if err := m.StandardModel.MarshalLogObject(enc); err != nil {
		return err
	}
	enc.AddString("path", m.Path)
	enc.AddString("component", m.Component)
	enc.AddObject("meta", &m.Meta)
	enc.AddString("name", m.Name)
	enc.AddUint32("arrange_order", m.ArrangeOrder)
	enc.AddBool("is_active", m.IsActive)
	enc.AddString("descr", m.Descr)
	if m.ParentID != nil {
		enc.AddUint32("parent_id", *m.ParentID)
	}
	enc.AddArray(PermissionIDsKey, zapcore.ArrayMarshalerFunc(func(ae zapcore.ArrayEncoder) error {
		for _, perm := range m.Permissions {
			ae.AppendUint32(perm.ID)
		}
		return nil
	}))
	return nil
}

func ListMenuModelToUint32s(mms *[]MenuModel) []uint32 {
	if mms == nil {
		return []uint32{}
	}
	ms := *mms
	if len(ms) == 0 {
		return []uint32{}
	}

	ids := make([]uint32, len(ms))
	for i, m := range ms {
		ids[i] = m.ID
	}
	return ids
}

type MenuRepo interface {
	CreateModel(context.Context, *MenuModel, *[]PermissionModel) error
	UpdateModel(context.Context, map[string]any, *[]PermissionModel, ...any) error
	DeleteModel(context.Context, ...any) error
	FindModel(context.Context, []string, ...any) (*MenuModel, error)
	ListModel(context.Context, database.QueryParams) (int64, *[]MenuModel, error)
	AddGroupPolicy(context.Context, *MenuModel) error
	RemoveGroupPolicy(context.Context, *MenuModel, bool) error
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
	if err := ctxutil.CheckContext(ctx); err != nil {
		return nil, errors.FromError(err)
	}

	if parentID == nil || *parentID == 0 {
		return nil, nil
	}

	uc.log.Info(
		"开始查询父菜单",
		zap.Uint32("parent_id", *parentID),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)

	m, err := uc.menuRepo.FindModel(ctx, nil, *parentID)
	if err != nil {
		uc.log.Error(
			"查询父菜单失败",
			zap.Error(err),
			zap.Uint32("parent_id", *parentID),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		return nil, database.NewGormError(err, map[string]any{"parent_id": *parentID})
	}

	uc.log.Info(
		"查询父菜单成功",
		zap.Uint32("parent_id", *parentID),
		zap.Object(database.ModelKey, m),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)
	return m, nil
}

func (uc *MenuUsecase) GetPermissions(
	ctx context.Context,
	permIDs []uint32,
) (*[]PermissionModel, *errors.Error) {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return nil, errors.FromError(err)
	}

	if len(permIDs) == 0 {
		return &[]PermissionModel{}, nil
	}

	uc.log.Info(
		"开始查询菜单关联的权限列表",
		zap.Uint32s(PermissionIDsKey, permIDs),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)

	qp := database.QueryParams{
		Query: map[string]any{"id in ?": permIDs},
	}
	_, ms, err := uc.permRepo.ListModel(ctx, qp)
	if err != nil {
		uc.log.Error(
			"查询菜单关联的权限列表失败",
			zap.Error(err),
			zap.Uint32s(PermissionIDsKey, permIDs),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		return nil, database.NewGormError(err, nil)
	}

	uc.log.Info(
		"查询菜单关联的权限列表成功",
		zap.Uint32s(PermissionIDsKey, permIDs),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)
	return ms, nil
}

func (uc *MenuUsecase) CreateMenu(
	ctx context.Context,
	permIDs []uint32,
	m MenuModel,
) (*MenuModel, *errors.Error) {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return nil, errors.FromError(err)
	}

	uc.log.Info(
		"开始创建菜单",
		zap.Uint32s(PermissionIDsKey, permIDs),
		zap.Object(database.ModelKey, &m),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)

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

	if err := uc.menuRepo.CreateModel(ctx, &m, perms); err != nil {
		uc.log.Error(
			"创建菜单失败",
			zap.Error(err),
			zap.Object(database.ModelKey, &m),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		return nil, database.NewGormError(err, nil)
	}

	if perms != nil && len(*perms) > 0 {
		m.Permissions = *perms
	}

	if err := uc.menuRepo.AddGroupPolicy(ctx, &m); err != nil {
		uc.log.Error(
			"添加菜单组策略失败",
			zap.Error(err),
			zap.Object(database.ModelKey, &m),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		return nil, ErrAddGroupPolicy.WithCause(err)
	}

	uc.log.Info(
		"创建菜单成功",
		zap.Object(database.ModelKey, &m),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)
	return &m, nil
}

func (uc *MenuUsecase) UpdateMenuByID(
	ctx context.Context,
	menuID uint32,
	permIDs []uint32,
	data map[string]any,
) (*MenuModel, *errors.Error) {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return nil, errors.FromError(err)
	}

	uc.log.Info(
		"开始更新菜单",
		zap.Uint32(MenuIDKey, menuID),
		zap.Uint32s(PermissionIDsKey, permIDs),
		zap.Any(database.UpdateDataKey, data),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)

	perms, err := uc.GetPermissions(ctx, permIDs)
	if err != nil {
		return nil, err
	}

	data["id"] = menuID
	if err := uc.menuRepo.UpdateModel(ctx, data, perms, "id = ?", menuID); err != nil {
		uc.log.Error(
			"更新菜单失败",
			zap.Error(err),
			zap.Uint32(MenuIDKey, menuID),
			zap.Any(database.UpdateDataKey, data),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		return nil, database.NewGormError(err, data)
	}

	m, rErr := uc.FindMenuByID(ctx, []string{"Parent", "Permissions"}, menuID)
	if rErr != nil {
		return nil, rErr
	}
	if err := uc.menuRepo.RemoveGroupPolicy(ctx, m, false); err != nil {
		uc.log.Error(
			"移除旧菜单组策略失败",
			zap.Error(err),
			zap.Uint32(MenuIDKey, menuID),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		return nil, ErrRemoveGroupPolicy.WithCause(err)
	}

	if err := uc.menuRepo.AddGroupPolicy(ctx, m); err != nil {
		uc.log.Error(
			"添加新菜单组策略失败",
			zap.Error(err),
			zap.Uint32(MenuIDKey, menuID),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		return nil, ErrAddGroupPolicy.WithCause(err)
	}

	uc.log.Info(
		"更新菜单成功",
		zap.Uint32(MenuIDKey, menuID),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)
	return m, nil
}

func (uc *MenuUsecase) DeleteMenuByID(
	ctx context.Context,
	menuID uint32,
) *errors.Error {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return errors.FromError(err)
	}

	uc.log.Info(
		"开始删除菜单",
		zap.Uint32(MenuIDKey, menuID),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)

	m, rErr := uc.FindMenuByID(ctx, []string{"Parent", "Permissions"}, menuID)
	if rErr != nil {
		return rErr
	}

	if err := uc.menuRepo.DeleteModel(ctx, menuID); err != nil {
		uc.log.Error(
			"删除菜单失败",
			zap.Error(err),
			zap.Uint32(MenuIDKey, menuID),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		return database.NewGormError(err, map[string]any{"id": menuID})
	}

	if err := uc.menuRepo.RemoveGroupPolicy(ctx, m, true); err != nil {
		uc.log.Error(
			"移除菜单组策略失败",
			zap.Error(err),
			zap.Uint32(MenuIDKey, menuID),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		return ErrRemoveGroupPolicy.WithCause(err)
	}

	uc.log.Info(
		"删除菜单成功",
		zap.Uint32(MenuIDKey, menuID),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)
	return nil
}

func (uc *MenuUsecase) FindMenuByID(
	ctx context.Context,
	preloads []string,
	menuID uint32,
) (*MenuModel, *errors.Error) {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return nil, errors.FromError(err)
	}

	uc.log.Info(
		"开始查询菜单",
		zap.Strings(database.PreloadKey, preloads),
		zap.Uint32(MenuIDKey, menuID),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)

	m, err := uc.menuRepo.FindModel(ctx, preloads, menuID)
	if err != nil {
		uc.log.Error(
			"查询菜单失败",
			zap.Error(err),
			zap.Uint32(MenuIDKey, menuID),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		return nil, database.NewGormError(err, map[string]any{"id": menuID})
	}

	uc.log.Info(
		"查询菜单成功",
		zap.Uint32(MenuIDKey, menuID),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)
	return m, nil
}

func (uc *MenuUsecase) ListMenu(
	ctx context.Context,
	qp database.QueryParams,
) (int64, *[]MenuModel, *errors.Error) {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return 0, nil, errors.FromError(err)
	}

	uc.log.Info(
		"开始查询菜单列表",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)

	count, ms, err := uc.menuRepo.ListModel(ctx, qp)
	if err != nil {
		uc.log.Error(
			"查询菜单列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		return 0, nil, database.NewGormError(err, nil)
	}

	uc.log.Info(
		"查询菜单列表成功",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)
	return count, ms, nil
}

func (uc *MenuUsecase) LoadMenuPolicy(ctx context.Context) *errors.Error {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return errors.FromError(err)
	}

	uc.log.Info(
		"开始加载菜单策略",
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)

	qp := database.QueryParams{
		Preloads: []string{"Permissions"},
		Columns:  []string{"id", "parent_id"},
	}
	_, mms, err := uc.ListMenu(ctx, qp)
	if err != nil {
		uc.log.Error(
			"加载菜单策略时查询菜单列表失败",
			zap.Error(err),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		return err
	}

	var policyCount int
	if mms != nil {
		ms := *mms
		policyCount = len(ms)
		for i := range ms {
			if err := uc.menuRepo.AddGroupPolicy(ctx, &ms[i]); err != nil {
				uc.log.Error(
					"加载菜单策略失败",
					zap.Error(err),
					zap.Uint32(MenuIDKey, ms[i].ID),
					zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
				)
				return ErrAddGroupPolicy.WithCause(err)
			}
		}
	}
	uc.log.Info(
		"加载菜单策略成功",
		zap.Int("policy_count", policyCount),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)
	return nil
}
