package data

import (
	"context"
	"time"

	"emperror.dev/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"gin-artweb/internal/business/mds/biz"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/log"
)

type mdsColonyRepo struct {
	log      *zap.Logger
	gormDB   *gorm.DB
	timeouts *config.DBTimeout
}

func NewMdsColonyRepo(
	log *zap.Logger,
	gormDB *gorm.DB,
	timeouts *config.DBTimeout,
) biz.MdsColonyRepo {
	return &mdsColonyRepo{
		log:      log,
		gormDB:   gormDB,
		timeouts: timeouts,
	}
}

func (r *mdsColonyRepo) CreateModel(ctx context.Context, m *biz.MdsColonyModel) error {
	// 检查参数
	if m == nil {
		err := errors.New("创建mds集群失败: 模型为空")
		r.log.Error(
			"创建mds集群失败: 模型为空",
			zap.Error(err),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return err
	}
	r.log.Debug(
		"开始创建mds集群",
		zap.Object(database.ModelKey, m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	now := time.Now()
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	if err := database.DBCreate(dbCtx, r.gormDB, &biz.MdsColonyModel{}, m, nil); err != nil {
		r.log.Error(
			"创建mds集群失败",
			zap.Error(err),
			zap.Object(database.ModelKey, m),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return errors.WrapIf(err, "创建mds集群失败")
	}
	r.log.Debug(
		"创建mds集群成功",
		zap.Object(database.ModelKey, m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(now)),
	)
	return nil
}

func (r *mdsColonyRepo) UpdateModel(ctx context.Context, data map[string]any, conds ...any) error {
	// 检查参数
	if len(data) == 0 {
		err := errors.New("更新mds集群失败: 更新数据为空")
		r.log.Error(
			"更新mds集群失败: 更新数据为空",
			zap.Error(err),
			zap.Any(database.UpdateDataKey, data),
			zap.Any(database.ConditionsKey, conds),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return err
	}

	r.log.Debug(
		"开始更新mds集群",
		zap.Any(database.UpdateDataKey, data),
		zap.Any(database.ConditionsKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	startTime := time.Now()
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	if err := database.DBUpdate(dbCtx, r.gormDB, &biz.MdsColonyModel{}, data, nil, conds...); err != nil {
		r.log.Error(
			"更新mds集群失败",
			zap.Error(err),
			zap.Any(database.UpdateDataKey, data),
			zap.Any(database.ConditionsKey, conds),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(startTime)),
		)
		return errors.WrapIf(err, "更新mds集群失败")
	}
	r.log.Debug(
		"更新mds集群成功",
		zap.Any(database.UpdateDataKey, data),
		zap.Any(database.ConditionsKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(startTime)),
	)
	return nil
}

func (r *mdsColonyRepo) DeleteModel(ctx context.Context, conds ...any) error {
	r.log.Debug(
		"开始删除mds集群",
		zap.Any(database.ConditionsKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	startTime := time.Now()
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	if err := database.DBDelete(dbCtx, r.gormDB, &biz.MdsColonyModel{}, conds...); err != nil {
		r.log.Error(
			"删除mds集群失败",
			zap.Error(err),
			zap.Any(database.ConditionsKey, conds),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(startTime)),
		)
		return errors.WrapIf(err, "删除mds集群失败")
	}
	r.log.Debug(
		"删除mds集群成功",
		zap.Any(database.ConditionsKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(startTime)),
	)
	return nil
}

func (r *mdsColonyRepo) GetModel(
	ctx context.Context,
	preloads []string,
	conds ...any,
) (*biz.MdsColonyModel, error) {
	r.log.Debug(
		"开始查询mds集群",
		zap.Any(database.ConditionsKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	startTime := time.Now()
	var m biz.MdsColonyModel
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()
	if err := database.DBGet(dbCtx, r.gormDB, preloads, &m, conds...); err != nil {
		r.log.Error(
			"查询mds集群失败",
			zap.Error(err),
			zap.Any(database.ConditionsKey, conds),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(startTime)),
		)
		return nil, errors.WrapIf(err, "查询mds集群失败")
	}
	r.log.Debug(
		"查询mds集群成功",
		zap.Object(database.ModelKey, &m),
		zap.Any(database.ConditionsKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(startTime)),
	)
	return &m, nil
}

func (r *mdsColonyRepo) ListModel(
	ctx context.Context,
	qp database.QueryParams,
) (int64, *[]biz.MdsColonyModel, error) {
	r.log.Debug(
		"开始查询mds集群列表",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	startTime := time.Now()
	var ms []biz.MdsColonyModel
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ListTimeout)
	defer cancel()
	count, err := database.DBList(dbCtx, r.gormDB, &biz.MdsColonyModel{}, &ms, qp)
	if err != nil {
		r.log.Error(
			"查询mds集群列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(startTime)),
		)
		return 0, nil, err
	}
	r.log.Debug(
		"查询mds集群列表成功",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(startTime)),
	)
	return count, &ms, nil
}
