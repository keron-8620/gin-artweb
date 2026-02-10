package data

import (
	"context"
	"time"

	"emperror.dev/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"gin-artweb/internal/infra/jobs/biz"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/log"
)

// scheduleRepo 计划任务仓库实现
// 负责计划任务模型的CRUD操作
// 使用GORM进行数据库操作
type scheduleRepo struct {
	log      *zap.Logger       // 日志记录器
	gormDB   *gorm.DB          // GORM数据库连接
	timeouts *config.DBTimeout // 数据库操作超时配置
}

// NewScheduleRepo 创建计划任务仓库实例
//
// 参数：
//
//	log: 日志记录器，用于记录操作日志
//	gormDB: GORM数据库连接，用于执行数据库操作
//	timeouts: 数据库操作超时配置，控制各类数据库操作的超时时间
//
// 返回值：
//
//	biz.ScheduleRepo: 计划任务仓库接口实现
func NewScheduleRepo(
	log *zap.Logger,
	gormDB *gorm.DB,
	timeouts *config.DBTimeout,
) biz.ScheduleRepo {
	return &scheduleRepo{
		log:      log,
		gormDB:   gormDB,
		timeouts: timeouts,
	}
}

// CreateModel 创建计划任务模型
//
// 参数：
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	m: 计划任务模型，包含计划任务的详细信息
//
// 返回值：
//
//	error: 操作错误信息，成功则返回nil
//
// 功能：
//  1. 检查计划任务模型是否为空
//  2. 执行数据库创建操作
//  3. 记录操作日志
func (r *scheduleRepo) CreateModel(ctx context.Context, m *biz.ScheduleModel) error {
	// 检查参数
	if m == nil {
		err := errors.New("创建计划任务模型失败: 模型为空")
		r.log.Error(
			"创建计划任务模型失败: 模型为空",
			zap.Error(err),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return err
	}
	r.log.Debug(
		"开始创建计划任务模型",
		zap.Object(database.ModelKey, m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	now := time.Now()
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	if err := database.DBCreate(dbCtx, r.gormDB, &biz.ScheduleModel{}, m, nil); err != nil {
		r.log.Error(
			"创建计划任务模型失败",
			zap.Error(err),
			zap.Object(database.ModelKey, m),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return errors.WrapIf(err, "创建计划任务模型失败")
	}
	r.log.Debug(
		"创建计划任务模型成功",
		zap.Object(database.ModelKey, m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(now)),
	)
	return nil
}

// UpdateModel 更新计划任务模型
//
// 参数：
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	data: 更新数据，包含要更新的字段和值
//	conds: 查询条件，用于指定要更新的记录
//
// 返回值：
//
//	error: 操作错误信息，成功则返回nil
//
// 功能：
//  1. 检查更新数据是否为空
//  2. 执行数据库更新操作
//  3. 记录操作日志
func (r *scheduleRepo) UpdateModel(ctx context.Context, data map[string]any, conds ...any) error {
	// 检查参数
	if len(data) == 0 {
		err := errors.New("更新计划任务模型失败: 更新数据为空")
		r.log.Error(
			"更新计划任务模型失败: 更新数据为空",
			zap.Error(err),
			zap.Any(database.UpdateDataKey, data),
			zap.Any(database.ConditionsKey, conds),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return err
	}
	r.log.Debug(
		"开始更新计划任务模型",
		zap.Any(database.UpdateDataKey, data),
		zap.Any(database.ConditionsKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	startTime := time.Now()
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	if err := database.DBUpdate(dbCtx, r.gormDB, &biz.ScheduleModel{}, data, nil, conds...); err != nil {
		r.log.Error(
			"更新计划任务模型失败",
			zap.Error(err),
			zap.Any(database.UpdateDataKey, data),
			zap.Any(database.ConditionsKey, conds),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(startTime)),
		)
		return errors.WrapIf(err, "更新计划任务模型失败")
	}
	r.log.Debug(
		"更新计划任务模型成功",
		zap.Any(database.UpdateDataKey, data),
		zap.Any(database.ConditionsKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(startTime)),
	)
	return nil
}

// DeleteModel 删除计划任务模型
//
// 参数：
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	conds: 查询条件，用于指定要删除的记录
//
// 返回值：
//
//	error: 操作错误信息，成功则返回nil
//
// 功能：
//  1. 执行数据库删除操作
//  2. 记录操作日志
func (r *scheduleRepo) DeleteModel(ctx context.Context, conds ...any) error {
	r.log.Debug(
		"开始删除计划任务模型",
		zap.Any(database.ConditionsKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	startTime := time.Now()
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	if err := database.DBDelete(dbCtx, r.gormDB, &biz.ScheduleModel{}, conds...); err != nil {
		r.log.Error(
			"删除计划任务模型失败",
			zap.Error(err),
			zap.Any(database.ConditionsKey, conds),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(startTime)),
		)
		return errors.WrapIf(err, "删除计划任务模型失败")
	}
	r.log.Debug(
		"删除计划任务模型成功",
		zap.Any(database.ConditionsKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(startTime)),
	)
	return nil
}

// GetModel 查询单个计划任务模型
//
// 参数：
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	preloads: 需要预加载的关联关系
//	conds: 查询条件，用于指定要查询的记录
//
// 返回值：
//
//	*biz.ScheduleModel: 计划任务模型指针，包含计划任务的详细信息
//	error: 操作错误信息，成功则返回nil
//
// 功能：
//  1. 执行数据库查询操作
//  2. 预加载关联字段
//  3. 获取单个计划任务模型
//  4. 记录操作日志
func (r *scheduleRepo) GetModel(
	ctx context.Context,
	preloads []string,
	conds ...any,
) (*biz.ScheduleModel, error) {
	r.log.Debug(
		"开始查询计划任务模型",
		zap.Any(database.ConditionsKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	startTime := time.Now()
	var m biz.ScheduleModel
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()
	if err := database.DBGet(dbCtx, r.gormDB, preloads, &m, conds...); err != nil {
		r.log.Error(
			"查询计划任务模型失败",
			zap.Error(err),
			zap.Any(database.ConditionsKey, conds),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(startTime)),
		)
		return nil, errors.WrapIf(err, "查询计划任务模型失败")
	}
	r.log.Debug(
		"查询计划任务模型成功",
		zap.Object(database.ModelKey, &m),
		zap.Any(database.ConditionsKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(startTime)),
	)
	return &m, nil
}

// ListModel 查询计划任务模型列表
//
// 参数：
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	qp: 查询参数，包含分页、排序、过滤等条件
//
// 返回值：
//
//	int64: 查询结果总数
//	*[]biz.ScheduleModel: 计划任务模型列表指针，包含查询到的计划任务详情
//	error: 操作错误信息，成功则返回nil
//
// 功能：
//  1. 执行数据库查询操作
//  2. 应用分页、排序、过滤等条件
//  3. 获取计划任务模型列表
//  4. 记录操作日志
func (r *scheduleRepo) ListModel(
	ctx context.Context,
	qp database.QueryParams,
) (int64, *[]biz.ScheduleModel, error) {
	r.log.Debug(
		"开始查询计划任务模型列表",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	startTime := time.Now()
	var ms []biz.ScheduleModel
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ListTimeout)
	defer cancel()
	count, err := database.DBList(dbCtx, r.gormDB, &biz.ScheduleModel{}, &ms, qp)
	if err != nil {
		r.log.Error(
			"查询计划任务模型列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(startTime)),
		)
		return 0, nil, errors.WrapIf(err, "查询计划任务模型列表失败")
	}
	r.log.Debug(
		"查询计划任务模型列表成功",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(startTime)),
	)
	return count, &ms, nil
}
