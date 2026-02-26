package customer

import (
	"context"

	"go.uber.org/zap"

	custmodel "gin-artweb/internal/model/customer"
	custsvc "gin-artweb/internal/repository/customer"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
)

type RoleService struct {
	log        *zap.Logger
	apiRepo    *custsvc.ApiRepo
	menuRepo   *custsvc.MenuRepo
	buttonRepo *custsvc.ButtonRepo
	roleRepo   *custsvc.RoleRepo
}

func NewRoleService(
	log *zap.Logger,
	apiRepo *custsvc.ApiRepo,
	menuRepo *custsvc.MenuRepo,
	buttonRepo *custsvc.ButtonRepo,
	roleRepo *custsvc.RoleRepo,
) *RoleService {
	return &RoleService{
		log:        log,
		apiRepo:    apiRepo,
		menuRepo:   menuRepo,
		buttonRepo: buttonRepo,
		roleRepo:   roleRepo,
	}
}

func (s *RoleService) GetApis(
	ctx context.Context,
	apiIDs []uint32,
) (*[]custmodel.ApiModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	if len(apiIDs) == 0 {
		return &[]custmodel.ApiModel{}, nil
	}

	s.log.Info(
		"开始查询角色关联的API列表",
		zap.Uint32s("api_ids", apiIDs),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	qp := database.QueryParams{
		Query: map[string]any{"id in ?": apiIDs},
	}
	_, ms, err := s.apiRepo.ListModel(ctx, qp)
	if err != nil {
		s.log.Error(
			"查询角色关联的API列表失败",
			zap.Error(err),
			zap.Uint32s("api_ids", apiIDs),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, nil)
	}

	s.log.Info(
		"查询角色关联的API列表成功",
		zap.Uint32s("api_ids", apiIDs),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return ms, nil
}

func (s *RoleService) GetMenus(
	ctx context.Context,
	menuIDs []uint32,
) (*[]custmodel.MenuModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	if len(menuIDs) == 0 {
		return &[]custmodel.MenuModel{}, nil
	}

	s.log.Info(
		"开始角色关联的菜单列表",
		zap.Uint32s("menu_ids", menuIDs),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	qp := database.QueryParams{
		Query: map[string]any{"id in ?": menuIDs},
	}
	_, ms, err := s.menuRepo.ListModel(ctx, qp)
	if err != nil {
		s.log.Error(
			"查询角色关联的菜单列表失败",
			zap.Error(err),
			zap.Uint32s("menu_ids", menuIDs),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, nil)
	}

	s.log.Info(
		"查询角色关联的菜单列表成功",
		zap.Uint32s("menu_ids", menuIDs),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return ms, nil
}

func (s *RoleService) GetButtons(
	ctx context.Context,
	buttonIDs []uint32,
) (*[]custmodel.ButtonModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	if len(buttonIDs) == 0 {
		return &[]custmodel.ButtonModel{}, nil
	}

	s.log.Info(
		"开始查询角色关联的按钮列表",
		zap.Uint32s("button_ids", buttonIDs),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	qp := database.QueryParams{
		Query: map[string]any{"id in ?": buttonIDs},
	}
	_, ms, err := s.buttonRepo.ListModel(ctx, qp)
	if err != nil {
		s.log.Error(
			"查询角色关联的按钮列表失败",
			zap.Error(err),
			zap.Uint32s("button_ids", buttonIDs),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, nil)
	}

	s.log.Info(
		"查询角色关联的按钮列表成功",
		zap.Uint32s("button_ids", buttonIDs),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return ms, nil
}

func (s *RoleService) CreateRole(
	ctx context.Context,
	apiIDs []uint32,
	menuIDs []uint32,
	buttonIDs []uint32,
	m custmodel.RoleModel,
) (*custmodel.RoleModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始创建角色",
		zap.Uint32s("api_ids", apiIDs),
		zap.Object(database.ModelKey, &m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	var (
		apis    *[]custmodel.ApiModel
		menus   *[]custmodel.MenuModel
		buttons *[]custmodel.ButtonModel
		rErr    *errors.Error
	)

	apis, rErr = s.GetApis(ctx, apiIDs)
	if rErr != nil {
		return nil, rErr
	}

	menus, rErr = s.GetMenus(ctx, menuIDs)
	if rErr != nil {
		return nil, rErr
	}

	buttons, rErr = s.GetButtons(ctx, buttonIDs)
	if rErr != nil {
		return nil, rErr
	}

	if err := s.roleRepo.CreateModel(ctx, &m, apis, menus, buttons); err != nil {
		s.log.Error(
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

	if err := s.roleRepo.AddGroupPolicy(ctx, &m); err != nil {
		s.log.Error(
			"添加角色组策略失败",
			zap.Error(err),
			zap.Object(database.ModelKey, &m),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.FromError(err)
	}

	s.log.Info(
		"创建角色成功",
		zap.Object(database.ModelKey, &m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return &m, nil
}

func (s *RoleService) UpdateRoleByID(
	ctx context.Context,
	roleID uint32,
	apiIDs []uint32,
	menuIDs []uint32,
	buttonIDs []uint32,
	data map[string]any,
) (*custmodel.RoleModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始更新角色",
		zap.Uint32("role_id", roleID),
		zap.Uint32s("api_ids", apiIDs),
		zap.Uint32s("menu_ids", menuIDs),
		zap.Uint32s("button_ids", buttonIDs),
		zap.Any(database.UpdateDataKey, data),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	var (
		apis    *[]custmodel.ApiModel
		menus   *[]custmodel.MenuModel
		buttons *[]custmodel.ButtonModel
		rErr    *errors.Error
	)

	apis, rErr = s.GetApis(ctx, apiIDs)
	if rErr != nil {
		return nil, rErr
	}

	menus, rErr = s.GetMenus(ctx, menuIDs)
	if rErr != nil {
		return nil, rErr
	}

	buttons, rErr = s.GetButtons(ctx, buttonIDs)
	if rErr != nil {
		return nil, rErr
	}

	data["id"] = roleID
	if err := s.roleRepo.UpdateModel(ctx, data, apis, menus, buttons, "id = ?", roleID); err != nil {
		s.log.Error(
			"更新角色失败",
			zap.Error(err),
			zap.Uint32("role_id", roleID),
			zap.Any(database.UpdateDataKey, data),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, data)
	}

	var m *custmodel.RoleModel
	m, rErr = s.FindRoleByID(ctx, []string{"Apis", "Menus", "Buttons"}, roleID)
	if rErr != nil {
		return nil, rErr
	}

	if err := s.roleRepo.RemoveGroupPolicy(ctx, m); err != nil {
		s.log.Error(
			"移除旧角色组策略失败",
			zap.Error(err),
			zap.Uint32("role_id", roleID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.FromError(err)
	}

	if err := s.roleRepo.AddGroupPolicy(ctx, m); err != nil {
		s.log.Error(
			"添加新角色组策略失败",
			zap.Error(err),
			zap.Uint32("role_id", roleID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.FromError(err)
	}

	s.log.Info(
		"更新角色成功",
		zap.Uint32("role_id", roleID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return m, nil
}

func (s *RoleService) DeleteRoleByID(
	ctx context.Context,
	roleID uint32,
) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始删除角色",
		zap.Uint32("role_id", roleID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, rErr := s.FindRoleByID(ctx, []string{"Apis", "Menus", "Buttons"}, roleID)
	if rErr != nil {
		return rErr
	}

	if err := s.roleRepo.DeleteModel(ctx, roleID); err != nil {
		s.log.Error(
			"删除角色失败",
			zap.Error(err),
			zap.Uint32("role_id", roleID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.NewGormError(err, map[string]any{"id": roleID})
	}

	if err := s.roleRepo.RemoveGroupPolicy(ctx, m); err != nil {
		s.log.Error(
			"移除角色组策略失败",
			zap.Error(err),
			zap.Uint32("role_id", roleID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.FromError(err)
	}

	s.log.Info(
		"删除角色成功",
		zap.Uint32("role_id", roleID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return nil
}

func (s *RoleService) FindRoleByID(
	ctx context.Context,
	preloads []string,
	roleID uint32,
) (*custmodel.RoleModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始查询角色",
		zap.Strings(database.PreloadKey, preloads),
		zap.Uint32("role_id", roleID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, err := s.roleRepo.GetModel(ctx, preloads, roleID)
	if err != nil {
		s.log.Error(
			"查询角色失败",
			zap.Error(err),
			zap.Uint32("role_id", roleID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, map[string]any{"id": roleID})
	}

	s.log.Info(
		"查询角色成功",
		zap.Uint32("role_id", roleID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return m, nil
}

func (s *RoleService) ListRole(
	ctx context.Context,
	qp database.QueryParams,
) (int64, *[]custmodel.RoleModel, *errors.Error) {
	if ctx.Err() != nil {
		return 0, nil, errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始查询角色列表",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	count, ms, err := s.roleRepo.ListModel(ctx, qp)
	if err != nil {
		s.log.Error(
			"查询角色列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}

	s.log.Info(
		"查询角色列表成功",
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return count, ms, nil
}

func (s *RoleService) LoadRolePolicy(ctx context.Context) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始加载角色策略",
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	qp := database.QueryParams{
		Preloads: []string{"Apis", "Menus", "Buttons"},
		Columns:  []string{"id"},
	}

	_, rms, rErr := s.ListRole(ctx, qp)
	if rErr != nil {
		s.log.Error(
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
			if err := s.roleRepo.AddGroupPolicy(ctx, &ms[i]); err != nil {
				s.log.Error(
					"加载角色策略失败",
					zap.Error(err),
					zap.Uint32("role_id", ms[i].ID),
					zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
				)
				return errors.FromError(err)
			}
		}
	}

	s.log.Info(
		"加载角色策略成功",
		zap.Int("policy_count", policyCount),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return nil
}

func (s *RoleService) GetRoleMenuTree(
	ctx context.Context,
	roleID uint32,
) ([]custmodel.MenuTreeNode, *errors.Error) {
	m, rErr := s.FindRoleByID(ctx, []string{"Apis", "Menus", "Buttons"}, roleID)
	if rErr != nil {
		return nil, rErr
	}
	var topMenus []custmodel.MenuModel
	roleMenuMap := make(map[uint32]custmodel.MenuModel)
	for _, menu := range m.Menus {
		roleMenuMap[menu.ID] = menu
		if menu.ParentID == nil {
			topMenus = append(topMenus, menu)
		}
	}
	roleButtonMap := make(map[uint32]custmodel.ButtonModel)
	for _, button := range m.Buttons {
		roleButtonMap[button.ID] = button
	}
	var result []custmodel.MenuTreeNode
	for _, menu := range topMenus {
		mt, err := s.buildMenuTree(menu, roleMenuMap, roleButtonMap)
		if err != nil {
			return nil, err
		}
		result = append(result, *mt)
	}
	return result, nil
}

func (s *RoleService) buildMenuTree(
	m custmodel.MenuModel,
	mp map[uint32]custmodel.MenuModel,
	bp map[uint32]custmodel.ButtonModel,
) (*custmodel.MenuTreeNode, *errors.Error) {
	var children []custmodel.MenuModel
	for _, menu := range mp {
		if menu.ParentID != nil && *menu.ParentID == m.ID {
			children = append(children, menu)
		}
	}
	var childTrees []custmodel.MenuTreeNode
	for _, child := range children {
		childTree, err := s.buildMenuTree(child, mp, bp)
		if err != nil {
			return nil, err
		}
		childTrees = append(childTrees, *childTree)
	}
	var buttons []custmodel.ButtonBaseOut
	for _, button := range bp {
		if button.MenuID == m.ID {
			buttons = append(buttons, *custmodel.ButtonModelToBaseOut(button))
		}
	}
	return &custmodel.MenuTreeNode{
		MenuBaseOut: *custmodel.MenuModelToBaseOut(m),
		Children:    childTrees,
		Buttons:     buttons,
	}, nil
}
