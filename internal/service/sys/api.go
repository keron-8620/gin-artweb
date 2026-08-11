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

	log.Info(
		"创建API:开始执行",
		zap.Object("create_api_dto", &dto),
	)

	m := dto.ToModel()

	if err := s.apiRepo.CreateModel(ctx, &m); err != nil {
		log.Error(
			"创建API:创建数据库模型失败",
			zap.Error(err),
			zap.Object("api_model", &m),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, nil)
	}

	rollback := func() {
		if err := s.apiRepo.DeleteModel(ctx, m.ID); err != nil {
			log.Error(
				"创建API:回滚数据库数据失败，请手动处理脏数据",
				zap.Error(err),
				zap.Uint32("api_id", m.ID),
				zap.Duration("total_duration", time.Since(startTime)),
			)
		}
	}

	if err := s.apiRepo.AddPolicy(ctx, m); err != nil {
		log.Error(
			"创建API:添加访问策略失败",
			zap.Error(err),
			zap.Object("api_model", &m),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		rollback()
		return nil, errors.FromError(err)
	}

	log.Info(
		"创建API:执行成功",
		zap.Uint32("api_id", m.ID),
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
		zap.Object("update_api_dto", &dto),
	)

	oldApi, fErr := s.FindApiByID(ctx, apiID)
	if fErr != nil {
		log.Error(
			"更新API:查询更新前的API失败",
			zap.Error(fErr),
			zap.Uint32("api_id", apiID),
		)
		return nil, fErr
	}

	if err := s.apiRepo.RemovePolicy(ctx, *oldApi, false); err != nil {
		log.Error(
			"更新API:移除旧API策略失败",
			zap.Error(err),
			zap.Object("api_model", oldApi),
			zap.Bool("remove_inherited", false),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.FromError(err)
	}

	recoverOldPolicy := func() {
		if err := s.apiRepo.AddPolicy(ctx, *oldApi); err != nil {
			log.Error(
				"更新API:恢复旧API策略失败,请手动添加策略",
				zap.Error(err),
				zap.Object("api_model", oldApi),
				zap.Duration("total_duration", time.Since(startTime)),
			)
		}
	}

	updateData := dto.ToUpdateMap()
	if err := s.apiRepo.UpdateModel(ctx, updateData, "id = ?", apiID); err != nil {
		log.Error(
			"更新API:更新数据库模型失败",
			zap.Error(err),
			zap.Uint32("api_id", apiID),
			zap.Any("update_data", updateData),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		recoverOldPolicy()
		return nil, errors.NewGormError(err, updateData)
	}

	rollback := func() {
		upData := oldApi.ToUpdateMap()
		if err := s.apiRepo.UpdateModel(ctx, upData, "id = ?", apiID); err != nil {
			log.Error(
				"更新API:回滚数据库数据失败，请手动处理脏数据",
				zap.Error(err),
				zap.Uint32("api_id", apiID),
				zap.Any("update_data", upData),
				zap.Duration("total_duration", time.Since(startTime)),
			)
		}
	}

	newApi, fErr := s.FindApiByID(ctx, apiID)
	if fErr != nil {
		log.Error(
			"更新API:查询更新后的API失败",
			zap.Error(fErr),
			zap.Uint32("api_id", apiID),
		)
		rollback()
		recoverOldPolicy()
		return nil, fErr
	}

	if err := s.apiRepo.AddPolicy(ctx, *newApi); err != nil {
		log.Error(
			"更新API:添加新API策略失败",
			zap.Error(err),
			zap.Object("api_model", newApi),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		rollback()
		recoverOldPolicy()
		return nil, errors.FromError(err)
	}

	log.Info(
		"更新API:执行成功",
		zap.Uint32("api_id", apiID),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return newApi, nil
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
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return rErr
	}

	if err := s.apiRepo.RemovePolicy(ctx, *m, true); err != nil {
		log.Error(
			"删除API:移除API策略失败",
			zap.Error(err),
			zap.Object("api_model", m),
			zap.Bool("remove_inherited", true),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.FromError(err)
	}

	recoverOldPolicy := func() {
		if err := s.apiRepo.AddPolicy(ctx, *m); err != nil {
			log.Error(
				"删除API:恢复旧API策略失败,请手动添加策略",
				zap.Error(err),
				zap.Object("api_model", m),
				zap.Duration("total_duration", time.Since(startTime)),
			)
		}
	}

	if err := s.apiRepo.DeleteModel(ctx, apiID); err != nil {
		log.Error(
			"删除API:删除数据库模型失败，请手动清理脏数据",
			zap.Error(err),
			zap.Object("api_model", m),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		recoverOldPolicy()
		return errors.NewGormError(err, map[string]any{"id": apiID})
	}

	log.Info(
		"删除API:执行成功",
		zap.Uint32("api_id", apiID),
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

	return m, nil
}

func (s *ApiService) ListApi(
	ctx context.Context,
	dto sysmodel.ListApiDTO,
) (int, int, int64, []sysmodel.ApiModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return 0, 0, 0, nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Debug(
		"查询API列表:入参详情",
		zap.Object("list_api_dto", &dto),
	)

	page, size := dto.StandardModelQuery.GetPageParam()
	limit, offset := common.Page2LimitOffset(page, size)
	qp := database.QueryParams{
		Limit:   limit,
		Offset:  offset,
		OrderBy: []string{"id ASC"},
		Query:   dto.ToQueryMap(),
	}

	count, err := s.apiRepo.CountModel(ctx, qp.Query)
	if err != nil {
		log.Error(
			"查询API列表:查询数据库模型总数失败",
			zap.Error(err),
			zap.Any("query", qp.Query),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return page, size, 0, nil, errors.NewGormError(err, nil)
	}

	if count == 0 {
		log.Warn(
			"查询API列表:数据库模型总数为0",
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return page, size, 0, nil, nil
	}

	ms, err := s.apiRepo.ListModel(ctx, qp)
	if err != nil {
		log.Error(
			"查询API列表:查询数据库模型失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return page, size, 0, nil, errors.NewGormError(err, nil)
	}

	return page, size, count, ms, nil
}

func (s *ApiService) LoadApiPolicy(
	ctx context.Context,
) error {
	startTime := time.Now()
	s.log.Debug("加载API策略:开始执行")

	qp := database.QueryParams{
		Columns: []string{"id", "url", "method"},
	}

	ms, err := s.apiRepo.ListModel(ctx, qp)
	if err != nil {
		s.log.Error(
			"加载API策略:查询数据库模型列表失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.NewGormError(err, nil)
	}

	apiCount := int64(len(ms))
	if apiCount == 0 {
		s.log.Warn(
			"加载API策略:数据库模型总数为0",
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil
	}

	failCount := int64(0)
	for i := range ms {
		if err := s.apiRepo.AddPolicy(ctx, ms[i]); err != nil {
			s.log.Error(
				"加载API策略:添加策略失败",
				zap.Error(err),
				zap.Object("api_model", &ms[i]),
				zap.Duration("total_duration", time.Since(startTime)),
			)
			failCount++
		}
	}

	// 最终日志
	totalDuration := time.Since(startTime)
	s.log.Info(
		"加载API策略:执行完成",
		zap.Int64("total_api", apiCount),
		zap.Int64("success_count", apiCount-failCount),
		zap.Int64("fail_count", failCount),
		zap.Duration("total_duration", totalDuration),
	)

	// 有失败但不阻断启动
	if failCount > 0 {
		return emperror.Errorf("部分API策略加载失败，失败数量：%d", failCount)
	}
	return nil
}
