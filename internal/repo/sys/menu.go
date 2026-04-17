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

// MenuRepo 菜单仓库实现
// 负责菜单模型的数据库操作和Casbin权限策略管理
type MenuRepo struct {
	log           *zap.Logger             // 日志记录器
	gormDB        *gorm.DB                // GORM数据库实例
	timeouts      *config.DBTimeout       // 数据库操作超时配置
	slowThreshold *config.DBSlowThreshold // 数据库操作慢查询阈值配置
	enforcer      *casbin.Enforcer        // Casbin权限管理器
}

// NewMenuRepo 创建菜单仓库实例
//
// 参数:
//
//	log: 日志记录器
//	gormDB: GORM数据库实例
//	timeouts: 数据库操作超时配置
//	enforcer: Casbin权限管理器
//
// 返回值:
//
//	sysmodel.MenuRepo: 菜单仓库接口实现
func NewMenuRepo(
	log *zap.Logger,
	gormDB *gorm.DB,
	timeouts *config.DBTimeout,
	slowThreshold *config.DBSlowThreshold,
	enforcer *casbin.Enforcer,
) *MenuRepo {
	return &MenuRepo{
		log:           log,
		gormDB:        gormDB,
		timeouts:      timeouts,
		slowThreshold: slowThreshold,
		enforcer:      enforcer,
	}
}

// CreateModel 创建菜单模型
//
// 参数:
//
//	ctx: 上下文，用于传递追踪信息和控制超时
//	m: 菜单模型指针
//	apis: 关联的API模型列表指针
//
// 返回值:
//
//	error: 操作过程中的错误
func (r *MenuRepo) CreateModel(
	ctx context.Context,
	m *sysmodel.MenuModel,
	apis []sysmodel.ApiModel,
) error {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	// 检查参数
	if m == nil {
		err := errors.New("创建菜单模型:模型不能为空")
		log.Error(
			"创建菜单模型:模型不能为空",
			zap.Error(err),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return err
	}

	m.CreatedAt = startTime
	m.UpdatedAt = startTime

	// 构建关联关系映射
	upmap := make(map[string]any, 1)
	if len(apis) > 0 {
		upmap["Apis"] = apis
	}

	log.Debug(
		"创建菜单模型:模型详情",
		zap.Object("menu_model", m),
		zap.Uint32s("apis", sysmodel.ListApiModelToUint32s(apis)),
	)

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()

	createStartTime := time.Now()
	err := database.DBCreate(dbCtx, r.gormDB, &sysmodel.MenuModel{}, m, upmap)
	createDuration := time.Since(createStartTime)
	if err != nil {
		log.Error(
			"创建菜单模型:数据库创建失败",
			zap.Error(err),
			zap.Object("menu_model", m),
			zap.Duration("create_duration", createDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "创建菜单模型:数据库创建失败")
	}

	if createDuration > r.slowThreshold.WriteSlow {
		log.Warn("创建菜单模型:数据库创建耗时超过慢查询阈值，可能影响性能",
			zap.Uint32("menu_id", m.ID),
			zap.Duration("create_duration", createDuration),
			zap.Duration("threshold", r.slowThreshold.WriteSlow),
		)
	}
	return nil
}

// UpdateModel 更新菜单模型
//
// 参数:
//
//	ctx: 上下文，用于传递追踪信息和控制超时
//	data: 更新数据映射
//	apis: 关联的API模型列表指针
//	conds: 查询条件
//
// 返回值:
//
//	error: 操作过程中的错误
func (r *MenuRepo) UpdateModel(
	ctx context.Context,
	updateData map[string]any,
	apis []sysmodel.ApiModel,
	conds ...any,
) error {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	if len(updateData) == 0 {
		err := errors.New("更新菜单模型:更新数据为空")
		log.Error(
			"更新菜单模型:更新数据为空",
			zap.Error(err),
			zap.Any("update_data", updateData),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return err
	}

	updateData["updated_at"] = startTime
	apiIDs := sysmodel.ListApiModelToUint32s(apis)

	// 构建关联关系映射
	var upmap map[string]any
	if len(apis) > 0 {
		upmap = map[string]any{"Apis": apis}
	}

	log.Debug(
		"更新菜单模型:更新数据",
		zap.Any("conds", conds),
		zap.Any("update_data", updateData),
		zap.Uint32s("apis", apiIDs),
	)

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()

	updateStartTime := time.Now()
	err := database.DBUpdate(dbCtx, r.gormDB, &sysmodel.MenuModel{}, updateData, upmap, conds...)
	updateDuration := time.Since(updateStartTime)
	if err != nil {
		log.Error(
			"更新菜单模型:数据库更新失败",
			zap.Error(err),
			zap.Any("update_data", updateData),
			zap.Uint32s("api_ids", apiIDs),
			zap.Any("conds", conds),
			zap.Duration("update_duration", updateDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "更新菜单模型:数据库更新失败")
	}

	if updateDuration > r.slowThreshold.WriteSlow {
		log.Warn("更新菜单模型:数据库更新耗时超过慢查询阈值，可能影响性能",
			zap.Duration("update_duration", updateDuration),
			zap.Duration("threshold", r.slowThreshold.WriteSlow),
		)
	}

	return nil
}

// DeleteModel 删除菜单模型
//
// 参数:
//
//	ctx: 上下文，用于传递追踪信息和控制超时
//	conds: 查询条件
//
// 返回值:
//
//	error: 操作过程中的错误
func (r *MenuRepo) DeleteModel(
	ctx context.Context,
	conds ...any,
) error {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"删除菜单模型:删除条件",
		zap.Any("conds", conds),
	)

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()

	deleteStartTime := time.Now()
	err := database.DBDelete(dbCtx, r.gormDB, &sysmodel.MenuModel{}, conds...)
	deleteDuration := time.Since(deleteStartTime)
	if err != nil {
		log.Error(
			"删除菜单模型:数据库删除失败",
			zap.Error(err),
			zap.Any("conds", conds),
			zap.Duration("delete_duration", deleteDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "删除菜单模型:数据库删除失败")
	}

	if deleteDuration > r.slowThreshold.WriteSlow {
		log.Warn("删除菜单模型:数据库删除耗时超过慢查询阈值，可能影响性能",
			zap.Duration("delete_duration", deleteDuration),
			zap.Duration("threshold", r.slowThreshold.WriteSlow),
		)
	}
	return nil
}

// GetModel 查询单个菜单模型
//
// 参数:
//
//	ctx: 上下文，用于传递追踪信息和控制超时
//	preloads: 需要预加载的关联关系
//	conds: 查询条件
//
// 返回值:
//
//	*sysmodel.MenuModel: 查询到的菜单模型指针
//	error: 操作过程中的错误
func (r *MenuRepo) GetModel(
	ctx context.Context,
	preloads []string,
	conds ...any,
) (*sysmodel.MenuModel, error) {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"查询菜单模型:查询条件",
		zap.Strings("preloads", preloads),
		zap.Any("conds", conds),
	)

	var m sysmodel.MenuModel

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()

	getStartTime := time.Now()
	err := database.DBGet(dbCtx, r.gormDB, preloads, &m, conds...)
	getDuration := time.Since(getStartTime)
	if err != nil {
		log.Error(
			"查询菜单模型:数据库查询失败",
			zap.Error(err),
			zap.Strings("preloads", preloads),
			zap.Any("conds", conds),
			zap.Duration("get_duration", getDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.WrapIf(err, "查询菜单模型:数据库查询失败")
	}

	log.Debug(
		"查询菜单模型:查询到的菜单模型详情",
		zap.Object("menu_model", &m),
	)

	if getDuration > r.slowThreshold.ReadSlow {
		log.Warn("查询菜单模型:数据库查询耗时超过慢查询阈值，可能影响性能",
			zap.Duration("get_duration", getDuration),
			zap.Duration("threshold", r.slowThreshold.ReadSlow),
		)
	}
	return &m, nil
}

// ListModel 查询菜单模型列表
//
// 参数:
//
//	ctx: 上下文，用于传递追踪信息和控制超时
//	qp: 查询参数，包含分页、排序等信息
//
// 返回值:
//
//	int64: 总记录数
//	[]sysmodel.MenuModel: 菜单模型列表指针
//	error: 操作过程中的错误
func (r *MenuRepo) ListModel(
	ctx context.Context,
	qp database.QueryParams,
) ([]sysmodel.MenuModel, error) {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"查询菜单模型列表:入参详情",
		zap.Object("query_params", &qp),
	)

	var ms []sysmodel.MenuModel

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ListTimeout)
	defer cancel()

	listStartTime := time.Now()
	err := database.DBList(dbCtx, r.gormDB, &sysmodel.MenuModel{}, &ms, qp)
	listDuration := time.Since(listStartTime)
	if err != nil {
		log.Error(
			"查询菜单模型列表:数据库查询失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("list_duration", listDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.WrapIf(err, "查询菜单模型列表:数据库查询失败")
	}

	log.Debug(
		"查询菜单模型列表:查询到的菜单模型列表",
		zap.Uint32s("menu_ids", sysmodel.ListMenuModelToUint32s(ms)),
	)

	if listDuration > r.slowThreshold.ReadSlow {
		log.Warn("查询菜单模型列表:数据库查询耗时超过慢查询阈值，可能影响性能",
			zap.Duration("list_duration", listDuration),
			zap.Duration("threshold", r.slowThreshold.ReadSlow),
		)
	}
	return ms, nil
}

func (r *MenuRepo) CountModel(
	ctx context.Context,
	query map[string]any,
) (int64, error) {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"查询菜单模型总数:查询条件",
		zap.Any("query", query),
	)

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()

	countStartTime := time.Now()
	count, err := database.DBCount(dbCtx, r.gormDB, &sysmodel.MenuModel{}, query)
	countDuration := time.Since(countStartTime)
	if err != nil {
		log.Error(
			"查询菜单模型总数:数据库查询失败",
			zap.Error(err),
			zap.Any("query", query),
			zap.Duration("count_duration", countDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, errors.WrapIf(err, "查询菜单模型总数:数据库查询失败")
	}

	log.Debug(
		"查询菜单模型总数:查询到的记录数",
		zap.Int64("count", count),
	)

	if countDuration > r.slowThreshold.ReadSlow {
		log.Warn("查询菜单模型总数:数据库查询耗时超过慢查询阈值，可能影响性能",
			zap.Duration("count_duration", countDuration),
			zap.Duration("threshold", r.slowThreshold.ReadSlow),
		)
	}
	return count, nil
}

// AddGroupPolicy 添加菜单的权限策略
//
// 参数:
//
//	ctx: 上下文，用于传递追踪信息和控制超时
//	m: 菜单模型，包含菜单的关联关系
//
// 返回值:
//
//	error: 操作过程中的错误
//
// 功能:
//  1. 检查上下文是否有效
//  2. 检查菜单模型是否为空
//  3. 检查菜单ID是否有效
//  4. 将菜单转换为Casbin策略
//  5. 添加菜单策略到Casbin
//  6. 记录操作日志
func (r *MenuRepo) AddGroupPolicy(
	ctx context.Context,
	m sysmodel.MenuModel,
) error {
	startTime := time.Now()

	// 检查上下文
	if ctx.Err() != nil {
		return errors.WrapIf(ctx.Err(), "添加菜单关联策略:上下文已取消或超时")
	}

	// 检查必要字段
	if m.ID == 0 {
		return errors.New("添加菜单关联策略:菜单ID不能为0")
	}

	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"添加菜单关联策略:入参详情",
		zap.Object("menu_model", &m),
	)

	rules := [][]string{}
	sub := auth.MenuToSubject(m.ID)

	// 处理父级关系
	if m.ParentID != nil {
		obj := auth.MenuToSubject(*m.ParentID)
		rules = append(rules, []string{sub, obj})
	}

	// 批量处理权限
	for _, o := range m.Apis {
		// 检查API模型的有效性
		if o.ID == 0 {
			log.Warn(
				"添加菜单关联策略:跳过API模型ID为0的关联策略",
				zap.Object("api_model", &o),
			)
			continue
		}

		obj := auth.ApiToSubject(o.ID)
		rules = append(rules, []string{sub, obj})
	}

	if err := auth.AddGroupPolicies(ctx, r.enforcer, rules); err != nil {
		log.Error(
			"添加菜单关联策略:Casbin添加策略失败",
			zap.Error(err),
			zap.Object("menu_model", &m),
			zap.Any("rules", rules),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "添加菜单关联策略:Casbin添加策略失败")
	}

	return nil
}

// RemoveGroupPolicy 删除菜单的权限策略
//
// 参数:
//
//	ctx: 上下文，用于传递追踪信息和控制超时
//	m: 菜单模型，包含菜单的关联关系
//	removeInherited: 是否删除继承该菜单的组策略
//
// 返回值:
//
//	error: 操作过程中的错误
//
// 功能:
//  1. 检查上下文是否有效
//  2. 检查菜单模型是否为空
//  3. 检查菜单ID是否有效
//  4. 删除该菜单作为子级的组策略（被其他策略继承）
//  5. 可选:删除该菜单作为父级的组策略（被其他菜单或API继承）
//  6. 记录操作日志
func (r *MenuRepo) RemoveGroupPolicy(
	ctx context.Context,
	m sysmodel.MenuModel,
	removeInherited bool,
) error {
	startTime := time.Now()

	// 检查上下文
	if ctx.Err() != nil {
		return errors.WrapIf(ctx.Err(), "删除菜单关联策略:上下文已取消或超时")
	}

	// 检查必要字段
	if m.ID == 0 {
		return errors.New("删除菜单关联策略:菜单ID不能为0")
	}

	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"删除菜单关联策略:入参详情",
		zap.Object("menu_model", &m),
		zap.Bool("removeInherited", removeInherited),
	)

	sub := auth.MenuToSubject(m.ID)

	if removeInherited {
		if err := auth.RemoveFilteredGroupingPolicy(ctx, r.enforcer, 1, sub); err != nil {
			log.Error(
				"删除菜单关联策略:Casbin删除策略失败",
				zap.Error(err),
				zap.Object("menu_model", &m),
				zap.String("sub", sub),
				zap.Duration("total_duration", time.Since(startTime)),
			)
			return errors.WrapIf(err, "删除菜单关联策略:Casbin删除策略失败")
		}
	}

	if err := auth.RemoveFilteredGroupingPolicy(ctx, r.enforcer, 0, sub); err != nil {
		log.Error(
			"删除菜单关联策略:Casbin删除策略失败",
			zap.Error(err),
			zap.Object("menu_model", &m),
			zap.String("sub", sub),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "删除菜单关联策略:Casbin删除策略失败")
	}

	return nil
}
