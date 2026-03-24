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
	log      *zap.Logger       // 日志记录器
	gormDB   *gorm.DB          // GORM数据库实例
	timeouts *config.DBTimeout // 数据库操作超时配置
	enforcer *casbin.Enforcer  // Casbin权限管理器
}

// NewMenuRepo 创建菜单仓库实例
//
// 参数：
//
//	log: 日志记录器
//	gormDB: GORM数据库实例
//	timeouts: 数据库操作超时配置
//	enforcer: Casbin权限管理器
//
// 返回值：
//
//	sysmodel.MenuRepo: 菜单仓库接口实现
func NewMenuRepo(
	log *zap.Logger,
	gormDB *gorm.DB,
	timeouts *config.DBTimeout,
	enforcer *casbin.Enforcer,
) *MenuRepo {
	return &MenuRepo{
		log:      log,
		gormDB:   gormDB,
		timeouts: timeouts,
		enforcer: enforcer,
	}
}

// CreateModel 创建菜单模型
//
// 参数：
//
//	ctx: 上下文，用于传递追踪信息和控制超时
//	m: 菜单模型指针
//	apis: 关联的API模型列表指针
//
// 返回值：
//
//	error: 操作过程中的错误
func (r *MenuRepo) CreateModel(
	ctx context.Context,
	m *sysmodel.MenuModel,
	apis []sysmodel.ApiModel,
) error {
	log := ctxutil.NewLogger(r.log, ctx)

	// 检查参数
	if m == nil {
		err := errors.New("创建菜单模型：模型不能为空")
		log.Error(
			"创建菜单模型：模型不能为空",
			zap.Error(err),
		)
		return err
	}

	log.Debug(
		"创建菜单模型：开始执行",
		zap.Object("menu_model", m),
	)
	now := time.Now()
	m.CreatedAt = now
	m.UpdatedAt = now

	// 构建关联关系映射
	upmap := make(map[string]any, 1)
	if len(apis) > 0 {
		upmap["Apis"] = apis
	}
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	if err := database.DBCreate(dbCtx, r.gormDB, &sysmodel.MenuModel{}, m, upmap); err != nil {
		log.Error(
			"创建菜单模型：数据库创建失败",
			zap.Error(err),
			zap.Object("menu_model", m),
			zap.Duration("create_menu_duration", time.Since(now)),
		)
		return errors.WrapIf(err, "创建菜单模型：数据库创建失败")
	}

	log.Debug(
		"创建菜单模型：执行成功",
		zap.Object("menu_model", m),
		zap.Duration("create_menu_duration", time.Since(now)),
	)
	return nil
}

// UpdateModel 更新菜单模型
//
// 参数：
//
//	ctx: 上下文，用于传递追踪信息和控制超时
//	data: 更新数据映射
//	apis: 关联的API模型列表指针
//	conds: 查询条件
//
// 返回值：
//
//	error: 操作过程中的错误
func (r *MenuRepo) UpdateModel(
	ctx context.Context,
	updateData map[string]any,
	apis []sysmodel.ApiModel,
	conds ...any,
) error {
	log := ctxutil.NewLogger(r.log, ctx)

	apiIDs := sysmodel.ListApiModelToUint32s(apis)
	log.Debug(
		"更新菜单模型：开始执行",
		zap.Any("update_data", updateData),
		zap.Uint32s("apis", apiIDs),
		zap.Any("conds", conds),
	)

	now := time.Now()
	// 构建关联关系映射
	var upmap map[string]any
	if len(apis) > 0 {
		upmap = map[string]any{"Apis": apis}
	}
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	if err := database.DBUpdate(dbCtx, r.gormDB, &sysmodel.MenuModel{}, updateData, upmap, conds...); err != nil {
		log.Error(
			"更新菜单模型：数据库更新失败",
			zap.Error(err),
			zap.Any("update_data", updateData),
			zap.Uint32s("apis", apiIDs),
			zap.Any("conds", conds),
			zap.Duration("update_menu_duration", time.Since(now)),
		)
		return errors.WrapIf(err, "更新菜单模型：数据库更新失败")
	}

	log.Debug(
		"更新菜单模型：执行成功",
		zap.Any("update_data", updateData),
		zap.Any("conds", conds),
		zap.Uint32s("apis", apiIDs),
		zap.Duration("update_menu_duration", time.Since(now)),
	)
	return nil
}

// DeleteModel 删除菜单模型
//
// 参数：
//
//	ctx: 上下文，用于传递追踪信息和控制超时
//	conds: 查询条件
//
// 返回值：
//
//	error: 操作过程中的错误
func (r *MenuRepo) DeleteModel(ctx context.Context, conds ...any) error {
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"删除菜单模型：开始执行",
		zap.Any("conds", conds),
	)

	now := time.Now()
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	if err := database.DBDelete(dbCtx, r.gormDB, &sysmodel.MenuModel{}, conds...); err != nil {
		log.Error(
			"删除菜单模型：数据库删除失败",
			zap.Error(err),
			zap.Any("conds", conds),
			zap.Duration("delete_menu_duration", time.Since(now)),
		)
		return errors.WrapIf(err, "删除菜单模型：数据库删除失败")
	}

	log.Debug(
		"删除菜单模型：执行成功",
		zap.Any("conds", conds),

		zap.Duration("delete_menu_duration", time.Since(now)),
	)
	return nil
}

// GetModel 查询单个菜单模型
//
// 参数：
//
//	ctx: 上下文，用于传递追踪信息和控制超时
//	preloads: 需要预加载的关联关系
//	conds: 查询条件
//
// 返回值：
//
//	*sysmodel.MenuModel: 查询到的菜单模型指针
//	error: 操作过程中的错误
func (r *MenuRepo) GetModel(
	ctx context.Context,
	preloads []string,
	conds ...any,
) (*sysmodel.MenuModel, error) {
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"查询菜单模型：开始执行",
		zap.Strings("preloads", preloads),
		zap.Any("conds", conds),
	)

	now := time.Now()
	var m sysmodel.MenuModel
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()
	if err := database.DBGet(dbCtx, r.gormDB, preloads, &m, conds...); err != nil {
		log.Error(
			"查询菜单模型：数据库查询失败",
			zap.Error(err),
			zap.Strings("preloads", preloads),
			zap.Any("conds", conds),
			zap.Duration("get_menu_duration", time.Since(now)),
		)
		return nil, errors.WrapIf(err, "查询菜单模型：数据库查询失败")
	}

	log.Debug(
		"查询菜单模型：执行成功",
		zap.Object("menu_model", &m),
		zap.Strings("preloads", preloads),
		zap.Any("conds", conds),

		zap.Duration("get_menu_duration", time.Since(now)),
	)
	return &m, nil
}

// ListModel 查询菜单模型列表
//
// 参数：
//
//	ctx: 上下文，用于传递追踪信息和控制超时
//	qp: 查询参数，包含分页、排序等信息
//
// 返回值：
//
//	int64: 总记录数
//	[]sysmodel.MenuModel: 菜单模型列表指针
//	error: 操作过程中的错误
func (r *MenuRepo) ListModel(
	ctx context.Context,
	qp database.QueryParams,
) ([]sysmodel.MenuModel, error) {
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"查询菜单模型列表：开始执行",
		zap.Object("query_params", &qp),
	)

	now := time.Now()
	var ms []sysmodel.MenuModel
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ListTimeout)
	defer cancel()
	if err := database.DBList(dbCtx, r.gormDB, &sysmodel.MenuModel{}, &ms, qp); err != nil {
		log.Error(
			"查询菜单模型列表：数据库查询失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("list_menu_duration", time.Since(now)),
		)
		return nil, errors.WrapIf(err, "查询菜单模型列表：数据库查询失败")
	}

	log.Debug(
		"查询菜单模型列表：执行成功",
		zap.Object("query_params", &qp),
		zap.Duration("list_menu_duration", time.Since(now)),
	)
	return ms, nil
}

func (r *MenuRepo) CountModel(
	ctx context.Context,
	query map[string]any,
) (int64, error) {
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"查询菜单模型总数：开始执行",
		zap.Any("query", query),
	)

	now := time.Now()
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()
	count, err := database.DBCount(dbCtx, r.gormDB, &sysmodel.MenuModel{}, query)
	if err != nil {
		log.Error(
			"查询菜单模型总数：数据库查询失败",
			zap.Error(err),
			zap.Any("query", query),
			zap.Duration("count_menu_duration", time.Since(now)),
		)
		return 0, errors.WrapIf(err, "查询菜单模型总数：数据库查询失败")
	}
	log.Debug(
		"查询菜单模型总数：执行成功",
		zap.Any("query", query),
		zap.Int64("count", count),
		zap.Duration("count_menu_duration", time.Since(now)),
	)
	return count, nil
}

// AddGroupPolicy 添加菜单的权限策略
//
// 参数：
//
//	ctx: 上下文，用于传递追踪信息和控制超时
//	menu: 菜单模型指针
//
// 返回值：
//
//	error: 操作过程中的错误
//
// 功能：
//  1. 检查上下文是否有效
//  2. 检查菜单模型是否为空
//  3. 检查菜单ID是否有效
//  4. 将菜单转换为Casbin策略
//  5. 添加菜单策略到Casbin
//  6. 记录操作日志
func (r *MenuRepo) AddGroupPolicy(
	ctx context.Context,
	menu *sysmodel.MenuModel,
) error {
	// 检查上下文
	if ctx.Err() != nil {
		return errors.WrapIf(ctx.Err(), "添加菜单关联策略：上下文已取消或超时")
	}

	// 检查参数
	if menu == nil {
		return errors.New("添加菜单关联策略：菜单模型不能为空")
	}

	m := *menu
	// 检查必要字段
	if m.ID == 0 {
		return errors.New("添加菜单关联策略：菜单ID不能为0")
	}

	log := ctxutil.NewLogger(r.log, ctx)

	apiIDs := sysmodel.ListApiModelToUint32s(m.Apis)
	log.Debug(
		"添加菜单关联策略：开始执行",
		zap.Object("menu_model", menu),
		zap.Uint32s("apis", apiIDs),
	)

	now := time.Now()
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
				"添加菜单关联策略：跳过API模型ID为0的关联策略",
				zap.Object("menu_model", menu),
			)
			continue
		}

		obj := auth.ApiToSubject(o.ID)
		rules = append(rules, []string{sub, obj})

	}
	if err := auth.AddGroupPolicies(ctx, r.enforcer, rules); err != nil {
		log.Error(
			"添加菜单关联策略：Casbin添加策略失败",
			zap.Error(err),
			zap.Object("menu_model", menu),
			zap.Any("rules", rules),
			zap.Duration("add_menu_policy_duration", time.Since(now)),
		)
		return errors.WrapIf(err, "添加菜单关联策略：Casbin添加策略失败")
	}

	log.Debug(
		"添加菜单关联策略：执行成功",
		zap.Object("menu_model", menu),
		zap.Uint32s("apis", apiIDs),
		zap.Duration("add_menu_policy_duration", time.Since(now)),
	)
	return nil
}

// RemoveGroupPolicy 删除菜单的权限策略
//
// 参数：
//
//	ctx: 上下文，用于传递追踪信息和控制超时
//	menu: 菜单模型指针
//	removeInherited: 是否删除继承该菜单的组策略
//
// 返回值：
//
//	error: 操作过程中的错误
//
// 功能：
//  1. 检查上下文是否有效
//  2. 检查菜单模型是否为空
//  3. 检查菜单ID是否有效
//  4. 删除该菜单作为子级的组策略（被其他策略继承）
//  5. 可选：删除该菜单作为父级的组策略（被其他菜单或API继承）
//  6. 记录操作日志
func (r *MenuRepo) RemoveGroupPolicy(
	ctx context.Context,
	menu *sysmodel.MenuModel,
	removeInherited bool,
) error {
	// 检查上下文
	if ctx.Err() != nil {
		return errors.WrapIf(ctx.Err(), "删除菜单关联策略：上下文已取消或超时")
	}

	// 检查参数
	if menu == nil {
		return errors.New("删除菜单关联策略：菜单模型不能为空")
	}

	m := *menu
	// 检查必要字段
	if m.ID == 0 {
		return errors.New("删除菜单关联策略：菜单ID不能为0")
	}

	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"删除菜单关联策略：开始执行",
		zap.Object("menu_model", &m),
		zap.Bool("removeInherited", removeInherited),
	)

	rmSubStartTime := time.Now()
	sub := auth.MenuToSubject(m.ID)
	if err := auth.RemoveFilteredGroupingPolicy(ctx, r.enforcer, 0, sub); err != nil {
		log.Error(
			"删除菜单关联策略：Casbin删除策略失败",
			zap.Error(err),
			zap.Object("menu_model", &m),
			zap.String("sub", sub),
			zap.Duration("remove_policy_duration", time.Since(rmSubStartTime)),
		)
		return errors.WrapIf(err, "删除菜单关联策略：Casbin删除策略失败")
	}
	log.Debug(
		"删除菜单关联策略：执行成功",
		zap.Object("menu_model", &m),
		zap.String("sub", sub),
		zap.Duration("remove_policy_duration", time.Since(rmSubStartTime)),
	)

	if removeInherited {
		rmObjStartTime := time.Now()
		if err := auth.RemoveFilteredGroupingPolicy(ctx, r.enforcer, 1, sub); err != nil {
			log.Error(
				"删除菜单关联策略：Casbin删除策略失败",
				zap.Error(err),
				zap.Object("menu_model", &m),
				zap.String("sub", sub),
				zap.Duration("remove_group_policy_duration", time.Since(rmObjStartTime)),
			)
			return errors.WrapIf(err, "删除菜单关联策略：Casbin删除策略失败")
		}

		log.Debug(
			"删除菜单关联策略：执行成功",
			zap.Object("menu_model", &m),
			zap.String("obj", sub),
			zap.Duration("remove_group_policy_duration", time.Since(rmObjStartTime)),
		)
	}
	return nil
}
