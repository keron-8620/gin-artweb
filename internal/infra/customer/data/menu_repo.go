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
//	model.MenuRepo: 菜单仓库接口实现
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
	m *model.MenuModel,
	apis *[]model.ApiModel,
) error {
	// 检查参数
	if m == nil {
		err := errors.New("创建菜单模型失败: 模型为空")
		r.log.Error(
			"创建菜单模型失败: 模型为空",
			zap.Error(err),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return err
	}

	r.log.Debug(
		"开始创建菜单模型",
		zap.Object(database.ModelKey, m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	now := time.Now()
	m.CreatedAt = now
	m.UpdatedAt = now

	// 构建关联关系映射
	upmap := make(map[string]any, 1)
	if apis != nil {
		if len(*apis) > 0 {
			upmap["Apis"] = *apis
		} else {
			upmap["Apis"] = []model.ApiModel{}
		}
	}

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	if err := database.DBCreate(dbCtx, r.gormDB, &model.MenuModel{}, m, upmap); err != nil {
		r.log.Error(
			"创建菜单模型失败",
			zap.Error(err),
			zap.Object(database.ModelKey, m),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return errors.WrapIf(err, "创建菜单模型失败")
	}

	r.log.Debug(
		"创建菜单模型成功",
		zap.Object(database.ModelKey, m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(now)),
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
	data map[string]any,
	apis *[]model.ApiModel,
	conds ...any,
) error {
	r.log.Debug(
		"开始更新菜单模型",
		zap.Any(database.UpdateDataKey, data),
		zap.Uint32s("apis", ListApiModelToUint32s(apis)),
		zap.Any(database.ConditionsKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	now := time.Now()
	// 构建关联关系映射
	upmap := make(map[string]any, 1)
	if apis != nil {
		if len(*apis) > 0 {
			upmap["Apis"] = *apis
		} else {
			upmap["Apis"] = []model.ApiModel{}
		}
	}

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	if err := database.DBUpdate(dbCtx, r.gormDB, &model.MenuModel{}, data, upmap, conds...); err != nil {
		r.log.Error(
			"更新菜单模型失败",
			zap.Error(err),
			zap.Any(database.UpdateDataKey, data),
			zap.Uint32s("apis", ListApiModelToUint32s(apis)),
			zap.Any(database.ConditionsKey, conds),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return errors.WrapIf(err, "更新菜单模型失败")
	}

	r.log.Debug(
		"更新菜单模型成功",
		zap.Any(database.UpdateDataKey, data),
		zap.Any(database.ConditionsKey, conds),
		zap.Uint32s("apis", ListApiModelToUint32s(apis)),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(now)),
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
	r.log.Debug(
		"开始删除菜单模型",
		zap.Any(database.ConditionsKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	now := time.Now()
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	if err := database.DBDelete(dbCtx, r.gormDB, &model.MenuModel{}, conds...); err != nil {
		r.log.Error(
			"删除菜单模型失败",
			zap.Error(err),
			zap.Any(database.ConditionsKey, conds),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return errors.WrapIf(err, "删除菜单模型失败")
	}

	r.log.Debug(
		"删除菜单模型成功",
		zap.Any(database.ConditionsKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(now)),
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
//	*model.MenuModel: 查询到的菜单模型指针
//	error: 操作过程中的错误
func (r *MenuRepo) GetModel(
	ctx context.Context,
	preloads []string,
	conds ...any,
) (*model.MenuModel, error) {
	r.log.Debug(
		"开始获取菜单模型",
		zap.Strings(database.PreloadKey, preloads),
		zap.Any(database.ConditionsKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	now := time.Now()
	var m model.MenuModel
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()
	if err := database.DBGet(dbCtx, r.gormDB, preloads, &m, conds...); err != nil {
		r.log.Error(
			"获取菜单模型失败",
			zap.Error(err),
			zap.Strings(database.PreloadKey, preloads),
			zap.Any(database.ConditionsKey, conds),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return nil, errors.WrapIf(err, "获取菜单模型失败")
	}

	r.log.Debug(
		"获取菜单模型成功",
		zap.Object(database.ModelKey, &m),
		zap.Strings(database.PreloadKey, preloads),
		zap.Any(database.ConditionsKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(now)),
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
//	*[]model.MenuModel: 菜单模型列表指针
//	error: 操作过程中的错误
func (r *MenuRepo) ListModel(
	ctx context.Context,
	qp database.QueryParams,
) (int64, *[]model.MenuModel, error) {
	r.log.Debug(
		"开始查询菜单模型列表",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	now := time.Now()
	var ms []model.MenuModel
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ListTimeout)
	defer cancel()
	count, err := database.DBList(dbCtx, r.gormDB, &model.MenuModel{}, &ms, qp)
	if err != nil {
		r.log.Error(
			"查询菜单模型列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return 0, nil, errors.WrapIf(err, "查询菜单模型列表失败")
	}

	r.log.Debug(
		"查询菜单模型列表成功",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(now)),
	)
	return count, &ms, nil
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
	menu *model.MenuModel,
) error {
	// 检查上下文
	if ctx.Err() != nil {
		return errors.WrapIf(ctx.Err(), "AddGroupPolicy操作时上下文已取消或超时")
	}

	// 检查参数
	if menu == nil {
		return errors.New("AddGroupPolicy操作时菜单模型不能为空")
	}

	m := *menu
	// 检查必要字段
	if m.ID == 0 {
		return errors.New("AddGroupPolicy操作时菜单ID不能为0")
	}

	r.log.Debug(
		"开始添加菜单关联策略",
		zap.Object(database.ModelKey, menu),
		zap.Uint32s("apis", ListApiModelToUint32s(&m.Apis)),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
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
			r.log.Warn(
				"跳过无效API",
				zap.Object(database.ModelKey, menu),
				zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			)
			continue
		}

		obj := auth.ApiToSubject(o.ID)
		rules = append(rules, []string{sub, obj})

	}
	if err := auth.AddGroupPolicies(ctx, r.enforcer, rules); err != nil {
		r.log.Error(
			"添加菜单关联策略失败",
			zap.Error(err),
			zap.Object(database.ModelKey, menu),
			zap.Any("rules", rules),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return errors.WrapIf(err, "添加菜单关联策略失败")
	}

	r.log.Debug(
		"添加菜单关联策略成功",
		zap.Object(database.ModelKey, menu),
		zap.Uint32s("apis", ListApiModelToUint32s(&m.Apis)),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(now)),
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
	menu *model.MenuModel,
	removeInherited bool,
) error {
	// 检查上下文
	if ctx.Err() != nil {
		return errors.WrapIf(ctx.Err(), "RemoveGroupPolicy操作时上下文已取消或超时")
	}

	// 检查参数
	if menu == nil {
		return errors.New("RemoveGroupPolicy操作时菜单模型不能为空")
	}

	m := *menu
	// 检查必要字段
	if m.ID == 0 {
		return errors.New("RemoveGroupPolicy操作时菜单ID不能为0")
	}

	r.log.Debug(
		"开始删除该菜单作为子级的组策略",
		zap.Object(database.ModelKey, &m),
		zap.Bool("removeInherited", removeInherited),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	rmSubStartTime := time.Now()
	sub := auth.MenuToSubject(m.ID)
	if err := auth.RemoveFilteredGroupingPolicy(ctx, r.enforcer, 0, sub); err != nil {
		r.log.Error(
			"删除该菜单作为子级的组策略失败",
			zap.Error(err),
			zap.Object(database.ModelKey, menu),
			zap.String(auth.GroupSubKey, sub),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(rmSubStartTime)),
		)
		return errors.WrapIf(err, "删除该菜单作为子级的组策略失败")
	}
	r.log.Debug(
		"删除该菜单作为子级的组策略成功",
		zap.Object(database.ModelKey, menu),
		zap.String(auth.GroupSubKey, sub),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(rmSubStartTime)),
	)

	if removeInherited {
		rmObjStartTime := time.Now()
		if err := auth.RemoveFilteredGroupingPolicy(ctx, r.enforcer, 1, sub); err != nil {
			r.log.Error(
				"删除该菜单作为父级的组策略失败",
				zap.Error(err),
				zap.Object(database.ModelKey, menu),
				zap.String(auth.GroupObjKey, sub),
				zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
				zap.Duration(log.DurationKey, time.Since(rmObjStartTime)),
			)
			return errors.WrapIf(err, "删除该菜单作为父级的组策略失败")
		}

		r.log.Debug(
			"删除该菜单作为父级的组策略成功",
			zap.Object(database.ModelKey, menu),
			zap.String(auth.GroupObjKey, sub),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(rmObjStartTime)),
		)
	}
	return nil
}

func ListMenuModelToUint32s(mms *[]model.MenuModel) []uint32 {
	if mms == nil {
		return []uint32{}
	}
	ms := *mms
	if len(ms) == 0 {
		return []uint32{}
	}

	ids := make([]uint32, len(ms))
	for i, m := range ms {
		ids[i] = m.ID
	}
	return ids
}
