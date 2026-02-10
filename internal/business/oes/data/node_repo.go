package data

import (
	"context"
	"time"

	"emperror.dev/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"gin-artweb/internal/business/oes/biz"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/log"
)

type oesNodeRepo struct {
	log      *zap.Logger
	gormDB   *gorm.DB
	timeouts *config.DBTimeout
}

func NewOesNodeRepo(
	log *zap.Logger,
	gormDB *gorm.DB,
	timeouts *config.DBTimeout,
) biz.OesNodeRepo {
	return &oesNodeRepo{
		log:      log,
		gormDB:   gormDB,
		timeouts: timeouts,
	}
}

func (r *oesNodeRepo) CreateModel(ctx context.Context, m *biz.OesNodeModel) error {
	// 检查参数
	if m == nil {
		err := errors.New("创建oes节点失败: 模型为空")
		r.log.Error(
			"创建oes节点失败: 模型为空",
			zap.Error(err),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return err
	}
	r.log.Debug(
		"开始创建oes节点",
		zap.Object(database.ModelKey, m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	now := time.Now()
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	if err := database.DBCreate(dbCtx, r.gormDB, &biz.OesNodeModel{}, m, nil); err != nil {
		r.log.Error(
			"创建oes节点失败",
			zap.Error(err),
			zap.Object(database.ModelKey, m),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return errors.WrapIf(err, "创建oes节点失败")
	}
	r.log.Debug(
		"创建oes节点成功",
		zap.Object(database.ModelKey, m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(now)),
	)
	return nil
}

func (r *oesNodeRepo) UpdateModel(ctx context.Context, data map[string]any, conds ...any) error {
	// 检查参数
	if data == nil || len(data) == 0 {
		err := errors.New("更新oes节点失败: 更新数据为空")
		r.log.Error(
			"更新oes节点失败: 更新数据为空",
			zap.Error(err),
			zap.Any(database.ConditionsKey, conds),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return err
	}
	r.log.Debug(
		"开始更新oes节点",
		zap.Any(database.UpdateDataKey, data),
		zap.Any(database.ConditionsKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	startTime := time.Now()
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	if err := database.DBUpdate(dbCtx, r.gormDB, &biz.OesNodeModel{}, data, nil, conds...); err != nil {
		r.log.Error(
			"更新oes节点失败",
			zap.Error(err),
			zap.Any(database.UpdateDataKey, data),
			zap.Any(database.ConditionsKey, conds),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(startTime)),
		)
		return errors.WrapIf(err, "更新oes节点失败")
	}
	r.log.Debug(
		"更新oes节点成功",
		zap.Any(database.UpdateDataKey, data),
		zap.Any(database.ConditionsKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(startTime)),
	)
	return nil
}

func (r *oesNodeRepo) DeleteModel(ctx context.Context, conds ...any) error {
	r.log.Debug(
		"开始删除oes节点",
		zap.Any(database.ConditionsKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	startTime := time.Now()
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	if err := database.DBDelete(dbCtx, r.gormDB, &biz.OesNodeModel{}, conds...); err != nil {
		r.log.Error(
			"删除oes节点失败",
			zap.Error(err),
			zap.Any(database.ConditionsKey, conds),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(startTime)),
		)
		return errors.WrapIf(err, "删除oes节点失败")
	}
	r.log.Debug(
		"删除oes节点成功",
		zap.Any(database.ConditionsKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(startTime)),
	)
	return nil
}

func (r *oesNodeRepo) GetModel(
	ctx context.Context,
	preloads []string,
	conds ...any,
) (*biz.OesNodeModel, error) {
	r.log.Debug(
		"开始查询oes节点",
		zap.Any(database.ConditionsKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	startTime := time.Now()
	var m biz.OesNodeModel
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()
	if err := database.DBGet(dbCtx, r.gormDB, preloads, &m, conds...); err != nil {
		r.log.Error(
			"查询oes节点失败",
			zap.Error(err),
			zap.Any(database.ConditionsKey, conds),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(startTime)),
		)
		return nil, errors.WrapIf(err, "查询oes节点失败")
	}
	r.log.Debug(
		"查询oes节点成功",
		zap.Object(database.ModelKey, &m),
		zap.Any(database.ConditionsKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(startTime)),
	)
	return &m, nil
}

func (r *oesNodeRepo) ListModel(
	ctx context.Context,
	qp database.QueryParams,
) (int64, *[]biz.OesNodeModel, error) {
	r.log.Debug(
		"开始查询oes节点列表",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	startTime := time.Now()
	var ms []biz.OesNodeModel
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ListTimeout)
	defer cancel()
	count, err := database.DBList(dbCtx, r.gormDB, &biz.OesNodeModel{}, &ms, qp)
	if err != nil {
		r.log.Error(
			"查询oes节点列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(startTime)),
		)
		return 0, nil, errors.WrapIf(err, "查询oes节点列表失败")
	}
	r.log.Debug(
		"查询oes节点列表成功",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(startTime)),
	)
	return count, &ms, nil
}
