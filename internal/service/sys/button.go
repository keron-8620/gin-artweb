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

func (s *ButtonService) GetMenu(
	ctx context.Context,
	menuID uint32,
) (*sysmodel.MenuModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Debug(
		"查询按钮关联的菜单：开始执行",
		zap.Uint32("menu_id", menuID),
	)

	m, err := s.menuRepo.GetModel(ctx, nil, menuID)
	if err != nil {
		log.Error(
			"查询按钮关联的菜单：查询数据库失败",
			zap.Error(err),
			zap.Uint32("menu_id", menuID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, map[string]any{"menu_id": menuID})
	}

	log.Info(
		"查询按钮关联的菜单：执行成功",
		zap.Uint32("menu_id", menuID),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return m, nil
}

func (s *ButtonService) GetApis(
	ctx context.Context,
	apiIDs []uint32,
) ([]sysmodel.ApiModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"查询按钮关联的API列表：开始执行",
		zap.Uint32s("api_ids", apiIDs),
	)

	if len(apiIDs) == 0 {
		log.Info(
			"查询按钮关联的API列表：API ID列表为空",
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
			"查询按钮关联的API列表：查询数据库失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, nil)
	}
	log.Debug(
		"查询按钮关联的API列表：查询数据库成功",
		zap.Uint32s("api_ids", sysmodel.ListApiModelToUint32s(ms)),
	)

	log.Info(
		"查询按钮关联的API列表：执行成功",
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return ms, nil
}

func (s *ButtonService) CreateButton(
	ctx context.Context,
	dto sysmodel.CreateButtonDTO,
) (*sysmodel.ButtonModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info("创建按钮：开始执行")

	log.Debug(
		"创建按钮：输入参数",
		zap.Object("create_button_dto", &dto),
	)

	m := sysmodel.ButtonModel{
		StandardModel: database.StandardModel{
			BaseModel: database.BaseModel{ID: dto.ID},
		},
		Name:     dto.Name,
		Sort:     dto.Sort,
		IsActive: dto.IsActive,
		Descr:    dto.Descr,
		MenuID:   dto.MenuID,
	}

	menu, rErr := s.GetMenu(ctx, m.MenuID)
	if rErr != nil {
		log.Error(
			"创建按钮：查询菜单失败",
			zap.Error(rErr),
			zap.Uint32("menu_id", m.MenuID),
		)
		return nil, rErr
	}
	log.Debug(
		"创建按钮：查询菜单成功",
		zap.Uint32("menu_id", m.MenuID),
	)
	m.Menu = *menu

	apis, rErr := s.GetApis(ctx, dto.ApiIDs)
	if rErr != nil {
		log.Error(
			"创建按钮：查询按钮关联的权限列表失败",
			zap.Error(rErr),
			zap.Uint32s("api_ids", dto.ApiIDs),
		)
		return nil, rErr
	}
	log.Debug("创建按钮：查询按钮关联的权限列表成功")

	createStepStart := time.Now()
	log.Debug(
		"创建按钮：开始创建数据库模型",
		zap.Object("button_model", &m),
	)
	if err := s.buttonRepo.CreateModel(ctx, &m, apis); err != nil {
		log.Error(
			"创建按钮：创建数据库模型失败",
			zap.Error(err),
			zap.Object("button_model", &m),
			zap.Duration("create_step_duration", time.Since(createStepStart)),
		)
		return nil, errors.NewGormError(err, nil)
	}
	createStepDuration := time.Since(createStepStart)
	log.Debug(
		"创建按钮：创建数据库模型成功",
		zap.Object("button_model", &m),
		zap.Duration("create_step_duration", createStepDuration),
	)
	if len(apis) > 0 {
		m.Apis = apis
	}

	addPolicyStepStart := time.Now()
	log.Debug(
		"创建按钮：开始添加按钮组策略",
		zap.Object("button_model", &m),
		zap.Uint32s("api_ids", dto.ApiIDs),
	)
	if err := s.buttonRepo.AddGroupPolicy(ctx, &m); err != nil {
		log.Error(
			"创建按钮：添加策略失败",
			zap.Error(err),
			zap.Object("button_model", &m),
			zap.Duration("add_policy_step_duration", time.Since(addPolicyStepStart)),
		)
		return nil, errors.FromError(err)
	}
	addPolicyStepDuration := time.Since(addPolicyStepStart)
	log.Debug(
		"创建按钮：添加策略耗时",
		zap.Object("button_model", &m),
		zap.Duration("add_policy_step_duration", addPolicyStepDuration),
	)

	log.Info(
		"创建按钮：执行成功",
		zap.Uint32("button_id", m.ID),
		zap.Duration("create_step_duration", createStepDuration),
		zap.Duration("add_policy_step_duration", addPolicyStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return &m, nil
}

func (s *ButtonService) UpdateButtonByID(
	ctx context.Context,
	buttonID uint32,
	dto sysmodel.UpdateButtonDTO,
) (*sysmodel.ButtonModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"更新按钮：开始执行",
		zap.Uint32("button_id", buttonID),
		zap.Object("update_button_dto", &dto),
	)

	log.Debug(
		"更新按钮：输入参数",
		zap.Uint32("button_id", buttonID),
		zap.Object("update_button_dto", &dto),
	)

	apis, rErr := s.GetApis(ctx, dto.ApiIDs)
	if rErr != nil {
		log.Error(
			"更新按钮：查询按钮关联的权限列表失败",
			zap.Error(rErr),
			zap.Uint32s("api_ids", dto.ApiIDs),
		)
		return nil, rErr
	}
	log.Debug("更新按钮：查询按钮关联的权限列表成功")

	updateData := dto.ToUpdateMap()
	updateStepStart := time.Now()
	log.Debug(
		"更新按钮：开始更新数据库模型",
		zap.Any("update_data", updateData),
		zap.Uint32("button_id", buttonID),
	)
	if err := s.buttonRepo.UpdateModel(ctx, updateData, apis, "id = ?", buttonID); err != nil {
		log.Error(
			"更新按钮：更新数据库模型失败",
			zap.Error(err),
			zap.Any("update_data", updateData),
			zap.Uint32("button_id", buttonID),
			zap.Duration("update_step_duration", time.Since(updateStepStart)),
		)
		return nil, errors.NewGormError(err, updateData)
	}
	updateStepDuration := time.Since(updateStepStart)
	log.Debug(
		"更新按钮：更新数据库模型成功",
		zap.Uint32("button_id", buttonID),
		zap.Duration("update_step_duration", updateStepDuration),
	)

	var m *sysmodel.ButtonModel
	m, rErr = s.FindButtonByID(ctx, []string{"Menu", "Apis"}, buttonID)
	if rErr != nil {
		log.Error(
			"更新按钮：查询更新后的按钮详情失败",
			zap.Error(rErr),
			zap.Uint32("button_id", buttonID),
		)
		return nil, rErr
	}

	removePolicyStepStart := time.Now()
	log.Debug(
		"更新按钮：开始移除旧策略",
		zap.Uint32("button_id", buttonID),
		zap.Bool("remove_inherited", false),
	)
	if err := s.buttonRepo.RemoveGroupPolicy(ctx, m, false); err != nil {
		log.Error(
			"更新按钮：移除旧策略失败",
			zap.Error(err),
			zap.Object("button_model", m),
			zap.Bool("remove_inherited", false),
			zap.Duration("remove_policy_step_duration", time.Since(removePolicyStepStart)),
		)
		return nil, errors.FromError(err)
	}
	removePolicyStepDuration := time.Since(removePolicyStepStart)
	log.Debug(
		"更新按钮：移除旧策略耗时",
		zap.Uint32("button_id", buttonID),
		zap.Duration("remove_policy_step_duration", removePolicyStepDuration),
	)

	addPolicyStepStart := time.Now()
	log.Debug(
		"更新按钮：开始添加新策略",
		zap.Object("button_model", m),
		zap.Uint32s("api_ids", dto.ApiIDs),
	)
	if err := s.buttonRepo.AddGroupPolicy(ctx, m); err != nil {
		log.Error(
			"更新按钮：添加新策略失败",
			zap.Error(err),
			zap.Object("button_model", m),
			zap.Duration("add_policy_step_duration", time.Since(addPolicyStepStart)),
		)
		return nil, errors.FromError(err)
	}
	addPolicyStepDuration := time.Since(addPolicyStepStart)
	log.Debug(
		"更新按钮：添加新策略成功",
		zap.Uint32("button_id", buttonID),
		zap.Duration("add_policy_step_duration", addPolicyStepDuration),
	)

	log.Info(
		"更新按钮：执行成功",
		zap.Uint32("button_id", buttonID),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return m, nil
}

func (s *ButtonService) DeleteButtonByID(
	ctx context.Context,
	buttonID uint32,
) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"删除按钮：开始执行",
		zap.Uint32("button_id", buttonID),
	)

	m, rErr := s.FindButtonByID(ctx, []string{"Menu", "Apis"}, buttonID)
	if rErr != nil {
		log.Error(
			"删除按钮：查询按钮详情失败",
			zap.Error(rErr),
			zap.Uint32("button_id", buttonID),
		)
		return rErr
	}

	deleteStepStart := time.Now()
	log.Debug(
		"删除按钮：开始删除数据库模型",
		zap.Uint32("button_id", buttonID),
	)
	if err := s.buttonRepo.DeleteModel(ctx, buttonID); err != nil {
		log.Error(
			"删除按钮：删除数据库模型失败",
			zap.Error(err),
			zap.Uint32("button_id", buttonID),
			zap.Duration("delete_step_duration", time.Since(deleteStepStart)),
		)
		return errors.NewGormError(err, map[string]any{"id": buttonID})
	}
	deleteStepDuration := time.Since(deleteStepStart)
	log.Debug(
		"删除按钮：删除数据库模型成功",
		zap.Uint32("button_id", buttonID),
		zap.Duration("delete_step_duration", deleteStepDuration),
	)

	removePolicyStepStart := time.Now()
	log.Debug(
		"删除按钮：开始移除策略",
		zap.Uint32("button_id", buttonID),
		zap.Bool("remove_inherited", true),
	)
	if err := s.buttonRepo.RemoveGroupPolicy(ctx, m, true); err != nil {
		log.Error(
			"删除按钮：移除策略失败",
			zap.Error(err),
			zap.Object("button_model", m),
			zap.Bool("remove_inherited", true),
			zap.Duration("remove_policy_step_duration", time.Since(removePolicyStepStart)),
		)
		return errors.FromError(err)
	}
	removePolicyStepDuration := time.Since(removePolicyStepStart)
	log.Debug(
		"删除按钮：移除策略成功",
		zap.Uint32("button_id", buttonID),
		zap.Duration("remove_policy_step_duration", removePolicyStepDuration),
	)

	log.Info(
		"删除按钮：执行成功",
		zap.Uint32("button_id", buttonID),
		zap.Duration("delete_step_duration", deleteStepDuration),
		zap.Duration("remove_policy_step_duration", removePolicyStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

func (s *ButtonService) FindButtonByID(
	ctx context.Context,
	preloads []string,
	buttonID uint32,
) (*sysmodel.ButtonModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"查询按钮：开始执行",
		zap.Strings("preloads", preloads),
		zap.Uint32("button_id", buttonID),
	)

	m, err := s.buttonRepo.GetModel(ctx, preloads, buttonID)
	if err != nil {
		log.Error(
			"查询按钮：查询数据库模型失败",
			zap.Error(err),
			zap.Strings("preloads", preloads),
			zap.Uint32("button_id", buttonID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, map[string]any{"id": buttonID})
	}
	log.Debug(
		"查询按钮：查询到的数据库模型详情",
		zap.Object("button_model", m),
	)

	log.Info(
		"查询按钮：执行成功",
		zap.Uint32("button_id", buttonID),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return m, nil
}

func (s *ButtonService) ListButton(
	ctx context.Context,
	page, size int,
	dto sysmodel.ListButtonDTO,
) (int64, []sysmodel.ButtonModel, *errors.Error) {
	if ctx.Err() != nil {
		return 0, nil, errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info("查询按钮列表：开始执行")

	log.Debug(
		"查询按钮列表：参数详情",
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

	log.Debug(
		"查询按钮列表：查询数据库模型参数",
		zap.Object("query_params", &qp),
	)

	countStepStart := time.Now()
	log.Debug(
		"查询按钮列表：开始查询数据库模型总数",
		zap.Object("query_params", &qp),
	)
	count, err := s.buttonRepo.CountModel(ctx, qp.Query)
	countStepDuration := time.Since(countStepStart)
	if err != nil {
		log.Error(
			"查询按钮列表：查询数据库模型总数失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("count_step_duration", countStepDuration),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}
	log.Debug(
		"查询按钮列表：查询数据库模型总数成功",
		zap.Int64("total_count", count),
		zap.Duration("count_step_duration", countStepDuration),
	)
	if count == 0 {
		log.Warn(
			"查询按钮列表：数据库模型总数为0",
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return count, nil, nil
	}

	ms, err := s.buttonRepo.ListModel(ctx, qp)
	if err != nil {
		log.Error(
			"查询按钮列表：查询数据库模型失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}

	log.Info(
		"查询按钮列表：执行成功",
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return count, ms, nil
}

func (s *ButtonService) LoadButtonPolicy(ctx context.Context) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	traceID := ctxutil.GetTraceID(ctx)
	log := s.log.With(zap.String("trace_id", traceID))

	log.Info(
		"加载按钮策略：开始执行",
	)

	qp := database.QueryParams{
		Preloads: []string{"Apis"},
		Columns:  []string{"id", "menu_id"},
	}

	listStepStart := time.Now()
	log.Debug(
		"加载按钮策略：开始查询数据库模型列表",
		zap.Object("query_params", &qp),
	)
	ms, err := s.buttonRepo.ListModel(ctx, qp)
	listStepDuration := time.Since(listStepStart)
	if err != nil {
		log.Error(
			"加载按钮策略：查询数据库模型列表失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("list_step_duration", listStepDuration),
		)
		return errors.NewGormError(err, nil)
	}
	log.Debug(
		"加载按钮策略：查询数据库模型列表成功",
		zap.Duration("list_step_duration", listStepDuration),
	)

	var policyCount int
	if len(ms) > 0 {
		policyStepStart := time.Now()
		log.Debug(
			"加载按钮策略：开始添加按钮组策略",
			zap.Object("query_params", &qp),
		)
		policyCount = len(ms)
		for i := range ms {
			if err := s.buttonRepo.AddGroupPolicy(ctx, &ms[i]); err != nil {
				log.Error(
					"加载按钮策略：添加策略失败",
					zap.Error(err),
					zap.Uint32("button_id", ms[i].ID),
				)
				return errors.FromError(err)
			}
		}
		policyStepDuration := time.Since(policyStepStart)
		log.Debug(
			"加载按钮策略：添加按钮组策略成功",
			zap.Int("policy_count", policyCount),
			zap.Duration("policy_step_duration", policyStepDuration),
		)
	}

	log.Info(
		"加载按钮策略：执行成功",
		zap.Int("policy_count", policyCount),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}
