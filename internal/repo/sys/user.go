package sys

import (
	"context"
	"time"

	"emperror.dev/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"

	sysmodel "gin-artweb/internal/model/sys"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/database"
)

// UserRepo 用户仓库实现
// 负责用户模型的CRUD操作
// 使用GORM进行数据库操作
type UserRepo struct {
	log      *zap.Logger       // 日志记录器
	gormDB   *gorm.DB          // GORM数据库连接
	timeouts *config.DBTimeout // 数据库操作超时配置
}

// NewUserRepo 创建用户仓库实例
//
// 参数：
//
//	log: 日志记录器，用于记录操作日志
//	gormDB: GORM数据库连接，用于执行数据库操作
//	timeouts: 数据库操作超时配置，控制各类数据库操作的超时时间
//
// 返回值：
//
//	UserRepo: 用户仓库接口实现
func NewUserRepo(
	log *zap.Logger,
	gormDB *gorm.DB,
	timeouts *config.DBTimeout,
) *UserRepo {
	return &UserRepo{
		log:      log,
		gormDB:   gormDB,
		timeouts: timeouts,
	}
}

// CreateModel 创建用户模型
//
// 参数：
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	m: 用户模型，包含用户的详细信息
//
// 返回值：
//
//	error: 操作错误信息，成功则返回nil
//
// 功能：
//  1. 检查用户模型是否为空
//  2. 设置创建时间和更新时间
//  3. 执行数据库创建操作
//  4. 记录操作日志
func (r *UserRepo) CreateModel(
	ctx context.Context,
	m *sysmodel.UserModel,
) error {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	// 检查参数
	if m == nil {
		err := errors.New("创建用户模型: 模型不能为空")
		log.Error(
			"创建用户模型: 模型不能为空",
			zap.Error(err),
		)
		return err
	}

	log.Debug(
		"创建用户模型：开始执行",
		zap.Object("user_model", m),
	)

	m.CreatedAt = startTime
	m.UpdatedAt = startTime

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	createUserStartTime := time.Now()
	err := database.DBCreate(dbCtx, r.gormDB, &sysmodel.UserModel{}, m, nil)
	createUserDuration := time.Since(createUserStartTime)
	if err != nil {
		log.Error(
			"创建用户模型：数据库创建失败",
			zap.Error(err),
			zap.Object("user_model", m),
			zap.Duration("create_user_duration", createUserDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "创建用户模型：数据库创建失败")
	}
	log.Debug(
		"创建用户模型：执行成功",
		zap.Object("user_model", m),
		zap.Duration("create_user_duration", createUserDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

// UpdateModel 更新用户模型
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
func (r *UserRepo) UpdateModel(
	ctx context.Context,
	data map[string]any,
	conds ...any,
) error {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	if len(data) == 0 {
		err := errors.New("更新用户模型：更新数据不能为空")
		log.Error(
			"更新用户模型：更新数据不能为空",
			zap.Error(err),
			zap.Any("update_data", data),
		)
		return err
	}
	log.Debug(
		"更新用户模型：开始执行",
		zap.Any("update_data", data),
		zap.Any("conds", conds),
	)

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	updateUserStartTime := time.Now()
	err := database.DBUpdate(dbCtx, r.gormDB, &sysmodel.UserModel{}, data, nil, conds...)
	updateUserDuration := time.Since(updateUserStartTime)
	if err != nil {
		log.Error(
			"更新用户模型：数据库更新失败",
			zap.Error(err),
			zap.Any("update_data", data),
			zap.Any("conds", conds),
			zap.Duration("update_user_duration", updateUserDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "更新用户模型：数据库更新失败")
	}
	log.Debug(
		"更新用户模型：执行成功",
		zap.Any("update_data", data),
		zap.Any("conds", conds),
		zap.Duration("update_user_duration", updateUserDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

// DeleteModel 删除用户模型
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
func (r *UserRepo) DeleteModel(
	ctx context.Context,
	conds ...any,
) error {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"删除用户模型：开始执行",
		zap.Any("conds", conds),
	)

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	deleteUserStartTime := time.Now()
	err := database.DBDelete(dbCtx, r.gormDB, &sysmodel.UserModel{}, conds...)
	deleteUserDuration := time.Since(deleteUserStartTime)
	if err != nil {
		log.Error(
			"删除用户模型：数据库删除失败",
			zap.Error(err),
			zap.Any("conds", conds),
			zap.Duration("delete_user_duration", deleteUserDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "删除用户模型：数据库删除失败")
	}
	log.Debug(
		"删除用户模型：执行成功",
		zap.Any("conds", conds),
		zap.Duration("delete_user_duration", deleteUserDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

// GetModel 查询单个用户模型
//
// 参数：
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	preloads: 需要预加载的关联关系
//	conds: 查询条件，用于指定要查询的记录
//
// 返回值：
//
//	*sysmodel.UserModel: 用户模型指针，包含用户的详细信息
//	error: 操作错误信息，成功则返回nil
//
// 功能：
//  1. 执行数据库查询操作
//  2. 预加载关联字段
//  3. 获取单个用户模型
//  4. 记录操作日志
func (r *UserRepo) GetModel(
	ctx context.Context,
	preloads []string,
	conds ...any,
) (*sysmodel.UserModel, error) {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"查询用户模型：开始执行",
		zap.Strings("preloads", preloads),
		zap.Any("conds", conds),
	)

	var m sysmodel.UserModel
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()
	getUserStartTime := time.Now()
	err := database.DBGet(dbCtx, r.gormDB, preloads, &m, conds...)
	getUserDuration := time.Since(getUserStartTime)
	if err != nil {
		log.Error(
			"查询用户模型：数据库查询失败",
			zap.Error(err),
			zap.Strings("preloads", preloads),
			zap.Any("conds", conds),
			zap.Duration("get_user_duration", getUserDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.WrapIf(err, "查询用户模型：数据库查询失败")
	}
	log.Debug(
		"查询用户模型：执行成功",
		zap.Object("user_model", &m),
		zap.Strings("preloads", preloads),
		zap.Any("conds", conds),
		zap.Duration("get_user_duration", getUserDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return &m, nil
}

// ListModel 查询用户模型列表
//
// 参数：
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	qp: 查询参数，包含分页、排序等查询条件
//
// 返回值：
//
//	int64: 总记录数
//	[]sysmodel.UserModel: 用户模型列表指针，包含符合条件的用户模型
//	error: 操作错误信息，成功则返回nil
//
// 功能：
//  1. 执行数据库查询操作
//  2. 获取用户模型列表
//  3. 返回总记录数和模型列表
//  4. 记录操作日志
func (r *UserRepo) ListModel(
	ctx context.Context,
	qp database.QueryParams,
) ([]sysmodel.UserModel, error) {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"查询用户模型列表：开始执行",
		zap.Object("query_params", &qp),
	)

	var ms []sysmodel.UserModel
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()
	listUserStartTime := time.Now()
	err := database.DBList(dbCtx, r.gormDB, &sysmodel.UserModel{}, &ms, qp)
	listUserDuration := time.Since(listUserStartTime)
	if err != nil {
		log.Error(
			"查询用户模型列表：数据库查询失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("list_user_duration", listUserDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.WrapIf(err, "查询用户模型列表：数据库查询失败")
	}
	log.Debug(
		"查询用户模型列表：执行成功",
		zap.Object("query_params", &qp),
		zap.Duration("list_user_duration", listUserDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return ms, nil
}

func (r *UserRepo) CountModel(
	ctx context.Context,
	query map[string]any,
) (int64, error) {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"查询用户模型总数：开始执行",
		zap.Any("query", query),
	)
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()
	countUserStartTime := time.Now()
	count, err := database.DBCount(dbCtx, r.gormDB, &sysmodel.UserModel{}, query)
	countUserDuration := time.Since(countUserStartTime)
	if err != nil {
		log.Error(
			"查询用户模型总数：数据库查询失败",
			zap.Error(err),
			zap.Any("query", query),
			zap.Duration("count_user_duration", countUserDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, errors.WrapIf(err, "查询用户模型总数：数据库查询失败")
	}
	log.Debug(
		"查询用户模型总数：执行成功",
		zap.Any("query", query),
		zap.Int64("count", count),
		zap.Duration("count_user_duration", countUserDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return count, nil
}
