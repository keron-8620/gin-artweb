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
	RoleTableName = "customer_role"
	RoleIDKey     = "role_id"
	RoleBase      = "role_base"
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
	return RoleTableName
}

func (m *RoleModel) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if m == nil {
		return database.GormModelIsNil(RoleTableName)
	}
	if err := m.StandardModel.MarshalLogObject(enc); err != nil {
		return err
	}
	enc.AddString("name", m.Name)
	enc.AddString("descr", m.Descr)
	enc.AddArray(PermissionIDsKey, zapcore.ArrayMarshalerFunc(func(ae zapcore.ArrayEncoder) error {
		for _, perm := range m.Permissions {
			ae.AppendUint32(perm.ID)
		}
		return nil
	}))
	enc.AddArray(MenuIDsKey, zapcore.ArrayMarshalerFunc(func(ae zapcore.ArrayEncoder) error {
		for _, menu := range m.Menus {
			ae.AppendUint32(menu.ID)
		}
		return nil
	}))
	enc.AddArray(ButtonIDsKey, zapcore.ArrayMarshalerFunc(func(ae zapcore.ArrayEncoder) error {
		for _, button := range m.Buttons {
			ae.AppendUint32(button.ID)
		}
		return nil
	}))
	return nil
}

type MenuTreeNode struct {
	MenuModel MenuModel
	Children  []*MenuTreeNode
	Buttons   []ButtonModel
}

type RoleRepo interface {
	CreateModel(context.Context, *RoleModel, *[]PermissionModel, *[]MenuModel, *[]ButtonModel) error
	UpdateModel(context.Context, map[string]any, *[]PermissionModel, *[]MenuModel, *[]ButtonModel, ...any) error
	DeleteModel(context.Context, ...any) error
	FindModel(context.Context, []string, ...any) (*RoleModel, error)
	ListModel(context.Context, database.QueryParams) (int64, *[]RoleModel, error)
	AddGroupPolicy(context.Context, *RoleModel) error
	RemoveGroupPolicy(context.Context, *RoleModel) error
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
) (*[]PermissionModel, *errors.Error) {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return nil, errors.FromError(err)
	}

	if len(permIDs) == 0 {
		return &[]PermissionModel{}, nil
	}

	uc.log.Info(
		"开始查询角色关联的权限列表",
		zap.Uint32s(PermissionIDsKey, permIDs),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)

	qp := database.QueryParams{
		Query: map[string]any{"id in ?": permIDs},
	}
	_, ms, err := uc.permRepo.ListModel(ctx, qp)
	if err != nil {
		uc.log.Error(
			"查询角色关联的权限列表失败",
			zap.Error(err),
			zap.Uint32s(PermissionIDsKey, permIDs),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		return nil, database.NewGormError(err, nil)
	}

	uc.log.Info(
		"查询角色关联的权限列表成功",
		zap.Uint32s(PermissionIDsKey, permIDs),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)
	return ms, nil
}

func (uc *RoleUsecase) GetMenus(
	ctx context.Context,
	menuIDs []uint32,
) (*[]MenuModel, *errors.Error) {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return nil, errors.FromError(err)
	}

	if len(menuIDs) == 0 {
		return &[]MenuModel{}, nil
	}

	uc.log.Info(
		"开始角色关联的菜单列表",
		zap.Uint32s(MenuIDsKey, menuIDs),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)

	qp := database.QueryParams{
		Query: map[string]any{"id in ?": menuIDs},
	}
	_, ms, err := uc.menuRepo.ListModel(ctx, qp)
	if err != nil {
		uc.log.Error(
			"查询角色关联的菜单列表失败",
			zap.Error(err),
			zap.Uint32s(MenuIDsKey, menuIDs),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		return nil, database.NewGormError(err, nil)
	}

	uc.log.Info(
		"查询角色关联的菜单列表成功",
		zap.Uint32s(MenuIDsKey, menuIDs),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)
	return ms, nil
}

func (uc *RoleUsecase) GetButtons(
	ctx context.Context,
	buttonIDs []uint32,
) (*[]ButtonModel, *errors.Error) {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return nil, errors.FromError(err)
	}

	if len(buttonIDs) == 0 {
		return &[]ButtonModel{}, nil
	}

	uc.log.Info(
		"开始查询角色关联的按钮列表",
		zap.Uint32s(ButtonIDsKey, buttonIDs),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)

	qp := database.QueryParams{
		Query: map[string]any{"id in ?": buttonIDs},
	}
	_, ms, err := uc.buttonRepo.ListModel(ctx, qp)
	if err != nil {
		uc.log.Error(
			"查询角色关联的按钮列表失败",
			zap.Error(err),
			zap.Uint32s(ButtonIDsKey, buttonIDs),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		return nil, database.NewGormError(err, nil)
	}

	uc.log.Info(
		"查询角色关联的按钮列表成功",
		zap.Uint32s(ButtonIDsKey, buttonIDs),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)
	return ms, nil
}

func (uc *RoleUsecase) CreateRole(
	ctx context.Context,
	permIDs []uint32,
	menuIDs []uint32,
	buttonIDs []uint32,
	m RoleModel,
) (*RoleModel, *errors.Error) {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return nil, errors.FromError(err)
	}

	uc.log.Info(
		"开始创建角色",
		zap.Uint32s(PermissionIDsKey, permIDs),
		zap.Object(database.ModelKey, &m),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)

	perms, err := uc.GetPermissions(ctx, permIDs)
	if err != nil {
		return nil, err
	}

	menus, err := uc.GetMenus(ctx, menuIDs)
	if err != nil {
		return nil, err
	}

	buttons, err := uc.GetButtons(ctx, buttonIDs)
	if err != nil {
		return nil, err
	}

	if err := uc.roleRepo.CreateModel(ctx, &m, perms, menus, buttons); err != nil {
		uc.log.Error(
			"创建角色失败",
			zap.Error(err),
			zap.Object(database.ModelKey, &m),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		return nil, database.NewGormError(err, nil)
	}

	if perms != nil {
		if len(*perms) > 0 {
			m.Permissions = *perms
		}
	}
	if menus != nil {
		if len(*menus) > 0 {
			m.Menus = *menus
		}
	}
	if buttons != nil {
		if len(*buttons) > 0 {
			m.Buttons = *buttons
		}
	}

	if err := uc.roleRepo.AddGroupPolicy(ctx, &m); err != nil {
		uc.log.Error(
			"添加角色组策略失败",
			zap.Error(err),
			zap.Object(database.ModelKey, &m),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		return nil, ErrAddGroupPolicy.WithCause(err)
	}

	uc.log.Info(
		"创建角色成功",
		zap.Object(database.ModelKey, &m),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)
	return &m, nil
}

func (uc *RoleUsecase) UpdateRoleByID(
	ctx context.Context,
	roleID uint32,
	permIDs []uint32,
	menuIDs []uint32,
	buttonIDs []uint32,
	data map[string]any,
) (*RoleModel, *errors.Error) {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return nil, errors.FromError(err)
	}

	uc.log.Info(
		"开始更新角色",
		zap.Uint32(RoleIDKey, roleID),
		zap.Uint32s(PermissionIDsKey, permIDs),
		zap.Uint32s(MenuIDsKey, menuIDs),
		zap.Uint32s(ButtonIDsKey, buttonIDs),
		zap.Any(database.UpdateDataKey, data),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)

	perms, err := uc.GetPermissions(ctx, permIDs)
	if err != nil {
		return nil, err
	}

	menus, err := uc.GetMenus(ctx, menuIDs)
	if err != nil {
		return nil, err
	}

	buttons, err := uc.GetButtons(ctx, buttonIDs)
	if err != nil {
		return nil, err
	}

	data["id"] = roleID
	if err := uc.roleRepo.UpdateModel(ctx, data, perms, menus, buttons, "id = ?", roleID); err != nil {
		uc.log.Error(
			"更新角色失败",
			zap.Error(err),
			zap.Uint32(RoleIDKey, roleID),
			zap.Any(database.UpdateDataKey, data),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		return nil, database.NewGormError(err, data)
	}

	m, rErr := uc.FindRoleByID(ctx, []string{"Permissions", "Menus", "Buttons"}, roleID)
	if rErr != nil {
		return nil, rErr
	}

	if err := uc.roleRepo.RemoveGroupPolicy(ctx, m); err != nil {
		uc.log.Error(
			"移除旧角色组策略失败",
			zap.Error(err),
			zap.Uint32(RoleIDKey, roleID),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		return nil, ErrRemoveGroupPolicy.WithCause(err)
	}

	if err := uc.roleRepo.AddGroupPolicy(ctx, m); err != nil {
		uc.log.Error(
			"添加新角色组策略失败",
			zap.Error(err),
			zap.Uint32(RoleIDKey, roleID),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		return nil, ErrAddGroupPolicy.WithCause(err)
	}

	uc.log.Info(
		"更新角色成功",
		zap.Uint32(RoleIDKey, roleID),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)
	return m, nil
}

func (uc *RoleUsecase) DeleteRoleByID(
	ctx context.Context,
	roleID uint32,
) *errors.Error {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return errors.FromError(err)
	}

	uc.log.Info(
		"开始删除角色",
		zap.Uint32(RoleIDKey, roleID),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)

	m, rErr := uc.FindRoleByID(ctx, []string{"Permissions", "Menus", "Buttons"}, roleID)
	if rErr != nil {
		return rErr
	}

	if err := uc.roleRepo.DeleteModel(ctx, roleID); err != nil {
		uc.log.Error(
			"删除角色失败",
			zap.Error(err),
			zap.Uint32(RoleIDKey, roleID),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		return database.NewGormError(err, map[string]any{"id": roleID})
	}

	if err := uc.roleRepo.RemoveGroupPolicy(ctx, m); err != nil {
		uc.log.Error(
			"移除角色组策略失败",
			zap.Error(err),
			zap.Uint32(RoleIDKey, roleID),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		return ErrRemoveGroupPolicy.WithCause(err)
	}

	uc.log.Info(
		"删除角色成功",
		zap.Uint32(RoleIDKey, roleID),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)
	return nil
}

func (uc *RoleUsecase) FindRoleByID(
	ctx context.Context,
	preloads []string,
	roleID uint32,
) (*RoleModel, *errors.Error) {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return nil, errors.FromError(err)
	}

	uc.log.Info(
		"开始查询角色",
		zap.Strings(database.PreloadKey, preloads),
		zap.Uint32(RoleIDKey, roleID),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)

	m, err := uc.roleRepo.FindModel(ctx, preloads, roleID)
	if err != nil {
		uc.log.Error(
			"查询角色失败",
			zap.Error(err),
			zap.Uint32(RoleIDKey, roleID),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		return nil, database.NewGormError(err, map[string]any{"id": roleID})
	}

	uc.log.Info(
		"查询角色成功",
		zap.Uint32(RoleIDKey, roleID),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)
	return m, nil
}

func (uc *RoleUsecase) ListRole(
	ctx context.Context,
	qp database.QueryParams,
) (int64, *[]RoleModel, *errors.Error) {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return 0, nil, errors.FromError(err)
	}

	uc.log.Info(
		"开始查询角色列表",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)

	count, ms, err := uc.roleRepo.ListModel(ctx, qp)
	if err != nil {
		uc.log.Error(
			"查询角色列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		return 0, nil, database.NewGormError(err, nil)
	}

	uc.log.Info(
		"查询角色列表成功",
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)
	return count, ms, nil
}

func (uc *RoleUsecase) LoadRolePolicy(ctx context.Context) *errors.Error {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return errors.FromError(err)
	}

	uc.log.Info(
		"开始加载角色策略",
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)

	qp := database.QueryParams{
		Preloads: []string{"Permissions", "Menus", "Buttons"},
		Columns:  []string{"id"},
	}

	_, rms, err := uc.ListRole(ctx, qp)
	if err != nil {
		uc.log.Error(
			"加载角色策略时查询角色列表失败",
			zap.Error(err),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		return err
	}

	var policyCount int
	if rms != nil {
		ms := *rms
		policyCount = len(ms)
		for i := range ms {
			if err := uc.roleRepo.AddGroupPolicy(ctx, &ms[i]); err != nil {
				uc.log.Error(
					"加载角色策略失败",
					zap.Error(err),
					zap.Uint32(MenuIDKey, ms[i].ID),
					zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
				)
				return ErrAddGroupPolicy.WithCause(err)
			}
		}
	}

	uc.log.Info(
		"加载角色策略成功",
		zap.Int("policy_count", policyCount),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)
	return nil
}

func (uc *RoleUsecase) GetRoleMenuTree(
	ctx context.Context,
	roleID uint32,
) ([]*MenuTreeNode, *errors.Error) {
	m, rErr := uc.FindRoleByID(ctx, []string{"Permissions", "Menus", "Buttons"}, roleID)
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
