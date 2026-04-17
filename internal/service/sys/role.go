package sys

import (
	"context"
	"time"

	emperror "emperror.dev/errors"
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
	startTime := time.Now()
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	if len(apiIDs) == 0 {
		return []sysmodel.ApiModel{}, nil
	}

	qp := database.QueryParams{
		Query: map[string]any{"id in ?": apiIDs},
	}

	log.Debug(
		"查询角色关联的API列表:查询参数",
		zap.Uint32s("api_ids", apiIDs),
	)

	ms, err := s.apiRepo.ListModel(ctx, qp)
	if err != nil {
		log.Error(
			"查询角色关联的API列表:查询数据库失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, nil)
	}

	return ms, nil
}

func (s *RoleService) GetMenus(
	ctx context.Context,
	menuIDs []uint32,
) ([]sysmodel.MenuModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	if len(menuIDs) == 0 {
		return []sysmodel.MenuModel{}, nil
	}

	qp := database.QueryParams{
		Query: map[string]any{"id in ?": menuIDs},
	}

	log.Debug(
		"查询角色关联的菜单列表:查询参数",
		zap.Uint32s("menu_ids", menuIDs),
	)

	ms, err := s.menuRepo.ListModel(ctx, qp)
	if err != nil {
		log.Error(
			"查询角色关联的菜单列表:查询数据库失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, nil)
	}

	return ms, nil
}

func (s *RoleService) GetButtons(
	ctx context.Context,
	buttonIDs []uint32,
) ([]sysmodel.ButtonModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	if len(buttonIDs) == 0 {
		return []sysmodel.ButtonModel{}, nil
	}

	qp := database.QueryParams{
		Query: map[string]any{"id in ?": buttonIDs},
	}

	log.Debug(
		"查询角色关联的按钮列表:查询参数",
		zap.Uint32s("button_ids", buttonIDs),
	)

	ms, err := s.buttonRepo.ListModel(ctx, qp)
	if err != nil {
		log.Error(
			"查询角色关联的按钮列表:查询数据库失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, nil)
	}

	return ms, nil
}

func (s *RoleService) CreateRole(
	ctx context.Context,
	dto sysmodel.RoleUpsertDTO,
) (*sysmodel.RoleModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"创建角色:开始执行",
		zap.Object("role_upsert_dto", &dto),
	)

	m := dto.ToModel()

	apis, rErr := s.GetApis(ctx, dto.ApiIDs)
	if rErr != nil {
		log.Error(
			"创建角色:查询关联的API列表失败",
			zap.Error(rErr),
			zap.Uint32s("api_ids", dto.ApiIDs),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, rErr
	}

	menus, rErr := s.GetMenus(ctx, dto.MenuIDs)
	if rErr != nil {
		log.Error(
			"创建角色:查询关联的菜单列表失败",
			zap.Error(rErr),
			zap.Uint32s("menu_ids", dto.MenuIDs),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, rErr
	}

	buttons, rErr := s.GetButtons(ctx, dto.ButtonIDs)
	if rErr != nil {
		log.Error(
			"创建角色:查询关联的按钮列表失败",
			zap.Error(rErr),
			zap.Uint32s("button_ids", dto.ButtonIDs),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, rErr
	}

	if err := s.roleRepo.CreateModel(ctx, &m, apis, menus, buttons); err != nil {
		log.Error(
			"创建角色:创建数据库模型失败",
			zap.Error(err),
			zap.Object("role_model", &m),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, nil)
	}

	rollback := func() {
		if err := s.roleRepo.DeleteModel(ctx, m.ID); err != nil {
			log.Error(
				"创建角色:回滚数据库数据失败，请手动处理脏数据",
				zap.Error(err),
				zap.Uint32("role_id", m.ID),
				zap.Duration("total_duration", time.Since(startTime)),
			)
		}
	}

	if len(apis) > 0 {
		m.Apis = apis
	}
	if len(menus) > 0 {
		m.Menus = menus
	}
	if len(buttons) > 0 {
		m.Buttons = buttons
	}

	if err := s.roleRepo.AddGroupPolicy(ctx, m); err != nil {
		log.Error(
			"创建角色:添加角色组策略失败",
			zap.Error(err),
			zap.Object("role_model", &m),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		rollback()
		return nil, errors.FromError(err)
	}

	log.Info(
		"创建角色:执行成功",
		zap.Uint32("role_id", m.ID),
		zap.Duration("total_step_duration", time.Since(startTime)),
	)
	return &m, nil
}

func (s *RoleService) UpdateRoleByID(
	ctx context.Context,
	roleID uint32,
	dto sysmodel.RoleUpsertDTO,
) (*sysmodel.RoleModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"更新角色:开始执行",
		zap.Uint32("role_id", roleID),
		zap.Object("update_role_dto", &dto),
	)

	apis, rErr := s.GetApis(ctx, dto.ApiIDs)
	if rErr != nil {
		log.Error(
			"更新角色:查询关联的API列表失败",
			zap.Error(rErr),
			zap.Uint32s("api_ids", dto.ApiIDs),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, rErr
	}

	menus, rErr := s.GetMenus(ctx, dto.MenuIDs)
	if rErr != nil {
		log.Error(
			"更新角色:查询关联的菜单列表失败",
			zap.Error(rErr),
			zap.Uint32s("menu_ids", dto.MenuIDs),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, rErr
	}

	buttons, rErr := s.GetButtons(ctx, dto.ButtonIDs)
	if rErr != nil {
		log.Error(
			"更新角色:查询关联的按钮列表失败",
			zap.Error(rErr),
			zap.Uint32s("button_ids", dto.ButtonIDs),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, rErr
	}

	om, rErr := s.FindRoleByID(ctx, []string{"Apis", "Menus", "Buttons"}, roleID)
	if rErr != nil {
		log.Error(
			"更新角色:查询更新前的角色详情失败",
			zap.Error(rErr),
			zap.Uint32("role_id", roleID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, rErr
	}

	if err := s.roleRepo.RemoveGroupPolicy(ctx, *om); err != nil {
		log.Error(
			"更新角色:移除旧策略失败",
			zap.Error(err),
			zap.Object("role_model", om),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.FromError(err)
	}

	recoverOldPolicy := func() {
		if err := s.roleRepo.AddGroupPolicy(ctx, *om); err != nil {
			log.Error(
				"更新角色:恢复旧角色组策略失败,请手动添加策略",
				zap.Error(err),
				zap.Object("role_model", om),
				zap.Duration("total_duration", time.Since(startTime)),
			)
		}
	}

	updateData := dto.ToUpdateMap()
	if err := s.roleRepo.UpdateModel(ctx, updateData, apis, menus, buttons, "id = ?", roleID); err != nil {
		log.Error(
			"更新角色:更新数据库模型失败",
			zap.Error(err),
			zap.Any("update_data", updateData),
			zap.Uint32("role_id", roleID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		recoverOldPolicy()
		return nil, errors.NewGormError(err, updateData)
	}

	rollback := func() {
		upData := om.ToUpdateMap()
		if upErr := s.roleRepo.UpdateModel(ctx, upData, om.Apis, om.Menus, om.Buttons, "id = ?", roleID); upErr != nil {
			log.Error(
				"更新角色:回滚数据库数据失败，请手动处理脏数据",
				zap.Error(upErr),
				zap.Uint32("role_id", roleID),
				zap.Duration("total_duration", time.Since(startTime)),
			)
		}
	}

	m, rErr := s.FindRoleByID(ctx, []string{"Apis", "Menus", "Buttons"}, roleID)
	if rErr != nil {
		log.Error(
			"更新角色:查询更新后的角色详情失败",
			zap.Error(rErr),
			zap.Uint32("role_id", roleID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		rollback()
		recoverOldPolicy()
		return nil, rErr
	}

	if err := s.roleRepo.AddGroupPolicy(ctx, *m); err != nil {
		log.Error(
			"更新角色:添加新策略失败",
			zap.Error(err),
			zap.Object("role_model", m),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		rollback()
		recoverOldPolicy()
		return nil, errors.FromError(err)
	}

	log.Info(
		"更新角色:执行成功",
		zap.Uint32("role_id", roleID),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return m, nil
}

func (s *RoleService) DeleteRoleByID(
	ctx context.Context,
	roleID uint32,
) *errors.Error {
	startTime := time.Now()
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"删除角色:开始执行",
		zap.Uint32("role_id", roleID),
	)

	m, rErr := s.FindRoleByID(ctx, []string{"Apis", "Menus", "Buttons"}, roleID)
	if rErr != nil {
		log.Error(
			"删除角色:查询角色详情失败",
			zap.Error(rErr),
			zap.Uint32("role_id", roleID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return rErr
	}

	// 移除策略
	if err := s.roleRepo.RemoveGroupPolicy(ctx, *m); err != nil {
		log.Error("删除角色:移除角色组策略失败",
			zap.Error(err),
			zap.Object("role_model", m),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.FromError(err)
	}

	// 删库失败时恢复策略
	recoverOldPolicy := func() {
		if err := s.roleRepo.AddGroupPolicy(ctx, *m); err != nil {
			log.Error("删除角色:恢复旧角色组策略失败，请手动处理",
				zap.Error(err),
				zap.Object("role_model", m),
				zap.Duration("total_duration", time.Since(startTime)),
			)
		}
	}

	if err := s.roleRepo.DeleteModel(ctx, roleID); err != nil {
		log.Error(
			"删除角色:删除数据库模型失败",
			zap.Error(err),
			zap.Uint32("role_id", roleID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		recoverOldPolicy()
		return errors.NewGormError(err, map[string]any{"id": roleID})
	}

	log.Info(
		"删除角色:执行成功",
		zap.Uint32("role_id", roleID),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

func (s *RoleService) FindRoleByID(
	ctx context.Context,
	preloads []string,
	roleID uint32,
) (*sysmodel.RoleModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	m, err := s.roleRepo.GetModel(ctx, preloads, roleID)
	if err != nil {
		log.Error(
			"查询角色:查询数据库模型失败",
			zap.Error(err),
			zap.Strings("preloads", preloads),
			zap.Uint32("role_id", roleID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, map[string]any{"id": roleID})
	}

	log.Debug(
		"查询角色:查询到的数据库模型详情",
		zap.Object("role_model", m),
	)
	return m, nil
}

func (s *RoleService) ListRole(
	ctx context.Context,
	page, size int,
	dto sysmodel.ListRoleDTO,
) (int64, []sysmodel.RoleModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return 0, nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Debug(
		"查询角色列表:参数详情",
		zap.Int("page", page),
		zap.Int("size", size),
		zap.Object("list_role_dto", &dto),
	)

	limit, offset := common.Page2LimitOffset(page, size)
	qp := database.QueryParams{
		Preloads: []string{"Apis", "Menus", "Buttons"},
		Limit:    limit,
		Offset:   offset,
		OrderBy:  []string{"id ASC"},
		Query:    dto.ToQueryMap(),
	}

	count, err := s.roleRepo.CountModel(ctx, qp.Query)
	if err != nil {
		log.Error(
			"查询角色列表:查询数据库模型总数失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}

	if count == 0 {
		log.Warn(
			"查询角色列表:数据库模型总数为0",
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, nil, nil
	}

	ms, err := s.roleRepo.ListModel(ctx, qp)
	if err != nil {
		log.Error(
			"查询角色列表:查询数据库模型失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}
	return count, ms, nil
}

func (s *RoleService) LoadRolePolicy(
	ctx context.Context,
) error {
	startTime := time.Now()
	s.log.Debug("加载角色策略:开始执行")

	qp := database.QueryParams{
		Preloads: []string{"Apis", "Menus", "Buttons"},
		Columns:  []string{"id"},
	}

	ms, err := s.roleRepo.ListModel(ctx, qp)
	if err != nil {
		s.log.Error(
			"加载角色策略:查询数据库模型角色列表失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("list_step_duration", time.Since(startTime)),
		)
		return errors.NewGormError(err, nil)
	}

	roleCount := int64(len(ms))
	if roleCount == 0 {
		s.log.Warn(
			"加载角色策略:数据库模型总数为0",
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil
	}

	failCount := int64(0)
	for i := range ms {
		if err := s.roleRepo.AddGroupPolicy(ctx, ms[i]); err != nil {
			s.log.Error(
				"加载角色策略:添加角色组策略失败",
				zap.Error(err),
				zap.Uint32("role_id", ms[i].ID),
			)
			failCount++
		}
	}

	// 最终日志
	totalDuration := time.Since(startTime)
	s.log.Info(
		"加载角色策略:执行完成",
		zap.Int64("total_role", roleCount),
		zap.Int64("success_count", roleCount-failCount),
		zap.Int64("fail_count", failCount),
		zap.Duration("total_duration", totalDuration),
	)

	// 有失败但不阻断启动
	if failCount > 0 {
		return emperror.Errorf("部分角色策略加载失败，失败数量：%d", failCount)
	}
	return nil
}

func (s *RoleService) GetRoleMenuTree(
	ctx context.Context,
	roleID uint32,
) ([]sysmodel.MenuTreeNode, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	m, rErr := s.FindRoleByID(ctx, []string{"Apis", "Menus", "Buttons"}, roleID)
	if rErr != nil {
		log.Error(
			"获取角色菜单树:查询角色详情失败",
			zap.Error(rErr),
			zap.Uint32("role_id", roleID),
			zap.Duration("total_duration", time.Since(startTime)),
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
				"获取角色菜单树:构建菜单树失败",
				zap.Error(err),
				zap.Uint32("role_id", roleID),
				zap.Duration("total_duration", time.Since(startTime)),
			)
			return nil, err
		}
		result = append(result, *mt)
	}

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
