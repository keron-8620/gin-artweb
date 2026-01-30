package data

import (
	"context"
	"time"
	"errors"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"gin-artweb/internal/infra/jobs/biz"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/log"
	"gin-artweb/pkg/ctxutil"
)

type scriptRepo struct {
	log      *zap.Logger
	gormDB   *gorm.DB
	timeouts *config.DBTimeout
}

func NewScriptRepo(
	log *zap.Logger,
	gormDB *gorm.DB,
	timeouts *config.DBTimeout,
) biz.ScriptRepo {
	return &scriptRepo{
		log:      log,
		gormDB:   gormDB,
		timeouts: timeouts,
	}
}

func (r *scriptRepo) CreateModel(ctx context.Context, m *biz.ScriptModel) error {
	// 检查参数
	if m == nil {
		return errors.New("创建按钮模型失败: 按钮模型不能为空")
	}

	r.log.Debug(
		"开始创建脚本模型",
		zap.Object(database.ModelKey, m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	now := time.Now()
	m.CreatedAt = now
	m.UpdatedAt = now
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	if err := database.DBCreate(dbCtx, r.gormDB, &biz.ScriptModel{}, m, nil); err != nil {
		r.log.Error(
			"创建脚本模型失败",
			zap.Error(err),
			zap.Object(database.ModelKey, m),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return err
	}
	r.log.Debug(
		"创建脚本模型成功",
		zap.Object(database.ModelKey, m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(now)),
	)
	return nil
}

func (r *scriptRepo) UpdateModel(ctx context.Context, data map[string]any, conds ...any) error {
	r.log.Debug(
		"开始更新脚本模型",
		zap.Any(database.UpdateDataKey, data),
		zap.Any(database.ConditionKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	startTime := time.Now()
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	if err := database.DBUpdate(dbCtx, r.gormDB, &biz.ScriptModel{}, data, nil, conds...); err != nil {
		r.log.Error(
			"更新脚本模型失败",
			zap.Error(err),
			zap.Any(database.UpdateDataKey, data),
			zap.Any(database.ConditionKey, conds),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(startTime)),
		)
		return err
	}
	r.log.Debug(
		"更新脚本模型成功",
		zap.Any(database.UpdateDataKey, data),
		zap.Any(database.ConditionKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(startTime)),
	)
	return nil
}

func (r *scriptRepo) DeleteModel(ctx context.Context, conds ...any) error {
	r.log.Debug(
		"开始删除脚本模型",
		zap.Any(database.ConditionKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	startTime := time.Now()
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	if err := database.DBDelete(dbCtx, r.gormDB, &biz.ScriptModel{}, conds...); err != nil {
		r.log.Error(
			"删除脚本模型失败",
			zap.Error(err),
			zap.Any(database.ConditionKey, conds),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(startTime)),
		)
		return err
	}
	r.log.Debug(
		"删除脚本模型成功",
		zap.Any(database.ConditionKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(startTime)),
	)
	return nil
}

func (r *scriptRepo) FindModel(
	ctx context.Context,
	conds ...any,
) (*biz.ScriptModel, error) {
	r.log.Debug(
		"开始查询脚本模型",
		zap.Any(database.ConditionKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	startTime := time.Now()
	var m biz.ScriptModel
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()
	if err := database.DBFind(dbCtx, r.gormDB, nil, &m, conds...); err != nil {
		r.log.Error(
			"查询脚本模型失败",
			zap.Error(err),
			zap.Any(database.ConditionKey, conds),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(startTime)),
		)
		return nil, err
	}
	r.log.Debug(
		"查询脚本模型成功",
		zap.Object(database.ModelKey, &m),
		zap.Any(database.ConditionKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(startTime)),
	)
	return &m, nil
}

func (r *scriptRepo) ListModel(
	ctx context.Context,
	qp database.QueryParams,
) (int64, *[]biz.ScriptModel, error) {
	r.log.Debug(
		"开始查询脚本模型列表",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	startTime := time.Now()
	var ms []biz.ScriptModel
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ListTimeout)
	defer cancel()
	count, err := database.DBList(dbCtx, r.gormDB, &biz.ScriptModel{}, &ms, qp)
	if err != nil {
		r.log.Error(
			"查询脚本模型列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(startTime)),
		)
		return 0, nil, err
	}
	r.log.Debug(
		"查询脚本模型列表成功",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(startTime)),
	)
	return count, &ms, nil
}

func (r *scriptRepo) ListProjects(
	ctx context.Context,
) ([]string, error) {
	r.log.Debug(
		"开始查询脚本所有的项目名称",
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()

	var projects []string
	if err := r.gormDB.WithContext(dbCtx).Model(&biz.ScriptModel{}).Distinct("project").Pluck("project", &projects).Error; err != nil {
		r.log.Error(
			"查询项目名称失败",
			zap.Error(err),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, err
	}

	r.log.Debug(
		"查询项目名称成功",
		zap.Any("projects", projects),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	return projects, nil
}

func (r *scriptRepo) ListLabels(
	ctx context.Context,
) ([]string, error) {
	r.log.Debug(
		"开始查询所有标签名称",
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	var labels []string
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()

	// 查询指定项目下所有唯一的标签名称
	if err := r.gormDB.WithContext(dbCtx).Model(&biz.ScriptModel{}).Distinct("label").Pluck("label", &labels).Error; err != nil {
		r.log.Error(
			"查询标签名称失败",
			zap.Error(err),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, err
	}

	r.log.Debug(
		"查询标签名称成功",
		zap.Any("labels", labels),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	return labels, nil
}
