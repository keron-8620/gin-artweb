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

type ApiService struct {
	log     *zap.Logger
	apiRepo *sysrepo.ApiRepo
}

func NewApiService(
	log *zap.Logger,
	apiRepo *sysrepo.ApiRepo,
) *ApiService {
	return &ApiService{
		log:     log,
		apiRepo: apiRepo,
	}
}

func (s *ApiService) CreateApi(
	ctx context.Context,
	dto sysmodel.CreateApiDTO,
) (*sysmodel.ApiModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info("创建API:开始执行")

	log.Debug(
		"创建API:入参详情",
		zap.Object("create_api_dto", &dto),
	)

	m := sysmodel.ApiModel{
		StandardModel: database.StandardModel{
			BaseModel: database.BaseModel{ID: dto.ID},
		},
		URL:    dto.URL,
		Method: dto.Method,
		Label:  dto.Label,
		Descr:  dto.Descr,
	}

	createStepStart := time.Now()
	log.Debug(
		"创建API:开始创建数据库模型",
		zap.Object("api_model", &m),
	)
	if err := s.apiRepo.CreateModel(ctx, &m); err != nil {
		log.Error(
			"创建API:创建数据库模型失败",
			zap.Error(err),
			zap.Object("api_model", &m),
			zap.Duration("create_step_duration", time.Since(createStepStart)),
		)
		return nil, errors.NewGormError(err, nil)
	}
	createStepDuration := time.Since(createStepStart)
	log.Debug(
		"创建API:创建数据库模型成功",
		zap.Object("api_model", &m),
		zap.Duration("create_step_duration", createStepDuration),
	)

	policyStepStart := time.Now()
	log.Debug(
		"创建API:开始添加访问策略",
		zap.Uint32("api_id", m.ID),
	)
	if err := s.apiRepo.AddPolicy(ctx, m); err != nil {
		log.Error(
			"创建API:添加访问策略失败",
			zap.Error(err),
			zap.Object("api_model", &m),
			zap.Duration("policy_step_duration", time.Since(policyStepStart)),
		)
		return nil, errors.FromError(err)
	}
	policyStepDuration := time.Since(policyStepStart)
	log.Debug(
		"创建API:添加访问策略成功",
		zap.Uint32("api_id", m.ID),
		zap.Duration("policy_step_duration", policyStepDuration),
	)

	log.Info(
		"创建API:执行成功",
		zap.Uint32("api_id", m.ID),
		zap.Duration("create_step_duration", createStepDuration),
		zap.Duration("policy_step_duration", policyStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return &m, nil
}

func (s *ApiService) UpdateApiByID(
	ctx context.Context,
	apiID uint32,
	dto sysmodel.UpdateApiDTO,
) (*sysmodel.ApiModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"更新API:开始执行",
		zap.Uint32("api_id", apiID),
	)

	log.Debug(
		"更新API:入参详情",
		zap.Uint32("api_id", apiID),
		zap.Object("update_api_dto", &dto),
	)

	updateData := dto.ToUpdateMap()
	log.Debug(
		"更新API:转换为数据库更新参数",
		zap.Uint32("api_id", apiID),
		zap.Any("update_data", updateData),
	)

	updateStepStart := time.Now()
	log.Debug(
		"更新API:开始更新数据库模型",
		zap.Uint32("api_id", apiID),
		zap.Any("update_data", updateData),
	)
	if err := s.apiRepo.UpdateModel(ctx, updateData, "id = ?", apiID); err != nil {
		log.Error(
			"更新API:更新数据库模型失败",
			zap.Error(err),
			zap.Uint32("api_id", apiID),
			zap.Any("update_data", updateData),
			zap.Duration("update_step_duration", time.Since(updateStepStart)),
		)
		return nil, errors.NewGormError(err, updateData)
	}
	updateStepDuration := time.Since(updateStepStart)
	log.Debug(
		"更新API:更新数据库模型成功",
		zap.Uint32("api_id", apiID),
		zap.Duration("update_step_duration", updateStepDuration),
	)

	m, rErr := s.FindApiByID(ctx, apiID)
	if rErr != nil {
		log.Error(
			"更新API:查询更新后的API失败",
			zap.Error(rErr),
			zap.Uint32("api_id", apiID),
		)
		return nil, rErr
	}

	removePolicyStepStart := time.Now()
	log.Debug(
		"更新API:开始移除旧API策略",
		zap.Uint32("api_id", apiID),
		zap.Bool("remove_inherited", false),
	)
	if err := s.apiRepo.RemovePolicy(ctx, *m, false); err != nil {
		log.Error(
			"更新API:移除旧API策略失败",
			zap.Error(err),
			zap.Uint32("api_id", apiID),
			zap.Bool("remove_inherited", false),
			zap.Duration("remove_policy_step_duration", time.Since(removePolicyStepStart)),
		)
		return nil, errors.FromError(err)
	}
	removePolicyStepDuration := time.Since(removePolicyStepStart)
	log.Debug(
		"更新API:移除旧API策略成功",
		zap.Duration("remove_policy_step_duration", removePolicyStepDuration),
	)

	addPolicyStepStart := time.Now()
	log.Debug(
		"更新API:开始添加新API策略",
		zap.Uint32("api_id", apiID),
	)
	if err := s.apiRepo.AddPolicy(ctx, *m); err != nil {
		log.Error(
			"更新API:添加新API策略失败",
			zap.Error(err),
			zap.Uint32("api_id", apiID),
			zap.Duration("add_policy_step_duration", time.Since(addPolicyStepStart)),
		)
		return nil, errors.FromError(err)
	}
	addPolicyStepDuration := time.Since(addPolicyStepStart)
	log.Debug(
		"更新API:添加新API策略成功",
		zap.Uint32("api_id", apiID),
		zap.Duration("add_policy_step_duration", addPolicyStepDuration),
	)

	log.Info(
		"更新API:执行成功",
		zap.Uint32("api_id", apiID),
		zap.Duration("update_step_duration", updateStepDuration),
		zap.Duration("remove_policy_step_duration", removePolicyStepDuration),
		zap.Duration("add_policy_step_duration", addPolicyStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return m, nil
}

func (s *ApiService) DeleteApiByID(
	ctx context.Context,
	apiID uint32,
) *errors.Error {
	startTime := time.Now()
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"删除API:开始执行",
		zap.Uint32("api_id", apiID),
	)

	m, rErr := s.FindApiByID(ctx, apiID)
	if rErr != nil {
		log.Error(
			"删除API:查询待删除API失败",
			zap.Error(rErr),
			zap.Uint32("api_id", apiID),
		)
		return rErr
	}

	deleteStepStart := time.Now()
	log.Debug(
		"删除API:开始删除数据库模型",
		zap.Uint32("api_id", apiID),
	)
	if err := s.apiRepo.DeleteModel(ctx, apiID); err != nil {
		log.Error(
			"删除API:删除数据库模型失败",
			zap.Error(err),
			zap.Uint32("api_id", apiID),
			zap.Duration("delete_step_duration", time.Since(deleteStepStart)),
		)
		return errors.NewGormError(err, map[string]any{"id": apiID})
	}
	deleteStepDuration := time.Since(deleteStepStart)
	log.Debug(
		"删除API:删除数据库模型成功",
		zap.Uint32("api_id", apiID),
		zap.Duration("delete_step_duration", deleteStepDuration),
	)

	removePolicyStepStart := time.Now()
	log.Debug(
		"删除API:开始移除API策略",
		zap.Uint32("api_id", apiID),
		zap.Bool("remove_inherited", true),
	)
	if err := s.apiRepo.RemovePolicy(ctx, *m, true); err != nil {
		log.Error(
			"删除API:移除API策略失败",
			zap.Error(err),
			zap.Object("api_model", m),
			zap.Bool("remove_inherited", true),
			zap.Duration("remove_policy_step_duration", time.Since(removePolicyStepStart)),
		)
		return errors.FromError(err)
	}
	removePolicyStepDuration := time.Since(removePolicyStepStart)
	log.Debug(
		"删除API:移除API策略成功",
		zap.Uint32("api_id", apiID),
		zap.Duration("remove_policy_step_duration", removePolicyStepDuration),
	)

	log.Info(
		"删除API:执行成功",
		zap.Uint32("api_id", apiID),
		zap.Duration("delete_step_duration", deleteStepDuration),
		zap.Duration("remove_policy_step_duration", removePolicyStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

func (s *ApiService) FindApiByID(
	ctx context.Context,
	apiID uint32,
) (*sysmodel.ApiModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"查询API:开始执行",
		zap.Uint32("api_id", apiID),
	)

	m, err := s.apiRepo.GetModel(ctx, apiID)
	if err != nil {
		log.Error(
			"查询API:查询数据库模型失败",
			zap.Error(err),
			zap.Uint32("api_id", apiID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, map[string]any{"id": apiID})
	}
	log.Debug(
		"查询API:查询到的数据库模型详情",
		zap.Object("api_model", m),
	)

	log.Info(
		"查询API:执行成功",
		zap.Uint32("api_id", apiID),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return m, nil
}

func (s *ApiService) ListApi(
	ctx context.Context,
	page, size int,
	dto sysmodel.ListApiDTO,
) (int64, []sysmodel.ApiModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return 0, nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info("查询API列表:开始执行")

	log.Debug(
		"查询API列表:入参详情",
		zap.Int("page", page),
		zap.Int("size", size),
		zap.Object("list_api_dto", &dto),
	)

	limit, offset := common.Page2LimitOffset(page, size)
	qp := database.QueryParams{
		Limit:   limit,
		Offset:  offset,
		OrderBy: []string{"id ASC"},
		Query:   dto.ToQueryMap(),
	}

	log.Debug(
		"查询API列表:查询数据库模型参数",
		zap.Object("query_params", &qp),
	)

	countStepStart := time.Now()
	log.Debug(
		"查询API列表:开始查询数据库模型总数",
		zap.Any("query", qp.Query),
	)
	count, err := s.apiRepo.CountModel(ctx, qp.Query)
	if err != nil {
		log.Error(
			"查询API列表:查询数据库模型总数失败",
			zap.Error(err),
			zap.Any("query", qp.Query),
			zap.Duration("count_step_duration", time.Since(countStepStart)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}
	countStepDuration := time.Since(countStepStart)
	log.Debug(
		"查询API列表:查询数据库模型总数成功",
		zap.Any("query", qp.Query),
		zap.Duration("count_step_duration", countStepDuration),
	)
	if count == 0 {
		return 0, nil, nil
	}

	listStepStart := time.Now()
	log.Debug(
		"查询API列表:开始查询数据库模型",
		zap.Any("query", qp.Query),
	)
	ms, err := s.apiRepo.ListModel(ctx, qp)
	if err != nil {
		log.Error(
			"查询API列表:查询数据库模型失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("list_step_duration", time.Since(listStepStart)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}
	listStepDuration := time.Since(listStepStart)
	log.Info(
		"查询API列表:执行成功",
		zap.Duration("list_step_duration", listStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return count, ms, nil
}

func (s *ApiService) LoadApiPolicy(
	ctx context.Context,
) *errors.Error {
	startTime := time.Now()
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	s.log.Debug("加载API策略:开始执行")

	qp := database.QueryParams{
		Columns: []string{"id", "url", "method"},
	}

	dbStartTime := time.Now()
	s.log.Debug(
		"加载API策略:开始查询数据库模型列表",
		zap.Object("query_params", &qp),
	)
	ms, err := s.apiRepo.ListModel(ctx, qp)
	if err != nil {
		s.log.Error(
			"加载API策略:查询数据库模型列表失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("db_step_duration", time.Since(dbStartTime)),
		)
		return errors.NewGormError(err, nil)
	}
	dbStepDuration := time.Since(dbStartTime)
	s.log.Debug(
		"加载API策略:查询数据库模型列表成功",
		zap.Duration("db_step_duration", dbStepDuration),
	)

	var policyCount int
	if len(ms) > 0 {
		policyStepStart := time.Now()
		s.log.Debug(
			"加载API策略:开始添加策略",
			zap.Object("query_params", &qp),
		)
		policyCount = len(ms)
		for i := range ms {
			if err := s.apiRepo.AddPolicy(ctx, ms[i]); err != nil {
				s.log.Error(
					"加载API策略:添加策略失败",
					zap.Error(err),
					zap.Object("api_model", &ms[i]),
				)
				return errors.FromError(err)
			}
		}
		policyStepDuration := time.Since(policyStepStart)
		s.log.Debug(
			"加载API策略:添加策略成功",
			zap.Int("policy_count", policyCount),
			zap.Duration("policy_step_duration", policyStepDuration),
		)
	}

	s.log.Debug(
		"加载API策略:执行成功",
		zap.Int("policy_count", policyCount),
		zap.Duration("total_duration", time.Since(startTime)),
		zap.Duration("db_step_duration", dbStepDuration),
	)
	return nil
}
