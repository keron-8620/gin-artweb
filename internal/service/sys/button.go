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

type ButtonService struct {
	log        *zap.Logger
	apiRepo    *sysrepo.ApiRepo
	menuRepo   *sysrepo.MenuRepo
	buttonRepo *sysrepo.ButtonRepo
}

func NewButtonService(
	log *zap.Logger,
	apiRepo *sysrepo.ApiRepo,
	menuRepo *sysrepo.MenuRepo,
	buttonRepo *sysrepo.ButtonRepo,
) *ButtonService {
	return &ButtonService{
		log:        log,
		apiRepo:    apiRepo,
		menuRepo:   menuRepo,
		buttonRepo: buttonRepo,
	}
}

func (s *ButtonService) getMenu(
	ctx context.Context,
	menuID uint32,
) (*sysmodel.MenuModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	m, err := s.menuRepo.GetModel(ctx, nil, menuID)
	if err != nil {
		log.Error(
			"查询按钮关联的菜单:查询数据库失败",
			zap.Error(err),
			zap.Uint32("menu_id", menuID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, map[string]any{"menu_id": menuID})
	}

	return m, nil
}

func (s *ButtonService) getApis(
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
		"查询按钮关联的权限列表:查询参数",
		zap.Uint32s("api_ids", apiIDs),
	)

	ms, err := s.apiRepo.ListModel(ctx, qp)
	if err != nil {
		log.Error(
			"查询按钮关联的权限列表:查询数据库失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, nil)
	}

	return ms, nil
}

func (s *ButtonService) CreateButton(
	ctx context.Context,
	dto sysmodel.CreateButtonDTO,
) (*sysmodel.ButtonModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"创建按钮:开始执行",
		zap.Object("create_button_dto", &dto),
	)

	m := dto.ToModel()

	menu, err := s.getMenu(ctx, m.MenuID)
	if err != nil {
		log.Error(
			"创建按钮:查询关联菜单失败",
			zap.Error(err),
			zap.Uint32("menu_id", m.MenuID),
		)
		return nil, err
	}
	m.Menu = *menu

	apis, err := s.getApis(ctx, dto.ApiIDs)
	if err != nil {
		log.Error(
			"创建按钮:查询关联API列表失败",
			zap.Error(err),
			zap.Uint32s("api_ids", dto.ApiIDs),
		)
		return nil, err
	}

	if err := s.buttonRepo.CreateModel(ctx, &m, apis); err != nil {
		log.Error(
			"创建按钮:创建数据库模型失败",
			zap.Error(err),
			zap.Object("button_model", &m),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, nil)
	}

	rollback := func() {
		if err := s.buttonRepo.DeleteModel(ctx, m.ID); err != nil {
			log.Error(
				"创建按钮:回滚数据库数据失败，请手动处理脏数据",
				zap.Error(err),
				zap.Uint32("button_id", m.ID),
				zap.Duration("total_duration", time.Since(startTime)),
			)
		}
	}

	if len(apis) > 0 {
		m.Apis = apis
	}

	if err := s.buttonRepo.AddGroupPolicy(ctx, m); err != nil {
		log.Error(
			"创建按钮:添加按钮的组策略失败",
			zap.Error(err),
			zap.Object("button_model", &m),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		rollback()
		return nil, errors.FromError(err)
	}

	log.Info(
		"创建按钮:执行成功",
		zap.Uint32("button_id", m.ID),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return &m, nil
}

func (s *ButtonService) UpdateButtonByID(
	ctx context.Context,
	buttonID uint32,
	dto sysmodel.UpdateButtonDTO,
) (*sysmodel.ButtonModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"更新按钮:开始执行",
		zap.Uint32("button_id", buttonID),
		zap.Object("update_button_dto", &dto),
	)

	apis, fErr := s.getApis(ctx, dto.ApiIDs)
	if fErr != nil {
		log.Error(
			"更新按钮:查询按钮关联的权限列表失败",
			zap.Error(fErr),
			zap.Uint32s("api_ids", dto.ApiIDs),
		)
		return nil, fErr
	}

	om, rErr := s.FindButtonByID(ctx, []string{"Menu", "Apis"}, buttonID)
	if rErr != nil {
		log.Error(
			"更新按钮:查询更新前的按钮详情失败",
			zap.Error(rErr),
			zap.Uint32("button_id", buttonID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, rErr
	}

	if err := s.buttonRepo.RemoveGroupPolicy(ctx, *om, false); err != nil {
		log.Error(
			"更新按钮:删除原按钮组策略失败",
			zap.Error(err),
			zap.Object("button_model", om),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.FromError(err)
	}

	recoverOldPolicy := func() {
		if err := s.buttonRepo.AddGroupPolicy(ctx, *om); err != nil {
			log.Error(
				"更新按钮:恢复旧按钮组策略失败,请手动添加策略",
				zap.Error(err),
				zap.Object("button_model", om),
				zap.Duration("total_duration", time.Since(startTime)),
			)
		}
	}

	updateData := dto.ToUpdateMap()
	if err := s.buttonRepo.UpdateModel(ctx, updateData, apis, "id = ?", buttonID); err != nil {
		log.Error(
			"更新按钮:更新数据库模型失败",
			zap.Error(err),
			zap.Any("update_data", updateData),
			zap.Uint32("button_id", buttonID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		recoverOldPolicy()
		return nil, errors.NewGormError(err, updateData)
	}

	rollback := func() {
		upData := om.ToUpdateMap()
		if upErr := s.buttonRepo.UpdateModel(ctx, upData, om.Apis, "id = ?", buttonID); upErr != nil {
			log.Error(
				"更新按钮:回滚数据库数据失败，请手动处理脏数据",
				zap.Error(upErr),
				zap.Uint32("button_id", buttonID),
				zap.Duration("total_duration", time.Since(startTime)),
			)
		}
	}

	m, rErr := s.FindButtonByID(ctx, []string{"Menu", "Apis"}, buttonID)
	if rErr != nil {
		log.Error(
			"更新按钮:查询更新后的按钮详情失败",
			zap.Error(rErr),
			zap.Uint32("button_id", buttonID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		rollback()
		recoverOldPolicy()
		return nil, rErr
	}

	if err := s.buttonRepo.AddGroupPolicy(ctx, *m); err != nil {
		log.Error(
			"更新按钮:添加新按钮组策略失败",
			zap.Error(err),
			zap.Object("button_model", m),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		rollback()
		recoverOldPolicy()
		return nil, errors.FromError(err)
	}

	log.Info(
		"更新按钮:执行成功",
		zap.Uint32("button_id", buttonID),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return m, nil
}

func (s *ButtonService) DeleteButtonByID(
	ctx context.Context,
	buttonID uint32,
) *errors.Error {
	startTime := time.Now()
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"删除按钮:开始执行",
		zap.Uint32("button_id", buttonID),
	)

	m, rErr := s.FindButtonByID(ctx, []string{"Menu", "Apis"}, buttonID)
	if rErr != nil {
		log.Error(
			"删除按钮:查询按钮详情失败",
			zap.Error(rErr),
			zap.Uint32("button_id", buttonID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return rErr
	}

	if err := s.buttonRepo.RemoveGroupPolicy(ctx, *m, true); err != nil {
		log.Error(
			"删除按钮:移除策略失败",
			zap.Error(err),
			zap.Object("button_model", m),
			zap.Bool("remove_inherited", true),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.FromError(err)
	}

	recoverOldPolicy := func() {
		if err := s.buttonRepo.AddGroupPolicy(ctx, *m); err != nil {
			log.Error(
				"删除按钮:恢复旧按钮组策略失败,请手动添加策略",
				zap.Error(err),
				zap.Object("button_model", m),
				zap.Duration("total_duration", time.Since(startTime)),
			)
		}
	}

	if err := s.buttonRepo.DeleteModel(ctx, buttonID); err != nil {
		log.Error(
			"删除按钮:删除数据库模型失败",
			zap.Error(err),
			zap.Uint32("button_id", buttonID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		recoverOldPolicy()
		return errors.NewGormError(err, map[string]any{"id": buttonID})
	}

	log.Info(
		"删除按钮:执行成功",
		zap.Uint32("button_id", buttonID),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

func (s *ButtonService) FindButtonByID(
	ctx context.Context,
	preloads []string,
	buttonID uint32,
) (*sysmodel.ButtonModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	m, err := s.buttonRepo.GetModel(ctx, preloads, buttonID)
	if err != nil {
		log.Error(
			"查询按钮:查询数据库模型失败",
			zap.Error(err),
			zap.Strings("preloads", preloads),
			zap.Uint32("button_id", buttonID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, map[string]any{"id": buttonID})
	}

	log.Debug(
		"查询按钮:查询到的数据库模型详情",
		zap.Object("button_model", m),
	)
	return m, nil
}

func (s *ButtonService) ListButton(
	ctx context.Context,
	page, size int,
	dto sysmodel.ListButtonDTO,
) (int64, []sysmodel.ButtonModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return 0, nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Debug(
		"查询按钮列表:入参详情",
		zap.Int("page", page),
		zap.Int("size", size),
		zap.Object("list_button_dto", &dto),
	)

	limit, offset := common.Page2LimitOffset(page, size)
	qp := database.QueryParams{
		Limit:   limit,
		Offset:  offset,
		OrderBy: []string{"id ASC"},
		Query:   dto.ToQueryMap(),
	}

	count, err := s.buttonRepo.CountModel(ctx, qp.Query)
	if err != nil {
		log.Error(
			"查询按钮列表:查询数据库模型总数失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}

	if count == 0 {
		log.Warn(
			"查询按钮列表:数据库模型总数为0",
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, nil, nil
	}

	ms, err := s.buttonRepo.ListModel(ctx, qp)
	if err != nil {
		log.Error(
			"查询按钮列表:查询数据库模型失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}
	return count, ms, nil
}

func (s *ButtonService) LoadButtonPolicy(
	ctx context.Context,
) error {
	startTime := time.Now()
	s.log.Debug("加载按钮策略:开始执行")

	qp := database.QueryParams{
		Preloads: []string{"Apis"},
		Columns:  []string{"id", "menu_id"},
	}

	ms, err := s.buttonRepo.ListModel(ctx, qp)
	if err != nil {
		s.log.Error(
			"加载按钮策略:查询数据库模型列表失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.NewGormError(err, nil)
	}

	buttonCount := int64(len(ms))
	if buttonCount == 0 {
		s.log.Warn(
			"加载按钮策略:数据库模型总数为0",
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil
	}

	failCount := int64(0)
	for i := range ms {
		if err := s.buttonRepo.AddGroupPolicy(ctx, ms[i]); err != nil {
			s.log.Error(
				"加载按钮策略:添加按钮组组策略失败",
				zap.Error(err),
				zap.Uint32("button_id", ms[i].ID),
			)
			failCount++
		}
	}

	// 最终日志
	totalDuration := time.Since(startTime)
	s.log.Info(
		"加载按钮策略:执行完成",
		zap.Int64("total_button", buttonCount),
		zap.Int64("success_count", buttonCount-failCount),
		zap.Int64("fail_count", failCount),
		zap.Duration("total_duration", totalDuration),
	)

	// 有失败但不阻断启动
	if failCount > 0 {
		return emperror.Errorf("部分按钮策略加载失败，失败数量：%d", failCount)
	}
	return nil
}
