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

// buttonRepo 按钮仓库实现
// 负责按钮模型的CRUD操作和按钮权限策略的管理
// 使用GORM进行数据库操作，使用Casbin进行权限策略管理
type ButtonRepo struct {
	log      *zap.Logger       // 日志记录器
	gormDB   *gorm.DB          // GORM数据库连接
	timeouts *config.DBTimeout // 数据库操作超时配置
	enforcer *casbin.Enforcer  // Casbin权限管理器
}

// NewButtonRepo 创建按钮仓库实例
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
//	model.ButtonRepo: 按钮仓库接口实现
func NewButtonRepo(
	log *zap.Logger,
	gormDB *gorm.DB,
	timeouts *config.DBTimeout,
	enforcer *casbin.Enforcer,
) *ButtonRepo {
	return &ButtonRepo{
		log:      log,
		gormDB:   gormDB,
		timeouts: timeouts,
		enforcer: enforcer,
	}
}

// CreateModel 创建按钮模型
//
// 参数：
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	m: 按钮模型，包含按钮的详细信息
//	apis: API模型列表，包含与按钮关联的权限
//
// 返回值：
//
//	error: 操作错误信息，成功则返回nil
//
// 功能：
//  1. 检查按钮模型是否为空
//  2. 设置创建时间和更新时间
//  3. 处理关联的权限信息
//  4. 执行数据库创建操作
//  5. 记录操作日志
func (r *ButtonRepo) CreateModel(
	ctx context.Context,
	m *model.ButtonModel,
	apis *[]model.ApiModel,
) error {
	// 检查参数
	if m == nil {
		err := errors.New("创建按钮模型失败: 模型为空")
		r.log.Error(
			"创建按钮模型失败: 模型为空",
			zap.Error(err),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return err
	}

	r.log.Debug(
		"开始创建按钮模型",
		zap.Object(database.ModelKey, m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	now := time.Now()
	m.CreatedAt = now
	m.UpdatedAt = now

	upmap := make(map[string]any, 1)
	if apis != nil {
		if len(*apis) > 0 {
			upmap["Apis"] = *apis
		}
	}

	if err := database.DBCreate(ctx, r.gormDB, &model.ButtonModel{}, m, upmap); err != nil {
		r.log.Error(
			"创建按钮模型失败",
			zap.Error(err),
			zap.Object(database.ModelKey, m),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return errors.WrapIf(err, "创建按钮模型失败")
	}

	r.log.Debug(
		"创建按钮模型成功",
		zap.Object(database.ModelKey, m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(now)),
	)
	return nil
}

// UpdateModel 更新按钮模型
//
// 参数：
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	data: 更新数据，包含要更新的字段和值
//	apis: API模型列表，包含与按钮关联的权限
//	conds: 查询条件，用于指定要更新的记录
//
// 返回值：
//
//	error: 操作错误信息，成功则返回nil
//
// 功能：
//  1. 检查更新数据是否为空
//  2. 处理关联的权限信息
//  3. 执行数据库更新操作
//  4. 记录操作日志
func (r *ButtonRepo) UpdateModel(
	ctx context.Context,
	data map[string]any,
	apis *[]model.ApiModel,
	conds ...any,
) error {
	if len(data) == 0 {
		err := errors.New("更新按钮模型失败: 更新数据为空")
		r.log.Error(
			"更新按钮模型失败: 更新数据为空",
			zap.Error(err),
			zap.Any(database.UpdateDataKey, data),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return err
	}
	r.log.Debug(
		"开始更新按钮模型",
		zap.Any(database.UpdateDataKey, data),
		zap.Any(database.ConditionsKey, conds),
		zap.Uint32s("apis", ListApiModelToUint32s(apis)),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	now := time.Now()
	upmap := make(map[string]any, 1)
	if apis != nil {
		if len(*apis) > 0 {
			upmap["Apis"] = *apis
		} else {
			upmap["Apis"] = []model.ApiModel{}
		}
	}

	if err := database.DBUpdate(ctx, r.gormDB, &model.ButtonModel{}, data, upmap, conds...); err != nil {
		r.log.Error(
			"更新按钮模型失败",
			zap.Error(err),
			zap.Any(database.UpdateDataKey, data),
			zap.Uint32s("apis", ListApiModelToUint32s(apis)),
			zap.Any(database.ConditionsKey, conds),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return errors.WrapIf(err, "更新按钮模型失败")
	}

	r.log.Debug(
		"更新按钮模型成功",
		zap.Any(database.UpdateDataKey, data),
		zap.Uint32s("apis", ListApiModelToUint32s(apis)),
		zap.Any(database.ConditionsKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(now)),
	)
	return nil
}

// DeleteModel 删除按钮模型
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
func (r *ButtonRepo) DeleteModel(ctx context.Context, conds ...any) error {
	r.log.Debug(
		"开始删除按钮模型",
		zap.Any(database.ConditionsKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	now := time.Now()
	if err := database.DBDelete(ctx, r.gormDB, &model.ButtonModel{}, conds...); err != nil {
		r.log.Error(
			"删除按钮模型失败",
			zap.Error(err),
			zap.Any(database.ConditionsKey, conds),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return errors.WrapIf(err, "删除按钮模型失败")
	}

	r.log.Debug(
		"删除按钮模型成功",
		zap.Any(database.ConditionsKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(now)),
	)
	return nil
}

// GetModel 获取单个按钮模型
//
// 参数：
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	preloads: 预加载的关联字段列表
//	conds: 查询条件，用于指定要获取的记录
//
// 返回值：
//
//	*model.ButtonModel: 按钮模型指针，包含按钮的详细信息
//	error: 操作错误信息，成功则返回nil
//
// 功能：
//  1. 执行数据库查询操作
//  2. 预加载关联字段
//  3. 获取单个按钮模型
//  4. 记录操作日志
func (r *ButtonRepo) GetModel(
	ctx context.Context,
	preloads []string,
	conds ...any,
) (*model.ButtonModel, error) {
	r.log.Debug(
		"开始查询按钮模型",
		zap.Strings(database.PreloadKey, preloads),
		zap.Any(database.ConditionsKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	now := time.Now()
	var m model.ButtonModel
	if err := database.DBGet(ctx, r.gormDB, preloads, &m, conds...); err != nil {
		r.log.Error(
			"查询按钮模型失败",
			zap.Error(err),
			zap.Strings(database.PreloadKey, preloads),
			zap.Any(database.ConditionsKey, conds),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return nil, errors.WrapIf(err, "查询按钮模型失败")
	}

	r.log.Debug(
		"查询按钮模型成功",
		zap.Object(database.ModelKey, &m),
		zap.Strings(database.PreloadKey, preloads),
		zap.Any(database.ConditionsKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(now)),
	)
	return &m, nil
}

// ListModel 获取按钮模型列表
//
// 参数：
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	qp: 查询参数，包含分页、排序等查询条件
//
// 返回值：
//
//	int64: 总记录数
//	*[]model.ButtonModel: 按钮模型列表指针，包含符合条件的按钮模型
//	error: 操作错误信息，成功则返回nil
//
// 功能：
//  1. 执行数据库查询操作
//  2. 获取按钮模型列表
//  3. 返回总记录数和模型列表
//  4. 记录操作日志
func (r *ButtonRepo) ListModel(
	ctx context.Context,
	qp database.QueryParams,
) (int64, *[]model.ButtonModel, error) {
	r.log.Debug(
		"开始查询按钮模型列表",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	now := time.Now()
	var ms []model.ButtonModel
	count, err := database.DBList(ctx, r.gormDB, &model.ButtonModel{}, &ms, qp)
	if err != nil {
		r.log.Error(
			"查询按钮模型列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return 0, nil, errors.WrapIf(err, "查询按钮模型列表失败")
	}

	r.log.Debug(
		"查询按钮模型列表成功",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(now)),
	)
	return count, &ms, nil
}

// AddGroupPolicy 添加按钮组策略
//
// 参数：
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	button: 按钮模型，包含按钮的详细信息
//
// 返回值：
//
//	error: 操作错误信息，成功则返回nil
//
// 功能：
//  1. 检查上下文是否有效
//  2. 检查按钮模型是否为空
//  3. 检查按钮ID和菜单ID是否有效
//  4. 将按钮转换为Casbin策略
//  5. 添加按钮策略到Casbin
//  6. 记录操作日志
func (r *ButtonRepo) AddGroupPolicy(
	ctx context.Context,
	button *model.ButtonModel,
) error {
	// 检查上下文
	if ctx.Err() != nil {
		return errors.WrapIf(ctx.Err(), "AddGroupPolicy操作失败: 上下文错误")
	}

	// 检查参数
	if button == nil {
		return errors.New("AddGroupPolicy操作失败: 按钮模型不能为空")
	}

	m := *button

	// 检查必要字段
	if m.ID == 0 {
		return errors.New("AddGroupPolicy操作失败: 按钮ID不能为0")
	}
	if m.MenuID == 0 {
		return errors.New("AddGroupPolicy操作失败: 菜单ID不能为0")
	}

	r.log.Debug(
		"开始添加按钮关联策略",
		zap.Object(database.ModelKey, button),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	now := time.Now()
	sub := auth.ButtonToSubject(m.ID)
	menuObj := auth.MenuToSubject(m.MenuID)
	rules := [][]string{{sub, menuObj}}
	for i, o := range m.Apis {
		// 检查API模型的有效性
		if o.ID == 0 {
			r.log.Warn(
				"跳过无效API",
				zap.Object(database.ModelKey, button),
				zap.Int("Api_index", i),
				zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			)
			continue
		}

		obj := auth.ApiToSubject(o.ID)
		rules = append(rules, []string{sub, obj})
	}
	if err := auth.AddGroupPolicies(ctx, r.enforcer, rules); err != nil {
		r.log.Error(
			"添加按钮关联策略失败",
			zap.Error(err),
			zap.Object(database.ModelKey, button),
			zap.String(auth.GroupSubKey, sub),
			zap.String(auth.GroupObjKey, menuObj),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return errors.WrapIf(err, "添加按钮关联策略失败")
	}
	r.log.Debug(
		"添加按钮关联策略成功",
		zap.Object(database.ModelKey, button),
		zap.String(auth.GroupSubKey, sub),
		zap.String(auth.GroupObjKey, menuObj),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(now)),
	)
	return nil
}

// RemoveGroupPolicy 删除按钮组策略
//
// 参数：
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	button: 按钮模型，包含按钮的详细信息
//	removeInherited: 是否删除继承该按钮的组策略
//
// 返回值：
//
//	error: 操作错误信息，成功则返回nil
//
// 功能：
//  1. 检查上下文是否有效
//  2. 检查按钮模型是否为空
//  3. 检查按钮ID是否有效
//  4. 删除该按钮作为子级的组策略（被其他策略继承）
//  5. 可选：删除该按钮作为父级的组策略（被其他菜单或权限继承）
//  6. 记录操作日志
func (r *ButtonRepo) RemoveGroupPolicy(
	ctx context.Context,
	button *model.ButtonModel,
	removeInherited bool,
) error {
	// 检查上下文
	if ctx.Err() != nil {
		return errors.WrapIf(ctx.Err(), "RemoveGroupPolicy操作失败: 上下文错误")
	}

	// 检查参数
	if button == nil {
		return errors.New("RemoveGroupPolicy操作失败: 按钮模型不能为空")
	}

	m := *button

	// 检查必要字段
	if m.ID == 0 {
		return errors.New("RemoveGroupPolicy操作失败: 按钮ID不能为0")
	}

	r.log.Debug(
		"开始删除该按钮作为子级的组策略",
		zap.Object(database.ModelKey, button),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	rmSubStartTime := time.Now()
	sub := auth.ButtonToSubject(m.ID)
	// 删除该按钮作为子级的策略（被其他策略继承）
	if err := auth.RemoveFilteredGroupingPolicy(ctx, r.enforcer, 0, sub); err != nil {
		r.log.Error(
			"删除按钮作为子级策略失败(该策略继承自其他策略)",
			zap.Error(err),
			zap.Object(database.ModelKey, button),
			zap.String(auth.GroupSubKey, sub),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(rmSubStartTime)),
		)
		return errors.WrapIf(err, "删除按钮作为子级策略失败(该策略继承自其他策略)")
	}
	r.log.Debug(
		"删除该按钮作为子级的组策略成功",
		zap.Object(database.ModelKey, button),
		zap.String(auth.GroupSubKey, sub),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(rmSubStartTime)),
	)

	// 删除该按钮作为父级的策略（被其他菜单或权限继承）
	if removeInherited {
		r.log.Debug(
			"开始删除该按钮作为父级的组策略",
			zap.Object(database.ModelKey, button),
			zap.String(auth.GroupObjKey, sub),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rmObjStartTime := time.Now()
		if err := auth.RemoveFilteredGroupingPolicy(ctx, r.enforcer, 1, sub); err != nil {
			r.log.Error(
				"删除按钮作为父级策略失败(该策略被其他策略继承)",
				zap.Error(err),
				zap.Object(database.ModelKey, button),
				zap.String(auth.GroupObjKey, sub),
				zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
				zap.Duration(log.DurationKey, time.Since(rmObjStartTime)),
			)
			return errors.WrapIf(err, "删除按钮作为父级策略失败(该策略被其他策略继承)")
		}
		r.log.Debug(
			"删除该按钮作为父级的组策略成功",
			zap.Object(database.ModelKey, button),
			zap.String(auth.GroupObjKey, sub),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(rmObjStartTime)),
		)
	}
	return nil
}

func ListButtonModelToUint32s(bms *[]model.ButtonModel) []uint32 {
	if bms == nil {
		return []uint32{}
	}
	ms := *bms
	if len(ms) == 0 {
		return []uint32{}
	}

	ids := make([]uint32, len(ms))
	for i, m := range ms {
		ids[i] = m.ID
	}
	return ids
}
