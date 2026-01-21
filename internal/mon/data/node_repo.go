package data

import (
	"context"
	"gin-artweb/internal/mon/biz"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/log"
	"gin-artweb/pkg/ctxutil"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type monNodeRepo struct {
	log      *zap.Logger
	gormDB   *gorm.DB
	timeouts *config.DBTimeout
}

func NewMonNodeRepo(
	log *zap.Logger,
	gormDB *gorm.DB,
	timeouts *config.DBTimeout,
) biz.MonNodeRepo {
	return &monNodeRepo{
		log:      log,
		gormDB:   gormDB,
		timeouts: timeouts,
	}
}

func (r *monNodeRepo) CreateModel(ctx context.Context, m *biz.MonNodeModel) error {
	r.log.Debug(
		"开始创建mon模型",
		zap.Object(database.ModelKey, m),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)
	now := time.Now()
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	if err := database.DBCreate(dbCtx, r.gormDB, &biz.MonNodeModel{}, m, nil); err != nil {
		r.log.Error(
			"创建mon模型失败",
			zap.Error(err),
			zap.Object(database.ModelKey, m),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return err
	}
	r.log.Debug(
		"创建mon模型成功",
		zap.Object(database.ModelKey, m),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(now)),
	)
	return nil
}

func (r *monNodeRepo) UpdateModel(ctx context.Context, data map[string]any, conds ...any) error {
	r.log.Debug(
		"开始更新mon模型",
		zap.Any(database.UpdateDataKey, data),
		zap.Any(database.ConditionKey, conds),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)
	startTime := time.Now()
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	if err := database.DBUpdate(dbCtx, r.gormDB, &biz.MonNodeModel{}, data, nil, conds...); err != nil {
		r.log.Error(
			"更新mon模型失败",
			zap.Error(err),
			zap.Any(database.UpdateDataKey, data),
			zap.Any(database.ConditionKey, conds),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(startTime)),
		)
		return err
	}
	r.log.Debug(
		"更新mon模型成功",
		zap.Any(database.UpdateDataKey, data),
		zap.Any(database.ConditionKey, conds),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(startTime)),
	)
	return nil
}

func (r *monNodeRepo) DeleteModel(ctx context.Context, conds ...any) error {
	r.log.Debug(
		"开始删除mon模型",
		zap.Any(database.ConditionKey, conds),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)
	startTime := time.Now()
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	if err := database.DBDelete(dbCtx, r.gormDB, &biz.MonNodeModel{}, conds...); err != nil {
		r.log.Error(
			"删除mon模型失败",
			zap.Error(err),
			zap.Any(database.ConditionKey, conds),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(startTime)),
		)
		return err
	}
	r.log.Debug(
		"删除mon模型成功",
		zap.Any(database.ConditionKey, conds),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(startTime)),
	)
	return nil
}

func (r *monNodeRepo) FindModel(
	ctx context.Context,
	preloads []string,
	conds ...any,
) (*biz.MonNodeModel, error) {
	r.log.Debug(
		"开始查询mon模型",
		zap.Any(database.ConditionKey, conds),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)
	startTime := time.Now()
	var m biz.MonNodeModel
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()
	if err := database.DBFind(dbCtx, r.gormDB, preloads, &m, conds...); err != nil {
		r.log.Error(
			"查询mon模型失败",
			zap.Error(err),
			zap.Any(database.ConditionKey, conds),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(startTime)),
		)
		return nil, err
	}
	r.log.Debug(
		"查询mon模型成功",
		zap.Object(database.ModelKey, &m),
		zap.Any(database.ConditionKey, conds),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(startTime)),
	)
	return &m, nil
}

func (r *monNodeRepo) ListModel(
	ctx context.Context,
	qp database.QueryParams,
) (int64, *[]biz.MonNodeModel, error) {
	r.log.Debug(
		"开始查询mon模型列表",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)
	startTime := time.Now()
	var ms []biz.MonNodeModel
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ListTimeout)
	defer cancel()
	count, err := database.DBList(dbCtx, r.gormDB, &biz.MonNodeModel{}, &ms, qp)
	if err != nil {
		r.log.Error(
			"查询mon模型列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(startTime)),
		)
		return 0, nil, err
	}
	r.log.Debug(
		"查询mon模型列表成功",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(startTime)),
	)
	return count, &ms, nil
}
