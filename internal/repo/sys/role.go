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
// 参数:
//
//	log: 日志记录器，用于记录操作日志
//	gormDB: GORM数据库连接，用于执行数据库操作
//	timeouts: 数据库操作超时配置，控制各类数据库操作的超时时间
//	enforcer: Casbin权限管理器，用于管理权限策略
//
// 返回值:
//
//	sysmodel.RoleRepo: 角色仓库接口实现
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
// 参数:
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	m: 角色模型，包含角色的详细信息
//	apis: API模型列表，包含与角色关联的权限
//	menus: 菜单模型列表，包含与角色关联的菜单
//	buttons: 按钮模型列表，包含与角色关联的按钮
//
// 返回值:
//
//	error: 操作错误信息，成功则返回nil
//
// 功能:
//  1. 检查角色模型是否为空
//  2. 设置创建时间和更新时间
//  3. 处理关联的权限、菜单和按钮信息
//  4. 执行数据库创建操作
//  5. 记录操作日志
func (r *RoleRepo) CreateModel(
	ctx context.Context,
	m *sysmodel.RoleModel,
	apis []sysmodel.ApiModel,
	menus []sysmodel.MenuModel,
	buttons []sysmodel.ButtonModel,
) error {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	// 检查参数
	if m == nil {
		err := errors.New("创建角色模型: 模型不能为空")
		log.Error(
			"创建角色模型: 模型不能为空",
			zap.Error(err),
		)
		return err
	}

	log.Debug(
		"创建角色模型:开始执行",
		zap.Object("role_model", m),
	)

	m.CreatedAt = startTime
	m.UpdatedAt = startTime

	upmap := map[string]any{
		"Apis":    apis,
		"Menus":   menus,
		"Buttons": buttons,
	}

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	createTime := time.Now()
	err := database.DBCreate(dbCtx, r.gormDB, &sysmodel.RoleModel{}, m, upmap)
	createDuration := time.Since(createTime)
	if err != nil {
		log.Error(
			"创建角色模型:数据库创建失败",
			zap.Error(err),
			zap.Object("role_model", m),
			zap.Duration("create_role_duration", createDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "创建角色模型:数据库创建失败")
	}

	log.Debug(
		"创建角色模型:执行成功",
		zap.Object("role_model", m),
		zap.Duration("create_role_duration", createDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

// UpdateModel 更新角色模型
//
// 参数:
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	data: 更新数据，包含要更新的字段和值
//	apis: API模型列表，包含与角色关联的权限
//	menus: 菜单模型列表，包含与角色关联的菜单
//	buttons: 按钮模型列表，包含与角色关联的按钮
//	conds: 查询条件，用于指定要更新的记录
//
// 返回值:
//
//	error: 操作错误信息，成功则返回nil
//
// 功能:
//  1. 检查更新数据是否为空
//  2. 处理关联的权限、菜单和按钮信息
//  3. 执行数据库更新操作
//  4. 记录操作日志
func (r *RoleRepo) UpdateModel(
	ctx context.Context,
	data map[string]any,
	apis []sysmodel.ApiModel,
	menus []sysmodel.MenuModel,
	buttons []sysmodel.ButtonModel,
	conds ...any,
) error {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	if len(data) == 0 {
		err := errors.New("更新角色模型: 更新数据不能为空")
		log.Error(
			"更新角色模型: 更新数据不能为空",
			zap.Error(err),
			zap.Any("update_data", data),
		)
		return err
	}

	apiIDs := sysmodel.ListApiModelToUint32s(apis)
	menuIDs := sysmodel.ListMenuModelToUint32s(menus)
	buttonIDs := sysmodel.ListButtonModelToUint32s(buttons)
	log.Debug(
		"更新角色模型:开始执行",
		zap.Any("update_data", data),
		zap.Any("conds", conds),
		zap.Uint32s("apis", apiIDs),
		zap.Uint32s("menus", menuIDs),
		zap.Uint32s("buttons", buttonIDs),
	)

	upmap := make(map[string]any, 3)
	if len(apis) > 0 {
		upmap["Apis"] = apis
	}
	if len(menus) > 0 {
		upmap["Menus"] = menus
	}
	if len(buttons) > 0 {
		upmap["Buttons"] = buttons
	}
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()

	updateTime := time.Now()
	err := database.DBUpdate(dbCtx, r.gormDB, &sysmodel.RoleModel{}, data, upmap, conds...)
	updateDuration := time.Since(updateTime)
	if err != nil {
		log.Error(
			"更新角色模型:数据库更新失败",
			zap.Error(err),
			zap.Any("update_data", data),
			zap.Uint32s("apis", apiIDs),
			zap.Uint32s("menus", menuIDs),
			zap.Uint32s("buttons", buttonIDs),
			zap.Any("conds", conds),
			zap.Duration("update_role_duration", updateDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "更新角色模型:数据库更新失败")
	}
	log.Debug(
		"更新角色模型:执行成功",
		zap.Any("update_data", data),
		zap.Uint32s("apis", apiIDs),
		zap.Uint32s("menus", menuIDs),
		zap.Uint32s("buttons", buttonIDs),
		zap.Any("conds", conds),
		zap.Duration("update_role_duration", updateDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

// DeleteModel 删除角色模型
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
func (r *RoleRepo) DeleteModel(
	ctx context.Context,
	conds ...any,
) error {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"删除角色模型:开始执行",
		zap.Any("conds", conds),
	)
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	deleteTime := time.Now()
	err := database.DBDelete(dbCtx, r.gormDB, &sysmodel.RoleModel{}, conds...)
	deleteDuration := time.Since(deleteTime)
	if err != nil {
		log.Error(
			"删除角色模型:数据库删除失败",
			zap.Error(err),
			zap.Any("conds", conds),
			zap.Duration("delete_role_duration", deleteDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "删除角色模型:数据库删除失败")
	}
	log.Debug(
		"删除角色模型:执行成功",
		zap.Any("conds", conds),
		zap.Duration("delete_role_duration", deleteDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

// GetModel 获取单个角色模型
//
// 参数:
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	preloads: 预加载的关联字段列表
//	conds: 查询条件，用于指定要获取的记录
//
// 返回值:
//
//	*sysmodel.RoleModel: 角色模型指针，包含角色的详细信息
//	error: 操作错误信息，成功则返回nil
//
// 功能:
//  1. 执行数据库查询操作
//  2. 预加载关联字段
//  3. 获取单个角色模型
//  4. 记录操作日志
func (r *RoleRepo) GetModel(
	ctx context.Context,
	preloads []string,
	conds ...any,
) (*sysmodel.RoleModel, error) {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"查询角色模型:开始执行",
		zap.Strings("preloads", preloads),
		zap.Any("conds", conds),
	)
	var m sysmodel.RoleModel
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()
	getTime := time.Now()
	err := database.DBGet(dbCtx, r.gormDB, preloads, &m, conds...)
	getDuration := time.Since(getTime)
	if err != nil {
		log.Error(
			"查询角色模型:数据库查询失败",
			zap.Error(err),
			zap.Strings("preloads", preloads),
			zap.Any("conds", conds),
			zap.Duration("get_role_duration", getDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.WrapIf(err, "查询角色模型:数据库查询失败")
	}
	log.Debug(
		"查询角色模型:执行成功",
		zap.Object("role_model", &m),
		zap.Strings("preloads", preloads),
		zap.Any("conds", conds),
		zap.Duration("get_role_duration", getDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return &m, nil
}

// ListModel 获取角色模型列表
//
// 参数:
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	qp: 查询参数，包含分页、排序等查询条件
//
// 返回值:
//
//	int64: 总记录数
//	[]sysmodel.RoleModel: 角色模型列表指针，包含符合条件的角色模型
//	error: 操作错误信息，成功则返回nil
//
// 功能:
//  1. 执行数据库查询操作
//  2. 获取角色模型列表
//  3. 返回总记录数和模型列表
//  4. 记录操作日志
func (r *RoleRepo) ListModel(
	ctx context.Context,
	qp database.QueryParams,
) ([]sysmodel.RoleModel, error) {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"查询角色模型列表:开始执行",
		zap.Object("query_params", &qp),
	)
	var ms []sysmodel.RoleModel
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()
	listTime := time.Now()
	err := database.DBList(dbCtx, r.gormDB, &sysmodel.RoleModel{}, &ms, qp)
	listDuration := time.Since(listTime)
	if err != nil {
		log.Error(
			"查询角色模型列表:数据库查询失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("list_role_duration", listDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.WrapIf(err, "查询角色模型列表:数据库查询失败")
	}
	log.Debug(
		"查询角色模型列表:执行成功",
		zap.Object("query_params", &qp),
		zap.Duration("list_role_duration", listDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return ms, nil
}

func (r *RoleRepo) CountModel(
	ctx context.Context,
	query map[string]any,
) (int64, error) {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"查询角色模型总数:开始执行",
		zap.Any("query", query),
	)
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()
	countTime := time.Now()
	count, err := database.DBCount(dbCtx, r.gormDB, &sysmodel.RoleModel{}, query)
	countDuration := time.Since(countTime)
	if err != nil {
		log.Error(
			"查询角色模型总数:数据库查询失败",
			zap.Error(err),
			zap.Any("query", query),
			zap.Duration("count_role_duration", countDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, errors.WrapIf(err, "查询角色模型总数:数据库查询失败")
	}
	log.Debug(
		"查询角色模型总数:执行成功",
		zap.Any("query", query),
		zap.Int64("count", count),
		zap.Duration("count_role_duration", countDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return count, nil
}

// AddGroupPolicy 添加角色组策略
//
// 参数:
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	role: 角色模型，包含角色的详细信息
//
// 返回值:
//
//	error: 操作错误信息，成功则返回nil
//
// 功能:
//  1. 检查角色模型是否为空
//  2. 检查角色ID是否有效
//  3. 将权限转换为Casbin策略
//  4. 将菜单转换为Casbin策略
//  5. 将按钮转换为Casbin策略
//  6. 添加角色策略到Casbin
//  7. 记录操作日志
func (r *RoleRepo) AddGroupPolicy(
	ctx context.Context,
	role *sysmodel.RoleModel,
) error {
	startTime := time.Now()

	// 检查参数
	if role == nil {
		return errors.New("添加角色关联策略:角色模型不能为空")
	}

	m := *role

	// 检查必要字段
	if m.ID == 0 {
		return errors.New("添加角色关联策略:角色ID不能为0")
	}

	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"添加角色关联策略:开始执行",
		zap.Object("role_model", role),
	)

	sub := auth.RoleToSubject(m.ID)
	rules := [][]string{}
	// 批量处理权限
	for i, o := range m.Apis {
		// 检查API模型的有效性
		if o.ID == 0 {
			log.Warn(
				"添加角色关联策略:跳过无效权限",
				zap.Object("role_model", role),
				zap.Int("api_index", i),
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
			log.Warn(
				"添加角色关联策略:跳过无效菜单",
				zap.Object("role_model", role),
				zap.Int("menu_index", i),
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
			log.Warn(
				"添加角色关联策略:跳过无效按钮",
				zap.Object("role_model", role),
				zap.Int("button_index", i),
			)
			continue
		}
		obj := auth.ButtonToSubject(o.ID)
		rules = append(rules, []string{sub, obj})
	}

	addGroupPolicyStartTime := time.Now()
	err := auth.AddGroupPolicies(ctx, r.enforcer, rules)
	addGroupPolicyDuration := time.Since(addGroupPolicyStartTime)
	if err != nil {
		log.Error(
			"添加角色关联策略:数据库操作失败",
			zap.Error(err),
			zap.Object("role_model", role),
			zap.Any("rules", rules),
			zap.Duration("add_group_policy_duration", addGroupPolicyDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "添加角色关联策略:数据库操作失败")
	}
	log.Debug(
		"添加角色关联策略:执行成功",
		zap.Object("role_model", role),
		zap.Duration("add_group_policy_duration", addGroupPolicyDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

// RemoveGroupPolicy 删除角色组策略
//
// 参数:
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	role: 角色模型，包含角色的详细信息
//
// 返回值:
//
//	error: 操作错误信息，成功则返回nil
//
// 功能:
//  1. 检查角色模型是否为空
//  2. 检查角色ID是否有效
//  3. 删除该角色作为子级的组策略（被其他策略继承）
//  4. 记录操作日志
func (r *RoleRepo) RemoveGroupPolicy(
	ctx context.Context,
	role *sysmodel.RoleModel,
) error {
	startTime := time.Now()

	// 检查参数
	if role == nil {
		return errors.New("删除角色关联策略:角色模型不能为空")
	}

	m := *role
	// 检查必要字段
	if m.ID == 0 {
		return errors.New("删除角色关联策略:角色ID不能为0")
	}

	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"删除角色关联策略:开始执行",
		zap.Object("role_model", role),
	)

	sub := auth.RoleToSubject(m.ID)

	// 删除该角色作为子级的策略（被其他策略继承）
	removeGroupPolicyStartTime := time.Now()
	err := auth.RemoveFilteredGroupingPolicy(ctx, r.enforcer, 0, sub)
	removeGroupPolicyDuration := time.Since(removeGroupPolicyStartTime)
	if err != nil {
		log.Error(
			"删除角色关联策略:删除角色作为子级策略失败(该策略继承自其他策略)",
			zap.Error(err),
			zap.Object("role_model", role),
			zap.Int("index", 0),
			zap.String("value", sub),
			zap.Duration("remove_group_policy_duration", removeGroupPolicyDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "删除角色关联策略:删除角色作为子级策略失败(该策略继承自其他策略)")
	}
	log.Debug(
		"删除角色关联策略:删除角色作为子级策略成功",
		zap.Object("role_model", role),
		zap.Int("index", 0),
		zap.String("value", sub),
		zap.Duration("remove_group_policy_duration", removeGroupPolicyDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}
