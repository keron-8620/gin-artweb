package data

import (
	"context"
	"time"

	"emperror.dev/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"gin-artweb/internal/infra/jobs/model"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/log"
)

// ScriptRepo 脚本仓库实现
// 负责脚本模型的CRUD操作
// 使用GORM进行数据库操作
type ScriptRepo struct {
	log      *zap.Logger       // 日志记录器
	gormDB   *gorm.DB          // GORM数据库连接
	timeouts *config.DBTimeout // 数据库操作超时配置
}

// NewScriptRepo 创建脚本仓库实例
//
// 参数：
//
//	log: 日志记录器，用于记录操作日志
//	gormDB: GORM数据库连接，用于执行数据库操作
//	timeouts: 数据库操作超时配置，控制各类数据库操作的超时时间
//
// 返回值：
//
//	*ScriptRepo: 脚本仓库实例
func NewScriptRepo(
	log *zap.Logger,
	gormDB *gorm.DB,
	timeouts *config.DBTimeout,
) *ScriptRepo {
	return &ScriptRepo{
		log:      log,
		gormDB:   gormDB,
		timeouts: timeouts,
	}
}

// CreateModel 创建脚本模型
//
// 参数：
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	m: 脚本模型，包含脚本的详细信息
//
// 返回值：
//
//	error: 操作错误信息，成功则返回nil
//
// 功能：
//  1. 检查脚本模型是否为空
//  2. 设置创建时间和更新时间
//  3. 执行数据库创建操作
//  4. 记录操作日志
func (r *ScriptRepo) CreateModel(ctx context.Context, m *model.ScriptModel) error {
	// 检查参数
	if m == nil {
		err := errors.New("创建脚本模型失败: 模型为空")
		r.log.Error(
			"创建脚本模型失败: 模型为空",
			zap.Error(err),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return err
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
	if err := database.DBCreate(dbCtx, r.gormDB, &model.ScriptModel{}, m, nil); err != nil {
		r.log.Error(
			"创建脚本模型失败",
			zap.Error(err),
			zap.Object(database.ModelKey, m),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return errors.WrapIf(err, "创建脚本模型失败")
	}
	r.log.Debug(
		"创建脚本模型成功",
		zap.Object(database.ModelKey, m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(now)),
	)
	return nil
}

// UpdateModel 更新脚本模型
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
func (r *ScriptRepo) UpdateModel(ctx context.Context, data map[string]any, conds ...any) error {
	// 检查参数
	if len(data) == 0 {
		err := errors.New("更新脚本模型失败: 更新数据为空")
		r.log.Error(
			"更新脚本模型失败: 更新数据为空",
			zap.Error(err),
			zap.Any(database.UpdateDataKey, data),
			zap.Any(database.ConditionsKey, conds),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return err
	}
	r.log.Debug(
		"开始更新脚本模型",
		zap.Any(database.UpdateDataKey, data),
		zap.Any(database.ConditionsKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	startTime := time.Now()
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	if err := database.DBUpdate(dbCtx, r.gormDB, &model.ScriptModel{}, data, nil, conds...); err != nil {
		r.log.Error(
			"更新脚本模型失败",
			zap.Error(err),
			zap.Any(database.UpdateDataKey, data),
			zap.Any(database.ConditionsKey, conds),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(startTime)),
		)
		return errors.WrapIf(err, "更新脚本模型失败")
	}
	r.log.Debug(
		"更新脚本模型成功",
		zap.Any(database.UpdateDataKey, data),
		zap.Any(database.ConditionsKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(startTime)),
	)
	return nil
}

// DeleteModel 删除脚本模型
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
func (r *ScriptRepo) DeleteModel(ctx context.Context, conds ...any) error {
	r.log.Debug(
		"开始删除脚本模型",
		zap.Any(database.ConditionsKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	startTime := time.Now()
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	if err := database.DBDelete(dbCtx, r.gormDB, &model.ScriptModel{}, conds...); err != nil {
		r.log.Error(
			"删除脚本模型失败",
			zap.Error(err),
			zap.Any(database.ConditionsKey, conds),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(startTime)),
		)
		return errors.WrapIf(err, "删除脚本模型失败")
	}
	r.log.Debug(
		"删除脚本模型成功",
		zap.Any(database.ConditionsKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(startTime)),
	)
	return nil
}

// GetModel 查询单个脚本模型
//
// 参数：
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	conds: 查询条件，用于指定要查询的记录
//
// 返回值：
//
//	*model.ScriptModel: 脚本模型指针，包含脚本的详细信息
//	error: 操作错误信息，成功则返回nil
//
// 功能：
//  1. 执行数据库查询操作
//  2. 获取单个脚本模型
//  3. 记录操作日志
func (r *ScriptRepo) GetModel(
	ctx context.Context,
	conds ...any,
) (*model.ScriptModel, error) {
	r.log.Debug(
		"开始查询脚本模型",
		zap.Any(database.ConditionsKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	startTime := time.Now()
	var m model.ScriptModel
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()
	if err := database.DBGet(dbCtx, r.gormDB, nil, &m, conds...); err != nil {
		r.log.Error(
			"查询脚本模型失败",
			zap.Error(err),
			zap.Any(database.ConditionsKey, conds),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(startTime)),
		)
		return nil, errors.WrapIf(err, "查询脚本模型失败")
	}
	r.log.Debug(
		"查询脚本模型成功",
		zap.Object(database.ModelKey, &m),
		zap.Any(database.ConditionsKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(startTime)),
	)
	return &m, nil
}

// ListModel 查询脚本模型列表
//
// 参数：
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	qp: 查询参数，包含分页、排序等查询条件
//
// 返回值：
//
//	int64: 总记录数
//	*[]model.ScriptModel: 脚本模型列表指针，包含符合条件的脚本模型
//	error: 操作错误信息，成功则返回nil
//
// 功能：
//  1. 执行数据库查询操作
//  2. 获取脚本模型列表
//  3. 返回总记录数和模型列表
//  4. 记录操作日志
func (r *ScriptRepo) ListModel(
	ctx context.Context,
	qp database.QueryParams,
) (int64, *[]model.ScriptModel, error) {
	r.log.Debug(
		"开始查询脚本模型列表",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	startTime := time.Now()
	var ms []model.ScriptModel
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ListTimeout)
	defer cancel()
	count, err := database.DBList(dbCtx, r.gormDB, &model.ScriptModel{}, &ms, qp)
	if err != nil {
		r.log.Error(
			"查询脚本模型列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(startTime)),
		)
		return 0, nil, errors.WrapIf(err, "查询脚本模型列表失败")
	}
	r.log.Debug(
		"查询脚本模型列表成功",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(startTime)),
	)
	return count, &ms, nil
}

// ListProjects 查询所有脚本的项目名称
//
// 参数：
//
//	ctx: 上下文，用于传递请求信息和控制超时
//
// 返回值：
//
//	[]string: 项目名称列表
//	error: 操作错误信息，成功则返回nil
//
// 功能：
//  1. 执行数据库查询操作
//  2. 获取所有脚本的项目名称（去重）
//  3. 记录操作日志
func (r *ScriptRepo) ListProjects(
	ctx context.Context,
	query map[string]any,
) ([]string, error) {
	r.log.Debug(
		"开始查询脚本所有的项目名称",
		zap.Any(database.QueryParamsKey, query),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()

	var projects []string
	if err := r.gormDB.WithContext(dbCtx).Model(&model.ScriptModel{}).Where(query).Distinct("project").Pluck("project", &projects).Error; err != nil {
		r.log.Error(
			"查询项目名称失败",
			zap.Error(err),
			zap.Any(database.QueryParamsKey, query),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.WrapIf(err, "查询项目名称失败")
	}

	r.log.Debug(
		"查询项目名称成功",
		zap.Any("projects", projects),
		zap.Any(database.QueryParamsKey, query),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	return projects, nil
}

// ListLabels 查询所有脚本的标签名称
//
// 参数：
//
//	ctx: 上下文，用于传递请求信息和控制超时
//
// 返回值：
//
//	[]string: 标签名称列表
//	error: 操作错误信息，成功则返回nil
//
// 功能：
//  1. 执行数据库查询操作
//  2. 获取所有脚本的标签名称（去重）
//  3. 记录操作日志
func (r *ScriptRepo) ListLabels(
	ctx context.Context,
	query map[string]any,
) ([]string, error) {
	r.log.Debug(
		"开始查询所有标签名称",
		zap.Any(database.QueryParamsKey, query),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	var labels []string
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()

	// 查询所有唯一的标签名称
	if err := r.gormDB.WithContext(dbCtx).Model(&model.ScriptModel{}).Where(query).Distinct("label").Pluck("label", &labels).Error; err != nil {
		r.log.Error(
			"查询标签名称失败",
			zap.Error(err),
			zap.Any(database.QueryParamsKey, query),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.WrapIf(err, "查询标签名称失败")
	}

	r.log.Debug(
		"查询标签名称成功",
		zap.Any("labels", labels),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	return labels, nil
}
