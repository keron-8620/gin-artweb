package biz

import (
	"context"

	"go.uber.org/zap"

	"gin-artweb/internal/infra/customer/data"
	"gin-artweb/internal/infra/customer/model"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
)

type MenuTreeNode struct {
	Menu     model.MenuModel
	Children []*MenuTreeNode
	Buttons  []model.ButtonModel
}

type RoleUsecase struct {
	log        *zap.Logger
	apiRepo    *data.ApiRepo
	menuRepo   *data.MenuRepo
	buttonRepo *data.ButtonRepo
	roleRepo   *data.RoleRepo
}

func NewRoleUsecase(
	log *zap.Logger,
	apiRepo *data.ApiRepo,
	menuRepo *data.MenuRepo,
	buttonRepo *data.ButtonRepo,
	roleRepo *data.RoleRepo,
) *RoleUsecase {
	return &RoleUsecase{
		log:        log,
		apiRepo:    apiRepo,
		menuRepo:   menuRepo,
		buttonRepo: buttonRepo,
		roleRepo:   roleRepo,
	}
}

func (uc *RoleUsecase) GetApis(
	ctx context.Context,
	apiIDs []uint32,
) (*[]model.ApiModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	if len(apiIDs) == 0 {
		return &[]model.ApiModel{}, nil
	}

	uc.log.Info(
		"开始查询角色关联的API列表",
		zap.Uint32s("api_ids", apiIDs),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	qp := database.QueryParams{
		Query: map[string]any{"id in ?": apiIDs},
	}
	_, ms, err := uc.apiRepo.ListModel(ctx, qp)
	if err != nil {
		uc.log.Error(
			"查询角色关联的API列表失败",
			zap.Error(err),
			zap.Uint32s("api_ids", apiIDs),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, nil)
	}

	uc.log.Info(
		"查询角色关联的API列表成功",
		zap.Uint32s("api_ids", apiIDs),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return ms, nil
}

func (uc *RoleUsecase) GetMenus(
	ctx context.Context,
	menuIDs []uint32,
) (*[]model.MenuModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	if len(menuIDs) == 0 {
		return &[]model.MenuModel{}, nil
	}

	uc.log.Info(
		"开始角色关联的菜单列表",
		zap.Uint32s("menu_ids", menuIDs),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	qp := database.QueryParams{
		Query: map[string]any{"id in ?": menuIDs},
	}
	_, ms, err := uc.menuRepo.ListModel(ctx, qp)
	if err != nil {
		uc.log.Error(
			"查询角色关联的菜单列表失败",
			zap.Error(err),
			zap.Uint32s("menu_ids", menuIDs),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, nil)
	}

	uc.log.Info(
		"查询角色关联的菜单列表成功",
		zap.Uint32s("menu_ids", menuIDs),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return ms, nil
}

func (uc *RoleUsecase) GetButtons(
	ctx context.Context,
	buttonIDs []uint32,
) (*[]model.ButtonModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	if len(buttonIDs) == 0 {
		return &[]model.ButtonModel{}, nil
	}

	uc.log.Info(
		"开始查询角色关联的按钮列表",
		zap.Uint32s("button_ids", buttonIDs),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	qp := database.QueryParams{
		Query: map[string]any{"id in ?": buttonIDs},
	}
	_, ms, err := uc.buttonRepo.ListModel(ctx, qp)
	if err != nil {
		uc.log.Error(
			"查询角色关联的按钮列表失败",
			zap.Error(err),
			zap.Uint32s("button_ids", buttonIDs),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, nil)
	}

	uc.log.Info(
		"查询角色关联的按钮列表成功",
		zap.Uint32s("button_ids", buttonIDs),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return ms, nil
}

func (uc *RoleUsecase) CreateRole(
	ctx context.Context,
	apiIDs []uint32,
	menuIDs []uint32,
	buttonIDs []uint32,
	m model.RoleModel,
) (*model.RoleModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	uc.log.Info(
		"开始创建角色",
		zap.Uint32s("api_ids", apiIDs),
		zap.Object(database.ModelKey, &m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	var (
		apis    *[]model.ApiModel
		menus   *[]model.MenuModel
		buttons *[]model.ButtonModel
		rErr    *errors.Error
	)

	apis, rErr = uc.GetApis(ctx, apiIDs)
	if rErr != nil {
		return nil, rErr
	}

	menus, rErr = uc.GetMenus(ctx, menuIDs)
	if rErr != nil {
		return nil, rErr
	}

	buttons, rErr = uc.GetButtons(ctx, buttonIDs)
	if rErr != nil {
		return nil, rErr
	}

	if err := uc.roleRepo.CreateModel(ctx, &m, apis, menus, buttons); err != nil {
		uc.log.Error(
			"创建角色失败",
			zap.Error(err),
			zap.Object(database.ModelKey, &m),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, nil)
	}

	if apis != nil {
		if len(*apis) > 0 {
			m.Apis = *apis
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
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.FromError(err)
	}

	uc.log.Info(
		"创建角色成功",
		zap.Object(database.ModelKey, &m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return &m, nil
}

func (uc *RoleUsecase) UpdateRoleByID(
	ctx context.Context,
	roleID uint32,
	apiIDs []uint32,
	menuIDs []uint32,
	buttonIDs []uint32,
	data map[string]any,
) (*model.RoleModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	uc.log.Info(
		"开始更新角色",
		zap.Uint32("role_id", roleID),
		zap.Uint32s("api_ids", apiIDs),
		zap.Uint32s("menu_ids", menuIDs),
		zap.Uint32s("button_ids", buttonIDs),
		zap.Any(database.UpdateDataKey, data),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	var (
		apis    *[]model.ApiModel
		menus   *[]model.MenuModel
		buttons *[]model.ButtonModel
		rErr    *errors.Error
	)

	apis, rErr = uc.GetApis(ctx, apiIDs)
	if rErr != nil {
		return nil, rErr
	}

	menus, rErr = uc.GetMenus(ctx, menuIDs)
	if rErr != nil {
		return nil, rErr
	}

	buttons, rErr = uc.GetButtons(ctx, buttonIDs)
	if rErr != nil {
		return nil, rErr
	}

	data["id"] = roleID
	if err := uc.roleRepo.UpdateModel(ctx, data, apis, menus, buttons, "id = ?", roleID); err != nil {
		uc.log.Error(
			"更新角色失败",
			zap.Error(err),
			zap.Uint32("role_id", roleID),
			zap.Any(database.UpdateDataKey, data),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, data)
	}

	var m *model.RoleModel
	m, rErr = uc.FindRoleByID(ctx, []string{"Apis", "Menus", "Buttons"}, roleID)
	if rErr != nil {
		return nil, rErr
	}

	if err := uc.roleRepo.RemoveGroupPolicy(ctx, m); err != nil {
		uc.log.Error(
			"移除旧角色组策略失败",
			zap.Error(err),
			zap.Uint32("role_id", roleID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.FromError(err)
	}

	if err := uc.roleRepo.AddGroupPolicy(ctx, m); err != nil {
		uc.log.Error(
			"添加新角色组策略失败",
			zap.Error(err),
			zap.Uint32("role_id", roleID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.FromError(err)
	}

	uc.log.Info(
		"更新角色成功",
		zap.Uint32("role_id", roleID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return m, nil
}

func (uc *RoleUsecase) DeleteRoleByID(
	ctx context.Context,
	roleID uint32,
) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	uc.log.Info(
		"开始删除角色",
		zap.Uint32("role_id", roleID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, rErr := uc.FindRoleByID(ctx, []string{"Apis", "Menus", "Buttons"}, roleID)
	if rErr != nil {
		return rErr
	}

	if err := uc.roleRepo.DeleteModel(ctx, roleID); err != nil {
		uc.log.Error(
			"删除角色失败",
			zap.Error(err),
			zap.Uint32("role_id", roleID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.NewGormError(err, map[string]any{"id": roleID})
	}

	if err := uc.roleRepo.RemoveGroupPolicy(ctx, m); err != nil {
		uc.log.Error(
			"移除角色组策略失败",
			zap.Error(err),
			zap.Uint32("role_id", roleID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.FromError(err)
	}

	uc.log.Info(
		"删除角色成功",
		zap.Uint32("role_id", roleID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return nil
}

func (uc *RoleUsecase) FindRoleByID(
	ctx context.Context,
	preloads []string,
	roleID uint32,
) (*model.RoleModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	uc.log.Info(
		"开始查询角色",
		zap.Strings(database.PreloadKey, preloads),
		zap.Uint32("role_id", roleID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, err := uc.roleRepo.GetModel(ctx, preloads, roleID)
	if err != nil {
		uc.log.Error(
			"查询角色失败",
			zap.Error(err),
			zap.Uint32("role_id", roleID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, map[string]any{"id": roleID})
	}

	uc.log.Info(
		"查询角色成功",
		zap.Uint32("role_id", roleID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return m, nil
}

func (uc *RoleUsecase) ListRole(
	ctx context.Context,
	qp database.QueryParams,
) (int64, *[]model.RoleModel, *errors.Error) {
	if ctx.Err() != nil {
		return 0, nil, errors.FromError(ctx.Err())
	}

	uc.log.Info(
		"开始查询角色列表",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	count, ms, err := uc.roleRepo.ListModel(ctx, qp)
	if err != nil {
		uc.log.Error(
			"查询角色列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}

	uc.log.Info(
		"查询角色列表成功",
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return count, ms, nil
}

func (uc *RoleUsecase) LoadRolePolicy(ctx context.Context) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	uc.log.Info(
		"开始加载角色策略",
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	qp := database.QueryParams{
		Preloads: []string{"Apis", "Menus", "Buttons"},
		Columns:  []string{"id"},
	}

	_, rms, rErr := uc.ListRole(ctx, qp)
	if rErr != nil {
		uc.log.Error(
			"加载角色策略时查询角色列表失败",
			zap.Error(rErr),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return rErr
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
					zap.Uint32("role_id", ms[i].ID),
					zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
				)
				return errors.FromError(err)
			}
		}
	}

	uc.log.Info(
		"加载角色策略成功",
		zap.Int("policy_count", policyCount),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return nil
}

func (uc *RoleUsecase) GetRoleMenuTree(
	ctx context.Context,
	roleID uint32,
) ([]*MenuTreeNode, *errors.Error) {
	m, rErr := uc.FindRoleByID(ctx, []string{"Apis", "Menus", "Buttons"}, roleID)
	if rErr != nil {
		return nil, rErr
	}
	roleMenuMap := make(map[uint32]model.MenuModel)
	for _, menu := range m.Menus {
		roleMenuMap[menu.ID] = menu
	}
	roleButtonMap := make(map[uint32]model.ButtonModel)
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
	m model.MenuModel,
	mp map[uint32]model.MenuModel,
	bp map[uint32]model.ButtonModel,
) (*MenuTreeNode, *errors.Error) {
	var children []model.MenuModel
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
	var buttons []model.ButtonModel
	for _, button := range bp {
		if button.MenuID == m.ID {
			buttons = append(buttons, button)
		}
	}
	return &MenuTreeNode{
		Menu:     m,
		Children: childTrees,
		Buttons:  buttons,
	}, nil
}
