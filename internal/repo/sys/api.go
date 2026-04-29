package sys

import (
	"context"
	"time"

	"emperror.dev/errors"
	"github.com/casbin/casbin/v2"
	"go.uber.org/zap"
	"gorm.io/gorm"

	sysmodel "gin-artweb/internal/model/sys"
	"gin-artweb/internal/shared/auth"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/database"
)

// ApiRepo API仓库实现
// 负责API模型的CRUD操作和API策略的管理
// 使用GORM进行数据库操作，使用Casbin进行API策略管理
type ApiRepo struct {
	log           *zap.Logger             // 日志记录器
	gormDB        *gorm.DB                // GORM数据库连接
	timeouts      *config.DBTimeout       // 数据库操作超时配置
	slowThreshold *config.DBSlowThreshold // 数据库操作慢查询阈值配置
	enforcer      *casbin.Enforcer        // CasbinAPI管理器
}

// NewApiRepo 创建API仓库实例
//
// 参数:
//
//	log: 日志记录器，用于记录操作日志
//	gormDB: GORM数据库连接，用于执行数据库操作
//	timeouts: 数据库操作超时配置，控制各类数据库操作的超时时间
//	enforcer: CasbinAPI管理器，用于管理API策略
//
// 返回值:
//
//	ApiRepo: API仓库接口实现
func NewApiRepo(
	log *zap.Logger,
	gormDB *gorm.DB,
	timeouts *config.DBTimeout,
	slowThreshold *config.DBSlowThreshold,
	enforcer *casbin.Enforcer,
) *ApiRepo {
	return &ApiRepo{
		log:           log,
		gormDB:        gormDB,
		timeouts:      timeouts,
		slowThreshold: slowThreshold,
		enforcer:      enforcer,
	}
}

// CreateModel 创建API模型
//
// 参数:
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	m: API模型，包含API的详细信息
//
// 返回值:
//
//	error: 操作错误信息，成功则返回nil
//
// 功能:
//  1. 检查API模型是否为空
//  2. 设置创建时间和更新时间
//  3. 执行数据库创建操作
//  4. 记录操作日志
func (r *ApiRepo) CreateModel(
	ctx context.Context,
	m *sysmodel.ApiModel,
) error {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	// 检查参数
	if m == nil {
		err := errors.New("创建API模型:模型不能为空")
		log.Error(
			"创建API模型:模型不能为空",
			zap.Error(err),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return err
	}

	m.CreatedAt = startTime
	m.UpdatedAt = startTime

	log.Debug(
		"创建API模型:模型详情",
		zap.Object("api_model", m),
	)

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()

	createStartTime := time.Now()
	err := database.DBCreate(dbCtx, r.gormDB, &sysmodel.ApiModel{}, m)
	createDuration := time.Since(createStartTime)
	if err != nil {
		log.Error(
			"创建API模型:数据库创建失败",
			zap.Error(err),
			zap.Object("api_model", m),
			zap.Duration("create_duration", createDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "创建API模型:数据库创建失败")
	}

	if createDuration > r.slowThreshold.WriteSlow {
		log.Warn("创建API模型:数据库创建耗时超过慢查询阈值，可能影响性能",
			zap.Uint32("api_id", m.ID),
			zap.Duration("create_duration", createDuration),
			zap.Duration("threshold", r.slowThreshold.WriteSlow),
		)
	}

	return nil
}

// UpdateModel 更新API模型
//
// 参数:
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	data: 更新数据，包含要更新的字段和值
//	conds: 查询条件，用于指定要更新的记录
//
// 返回值:
//
//	error: 操作错误信息，成功则返回nil
//
// 功能:
//  1. 检查更新数据是否为空
//  2. 设置更新时间
//  3. 执行数据库更新操作
//  4. 记录操作日志
func (r *ApiRepo) UpdateModel(
	ctx context.Context,
	updateData map[string]any,
	conds ...any,
) error {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	if len(updateData) == 0 {
		err := errors.New("更新API模型:更新数据为空")
		log.Error(
			"更新API模型:更新数据为空",
			zap.Error(err),
			zap.Any("update_data", updateData),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return err
	}

	updateData["updated_at"] = startTime

	log.Debug(
		"更新API模型:更新数据",
		zap.Any("update_data", updateData),
		zap.Any("conds", conds),
	)

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()

	updateStartTime := time.Now()
	err := database.DBUpdateTx(dbCtx, r.gormDB, &sysmodel.ApiModel{}, updateData, conds...)
	updateDuration := time.Since(updateStartTime)
	if err != nil {
		log.Error(
			"更新API模型:数据库更新失败",
			zap.Error(err),
			zap.Any("update_data", updateData),
			zap.Any("conds", conds),
			zap.Duration("update_duration", updateDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "更新API模型:数据库更新失败")
	}

	if updateDuration > r.slowThreshold.WriteSlow {
		log.Warn("更新API模型:数据库更新耗时超过慢查询阈值，可能影响性能",
			zap.Duration("update_duration", updateDuration),
			zap.Duration("threshold", r.slowThreshold.WriteSlow),
		)
	}
	return nil
}

// DeleteModel 删除API模型
//
// 参数:
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	conds: 查询条件，用于指定要删除的记录
//
// 返回值:
//
//	error: 操作错误信息，成功则返回nil
//
// 功能:
//  1. 执行数据库删除操作
//  2. 记录操作日志
func (r *ApiRepo) DeleteModel(
	ctx context.Context,
	conds ...any,
) error {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"删除API模型:删除条件",
		zap.Any("conds", conds),
	)

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()

	deleteStartTime := time.Now()
	err := database.DBDeleteTx(dbCtx, r.gormDB, &sysmodel.ApiModel{}, conds...)
	deleteDuration := time.Since(deleteStartTime)
	if err != nil {
		log.Error(
			"删除API模型:数据库删除失败",
			zap.Error(err),
			zap.Any("conds", conds),
			zap.Duration("delete_duration", deleteDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "删除API模型:数据库删除失败")
	}

	if deleteDuration > r.slowThreshold.WriteSlow {
		log.Warn("删除API模型:数据库删除耗时超过慢查询阈值，可能影响性能",
			zap.Duration("delete_duration", deleteDuration),
			zap.Duration("threshold", r.slowThreshold.WriteSlow),
		)
	}

	return nil
}

// GetModel 查询单个API模型
//
// 参数:
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	conds: 查询条件，用于指定要查询的记录
//
// 返回值:
//
//	*sysmodel.ApiModel: API模型指针，包含API的详细信息
//	error: 操作错误信息，成功则返回nil
//
// 功能:
//  1. 执行数据库查询操作
//  2. 查询单个API模型
//  3. 记录操作日志
func (r *ApiRepo) GetModel(
	ctx context.Context,
	conds ...any,
) (*sysmodel.ApiModel, error) {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"查询API模型:查询条件",
		zap.Any("conds", conds),
	)

	var m sysmodel.ApiModel

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()

	getStartTime := time.Now()
	err := database.DBGet(dbCtx, r.gormDB, nil, &m, conds...)
	getDuration := time.Since(getStartTime)
	if err != nil {
		log.Error(
			"查询API模型:数据库查询失败",
			zap.Error(err),
			zap.Any("conds", conds),
			zap.Duration("get_duration", getDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.WrapIf(err, "查询API模型:数据库查询失败")
	}

	log.Debug(
		"查询API模型:查询到的模型详情",
		zap.Object("api_model", &m),
	)

	if getDuration > r.slowThreshold.ReadSlow {
		log.Warn("查询API模型:数据库查询耗时超过慢查询阈值，可能影响性能",
			zap.Duration("get_duration", getDuration),
			zap.Duration("threshold", r.slowThreshold.ReadSlow),
		)
	}

	return &m, nil
}

// ListModel 查询API模型列表
//
// 参数:
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	qp: 查询参数，包含分页、排序等查询条件
//
// 返回值:
//
//	int64: 总记录数
//	[]sysmodel.ApiModel: API模型列表指针，包含符合条件的API模型
//	error: 操作错误信息，成功则返回nil
//
// 功能:
//  1. 执行数据库查询操作
//  2. 查询API模型列表
//  3. 返回总记录数和模型列表
//  4. 记录操作日志
func (r *ApiRepo) ListModel(
	ctx context.Context,
	qp database.QueryParams,
) ([]sysmodel.ApiModel, error) {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"查询API模型列表:入参详情",
		zap.Object("query_params", &qp),
	)

	var ms []sysmodel.ApiModel

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ListTimeout)
	defer cancel()

	listStartTime := time.Now()
	err := database.DBList(dbCtx, r.gormDB, &sysmodel.ApiModel{}, &ms, qp)
	listDuration := time.Since(listStartTime)
	if err != nil {
		log.Error(
			"查询API模型列表:数据库查询失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("list_duration", listDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.WrapIf(err, "查询API模型列表:数据库查询失败")
	}

	log.Debug(
		"查询API模型列表:查询到的模型列表",
		zap.Uint32s("api_ids", sysmodel.ListApiModelToUint32s(ms)),
	)

	if listDuration > r.slowThreshold.ListSlow {
		log.Warn("查询API模型列表:数据库查询耗时超过慢查询阈值，可能影响性能",
			zap.Duration("list_duration", listDuration),
			zap.Duration("threshold", r.slowThreshold.ListSlow),
		)
	}
	return ms, nil
}

func (r *ApiRepo) CountModel(
	ctx context.Context,
	query map[string]any,
) (int64, error) {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"查询API模型总数:查询条件",
		zap.Any("query", query),
	)

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()

	countStartTime := time.Now()
	count, err := database.DBCount(dbCtx, r.gormDB, &sysmodel.ApiModel{}, query)
	countDuration := time.Since(countStartTime)
	if err != nil {
		log.Error(
			"查询API模型总数:数据库查询失败",
			zap.Error(err),
			zap.Any("query", query),
			zap.Duration("count_duration", countDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, errors.WrapIf(err, "查询API模型总数:数据库查询失败")
	}

	log.Debug(
		"查询API模型总数:查询到的记录数",
		zap.Int64("count", count),
	)

	if countDuration > r.slowThreshold.ReadSlow {
		log.Warn("查询API模型总数:数据库查询耗时超过慢查询阈值，可能影响性能",
			zap.Duration("count_duration", countDuration),
			zap.Duration("threshold", r.slowThreshold.ReadSlow),
		)
	}
	return count, nil
}

// AddPolicy 添加API策略
//
// 参数:
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	m: API模型，包含API的详细信息
//
// 返回值:
//
//	error: 操作错误信息，成功则返回nil
//
// 功能:
//  1. 检查上下文是否有效
//  2. 检查API模型的有效性（ID、URL、Method不能为空）
//  3. 将API转换为Casbin策略
//  4. 添加API策略到Casbin
//  5. 记录操作日志
func (r *ApiRepo) AddPolicy(
	ctx context.Context,
	m sysmodel.ApiModel,
) error {
	startTime := time.Now()

	// 检查上下文
	if ctx.Err() != nil {
		return errors.WrapIf(ctx.Err(), "添加API策略: 上下文错误")
	}

	// 检查API模型的有效性
	if m.ID == 0 {
		return errors.New("添加API策略: APIID不能为0")
	}
	if m.URL == "" {
		return errors.New("添加API策略: URL不能为空")
	}
	if m.Method == "" {
		return errors.New("添加API策略: 请求方法不能为空")
	}

	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"添加API策略:入参详情",
		zap.Object("api_model", &m),
	)

	sub := auth.ApiToSubject(m.ID)
	rules := [][]string{{sub, m.URL, m.Method}}

	if err := auth.AddPolicies(ctx, r.enforcer, rules); err != nil {
		log.Error(
			"添加API策略:添加策略缓存失败",
			zap.Error(err),
			zap.Object("api_model", &m),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "添加API策略: 添加策略缓存失败")
	}

	return nil
}

// RemovePolicy 删除API策略
//
// 参数:
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	m: API模型，包含API的详细信息
//	removeInherited: 是否删除继承该API的组策略
//
// 返回值:
//
//	error: 操作错误信息，成功则返回nil
//
// 功能:
//  1. 检查上下文是否有效
//  2. 检查API模型的有效性（ID、URL、Method不能为空）
//  3. 将API转换为Casbin策略
//  4. 从Casbin中删除API策略
//  5. 可选:删除继承该API的组策略
//  6. 记录操作日志
func (r *ApiRepo) RemovePolicy(
	ctx context.Context,
	m sysmodel.ApiModel,
	removeInherited bool,
) error {
	startTime := time.Now()

	// 检查上下文
	if ctx.Err() != nil {
		return errors.WrapIf(ctx.Err(), "删除API策略: 上下文错误")
	}

	// 检查API模型的有效性
	if m.ID == 0 {
		return errors.New("删除API策略: APIID不能为0")
	}
	if m.URL == "" {
		return errors.New("删除API策略: URL不能为空")
	}
	if m.Method == "" {
		return errors.New("删除API策略: 请求方法不能为空")
	}

	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"删除API策略:入参详情",
		zap.Object("api_model", &m),
		zap.Bool("removeInherited", removeInherited),
	)

	sub := auth.ApiToSubject(m.ID)

	// 如果需要删除继承该API的组策略
	if removeInherited {
		if err := auth.RemoveFilteredGroupingPolicy(ctx, r.enforcer, 1, sub); err != nil {
			log.Error(
				"删除API策略:删除继承该API的组策略缓存失败",
				zap.Error(err),
				zap.Int("index", 1),
				zap.String("value", sub),
				zap.Duration("total_duration", time.Since(startTime)),
			)
			return errors.WrapIf(err, "删除API策略: 删除继承该API的组策略缓存失败")
		}
	}

	rules := [][]string{{sub, m.URL, m.Method}}
	err := auth.RemovePolicies(ctx, r.enforcer, rules)
	if err != nil {
		log.Error(
			"删除API策略:删除策略缓存失败",
			zap.Error(err),
			zap.String("sub", sub),
			zap.String("obj", m.URL),
			zap.String("act", m.Method),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "删除API策略: 删除策略缓存失败")
	}

	return nil
}
