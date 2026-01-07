package data

import (
	"context"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"gin-artweb/internal/jobs/biz"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/log"
)

type recordRepo struct {
	log      *zap.Logger
	gormDB   *gorm.DB
	timeouts *config.DBTimeout
}

func NewRecordRepo(
	log *zap.Logger,
	gormDB *gorm.DB,
	timeouts *config.DBTimeout,
) biz.ScriptRecordRepo {
	return &recordRepo{
		log:      log,
		gormDB:   gormDB,
		timeouts: timeouts,
	}
}

func (r *recordRepo) CreateModel(ctx context.Context, m *biz.ScriptRecordModel) error {
	r.log.Debug(
		"开始创建脚本执行记录模型",
		zap.Object(database.ModelKey, m),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)
	now := time.Now()
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	if err := database.DBCreate(dbCtx, r.gormDB, &biz.ScriptRecordModel{}, m, nil); err != nil {
		r.log.Error(
			"创建脚本执行记录模型失败",
			zap.Error(err),
			zap.Object(database.ModelKey, m),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return err
	}
	r.log.Debug(
		"创建脚本执行记录模型成功",
		zap.Object(database.ModelKey, m),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(now)),
	)
	return nil
}

func (r *recordRepo) UpdateModel(ctx context.Context, data map[string]any, conds ...any) error {
	r.log.Debug(
		"开始更新脚本执行记录模型",
		zap.Any(database.UpdateDataKey, data),
		zap.Any(database.ConditionKey, conds),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)
	startTime := time.Now()
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	if err := database.DBUpdate(dbCtx, r.gormDB, &biz.ScriptRecordModel{}, data, nil, conds...); err != nil {
		r.log.Error(
			"更新脚本执行记录模型失败",
			zap.Error(err),
			zap.Any(database.UpdateDataKey, data),
			zap.Any(database.ConditionKey, conds),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(startTime)),
		)
		return err
	}
	r.log.Debug(
		"更新脚本执行记录模型成功",
		zap.Any(database.UpdateDataKey, data),
		zap.Any(database.ConditionKey, conds),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(startTime)),
	)
	return nil
}

func (r *recordRepo) DeleteModel(ctx context.Context, conds ...any) error {
	r.log.Debug(
		"开始删除脚本执行记录模型",
		zap.Any(database.ConditionKey, conds),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)
	startTime := time.Now()
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	if err := database.DBDelete(dbCtx, r.gormDB, &biz.ScriptRecordModel{}, conds...); err != nil {
		r.log.Error(
			"删除脚本执行记录模型失败",
			zap.Error(err),
			zap.Any(database.ConditionKey, conds),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(startTime)),
		)
		return err
	}
	r.log.Debug(
		"删除脚本执行记录模型成功",
		zap.Any(database.ConditionKey, conds),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(startTime)),
	)
	return nil
}

func (r *recordRepo) FindModel(
	ctx context.Context,
	preloads []string,
	conds ...any,
) (*biz.ScriptRecordModel, error) {
	r.log.Debug(
		"开始查询脚本执行记录模型",
		zap.Any(database.ConditionKey, conds),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)
	startTime := time.Now()
	var m biz.ScriptRecordModel
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()
	if err := database.DBFind(dbCtx, r.gormDB, preloads, &m, conds...); err != nil {
		r.log.Error(
			"查询脚本执行记录模型失败",
			zap.Error(err),
			zap.Any(database.ConditionKey, conds),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(startTime)),
		)
		return nil, err
	}
	r.log.Debug(
		"查询脚本执行记录模型成功",
		zap.Object(database.ModelKey, &m),
		zap.Any(database.ConditionKey, conds),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(startTime)),
	)
	return &m, nil
}

func (r *recordRepo) ListModel(
	ctx context.Context,
	qp database.QueryParams,
) (int64, *[]biz.ScriptRecordModel, error) {
	r.log.Debug(
		"开始查询脚本执行记录模型列表",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)
	startTime := time.Now()
	var ms []biz.ScriptRecordModel
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ListTimeout)
	defer cancel()
	count, err := database.DBList(dbCtx, r.gormDB, &biz.ScriptRecordModel{}, &ms, qp)
	if err != nil {
		r.log.Error(
			"查询脚本执行记录模型列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(startTime)),
		)
		return 0, nil, err
	}
	r.log.Debug(
		"查询脚本执行记录模型列表成功",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(startTime)),
	)
	return count, &ms, nil
}
