package data

import (
	"context"
	"time"

	"emperror.dev/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"gin-artweb/internal/infra/customer/biz"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/log"
)

// userRepo 用户仓库实现
// 负责用户模型的CRUD操作
// 使用GORM进行数据库操作
type userRepo struct {
	log      *zap.Logger       // 日志记录器
	gormDB   *gorm.DB          // GORM数据库连接
	timeouts *config.DBTimeout // 数据库操作超时配置
}

// NewUserRepo 创建用户仓库实例
//
// 参数：
//   log: 日志记录器，用于记录操作日志
//   gormDB: GORM数据库连接，用于执行数据库操作
//   timeouts: 数据库操作超时配置，控制各类数据库操作的超时时间
//
// 返回值：
//   biz.UserRepo: 用户仓库接口实现
func NewUserRepo(
	log *zap.Logger,
	gormDB *gorm.DB,
	timeouts *config.DBTimeout,
) biz.UserRepo {
	return &userRepo{
		log:      log,
		gormDB:   gormDB,
		timeouts: timeouts,
	}
}

// CreateModel 创建用户模型
//
// 参数：
//   ctx: 上下文，用于传递请求信息和控制超时
//   m: 用户模型，包含用户的详细信息
//
// 返回值：
//   error: 操作错误信息，成功则返回nil
//
// 功能：
//   1. 检查用户模型是否为空
//   2. 设置创建时间和更新时间
//   3. 执行数据库创建操作
//   4. 记录操作日志
func (r *userRepo) CreateModel(ctx context.Context, m *biz.UserModel) error {
	// 检查参数
	if m == nil {
		err := errors.New("创建用户模型失败: 模型为空")
		r.log.Error(
			"创建用户模型失败: 模型为空",
			zap.Error(err),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return err
	}

	r.log.Debug(
		"开始创建用户模型",
		zap.Object(database.ModelKey, m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	now := time.Now()
	m.CreatedAt = now
	m.UpdatedAt = now
	if err := database.DBCreate(ctx, r.gormDB, &biz.UserModel{}, m, nil); err != nil {
		r.log.Error(
			"创建用户模型失败",
			zap.Error(err),
			zap.Object(database.ModelKey, m),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return errors.WrapIf(err, "创建用户模型失败")
	}
	r.log.Debug(
		"创建用户模型成功",
		zap.Object(database.ModelKey, m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(now)),
	)
	return nil
}

// UpdateModel 更新用户模型
//
// 参数：
//   ctx: 上下文，用于传递请求信息和控制超时
//   data: 更新数据，包含要更新的字段和值
//   conds: 查询条件，用于指定要更新的记录
//
// 返回值：
//   error: 操作错误信息，成功则返回nil
//
// 功能：
//   1. 检查更新数据是否为空
//   2. 执行数据库更新操作
//   3. 记录操作日志
func (r *userRepo) UpdateModel(ctx context.Context, data map[string]any, conds ...any) error {
	if len(data) == 0 {
		err := errors.New("更新用户模型失败: 更新数据为空")
		r.log.Error(
			"更新用户模型失败: 更新数据为空",
			zap.Error(err),
			zap.Any(database.UpdateDataKey, data),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return err
	}
	r.log.Debug(
		"开始更新用户模型",
		zap.Any(database.UpdateDataKey, data),
		zap.Any(database.ConditionsKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	now := time.Now()
	if err := database.DBUpdate(ctx, r.gormDB, &biz.UserModel{}, data, nil, conds...); err != nil {
		r.log.Error(
			"更新用户模型失败",
			zap.Error(err),
			zap.Any(database.UpdateDataKey, data),
			zap.Any(database.ConditionsKey, conds),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return errors.WrapIf(err, "更新用户模型失败")
	}
	r.log.Debug(
		"更新用户模型成功",
		zap.Any(database.UpdateDataKey, data),
		zap.Any(database.ConditionsKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(now)),
	)
	return nil
}

// DeleteModel 删除用户模型
//
// 参数：
//   ctx: 上下文，用于传递请求信息和控制超时
//   conds: 查询条件，用于指定要删除的记录
//
// 返回值：
//   error: 操作错误信息，成功则返回nil
//
// 功能：
//   1. 执行数据库删除操作
//   2. 记录操作日志
func (r *userRepo) DeleteModel(ctx context.Context, conds ...any) error {
	r.log.Debug(
		"开始删除用户模型",
		zap.Any(database.ConditionsKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	now := time.Now()
	if err := database.DBDelete(ctx, r.gormDB, &biz.UserModel{}, conds...); err != nil {
		r.log.Error(
			"删除用户模型失败",
			zap.Error(err),
			zap.Any(database.ConditionsKey, conds),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return errors.WrapIf(err, "删除用户模型失败")
	}
	r.log.Debug(
		"删除用户模型成功",
		zap.Any(database.ConditionsKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(now)),
	)
	return nil
}

// GetModel 查询单个用户模型
//
// 参数：
//   ctx: 上下文，用于传递请求信息和控制超时
//   preloads: 需要预加载的关联关系
//   conds: 查询条件，用于指定要查询的记录
//
// 返回值：
//   *biz.UserModel: 用户模型指针，包含用户的详细信息
//   error: 操作错误信息，成功则返回nil
//
// 功能：
//   1. 执行数据库查询操作
//   2. 预加载关联字段
//   3. 获取单个用户模型
//   4. 记录操作日志
func (r *userRepo) GetModel(
	ctx context.Context,
	preloads []string,
	conds ...any,
) (*biz.UserModel, error) {
	r.log.Debug(
		"开始查询用户模型",
		zap.Strings(database.PreloadKey, preloads),
		zap.Any(database.ConditionsKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	now := time.Now()
	var m biz.UserModel
	if err := database.DBGet(ctx, r.gormDB, preloads, &m, conds...); err != nil {
		r.log.Error(
			"查询用户模型失败",
			zap.Error(err),
			zap.Strings(database.PreloadKey, preloads),
			zap.Any(database.ConditionsKey, conds),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return nil, errors.WrapIf(err, "查询用户模型失败")
	}
	r.log.Debug(
		"查询用户模型成功",
		zap.Object(database.ModelKey, &m),
		zap.Strings(database.PreloadKey, preloads),
		zap.Any(database.ConditionsKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(now)),
	)
	return &m, nil
}

// ListModel 查询用户模型列表
//
// 参数：
//   ctx: 上下文，用于传递请求信息和控制超时
//   qp: 查询参数，包含分页、排序等查询条件
//
// 返回值：
//   int64: 总记录数
//   *[]biz.UserModel: 用户模型列表指针，包含符合条件的用户模型
//   error: 操作错误信息，成功则返回nil
//
// 功能：
//   1. 执行数据库查询操作
//   2. 获取用户模型列表
//   3. 返回总记录数和模型列表
//   4. 记录操作日志
func (r *userRepo) ListModel(
	ctx context.Context,
	qp database.QueryParams,
) (int64, *[]biz.UserModel, error) {
	r.log.Debug(
		"开始查询用户模型列表",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	now := time.Now()
	var ms []biz.UserModel
	count, err := database.DBList(ctx, r.gormDB, &biz.UserModel{}, &ms, qp)
	if err != nil {
		r.log.Error(
			"查询用户列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return 0, nil, errors.WrapIf(err, "查询用户列表失败")
	}
	r.log.Debug(
		"查询用户模型列表成功",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(now)),
	)
	return count, &ms, nil
}
