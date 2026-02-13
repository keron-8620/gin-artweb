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

type ApiUsecase struct {
	log     *zap.Logger
	apiRepo *data.ApiRepo
}

func NewApiUsecase(
	log *zap.Logger,
	apiRepo *data.ApiRepo,
) *ApiUsecase {
	return &ApiUsecase{
		log:     log,
		apiRepo: apiRepo,
	}
}

func (uc *ApiUsecase) CreateApi(
	ctx context.Context,
	m model.ApiModel,
) (*model.ApiModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	uc.log.Info(
		"开始创建api",
		zap.Object(database.ModelKey, &m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	if err := uc.apiRepo.CreateModel(ctx, &m); err != nil {
		uc.log.Error(
			"创建api失败",
			zap.Error(err),
			zap.Object(database.ModelKey, &m),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)

		return nil, errors.NewGormError(err, nil)
	}

	if err := uc.apiRepo.AddPolicy(ctx, m); err != nil {
		uc.log.Error(
			"添加api策略失败",
			zap.Error(err),
			zap.Object(database.ModelKey, &m),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.FromError(err)
	}

	uc.log.Info(
		"创建api成功",
		zap.Object(database.ModelKey, &m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return &m, nil
}

func (uc *ApiUsecase) UpdateApiByID(
	ctx context.Context,
	apiID uint32,
	data map[string]any,
) (*model.ApiModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	uc.log.Info(
		"开始更新api",
		zap.Uint32("api_id", apiID),
		zap.Any(database.UpdateDataKey, data),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	if err := uc.apiRepo.UpdateModel(ctx, data, "id = ?", apiID); err != nil {
		uc.log.Error(
			"更新api失败",
			zap.Error(err),
			zap.Uint32("api_id", apiID),
			zap.Any(database.UpdateDataKey, data),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, data)
	}

	m, rErr := uc.FindApiByID(ctx, apiID)
	if rErr != nil {
		uc.log.Error(
			"查询更新后的api失败",
			zap.Error(rErr),
			zap.Uint32("api_id", apiID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, rErr
	}

	if err := uc.apiRepo.RemovePolicy(ctx, *m, false); err != nil {
		uc.log.Error(
			"移除旧api策略失败",
			zap.Error(err),
			zap.Uint32("api_id", apiID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.FromError(err)
	}

	if err := uc.apiRepo.AddPolicy(ctx, *m); err != nil {
		uc.log.Error(
			"添加新api策略失败",
			zap.Error(err),
			zap.Uint32("api_id", apiID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.FromError(err)
	}

	uc.log.Info(
		"更新api成功",
		zap.Uint32("api_id", apiID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return m, nil
}

func (uc *ApiUsecase) DeleteApiByID(
	ctx context.Context,
	apiID uint32,
) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	uc.log.Info(
		"开始删除api",
		zap.Uint32("api_id", apiID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, rErr := uc.FindApiByID(ctx, apiID)
	if rErr != nil {
		uc.log.Error(
			"查询待删除api失败",
			zap.Error(rErr),
			zap.Uint32("api_id", apiID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return rErr
	}

	if err := uc.apiRepo.DeleteModel(ctx, apiID); err != nil {
		uc.log.Error(
			"删除api失败",
			zap.Error(err),
			zap.Uint32("api_id", apiID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.NewGormError(err, map[string]any{"id": apiID})
	}

	if err := uc.apiRepo.RemovePolicy(ctx, *m, true); err != nil {
		uc.log.Error(
			"移除api策略失败",
			zap.Error(err),
			zap.Uint32("api_id", apiID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.FromError(err)
	}

	uc.log.Info(
		"删除api成功",
		zap.Uint32("api_id", apiID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return nil
}

func (uc *ApiUsecase) FindApiByID(
	ctx context.Context,
	apiID uint32,
) (*model.ApiModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	uc.log.Info(
		"开始查询api",
		zap.Uint32("api_id", apiID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, err := uc.apiRepo.GetModel(ctx, apiID)
	if err != nil {
		uc.log.Error(
			"查询api失败",
			zap.Error(err),
			zap.Uint32("api_id", apiID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, map[string]any{"id": apiID})
	}

	uc.log.Info(
		"查询api成功",
		zap.Uint32("api_id", apiID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return m, nil
}

func (uc *ApiUsecase) ListApi(
	ctx context.Context,
	qp database.QueryParams,
) (int64, *[]model.ApiModel, *errors.Error) {
	if ctx.Err() != nil {
		return 0, nil, errors.FromError(ctx.Err())
	}

	uc.log.Info(
		"开始查询api列表",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	count, ms, err := uc.apiRepo.ListModel(ctx, qp)
	if err != nil {
		uc.log.Error(
			"查询api列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}

	uc.log.Info(
		"查询api列表成功",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return count, ms, nil
}

func (uc *ApiUsecase) LoadApiPolicy(ctx context.Context) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	uc.log.Info(
		"开始加载api策略",
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	qp := database.QueryParams{
		Columns: []string{"id", "url", "method"},
	}

	_, pms, rErr := uc.ListApi(ctx, qp)
	if rErr != nil {
		uc.log.Error(
			"加载api策略时查询api列表失败",
			zap.Error(rErr),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return rErr
	}

	var policyCount int
	if pms != nil {
		ms := *pms
		policyCount = len(ms)
		for i := range ms {
			if err := uc.apiRepo.AddPolicy(ctx, ms[i]); err != nil {
				uc.log.Error(
					"加载api策略失败",
					zap.Error(err),
					zap.Uint32("api_id", ms[i].ID),
					zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
				)
				return errors.FromError(err)
			}
		}
	}

	uc.log.Info(
		"加载api策略成功",
		zap.Int("policy_count", policyCount),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return nil
}
