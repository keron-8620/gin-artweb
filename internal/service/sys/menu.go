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

type MenuService struct {
	log      *zap.Logger
	apiRepo  *sysrepo.ApiRepo
	menuRepo *sysrepo.MenuRepo
}

func NewMenuService(
	log *zap.Logger,
	apiRepo *sysrepo.ApiRepo,
	menuRepo *sysrepo.MenuRepo,
) *MenuService {
	return &MenuService{
		log:      log,
		apiRepo:  apiRepo,
		menuRepo: menuRepo,
	}
}

func (s *MenuService) GetParentMenu(
	ctx context.Context,
	parentID *uint32,
) (*sysmodel.MenuModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Debug(
		"查询父菜单：开始执行",
		zap.Uint32p("parent_id", parentID),
	)

	if parentID == nil || *parentID == 0 {
		log.Debug(
			"查询父菜单：父菜单ID为空",
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, nil
	}

	pid := *parentID
	m, err := s.menuRepo.GetModel(ctx, nil, pid)
	if err != nil {
		log.Error(
			"查询父菜单：查询数据库失败",
			zap.Error(err),
			zap.Uint32("parent_id", pid),
			zap.Duration("find_step_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, map[string]any{"parent_id": pid})
	}

	log.Debug(
		"查询父菜单：查询数据库成功",
		zap.Object("menu_model", m),
		zap.Duration("find_step_duration", time.Since(startTime)),
	)

	log.Info(
		"查询父菜单：执行成功",
		zap.Uint32("parent_id", pid),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return m, nil
}

func (s *MenuService) GetApis(
	ctx context.Context,
	apiIDs []uint32,
) ([]sysmodel.ApiModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Debug(
		"查询菜单关联的权限列表：开始执行",
		zap.Uint32s("api_ids", apiIDs),
	)

	if len(apiIDs) == 0 {
		log.Info(
			"查询菜单关联的权限列表：API ID列表为空",
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return []sysmodel.ApiModel{}, nil
	}

	qp := database.QueryParams{
		Query: map[string]any{"id in ?": apiIDs},
	}
	log.Debug(
		"查询菜单关联的权限列表：查询数据库参数",
		zap.Object("query_params", &qp),
	)

	ms, err := s.apiRepo.ListModel(ctx, qp)
	if err != nil {
		log.Error(
			"查询菜单关联的权限列表：查询数据库失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, nil)
	}
	log.Debug(
		"查询菜单关联的权限列表：查询数据库成功",
		zap.Uint32s("api_ids", sysmodel.ListApiModelToUint32s(ms)),
	)

	log.Info(
		"查询菜单关联的权限列表：执行成功",
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return ms, nil
}

func (s *MenuService) CreateMenu(
	ctx context.Context,
	dto sysmodel.CreateMenuDTO,
) (*sysmodel.MenuModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info("创建菜单：开始执行")

	log.Debug(
		"创建菜单：输入参数",
		zap.Object("create_menu_dto", &dto),
	)

	m := sysmodel.MenuModel{
		StandardModel: database.StandardModel{
			BaseModel: database.BaseModel{ID: dto.ID},
		},
		Path:      dto.Path,
		Component: dto.Component,
		Name:      dto.Name,
		Meta: sysmodel.MetaSchemas{
			Icon:  dto.Meta.Icon,
			Title: dto.Meta.Title,
		},
		Sort:     dto.Sort,
		IsActive: dto.IsActive,
		Descr:    dto.Descr,
		ParentID: dto.ParentID,
	}

	menu, rErr := s.GetParentMenu(ctx, dto.ParentID)
	if rErr != nil {
		log.Error(
			"创建菜单：查询父菜单失败",
			zap.Error(rErr),
			zap.Uint32p("parent_id", dto.ParentID),
		)
		return nil, rErr
	}
	if menu != nil {
		m.Parent = menu
	}

	apis, rErr := s.GetApis(ctx, dto.ApiIDs)
	if rErr != nil {
		log.Error(
			"创建菜单：查询菜单关联的权限列表失败",
			zap.Error(rErr),
			zap.Uint32s("api_ids", dto.ApiIDs),
		)
		return nil, rErr
	}

	createStepStart := time.Now()
	log.Debug(
		"创建菜单：开始创建数据库模型",
		zap.Object("menu_model", &m),
	)
	if err := s.menuRepo.CreateModel(ctx, &m, apis); err != nil {
		log.Error(
			"创建菜单：创建数据库模型失败",
			zap.Error(err),
			zap.Object("menu_model", &m),
			zap.Duration("create_step_duration", time.Since(createStepStart)),
		)
		return nil, errors.NewGormError(err, nil)
	}
	createStepDuration := time.Since(createStepStart)
	log.Debug(
		"创建菜单：创建数据库模型成功",
		zap.Object("menu_model", &m),
		zap.Duration("create_step_duration", createStepDuration),
	)
	if len(apis) > 0 {
		m.Apis = apis
	}

	addPolicyStepStart := time.Now()
	log.Debug(
		"创建菜单：开始添加菜单组策略",
		zap.Object("menu_model", &m),
		zap.Uint32s("api_ids", dto.ApiIDs),
	)
	if err := s.menuRepo.AddGroupPolicy(ctx, &m); err != nil {
		log.Error(
			"创建菜单：添加菜单组策略失败",
			zap.Error(err),
			zap.Object("menu_model", &m),
			zap.Uint32s("api_ids", dto.ApiIDs),
			zap.Duration("add_policy_step_duration", time.Since(addPolicyStepStart)),
		)
		return nil, errors.FromError(err)
	}
	addPolicyStepDuration := time.Since(addPolicyStepStart)
	log.Debug(
		"创建菜单：添加菜单组策略成功",
		zap.Object("menu_model", &m),
		zap.Duration("add_policy_step_duration", addPolicyStepDuration),
	)

	log.Info(
		"创建菜单：执行成功",
		zap.Uint32("menu_id", m.ID),
		zap.Duration("create_step_duration", createStepDuration),
		zap.Duration("add_policy_step_duration", addPolicyStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return &m, nil
}

func (s *MenuService) UpdateMenuByID(
	ctx context.Context,
	menuID uint32,
	dto sysmodel.UpdateMenuDTO,
) (*sysmodel.MenuModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"更新菜单：开始执行",
		zap.Uint32("menu_id", menuID),
	)

	log.Debug(
		"更新菜单：输入参数",
		zap.Uint32("menu_id", menuID),
		zap.Object("update_menu_dto", &dto),
	)

	apis, rErr := s.GetApis(ctx, dto.ApiIDs)
	if rErr != nil {
		log.Debug(
			"更新菜单：查询菜单关联的权限列表失败",
			zap.Error(rErr),
			zap.Uint32s("api_ids", dto.ApiIDs),
		)
		return nil, rErr
	}

	updateData := dto.ToUpdateMap()
	updateStepStart := time.Now()
	log.Debug(
		"更新菜单：开始更新数据库模型",
		zap.Any("update_data", updateData),
		zap.Uint32("menu_id", menuID),
	)
	if err := s.menuRepo.UpdateModel(ctx, updateData, apis, "id = ?", menuID); err != nil {
		log.Error(
			"更新菜单：更新数据库模型失败",
			zap.Error(err),
			zap.Any("update_data", updateData),
			zap.Uint32("menu_id", menuID),
			zap.Duration("update_step_duration", time.Since(updateStepStart)),
		)
		return nil, errors.NewGormError(err, updateData)
	}
	updateStepDuration := time.Since(updateStepStart)
	log.Debug(
		"更新菜单：更新数据库模型成功",
		zap.Uint32("menu_id", menuID),
		zap.Duration("update_step_duration", updateStepDuration),
	)

	var m *sysmodel.MenuModel
	m, rErr = s.FindMenuByID(ctx, []string{"Parent", "Apis"}, menuID)
	if rErr != nil {
		log.Error(
			"更新菜单：查询更新后的菜单详情失败",
			zap.Error(rErr),
			zap.Uint32("menu_id", menuID),
		)
		return nil, rErr
	}

	removePolicyStepStart := time.Now()
	log.Debug(
		"更新菜单：开始移除旧菜单组策略",
		zap.Object("menu_model", m),
		zap.Bool("remove_inherited", false),
	)
	if err := s.menuRepo.RemoveGroupPolicy(ctx, m, false); err != nil {
		log.Error(
			"更新菜单：移除旧菜单组策略失败",
			zap.Error(err),
			zap.Object("menu_model", m),
			zap.Bool("remove_inherited", false),
		)
		return nil, errors.FromError(err)
	}
	removePolicyStepDuration := time.Since(removePolicyStepStart)
	log.Debug(
		"更新菜单：移除旧菜单组策略成功",
		zap.Uint32("menu_id", menuID),
		zap.Duration("remove_policy_step_duration", removePolicyStepDuration),
	)

	addPolicyStepStart := time.Now()
	log.Debug(
		"更新菜单：开始添加新菜单组策略",
		zap.Object("menu_model", m),
	)
	if err := s.menuRepo.AddGroupPolicy(ctx, m); err != nil {
		log.Error(
			"更新菜单：添加新菜单组策略失败",
			zap.Error(err),
			zap.Object("menu_model", m),
		)
		return nil, errors.FromError(err)
	}
	addPolicyStepDuration := time.Since(addPolicyStepStart)
	log.Debug(
		"更新菜单：添加新菜单组策略成功",
		zap.Uint32("menu_id", menuID),
		zap.Duration("add_policy_step_duration", addPolicyStepDuration),
	)

	log.Info(
		"更新菜单：执行成功",
		zap.Uint32("menu_id", menuID),
		zap.Duration("update_step_duration", updateStepDuration),
		zap.Duration("remove_policy_step_duration", removePolicyStepDuration),
		zap.Duration("add_policy_step_duration", addPolicyStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return m, nil
}

func (s *MenuService) DeleteMenuByID(
	ctx context.Context,
	menuID uint32,
) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"删除菜单：开始执行",
		zap.Uint32("menu_id", menuID),
	)

	m, rErr := s.FindMenuByID(ctx, []string{"Parent", "Apis"}, menuID)
	if rErr != nil {
		log.Error(
			"删除菜单：查询菜单详情失败",
			zap.Error(rErr),
			zap.Uint32("menu_id", menuID),
		)
		return rErr
	}

	deleteStepStart := time.Now()
	log.Debug(
		"删除菜单：开始删除数据库模型",
		zap.Uint32("menu_id", menuID),
	)
	if err := s.menuRepo.DeleteModel(ctx, menuID); err != nil {
		log.Error(
			"删除菜单：删除数据库模型失败",
			zap.Error(err),
			zap.Uint32("menu_id", menuID),
			zap.Duration("delete_step_duration", time.Since(deleteStepStart)),
		)
		return errors.NewGormError(err, map[string]any{"id": menuID})
	}
	deleteStepDuration := time.Since(deleteStepStart)
	log.Debug(
		"删除菜单：删除数据库模型成功",
		zap.Uint32("menu_id", menuID),
		zap.Duration("delete_step_duration", deleteStepDuration),
	)

	removePolicyStepStart := time.Now()
	log.Debug(
		"删除菜单：开始移除菜单组策略",
		zap.Object("menu_model", m),
		zap.Bool("remove_inherited", true),
	)
	if err := s.menuRepo.RemoveGroupPolicy(ctx, m, true); err != nil {
		log.Error(
			"删除菜单：移除菜单组策略失败",
			zap.Error(err),
			zap.Object("menu_model", m),
			zap.Bool("remove_inherited", true),
			zap.Duration("remove_policy_step_duration", time.Since(removePolicyStepStart)),
		)
		return errors.FromError(err)
	}
	removePolicyStepDuration := time.Since(removePolicyStepStart)
	log.Debug(
		"删除菜单：移除菜单组策略成功",
		zap.Uint32("menu_id", menuID),
		zap.Duration("remove_policy_step_duration", removePolicyStepDuration),
	)

	log.Info(
		"删除菜单成功",
		zap.Uint32("menu_id", menuID),
		zap.Duration("delete_step_duration", deleteStepDuration),
		zap.Duration("remove_policy_step_duration", removePolicyStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

func (s *MenuService) FindMenuByID(
	ctx context.Context,
	preloads []string,
	menuID uint32,
) (*sysmodel.MenuModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"查询菜单：开始执行",
		zap.Strings("preloads", preloads),
		zap.Uint32("menu_id", menuID),
	)

	m, err := s.menuRepo.GetModel(ctx, preloads, menuID)
	if err != nil {
		log.Error(
			"查询菜单：查询数据库模型失败",
			zap.Error(err),
			zap.Strings("preloads", preloads),
			zap.Uint32("menu_id", menuID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, map[string]any{"id": menuID})
	}

	log.Debug(
		"查询菜单：查询到的数据库模型详情",
		zap.Object("menu_model", m),
	)

	log.Info(
		"查询菜单：执行成功",
		zap.Strings("preloads", preloads),
		zap.Uint32("menu_id", menuID),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return m, nil
}

func (s *MenuService) ListMenu(
	ctx context.Context,
	page, size int,
	dto sysmodel.ListMenuDTO,
) (int64, []sysmodel.MenuModel, *errors.Error) {
	if ctx.Err() != nil {
		return 0, nil, errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info("查询菜单列表：开始执行")

	log.Debug(
		"查询菜单列表：参数详情",
		zap.Int("page", page),
		zap.Int("size", size),
		zap.Object("list_menu_dto", &dto),
	)

	limit, offset := common.Page2LimitOffset(page, size)
	qp := database.QueryParams{
		Limit:   limit,
		Offset:  offset,
		OrderBy: []string{"id ASC"},
		Query:   dto.ToQueryMap(),
	}

	log.Debug(
		"查询菜单列表：查询数据库模型参数",
		zap.Object("query_params", &qp),
	)

	countStepStart := time.Now()
	log.Debug(
		"加载菜单策略：开始查询数据库模型总数",
		zap.Object("query_params", &qp),
	)
	count, err := s.menuRepo.CountModel(ctx, qp.Query)
	countStepDuration := time.Since(countStepStart)
	if err != nil {
		log.Error(
			"查询菜单列表：查询数据库模型总数失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("count_step_duration", countStepDuration),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}
	log.Debug(
		"查询菜单列表：查询数据库模型总数成功",
		zap.Int64("total_count", count),
		zap.Duration("count_step_duration", countStepDuration),
	)
	if count == 0 {
		log.Warn(
			"查询菜单列表：数据库模型总数为0",
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return count, nil, nil
	}

	listStepStart := time.Now()
	log.Debug(
		"查询菜单列表：开始查询数据库模型",
		zap.Any("query", qp.Query),
	)
	ms, err := s.menuRepo.ListModel(ctx, qp)
	if err != nil {
		log.Error(
			"查询菜单列表：查询数据库模型失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("list_step_duration", time.Since(listStepStart)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}
	listStepDuration := time.Since(listStepStart)
	log.Debug(
		"查询菜单列表：查询数据库模型成功",
		zap.Int("menu_count", len(ms)),
		zap.Duration("list_step_duration", listStepDuration),
	)

	log.Info(
		"查询菜单列表：执行成功",
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return count, ms, nil
}

func (s *MenuService) LoadMenuPolicy(ctx context.Context) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := s.log.With(
		zap.String("trace_id", ctxutil.GetTraceID(ctx)),
	)

	log.Info("加载菜单策略：开始执行")

	qp := database.QueryParams{
		Preloads: []string{"Apis"},
		Columns:  []string{"id", "parent_id"},
	}

	listStepStart := time.Now()
	log.Debug(
		"加载菜单策略：开始查询数据库模型列表",
		zap.Object("query_params", &qp),
	)
	ms, err := s.menuRepo.ListModel(ctx, qp)
	listStepDuration := time.Since(listStepStart)
	if err != nil {
		log.Error(
			"加载菜单策略：查询数据库模型列表失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("list_step_duration", listStepDuration),
		)
		return errors.NewGormError(err, nil)
	}
	log.Debug(
		"加载菜单策略：查询数据库模型列表成功",
		zap.Duration("list_step_duration", listStepDuration),
	)

	var policyCount int
	if len(ms) > 0 {
		policyStepStart := time.Now()
		log.Debug(
			"加载菜单策略：开始添加菜单组策略",
			zap.Object("query_params", &qp),
		)
		policyCount = len(ms)
		for i := range ms {
			if err := s.menuRepo.AddGroupPolicy(ctx, &ms[i]); err != nil {
				log.Error(
					"加载菜单策略：添加菜单组策略失败",
					zap.Error(err),
					zap.Uint32("menu_id", ms[i].ID),
				)
				return errors.FromError(err)
			}
		}
		policyStepDuration := time.Since(policyStepStart)
		log.Debug(
			"加载菜单策略：添加菜单组策略成功",
			zap.Int("policy_count", policyCount),
			zap.Duration("policy_step_duration", policyStepDuration),
		)
	}
	log.Info(
		"加载菜单策略：执行成功",
		zap.Int("policy_count", policyCount),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}
