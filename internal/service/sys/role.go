package sys

import (
	"context"
	"time"

	"go.uber.org/zap"

	sysmodel "gin-artweb/internal/model/sys"
	sysrepo "gin-artweb/internal/repo/sys"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
)

type RoleService struct {
	log        *zap.Logger
	apiRepo    *sysrepo.ApiRepo
	menuRepo   *sysrepo.MenuRepo
	buttonRepo *sysrepo.ButtonRepo
	roleRepo   *sysrepo.RoleRepo
}

func NewRoleService(
	log *zap.Logger,
	apiRepo *sysrepo.ApiRepo,
	menuRepo *sysrepo.MenuRepo,
	buttonRepo *sysrepo.ButtonRepo,
	roleRepo *sysrepo.RoleRepo,
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
) ([]sysmodel.ApiModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"查询角色关联的API列表：开始执行",
		zap.Uint32s("api_ids", apiIDs),
	)

	if len(apiIDs) == 0 {
		log.Info(
			"查询角色关联的API列表：API ID列表为空",
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return []sysmodel.ApiModel{}, nil
	}

	qp := database.QueryParams{
		Query: map[string]any{"id in ?": apiIDs},
	}
	ms, err := s.apiRepo.ListModel(ctx, qp)
	if err != nil {
		log.Error(
			"查询角色关联的API列表：查询数据库失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, nil)
	}
	log.Debug(
		"查询角色关联的API列表：查询数据库成功",
		zap.Uint32s("api_ids", sysmodel.ListApiModelToUint32s(ms)),
	)

	log.Info(
		"查询角色关联的API列表：执行成功",
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return ms, nil
}

func (s *RoleService) GetMenus(
	ctx context.Context,
	menuIDs []uint32,
) ([]sysmodel.MenuModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"查询角色关联的菜单列表：开始执行",
		zap.Uint32s("menu_ids", menuIDs),
	)

	if len(menuIDs) == 0 {
		log.Info(
			"查询角色关联的菜单列表：菜单 ID列表为空",
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return []sysmodel.MenuModel{}, nil
	}

	qp := database.QueryParams{
		Query: map[string]any{"id in ?": menuIDs},
	}
	ms, err := s.menuRepo.ListModel(ctx, qp)
	if err != nil {
		log.Error(
			"查询角色关联的菜单列表：查询数据库失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, nil)
	}
	log.Debug(
		"查询角色关联的菜单列表：查询数据库成功",
		zap.Uint32s("menu_ids", sysmodel.ListMenuModelToUint32s(ms)),
	)

	log.Info(
		"查询角色关联的菜单列表：执行成功",
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return ms, nil
}

func (s *RoleService) GetButtons(
	ctx context.Context,
	buttonIDs []uint32,
) ([]sysmodel.ButtonModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"查询角色关联的按钮列表：开始执行",
		zap.Uint32s("button_ids", buttonIDs),
	)

	if len(buttonIDs) == 0 {
		log.Info(
			"查询角色关联的按钮列表：按钮 ID列表为空",
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return []sysmodel.ButtonModel{}, nil
	}

	qp := database.QueryParams{
		Query: map[string]any{"id in ?": buttonIDs},
	}
	ms, err := s.buttonRepo.ListModel(ctx, qp)
	if err != nil {
		log.Error(
			"查询角色关联的按钮列表：查询数据库失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, nil)
	}
	log.Debug(
		"查询角色关联的按钮列表：查询数据库成功",
		zap.Uint32s("button_ids", sysmodel.ListButtonModelToUint32s(ms)),
	)

	log.Info(
		"查询角色关联的按钮列表：执行成功",
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return ms, nil
}

func (s *RoleService) CreateRole(
	ctx context.Context,
	dto sysmodel.RoleUpsertDTO,
) (*sysmodel.RoleModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info("开始创建角色")

	log.Debug(
		"创建角色：输入参数",
		zap.Object("role_upsert_dto", &dto),
	)

	m := sysmodel.RoleModel{
		Name:  dto.Name,
		Descr: dto.Descr,
	}

	apis, rErr := s.GetApis(ctx, dto.ApiIDs)
	if rErr != nil {
		log.Error(
			"创建角色：查询角色关联的权限列表失败",
			zap.Error(rErr),
			zap.Uint32s("api_ids", dto.ApiIDs),
		)
		return nil, rErr
	}
	log.Debug("创建角色：查询角色关联的权限列表成功")

	menus, rErr := s.GetMenus(ctx, dto.MenuIDs)
	if rErr != nil {
		log.Error(
			"创建角色：查询菜单关联的菜单列表失败",
			zap.Error(rErr),
			zap.Uint32s("menu_ids", dto.MenuIDs),
		)
		return nil, rErr
	}
	log.Debug("创建角色：查询菜单关联的菜单列表成功")

	buttons, rErr := s.GetButtons(ctx, dto.ButtonIDs)
	if rErr != nil {
		log.Error(
			"创建角色：查询按钮关联的按钮列表失败",
			zap.Error(rErr),
			zap.Uint32s("button_ids", dto.ButtonIDs),
		)
		return nil, rErr
	}
	log.Debug("创建角色：查询按钮关联的按钮列表成功")

	createStepStart := time.Now()
	log.Debug(
		"创建角色：开始创建数据库模型",
		zap.Object("role_model", &m),
	)
	if err := s.roleRepo.CreateModel(ctx, &m, apis, menus, buttons); err != nil {
		log.Error(
			"创建角色：创建数据库模型失败",
			zap.Error(err),
			zap.Object("role_model", &m),
			zap.Duration("create_step_duration", time.Since(createStepStart)),
		)
		return nil, errors.NewGormError(err, nil)
	}
	createStepDuration := time.Since(createStepStart)
	log.Debug(
		"创建角色：创建数据库模型成功",
		zap.Object("role_model", &m),
		zap.Duration("create_step_duration", createStepDuration),
	)

	if len(apis) > 0 {
		m.Apis = apis
	}
	if len(menus) > 0 {
		m.Menus = menus
	}
	if len(buttons) > 0 {
		m.Buttons = buttons
	}

	addPolicyStepStart := time.Now()
	log.Debug(
		"创建角色：开始添加角色组策略",
		zap.Object("role_model", &m),
	)
	if err := s.roleRepo.AddGroupPolicy(ctx, &m); err != nil {
		log.Error(
			"创建角色：添加角色组策略失败",
			zap.Error(err),
			zap.Object("role_model", &m),
			zap.Duration("add_policy_step_duration", time.Since(addPolicyStepStart)),
		)
		return nil, errors.FromError(err)
	}
	addPolicyStepDuration := time.Since(addPolicyStepStart)
	log.Debug(
		"创建角色：添加角色组策略成功",
		zap.Object("role_model", &m),
		zap.Duration("add_policy_step_duration", addPolicyStepDuration),
	)

	log.Info(
		"创建角色：执行成功",
		zap.Uint32("role_id", m.ID),
		zap.Duration("create_step_duration", createStepDuration),
		zap.Duration("add_policy_step_duration", addPolicyStepDuration),
		zap.Duration("total_step_duration", time.Since(startTime)),
	)
	return &m, nil
}

func (s *RoleService) UpdateRoleByID(
	ctx context.Context,
	roleID uint32,
	dto sysmodel.RoleUpsertDTO,
) (*sysmodel.RoleModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"更新角色：开始执行",
		zap.Uint32("role_id", roleID),
	)

	log.Debug(
		"更新角色：输入参数",
		zap.Uint32("role_id", roleID),
		zap.Object("update_role_dto", &dto),
	)

	apis, rErr := s.GetApis(ctx, dto.ApiIDs)
	if rErr != nil {
		log.Error(
			"更新角色：查询角色关联的权限列表失败",
			zap.Error(rErr),
			zap.Uint32s("api_ids", dto.ApiIDs),
		)
		return nil, rErr
	}
	log.Debug("更新角色：查询角色关联的权限列表成功")

	menus, rErr := s.GetMenus(ctx, dto.MenuIDs)
	if rErr != nil {
		log.Error(
			"更新角色：查询菜单关联的菜单列表失败",
			zap.Error(rErr),
			zap.Uint32s("menu_ids", dto.MenuIDs),
		)
		return nil, rErr
	}
	log.Debug("更新角色：查询菜单关联的菜单列表成功")

	buttons, rErr := s.GetButtons(ctx, dto.ButtonIDs)
	if rErr != nil {
		log.Error(
			"更新角色：查询按钮关联的按钮列表失败",
			zap.Error(rErr),
			zap.Uint32s("button_ids", dto.ButtonIDs),
		)
		return nil, rErr
	}
	log.Debug("更新角色：查询按钮关联的按钮列表成功")

	updateData := dto.ToUpdateMap()
	updateStepStart := time.Now()
	log.Debug(
		"更新角色：开始更新数据库模型",
		zap.Any("update_data", updateData),
		zap.Uint32("role_id", roleID),
	)
	if err := s.roleRepo.UpdateModel(ctx, updateData, apis, menus, buttons, "id = ?", roleID); err != nil {
		log.Error(
			"更新角色：更新数据库模型失败",
			zap.Error(err),
			zap.Any("update_data", updateData),
			zap.Uint32("role_id", roleID),
		)
		return nil, errors.NewGormError(err, updateData)
	}
	updateStepDuration := time.Since(updateStepStart)
	log.Debug(
		"更新角色：更新数据库模型成功",
		zap.Uint32("role_id", roleID),
		zap.Duration("update_step_duration", updateStepDuration),
	)

	var m *sysmodel.RoleModel
	m, rErr = s.FindRoleByID(ctx, []string{"Apis", "Menus", "Buttons"}, roleID)
	if rErr != nil {
		log.Error(
			"更新角色：查询更新后的角色详情失败",
			zap.Error(rErr),
			zap.Uint32("role_id", roleID),
		)
		return nil, rErr
	}

	removePolicyStepStart := time.Now()
	log.Debug(
		"更新角色：开始移除旧策略",
		zap.Uint32("role_id", roleID),
		zap.Bool("remove_inherited", false),
	)
	if err := s.roleRepo.RemoveGroupPolicy(ctx, m); err != nil {
		log.Error(
			"更新角色：移除旧策略失败",
			zap.Error(err),
			zap.Object("role_model", m),
			zap.Bool("remove_inherited", false),
			zap.Duration("remove_policy_step_duration", time.Since(removePolicyStepStart)),
		)
		return nil, errors.FromError(err)
	}
	removePolicyStepDuration := time.Since(removePolicyStepStart)
	log.Debug(
		"更新角色：移除旧策略耗时",
		zap.Object("role_model", m),
		zap.Duration("remove_policy_step_duration", removePolicyStepDuration),
	)

	addPolicyStepStart := time.Now()
	if err := s.roleRepo.AddGroupPolicy(ctx, m); err != nil {
		log.Error(
			"更新角色：添加新策略失败",
			zap.Error(err),
			zap.Object("role_model", m),
			zap.Duration("add_policy_step_duration", time.Since(addPolicyStepStart)),
		)
		return nil, errors.FromError(err)
	}
	addPolicyStepDuration := time.Since(addPolicyStepStart)
	log.Debug(
		"更新角色：添加新策略成功",
		zap.Object("role_model", m),
		zap.Duration("add_policy_step_duration", addPolicyStepDuration),
	)

	log.Info(
		"更新角色：执行成功",
		zap.Uint32("role_id", roleID),
		zap.Duration("total_duration", time.Since(startTime)),
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

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"删除角色：执行开始",
		zap.Uint32("role_id", roleID),
	)

	m, rErr := s.FindRoleByID(ctx, []string{"Apis", "Menus", "Buttons"}, roleID)
	if rErr != nil {
		log.Error(
			"删除角色：查询角色详情失败",
			zap.Error(rErr),
			zap.Uint32("role_id", roleID),
		)
		return rErr
	}

	deleteStepStart := time.Now()
	log.Debug(
		"删除角色：开始删除数据库模型",
		zap.Uint32("role_id", roleID),
	)
	if err := s.roleRepo.DeleteModel(ctx, roleID); err != nil {
		log.Error(
			"删除角色：删除数据库模型失败",
			zap.Error(err),
			zap.Uint32("role_id", roleID),
			zap.Duration("delete_step_duration", time.Since(deleteStepStart)),
		)
		return errors.NewGormError(err, map[string]any{"id": roleID})
	}
	deleteStepDuration := time.Since(deleteStepStart)
	log.Debug(
		"删除角色：删除数据库模型成功",
		zap.Uint32("role_id", roleID),
		zap.Duration("delete_step_duration", deleteStepDuration),
	)

	removePolicyStepStart := time.Now()
	log.Debug(
		"删除角色：开始移除策略",
		zap.Uint32("role_id", roleID),
		zap.Bool("remove_inherited", true),
	)
	if err := s.roleRepo.RemoveGroupPolicy(ctx, m); err != nil {
		log.Error(
			"删除角色：移除策略失败",
			zap.Error(err),
			zap.Object("role_model", m),
			zap.Bool("remove_inherited", true),
			zap.Duration("remove_policy_step_duration", time.Since(removePolicyStepStart)),
		)
		return errors.FromError(err)
	}
	removePolicyStepDuration := time.Since(removePolicyStepStart)
	log.Debug(
		"删除角色：移除策略成功",
		zap.Uint32("role_id", roleID),
		zap.Duration("remove_policy_step_duration", removePolicyStepDuration),
	)

	log.Info(
		"删除角色：执行成功",
		zap.Uint32("role_id", roleID),
		zap.Duration("delete_step_duration", deleteStepDuration),
		zap.Duration("remove_policy_step_duration", removePolicyStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

func (s *RoleService) FindRoleByID(
	ctx context.Context,
	preloads []string,
	roleID uint32,
) (*sysmodel.RoleModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"查询角色：开始执行",
		zap.Strings("preloads", preloads),
		zap.Uint32("role_id", roleID),
	)

	m, err := s.roleRepo.GetModel(ctx, preloads, roleID)
	if err != nil {
		log.Error(
			"查询角色：查询数据库模型失败",
			zap.Error(err),
			zap.Strings("preloads", preloads),
			zap.Uint32("role_id", roleID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, map[string]any{"id": roleID})
	}
	log.Debug(
		"查询角色：查询到的数据库模型详情",
		zap.Object("role_model", m),
	)

	log.Info(
		"查询角色：执行成功",
		zap.Uint32("role_id", roleID),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return m, nil
}

func (s *RoleService) ListRole(
	ctx context.Context,
	page, size int,
	dto sysmodel.ListRoleDTO,
) (int64, []sysmodel.RoleModel, *errors.Error) {
	if ctx.Err() != nil {
		return 0, nil, errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info("查询角色列表：开始执行")

	log.Debug(
		"查询角色列表：参数详情",
		zap.Int("page", page),
		zap.Int("size", size),
		zap.Object("list_role_dto", &dto),
	)

	limit, offset := common.Page2LimitOffset(page, size)
	qp := database.QueryParams{
		Preloads: []string{"Apis", "Menus", "Buttons"},
		Columns:  []string{"id"},
		Limit:    limit,
		Offset:   offset,
		OrderBy:  []string{"id ASC"},
		Query:    dto.ToQueryMap(),
	}

	log.Debug(
		"查询角色列表：查询数据库模型参数",
		zap.Object("query_params", &qp),
	)

	countStepStart := time.Now()
	log.Debug(
		"查询角色列表：开始查询数据库模型总数",
		zap.Object("query_params", &qp),
	)
	count, err := s.roleRepo.CountModel(ctx, qp.Query)
	countStepDuration := time.Since(countStepStart)
	if err != nil {
		log.Error(
			"查询角色列表：查询数据库模型总数失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("count_step_duration", countStepDuration),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}
	log.Debug(
		"查询角色列表：查询数据库模型总数成功",
		zap.Int64("total_count", count),
		zap.Duration("count_step_duration", countStepDuration),
	)
	if count == 0 {
		log.Warn(
			"查询角色列表：数据库模型总数为0",
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return count, nil, nil
	}

	ms, err := s.roleRepo.ListModel(ctx, qp)
	if err != nil {
		log.Error(
			"查询角色列表：查询数据库模型失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}

	log.Info(
		"查询角色列表：执行成功",
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return count, ms, nil
}

func (s *RoleService) LoadRolePolicy(ctx context.Context) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := s.log.With(zap.String("trace_id", ctxutil.GetTraceID(ctx)))

	log.Debug(
		"加载角色策略：开始执行",
	)

	qp := database.QueryParams{
		Preloads: []string{"Apis", "Menus", "Buttons"},
		Columns:  []string{"id"},
	}

	listStepStart := time.Now()
	ms, err := s.roleRepo.ListModel(ctx, qp)
	listStepDuration := time.Since(listStepStart)
	if err != nil {
		log.Error(
			"加载角色策略：查询数据库模型角色列表失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("list_step_duration", listStepDuration),
		)
		return errors.NewGormError(err, nil)
	}

	var policyCount int
	if len(ms) > 0 {
		policyStepStart := time.Now()
		log.Debug("加载角色策略：开始添加角色组策略")
		policyCount = len(ms)
		for i := range ms {
			if err := s.roleRepo.AddGroupPolicy(ctx, &ms[i]); err != nil {
				log.Error(
					"加载角色策略：添加角色组策略失败",
					zap.Error(err),
					zap.Uint32("role_id", ms[i].ID),
				)
				return errors.FromError(err)
			}
		}
		policyStepDuration := time.Since(policyStepStart)
		log.Debug(
			"加载角色策略：添加角色组策略耗时",
			zap.Int("policy_count", policyCount),
			zap.Duration("policy_step_duration", policyStepDuration),
		)
	}

	log.Debug(
		"加载角色策略：执行成功",
		zap.Int("policy_count", policyCount),
		zap.Duration("list_step_duration", listStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

func (s *RoleService) GetRoleMenuTree(
	ctx context.Context,
	roleID uint32,
) ([]sysmodel.MenuTreeNode, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"获取角色菜单树：开始执行",
		zap.Uint32("role_id", roleID),
	)

	m, rErr := s.FindRoleByID(ctx, []string{"Apis", "Menus", "Buttons"}, roleID)
	if rErr != nil {
		log.Error(
			"获取角色菜单树：查询角色详情失败",
			zap.Error(rErr),
			zap.Uint32("role_id", roleID),
		)
		return nil, rErr
	}
	var topMenus []sysmodel.MenuModel
	roleMenuMap := make(map[uint32]sysmodel.MenuModel)
	for _, menu := range m.Menus {
		roleMenuMap[menu.ID] = menu
		if menu.ParentID == nil {
			topMenus = append(topMenus, menu)
		}
	}
	roleButtonMap := make(map[uint32]sysmodel.ButtonModel)
	for _, button := range m.Buttons {
		roleButtonMap[button.ID] = button
	}
	var result []sysmodel.MenuTreeNode
	for _, menu := range topMenus {
		mt, err := s.buildMenuTree(menu, roleMenuMap, roleButtonMap)
		if err != nil {
			log.Error(
				"获取角色菜单树：构建菜单树失败",
				zap.Error(err),
				zap.Uint32("role_id", roleID),
			)
			return nil, err
		}
		result = append(result, *mt)
	}

	log.Info(
		"获取角色菜单树：执行成功",
		zap.Uint32("role_id", roleID),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return result, nil
}

func (s *RoleService) buildMenuTree(
	m sysmodel.MenuModel,
	mp map[uint32]sysmodel.MenuModel,
	bp map[uint32]sysmodel.ButtonModel,
) (*sysmodel.MenuTreeNode, *errors.Error) {
	var children []sysmodel.MenuModel
	for _, menu := range mp {
		if menu.ParentID != nil && *menu.ParentID == m.ID {
			children = append(children, menu)
		}
	}
	var childTrees []sysmodel.MenuTreeNode
	for _, child := range children {
		childTree, err := s.buildMenuTree(child, mp, bp)
		if err != nil {
			return nil, err
		}
		childTrees = append(childTrees, *childTree)
	}
	var buttons []sysmodel.ButtonBaseOut
	for _, button := range bp {
		if button.MenuID == m.ID {
			buttons = append(buttons, *sysmodel.ButtonModelToBaseOut(button))
		}
	}
	return &sysmodel.MenuTreeNode{
		MenuBaseOut: *sysmodel.MenuModelToBaseOut(m),
		Children:    childTrees,
		Buttons:     buttons,
	}, nil
}
