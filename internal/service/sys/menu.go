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

func (s *MenuService) getParentMenu(
	ctx context.Context,
	parentID *uint32,
) (*sysmodel.MenuModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	if parentID == nil || *parentID == 0 {
		return nil, nil
	}
	pid := *parentID

	m, err := s.menuRepo.GetModel(ctx, nil, pid)
	if err != nil {
		log.Error(
			"查询父菜单:查询数据库失败",
			zap.Error(err),
			zap.Uint32("parent_id", pid),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, map[string]any{"parent_id": pid})
	}
	return m, nil
}

func (s *MenuService) getApis(
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
		"查询菜单关联的权限列表:查询参数",
		zap.Uint32s("api_ids", apiIDs),
	)

	ms, err := s.apiRepo.ListModel(ctx, qp)
	if err != nil {
		log.Error(
			"查询菜单关联的权限列表:查询数据库失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, nil)
	}

	return ms, nil
}

func (s *MenuService) CreateMenu(
	ctx context.Context,
	dto sysmodel.CreateMenuDTO,
) (*sysmodel.MenuModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"创建菜单:开始执行",
		zap.Object("create_menu_dto", &dto),
	)

	m := dto.ToModel()

	parent, rErr := s.getParentMenu(ctx, dto.ParentID)
	if rErr != nil {
		log.Error(
			"创建菜单:查询父菜单失败",
			zap.Error(rErr),
			zap.Uint32p("parent_id", dto.ParentID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, rErr
	}
	if parent != nil {
		m.Parent = parent
	}

	apis, rErr := s.getApis(ctx, dto.ApiIDs)
	if rErr != nil {
		log.Error(
			"创建菜单:查询菜单关联的权限列表失败",
			zap.Error(rErr),
			zap.Uint32s("api_ids", dto.ApiIDs),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, rErr
	}

	if err := s.menuRepo.CreateModel(ctx, &m, apis); err != nil {
		log.Error(
			"创建菜单:创建数据库模型失败",
			zap.Error(err),
			zap.Object("menu_model", &m),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, nil)
	}

	rollback := func() {
		if err := s.menuRepo.DeleteModel(ctx, m.ID); err != nil {
			log.Error(
				"创建菜单:回滚数据库数据失败，请手动处理脏数据",
				zap.Error(err),
				zap.Uint32("menu_id", m.ID),
				zap.Duration("total_duration", time.Since(startTime)),
			)
		}
	}

	if len(apis) > 0 {
		m.Apis = apis
	}

	if err := s.menuRepo.AddGroupPolicy(ctx, m); err != nil {
		log.Error(
			"创建菜单:添加菜单组策略失败",
			zap.Error(err),
			zap.Object("menu_model", &m),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		rollback()
		return nil, errors.FromError(err)
	}

	log.Info(
		"创建菜单:执行成功",
		zap.Uint32("menu_id", m.ID),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return &m, nil
}

func (s *MenuService) UpdateMenuByID(
	ctx context.Context,
	menuID uint32,
	dto sysmodel.UpdateMenuDTO,
) (*sysmodel.MenuModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"更新菜单:开始执行",
		zap.Uint32("menu_id", menuID),
		zap.Object("update_menu_dto", &dto),
	)

	apis, rErr := s.getApis(ctx, dto.ApiIDs)
	if rErr != nil {
		log.Error(
			"更新菜单:查询菜单关联的权限列表失败",
			zap.Error(rErr),
			zap.Uint32s("api_ids", dto.ApiIDs),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, rErr
	}

	om, rErr := s.FindMenuByID(ctx, []string{"Parent", "Apis"}, menuID)
	if rErr != nil {
		log.Error(
			"更新菜单:查询更新前的菜单详情失败",
			zap.Error(rErr),
			zap.Uint32("menu_id", menuID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, rErr
	}

	if err := s.menuRepo.RemoveGroupPolicy(ctx, *om, false); err != nil {
		log.Error(
			"更新菜单:删除原菜单组策略失败",
			zap.Error(err),
			zap.Object("menu_model", om),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.FromError(err)
	}

	recoverOldPolicy := func() {
		if err := s.menuRepo.AddGroupPolicy(ctx, *om); err != nil {
			log.Error(
				"更新菜单:恢复旧菜单组策略失败,请手动添加策略",
				zap.Error(err),
				zap.Object("menu_model", om),
				zap.Duration("total_duration", time.Since(startTime)),
			)
		}
	}

	updateData := dto.ToUpdateMap()
	if err := s.menuRepo.UpdateModel(ctx, updateData, apis, "id = ?", menuID); err != nil {
		log.Error(
			"更新菜单:更新数据库模型失败",
			zap.Error(err),
			zap.Uint32("menu_id", menuID),
			zap.Any("update_data", updateData),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		recoverOldPolicy()
		return nil, errors.NewGormError(err, updateData)
	}

	rollback := func() {
		upData := om.ToUpdateMap()
		if upErr := s.menuRepo.UpdateModel(ctx, upData, om.Apis, "id = ?", menuID); upErr != nil {
			log.Error(
				"更新菜单:回滚数据库数据失败，请手动处理脏数据",
				zap.Error(upErr),
				zap.Uint32("menu_id", menuID),
				zap.Duration("total_duration", time.Since(startTime)),
			)
		}
	}

	nm, rErr := s.FindMenuByID(ctx, []string{"Parent", "Apis"}, menuID)
	if rErr != nil {
		log.Error(
			"更新菜单:查询更新后的菜单详情失败",
			zap.Error(rErr),
			zap.Uint32("menu_id", menuID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		rollback()
		recoverOldPolicy()
		return nil, rErr
	}

	if err := s.menuRepo.AddGroupPolicy(ctx, *nm); err != nil {
		log.Error(
			"更新菜单:添加新菜单组策略失败",
			zap.Error(err),
			zap.Object("menu_model", nm),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		rollback()
		recoverOldPolicy()
		return nil, errors.FromError(err)
	}

	log.Info(
		"更新菜单:执行成功",
		zap.Uint32("menu_id", menuID),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nm, nil
}

func (s *MenuService) DeleteMenuByID(
	ctx context.Context,
	menuID uint32,
) *errors.Error {
	startTime := time.Now()
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"删除菜单:开始执行",
		zap.Uint32("menu_id", menuID),
	)

	m, fErr := s.FindMenuByID(ctx, []string{"Parent", "Apis"}, menuID)
	if fErr != nil {
		log.Error(
			"删除菜单:查询菜单详情失败",
			zap.Error(fErr),
			zap.Uint32("menu_id", menuID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return fErr
	}

	if err := s.menuRepo.RemoveGroupPolicy(ctx, *m, true); err != nil {
		log.Error(
			"删除菜单:移除按钮组策略失败",
			zap.Error(err),
			zap.Object("menu_model", m),
			zap.Bool("remove_inherited", true),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.FromError(err)
	}

	recoverOldPolicy := func() {
		if err := s.menuRepo.AddGroupPolicy(ctx, *m); err != nil {
			log.Error(
				"删除菜单:恢复旧菜单组策略失败,请手动添加策略",
				zap.Error(err),
				zap.Object("menu_model", m),
				zap.Duration("total_duration", time.Since(startTime)),
			)
		}
	}

	if err := s.menuRepo.DeleteModel(ctx, menuID); err != nil {
		log.Error(
			"删除菜单:删除数据库模型失败",
			zap.Error(err),
			zap.Uint32("menu_id", menuID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		recoverOldPolicy()
		return errors.NewGormError(err, map[string]any{"id": menuID})
	}

	log.Info(
		"删除菜单:执行成功",
		zap.Uint32("menu_id", menuID),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

func (s *MenuService) FindMenuByID(
	ctx context.Context,
	preloads []string,
	menuID uint32,
) (*sysmodel.MenuModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	m, err := s.menuRepo.GetModel(ctx, preloads, menuID)
	if err != nil {
		log.Error(
			"查询菜单:查询数据库模型失败",
			zap.Error(err),
			zap.Strings("preloads", preloads),
			zap.Uint32("menu_id", menuID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, map[string]any{"id": menuID})
	}

	log.Debug(
		"查询菜单:查询到的数据库模型详情",
		zap.Object("menu_model", m),
	)
	return m, nil
}

func (s *MenuService) ListMenu(
	ctx context.Context,
	dto sysmodel.ListMenuDTO,
) (int, int, int64, []sysmodel.MenuModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return 0, 0, 0, nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Debug(
		"查询菜单列表:入参详情",
		zap.Object("list_menu_dto", &dto),
	)

	page, size := dto.StandardModelQuery.GetPageParam()
	limit, offset := common.Page2LimitOffset(page, size)
	qp := database.QueryParams{
		Limit:   limit,
		Offset:  offset,
		OrderBy: []string{"id ASC"},
		Query:   dto.ToQueryMap(),
	}

	count, err := s.menuRepo.CountModel(ctx, qp.Query)
	if err != nil {
		log.Error(
			"查询菜单列表:查询数据库模型总数失败",
			zap.Error(err),
			zap.Any("query", qp.Query),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return page, size, 0, nil, errors.NewGormError(err, nil)
	}

	if count == 0 {
		log.Warn(
			"查询菜单列表:数据库模型总数为0",
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return page, size, 0, nil, nil
	}

	ms, err := s.menuRepo.ListModel(ctx, qp)
	if err != nil {
		log.Error(
			"查询菜单列表:查询数据库模型失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return page, size, 0, nil, errors.NewGormError(err, nil)
	}

	return page, size, count, ms, nil
}

func (s *MenuService) LoadMenuPolicy(
	ctx context.Context,
) error {
	startTime := time.Now()
	s.log.Debug("加载菜单策略:开始执行")

	qp := database.QueryParams{
		Preloads: []string{"Apis"},
		Columns:  []string{"id", "parent_id"},
	}

	ms, err := s.menuRepo.ListModel(ctx, qp)
	if err != nil {
		s.log.Error(
			"加载菜单策略:查询数据库模型列表失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.NewGormError(err, nil)
	}

	menuCount := int64(len(ms))
	if menuCount == 0 {
		s.log.Warn(
			"加载菜单策略:数据库模型总数为0",
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil
	}

	failCount := int64(0)
	for i := range ms {
		if err := s.menuRepo.AddGroupPolicy(ctx, ms[i]); err != nil {
			s.log.Error(
				"加载菜单策略:添加菜单组策略失败",
				zap.Error(err),
				zap.Uint32("menu_id", ms[i].ID),
				zap.Duration("total_duration", time.Since(startTime)),
			)
			failCount++
		}
	}

	// 最终日志
	totalDuration := time.Since(startTime)
	s.log.Info(
		"加载菜单策略:执行完成",
		zap.Int64("total_menu", menuCount),
		zap.Int64("success_count", menuCount-failCount),
		zap.Int64("fail_count", failCount),
		zap.Duration("total_duration", totalDuration),
	)

	// 有失败但不阻断启动
	if failCount > 0 {
		return emperror.Errorf("部分菜单策略加载失败，失败数量：%d", failCount)
	}
	return nil
}
