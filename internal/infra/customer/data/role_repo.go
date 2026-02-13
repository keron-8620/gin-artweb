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

// RoleRepo 角色仓库实现
// 负责角色模型的CRUD操作和角色权限策略的管理
// 使用GORM进行数据库操作，使用Casbin进行权限策略管理
type RoleRepo struct {
	log      *zap.Logger       // 日志记录器
	gormDB   *gorm.DB          // GORM数据库连接
	timeouts *config.DBTimeout // 数据库操作超时配置
	enforcer *casbin.Enforcer  // Casbin权限管理器
}

// NewRoleRepo 创建角色仓库实例
//
// 参数：
//
//	log: 日志记录器，用于记录操作日志
//	gormDB: GORM数据库连接，用于执行数据库操作
//	timeouts: 数据库操作超时配置，控制各类数据库操作的超时时间
//	enforcer: Casbin权限管理器，用于管理权限策略
//
// 返回值：
//
//	model.RoleRepo: 角色仓库接口实现
func NewRoleRepo(
	log *zap.Logger,
	gormDB *gorm.DB,
	timeouts *config.DBTimeout,
	enforcer *casbin.Enforcer,
) *RoleRepo {
	return &RoleRepo{
		log:      log,
		gormDB:   gormDB,
		timeouts: timeouts,
		enforcer: enforcer,
	}
}

// CreateModel 创建角色模型
//
// 参数：
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	m: 角色模型，包含角色的详细信息
//	apis: API模型列表，包含与角色关联的权限
//	menus: 菜单模型列表，包含与角色关联的菜单
//	buttons: 按钮模型列表，包含与角色关联的按钮
//
// 返回值：
//
//	error: 操作错误信息，成功则返回nil
//
// 功能：
//  1. 检查角色模型是否为空
//  2. 设置创建时间和更新时间
//  3. 处理关联的权限、菜单和按钮信息
//  4. 执行数据库创建操作
//  5. 记录操作日志
func (r *RoleRepo) CreateModel(
	ctx context.Context,
	m *model.RoleModel,
	apis *[]model.ApiModel,
	menus *[]model.MenuModel,
	buttons *[]model.ButtonModel,
) error {
	// 检查参数
	if m == nil {
		err := errors.New("创建角色模型失败: 模型为空")
		r.log.Error(
			"创建角色模型失败: 模型为空",
			zap.Error(err),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return err
	}

	r.log.Debug(
		"开始创建角色模型",
		zap.Object(database.ModelKey, m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	now := time.Now()
	m.CreatedAt = now
	m.UpdatedAt = now

	upmap := make(map[string]any, 3)
	if apis != nil {
		if len(*apis) > 0 {
			upmap["Apis"] = *apis
		}
	}
	if menus != nil {
		if len(*menus) > 0 {
			upmap["Menus"] = *menus
		}
	}
	if buttons != nil {
		if len(*buttons) > 0 {
			upmap["Buttons"] = *buttons
		}
	}

	if err := database.DBCreate(ctx, r.gormDB, &model.RoleModel{}, m, upmap); err != nil {
		r.log.Error(
			"创建角色模型失败",
			zap.Error(err),
			zap.Object(database.ModelKey, m),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return errors.WrapIf(err, "创建角色模型失败")
	}

	r.log.Debug(
		"创建角色模型成功",
		zap.Object(database.ModelKey, m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(now)),
	)
	return nil
}

// UpdateModel 更新角色模型
//
// 参数：
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	data: 更新数据，包含要更新的字段和值
//	apis: API模型列表，包含与角色关联的权限
//	menus: 菜单模型列表，包含与角色关联的菜单
//	buttons: 按钮模型列表，包含与角色关联的按钮
//	conds: 查询条件，用于指定要更新的记录
//
// 返回值：
//
//	error: 操作错误信息，成功则返回nil
//
// 功能：
//  1. 检查更新数据是否为空
//  2. 处理关联的权限、菜单和按钮信息
//  3. 执行数据库更新操作
//  4. 记录操作日志
func (r *RoleRepo) UpdateModel(
	ctx context.Context,
	data map[string]any,
	apis *[]model.ApiModel,
	menus *[]model.MenuModel,
	buttons *[]model.ButtonModel,
	conds ...any,
) error {
	if len(data) == 0 {
		err := errors.New("更新角色模型失败: 更新数据为空")
		r.log.Error(
			"更新角色模型失败: 更新数据为空",
			zap.Error(err),
			zap.Any(database.UpdateDataKey, data),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return err
	}
	r.log.Debug(
		"开始更新角色模型",
		zap.Any(database.UpdateDataKey, data),
		zap.Any(database.ConditionsKey, conds),
		zap.Uint32s("apis", ListApiModelToUint32s(apis)),
		zap.Uint32s("menus", ListMenuModelToUint32s(menus)),
		zap.Uint32s("buttons", ListButtonModelToUint32s(buttons)),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	now := time.Now()
	upmap := make(map[string]any, 3)
	if apis != nil {
		if len(*apis) > 0 {
			upmap["Apis"] = *apis
		} else {
			upmap["Apis"] = []model.ApiModel{}
		}
	}
	if menus != nil {
		if len(*menus) > 0 {
			upmap["Menus"] = *menus
		} else {
			upmap["Menus"] = []model.MenuModel{}
		}
	}
	if buttons != nil {
		if len(*buttons) > 0 {
			upmap["Buttons"] = *buttons
		} else {
			upmap["Buttons"] = []model.ButtonModel{}
		}
	}
	if err := database.DBUpdate(ctx, r.gormDB, &model.RoleModel{}, data, upmap, conds...); err != nil {
		r.log.Error(
			"更新角色模型失败",
			zap.Error(err),
			zap.Any(database.UpdateDataKey, data),
			zap.Uint32s("apis", ListApiModelToUint32s(apis)),
			zap.Uint32s("menus", ListMenuModelToUint32s(menus)),
			zap.Uint32s("buttons", ListButtonModelToUint32s(buttons)),
			zap.Any(database.ConditionsKey, conds),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return errors.WrapIf(err, "更新角色模型失败")
	}
	r.log.Debug(
		"更新角色模型成功",
		zap.Any(database.UpdateDataKey, data),
		zap.Uint32s("apis", ListApiModelToUint32s(apis)),
		zap.Uint32s("menus", ListMenuModelToUint32s(menus)),
		zap.Uint32s("buttons", ListButtonModelToUint32s(buttons)),
		zap.Any(database.ConditionsKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(now)),
	)
	return nil
}

// DeleteModel 删除角色模型
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
func (r *RoleRepo) DeleteModel(ctx context.Context, conds ...any) error {
	r.log.Debug(
		"开始删除角色模型",
		zap.Any(database.ConditionsKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	now := time.Now()
	if err := database.DBDelete(ctx, r.gormDB, &model.RoleModel{}, conds...); err != nil {
		r.log.Error(
			"删除角色模型失败",
			zap.Error(err),
			zap.Any(database.ConditionsKey, conds),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return errors.WrapIf(err, "删除角色模型失败")
	}
	r.log.Debug(
		"删除角色模型成功",
		zap.Any(database.ConditionsKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(now)),
	)
	return nil
}

// GetModel 获取单个角色模型
//
// 参数：
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	preloads: 预加载的关联字段列表
//	conds: 查询条件，用于指定要获取的记录
//
// 返回值：
//
//	*model.RoleModel: 角色模型指针，包含角色的详细信息
//	error: 操作错误信息，成功则返回nil
//
// 功能：
//  1. 执行数据库查询操作
//  2. 预加载关联字段
//  3. 获取单个角色模型
//  4. 记录操作日志
func (r *RoleRepo) GetModel(
	ctx context.Context,
	preloads []string,
	conds ...any,
) (*model.RoleModel, error) {
	r.log.Debug(
		"开始查询角色模型",
		zap.Strings(database.PreloadKey, preloads),
		zap.Any(database.ConditionsKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	now := time.Now()
	var m model.RoleModel
	if err := database.DBGet(ctx, r.gormDB, preloads, &m, conds...); err != nil {
		r.log.Error(
			"查询角色模型失败",
			zap.Error(err),
			zap.Strings(database.PreloadKey, preloads),
			zap.Any(database.ConditionsKey, conds),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return nil, errors.WrapIf(err, "查询角色模型失败")
	}
	r.log.Debug(
		"查询角色模型成功",
		zap.Object(database.ModelKey, &m),
		zap.Strings(database.PreloadKey, preloads),
		zap.Any(database.ConditionsKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(now)),
	)
	return &m, nil
}

// ListModel 获取角色模型列表
//
// 参数：
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	qp: 查询参数，包含分页、排序等查询条件
//
// 返回值：
//
//	int64: 总记录数
//	*[]model.RoleModel: 角色模型列表指针，包含符合条件的角色模型
//	error: 操作错误信息，成功则返回nil
//
// 功能：
//  1. 执行数据库查询操作
//  2. 获取角色模型列表
//  3. 返回总记录数和模型列表
//  4. 记录操作日志
func (r *RoleRepo) ListModel(
	ctx context.Context,
	qp database.QueryParams,
) (int64, *[]model.RoleModel, error) {
	r.log.Debug(
		"开始查询角色模型列表",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	now := time.Now()
	var ms []model.RoleModel
	count, err := database.DBList(ctx, r.gormDB, &model.RoleModel{}, &ms, qp)
	if err != nil {
		r.log.Error(
			"查询角色模型列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return 0, nil, errors.WrapIf(err, "查询角色模型列表失败")
	}
	r.log.Debug(
		"查询角色模型列表成功",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(now)),
	)
	return count, &ms, nil
}

// AddGroupPolicy 添加角色组策略
//
// 参数：
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	role: 角色模型，包含角色的详细信息
//
// 返回值：
//
//	error: 操作错误信息，成功则返回nil
//
// 功能：
//  1. 检查角色模型是否为空
//  2. 检查角色ID是否有效
//  3. 将权限转换为Casbin策略
//  4. 将菜单转换为Casbin策略
//  5. 将按钮转换为Casbin策略
//  6. 添加角色策略到Casbin
//  7. 记录操作日志
func (r *RoleRepo) AddGroupPolicy(
	ctx context.Context,
	role *model.RoleModel,
) error {
	// 检查参数
	if role == nil {
		return errors.New("AddGroupPolicy操作失败: 角色模型不能为空")
	}

	m := *role

	// 检查必要字段
	if m.ID == 0 {
		return errors.New("AddGroupPolicy操作失败: 角色ID不能为0")
	}

	r.log.Debug(
		"开始添加角色关联策略",
		zap.Object(database.ModelKey, role),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	now := time.Now()
	sub := auth.RoleToSubject(m.ID)
	rules := [][]string{}
	// 批量处理权限
	for i, o := range m.Apis {
		// 检查API模型的有效性
		if o.ID == 0 {
			r.log.Warn(
				"跳过无效权限",
				zap.Object(database.ModelKey, role),
				zap.Int("api_index", i),
				zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			)
			continue
		}
		obj := auth.ApiToSubject(o.ID)
		rules = append(rules, []string{sub, obj})
	}

	// 批量处理菜单
	for i, o := range m.Menus {
		// 检查菜单模型的有效性
		if o.ID == 0 {
			r.log.Warn(
				"跳过无效菜单",
				zap.Object(database.ModelKey, role),
				zap.Int("menu_index", i),
				zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			)
			continue
		}
		obj := auth.MenuToSubject(o.ID)
		rules = append(rules, []string{sub, obj})
	}

	// 批量处理按钮
	for i, o := range m.Buttons {
		// 检查按钮模型的有效性
		if o.ID == 0 {
			r.log.Warn(
				"跳过无效按钮",
				zap.Object(database.ModelKey, role),
				zap.Int("button_index", i),
				zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			)
			continue
		}
		obj := auth.ButtonToSubject(o.ID)
		rules = append(rules, []string{sub, obj})
	}
	if err := auth.AddGroupPolicies(ctx, r.enforcer, rules); err != nil {
		r.log.Error(
			"添加角色关联策略失败",
			zap.Error(err),
			zap.Object(database.ModelKey, role),
			zap.Any("rules", rules),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return errors.WrapIf(err, "添加角色关联策略失败")
	}
	r.log.Debug(
		"添加角色关联策略成功",
		zap.Object(database.ModelKey, role),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(now)),
	)
	return nil
}

// RemoveGroupPolicy 删除角色组策略
//
// 参数：
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	role: 角色模型，包含角色的详细信息
//
// 返回值：
//
//	error: 操作错误信息，成功则返回nil
//
// 功能：
//  1. 检查角色模型是否为空
//  2. 检查角色ID是否有效
//  3. 删除该角色作为子级的组策略（被其他策略继承）
//  4. 记录操作日志
func (r *RoleRepo) RemoveGroupPolicy(
	ctx context.Context,
	role *model.RoleModel,
) error {
	// 检查参数
	if role == nil {
		return errors.New("RemoveGroupPolicy操作失败: 角色模型不能为空")
	}

	m := *role
	// 检查必要字段
	if m.ID == 0 {
		return errors.New("RemoveGroupPolicy操作失败: 角色ID不能为0")
	}

	r.log.Debug(
		"开始删除该角色作为子级的组策略",
		zap.Object(database.ModelKey, role),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	rmSubStartTime := time.Now()
	sub := auth.RoleToSubject(m.ID)

	// 删除该角色作为子级的策略（被其他策略继承）
	if err := auth.RemoveFilteredGroupingPolicy(ctx, r.enforcer, 0, sub); err != nil {
		r.log.Error(
			"删除角色作为子级策略失败(该策略继承自其他策略)",
			zap.Error(err),
			zap.Object(database.ModelKey, role),
			zap.String(auth.GroupSubKey, sub),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(rmSubStartTime)),
		)
		return errors.WrapIf(err, "删除角色作为子级策略失败(该策略继承自其他策略)")
	}
	r.log.Debug(
		"删除该角色作为子级的组策略成功",
		zap.Object(database.ModelKey, role),
		zap.String(auth.GroupSubKey, sub),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(rmSubStartTime)),
	)
	return nil
}
