package data

import (
	"context"
	"time"

	"emperror.dev/errors"
	"github.com/casbin/casbin/v2"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"gin-artweb/internal/infra/customer/model"
	"gin-artweb/internal/shared/auth"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/log"
)

// ApiRepo API仓库实现
// 负责API模型的CRUD操作和API策略的管理
// 使用GORM进行数据库操作，使用Casbin进行API策略管理
type ApiRepo struct {
	log      *zap.Logger       // 日志记录器
	gormDB   *gorm.DB          // GORM数据库连接
	timeouts *config.DBTimeout // 数据库操作超时配置
	enforcer *casbin.Enforcer  // CasbinAPI管理器
}

// NewApiRepo 创建API仓库实例
//
// 参数：
//
//	log: 日志记录器，用于记录操作日志
//	gormDB: GORM数据库连接，用于执行数据库操作
//	timeouts: 数据库操作超时配置，控制各类数据库操作的超时时间
//	enforcer: CasbinAPI管理器，用于管理API策略
//
// 返回值：
//
//	ApiRepo: API仓库接口实现
func NewApiRepo(
	log *zap.Logger,
	gormDB *gorm.DB,
	timeouts *config.DBTimeout,
	enforcer *casbin.Enforcer,
) *ApiRepo {
	return &ApiRepo{
		log:      log,
		gormDB:   gormDB,
		timeouts: timeouts,
		enforcer: enforcer,
	}
}

// CreateModel 创建API模型
//
// 参数：
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	m: API模型，包含API的详细信息
//
// 返回值：
//
//	error: 操作错误信息，成功则返回nil
//
// 功能：
//  1. 检查API模型是否为空
//  2. 设置创建时间和更新时间
//  3. 执行数据库创建操作
//  4. 记录操作日志
func (r *ApiRepo) CreateModel(ctx context.Context, m *model.ApiModel) error {
	if m == nil {
		err := errors.New("创建API模型失败: 模型为空")
		r.log.Error(
			"创建API模型失败: 模型为空",
			zap.Error(err),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return err
	}
	r.log.Debug(
		"开始创建API模型",
		zap.Object(database.ModelKey, m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	now := time.Now()
	m.CreatedAt = now
	m.UpdatedAt = now

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	if err := database.DBCreate(dbCtx, r.gormDB, &model.ApiModel{}, m, nil); err != nil {
		r.log.Error(
			"创建API模型失败",
			zap.Error(err),
			zap.Object(database.ModelKey, m),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return errors.WrapIf(err, "创建API模型失败")
	}

	r.log.Debug(
		"创建API模型成功",
		zap.Object(database.ModelKey, m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(now)),
	)
	return nil
}

// UpdateModel 更新API模型
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
//  2. 设置更新时间
//  3. 执行数据库更新操作
//  4. 记录操作日志
func (r *ApiRepo) UpdateModel(ctx context.Context, data map[string]any, conds ...any) error {
	if len(data) == 0 {
		err := errors.New("更新API模型失败: 更新数据为空")
		r.log.Error(
			"更新API模型失败: 更新数据为空",
			zap.Error(err),
			zap.Any(database.UpdateDataKey, data),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return err
	}

	r.log.Debug(
		"开始更新API模型",
		zap.Any(database.UpdateDataKey, data),
		zap.Any(database.ConditionsKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	now := time.Now()
	data["updated_at"] = now
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	if err := database.DBUpdate(dbCtx, r.gormDB, &model.ApiModel{}, data, nil, conds...); err != nil {
		r.log.Error(
			"更新API模型失败",
			zap.Error(err),
			zap.Any(database.UpdateDataKey, data),
			zap.Any(database.ConditionsKey, conds),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return errors.WrapIf(err, "更新API模型失败")
	}

	r.log.Debug(
		"更新API模型成功",
		zap.Any(database.UpdateDataKey, data),
		zap.Any(database.ConditionsKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(now)),
	)
	return nil
}

// DeleteModel 删除API模型
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
func (r *ApiRepo) DeleteModel(ctx context.Context, conds ...any) error {
	r.log.Debug(
		"开始删除API模型",
		zap.Any(database.ConditionsKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	now := time.Now()
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	if err := database.DBDelete(dbCtx, r.gormDB, &model.ApiModel{}, conds...); err != nil {
		r.log.Error(
			"删除API模型失败",
			zap.Error(err),
			zap.Any(database.ConditionsKey, conds),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return errors.WrapIf(err, "删除API模型失败")
	}

	r.log.Debug(
		"删除API模型成功",
		zap.Any(database.ConditionsKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(now)),
	)
	return nil
}

// GetModel 查询单个API模型
//
// 参数：
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	conds: 查询条件，用于指定要查询的记录
//
// 返回值：
//
//	*model.ApiModel: API模型指针，包含API的详细信息
//	error: 操作错误信息，成功则返回nil
//
// 功能：
//  1. 执行数据库查询操作
//  2. 查询单个API模型
//  3. 记录操作日志
func (r *ApiRepo) GetModel(
	ctx context.Context,
	conds ...any,
) (*model.ApiModel, error) {
	r.log.Debug(
		"开始获取API模型",
		zap.Any(database.ConditionsKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	now := time.Now()
	var m model.ApiModel
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()
	if err := database.DBGet(dbCtx, r.gormDB, nil, &m, conds...); err != nil {
		r.log.Error(
			"获取API模型失败",
			zap.Error(err),
			zap.Any(database.ConditionsKey, conds),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return nil, errors.WrapIf(err, "获取API模型失败")
	}

	r.log.Debug(
		"获取API模型成功",
		zap.Object(database.ModelKey, &m),
		zap.Any(database.ConditionsKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(now)),
	)
	return &m, nil
}

// ListModel 查询API模型列表
//
// 参数：
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	qp: 查询参数，包含分页、排序等查询条件
//
// 返回值：
//
//	int64: 总记录数
//	*[]model.ApiModel: API模型列表指针，包含符合条件的API模型
//	error: 操作错误信息，成功则返回nil
//
// 功能：
//  1. 执行数据库查询操作
//  2. 查询API模型列表
//  3. 返回总记录数和模型列表
//  4. 记录操作日志
func (r *ApiRepo) ListModel(
	ctx context.Context,
	qp database.QueryParams,
) (int64, *[]model.ApiModel, error) {
	r.log.Debug(
		"开始查询API模型列表",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	now := time.Now()
	var ms []model.ApiModel
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ListTimeout)
	defer cancel()
	count, err := database.DBList(dbCtx, r.gormDB, &model.ApiModel{}, &ms, qp)
	if err != nil {
		r.log.Error(
			"查询API模型列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return 0, nil, errors.WrapIf(err, "查询API模型列表失败")
	}

	r.log.Debug(
		"查询API模型列表成功",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(now)),
	)
	return count, &ms, nil
}

// AddPolicy 添加API策略
//
// 参数：
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	m: API模型，包含API的详细信息
//
// 返回值：
//
//	error: 操作错误信息，成功则返回nil
//
// 功能：
//  1. 检查上下文是否有效
//  2. 检查API模型的有效性（ID、URL、Method不能为空）
//  3. 将API转换为Casbin策略
//  4. 添加API策略到Casbin
//  5. 记录操作日志
func (r *ApiRepo) AddPolicy(
	ctx context.Context,
	m model.ApiModel,
) error {
	// 检查上下文
	if ctx.Err() != nil {
		return errors.WrapIf(ctx.Err(), "AddPolicy操作失败: 上下文错误")
	}

	r.log.Debug(
		"AddPolicy: 传入参数",
		zap.Object(database.ModelKey, &m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	// 检查API模型的有效性
	if m.ID == 0 {
		return errors.New("添加API策略失败: APIID不能为0")
	}
	if m.URL == "" {
		return errors.New("添加API策略失败: URL不能为空")
	}
	if m.Method == "" {
		return errors.New("添加API策略失败: 请求方法不能为空")
	}

	r.log.Debug(
		"开始添加API策略",
		zap.Object(database.ModelKey, &m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	now := time.Now()
	sub := auth.ApiToSubject(m.ID)
	rules := [][]string{{sub, m.URL, m.Method}}
	if err := auth.AddPolicies(ctx, r.enforcer, rules); err != nil {
		r.log.Error(
			"添加API策略失败",
			zap.Error(err),
			zap.Object(database.ModelKey, &m),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return errors.WrapIf(err, "添加API策略失败")
	}
	r.log.Debug(
		"添加API策略成功",
		zap.Object(database.ModelKey, &m),
		zap.String(auth.SubKey, sub),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(now)),
	)
	return nil
}

// RemovePolicy 删除API策略
//
// 参数：
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	m: API模型，包含API的详细信息
//	removeInherited: 是否删除继承该API的组策略
//
// 返回值：
//
//	error: 操作错误信息，成功则返回nil
//
// 功能：
//  1. 检查上下文是否有效
//  2. 检查API模型的有效性（ID、URL、Method不能为空）
//  3. 将API转换为Casbin策略
//  4. 从Casbin中删除API策略
//  5. 可选：删除继承该API的组策略
//  6. 记录操作日志
func (r *ApiRepo) RemovePolicy(
	ctx context.Context,
	m model.ApiModel,
	removeInherited bool,
) error {
	// 检查上下文
	if ctx.Err() != nil {
		return errors.WrapIf(ctx.Err(), "RemovePolicy操作失败: 上下文错误")
	}

	r.log.Debug(
		"RemovePolicy: 传入参数",
		zap.Object(database.ModelKey, &m),
		zap.Bool("removeInherited", removeInherited),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	// 检查API模型的有效性
	if m.ID == 0 {
		return errors.New("删除API策略失败: APIID不能为0")
	}
	if m.URL == "" {
		return errors.New("删除API策略失败: URL不能为空")
	}
	if m.Method == "" {
		return errors.New("删除API策略失败: 请求方法不能为空")
	}

	r.log.Debug(
		"开始删除API策略",
		zap.Object(database.ModelKey, &m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	now := time.Now()
	sub := auth.ApiToSubject(m.ID)
	rules := [][]string{{sub, m.URL, m.Method}}
	if err := auth.RemovePolicies(ctx, r.enforcer, rules); err != nil {
		r.log.Error(
			"删除API策略失败",
			zap.Error(err),
			zap.Object(database.ModelKey, &m),
			zap.String(auth.SubKey, sub),
			zap.String(auth.ObjKey, m.URL),
			zap.String(auth.ActKey, m.Method),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return errors.WrapIf(err, "删除API策略失败")
	}
	r.log.Debug(
		"删除API策略成功",
		zap.Object(database.ModelKey, &m),
		zap.String(auth.SubKey, sub),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(now)),
	)

	// 如果需要删除继承该API的组策略
	if removeInherited {
		now = time.Now()
		r.log.Debug(
			"开始删除继承该API的组策略",
			zap.Object(database.ModelKey, &m),
			zap.String(auth.GroupObjKey, sub),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		if err := auth.RemoveFilteredGroupingPolicy(ctx, r.enforcer, 1, sub); err != nil {
			r.log.Error(
				"删除继承该API的组策略失败",
				zap.Error(err),
				zap.Object(database.ModelKey, &m),
				zap.String(auth.GroupObjKey, sub),
				zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
				zap.Duration(log.DurationKey, time.Since(now)),
			)
			return errors.WrapIf(err, "删除继承该API的组策略失败")
		}
		r.log.Debug(
			"删除继承该API的组策略成功",
			zap.Object(database.ModelKey, &m),
			zap.String(auth.GroupObjKey, sub),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
	}

	return nil
}

func ListApiModelToUint32s(pms *[]model.ApiModel) []uint32 {
	if pms == nil {
		return []uint32{}
	}
	ms := *pms
	if len(ms) == 0 {
		return []uint32{}
	}

	ids := make([]uint32, len(ms))
	for i, m := range ms {
		ids[i] = m.ID
	}
	return ids
}
