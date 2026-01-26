package data

import (
	"context"
	goerrors "errors"
	"time"

	"github.com/casbin/casbin/v2"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"gin-artweb/internal/customer/biz"
	"gin-artweb/internal/shared/auth"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/log"
	"gin-artweb/pkg/ctxutil"
)

type roleRepo struct {
	log      *zap.Logger
	gormDB   *gorm.DB
	timeouts *config.DBTimeout
	enforcer *casbin.Enforcer
}

func NewRoleRepo(
	log *zap.Logger,
	gormDB *gorm.DB,
	timeouts *config.DBTimeout,
	enforcer *casbin.Enforcer,
) biz.RoleRepo {
	return &roleRepo{
		log:      log,
		gormDB:   gormDB,
		timeouts: timeouts,
		enforcer: enforcer,
	}
}

func (r *roleRepo) CreateModel(
	ctx context.Context,
	m *biz.RoleModel,
	perms *[]biz.PermissionModel,
	menus *[]biz.MenuModel,
	buttons *[]biz.ButtonModel,
) error {
	// 检查参数
	if m == nil {
		return goerrors.New("创建角色模型失败: 角色模型不能为空")
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
	if perms != nil {
		if len(*perms) > 0 {
			upmap["Permissions"] = *perms
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

	if err := database.DBCreate(ctx, r.gormDB, &biz.RoleModel{}, m, upmap); err != nil {
		r.log.Error(
			"创建角色模型失败",
			zap.Error(err),
			zap.Object(database.ModelKey, m),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return err
	}

	r.log.Debug(
		"创建角色模型成功",
		zap.Object(database.ModelKey, m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(now)),
	)
	return nil
}

func (r *roleRepo) UpdateModel(
	ctx context.Context,
	data map[string]any,
	perms *[]biz.PermissionModel,
	menus *[]biz.MenuModel,
	buttons *[]biz.ButtonModel,
	conds ...any,
) error {
	r.log.Debug(
		"开始更新角色模型",
		zap.Any(database.UpdateDataKey, data),
		zap.Any(database.ConditionKey, conds),
		zap.Uint32s(biz.PermissionIDsKey, biz.ListPermissionModelToUint32s(perms)),
		zap.Uint32s(biz.MenuIDsKey, biz.ListMenuModelToUint32s(menus)),
		zap.Uint32s(biz.ButtonIDsKey, biz.ListButtonModelToUint32s(buttons)),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	now := time.Now()
	upmap := make(map[string]any, 3)
	if perms != nil {
		if len(*perms) > 0 {
			upmap["Permissions"] = *perms
		} else {
			upmap["Permissions"] = []biz.PermissionModel{}
		}
	}
	if menus != nil {
		if len(*menus) > 0 {
			upmap["Menus"] = *menus
		} else {
			upmap["Menus"] = []biz.MenuModel{}
		}
	}
	if buttons != nil {
		if len(*buttons) > 0 {
			upmap["Buttons"] = *buttons
		} else {
			upmap["Buttons"] = []biz.ButtonModel{}
		}
	}
	if err := database.DBUpdate(ctx, r.gormDB, &biz.RoleModel{}, data, upmap, conds...); err != nil {
		r.log.Error(
			"更新角色模型失败",
			zap.Error(err),
			zap.Any(database.UpdateDataKey, data),
			zap.Uint32s(biz.PermissionIDsKey, biz.ListPermissionModelToUint32s(perms)),
			zap.Uint32s(biz.MenuIDsKey, biz.ListMenuModelToUint32s(menus)),
			zap.Uint32s(biz.ButtonIDsKey, biz.ListButtonModelToUint32s(buttons)),
			zap.Any(database.ConditionKey, conds),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return err
	}
	r.log.Debug(
		"更新角色模型成功",
		zap.Any(database.UpdateDataKey, data),
		zap.Uint32s(biz.PermissionIDsKey, biz.ListPermissionModelToUint32s(perms)),
		zap.Uint32s(biz.MenuIDsKey, biz.ListMenuModelToUint32s(menus)),
		zap.Uint32s(biz.ButtonIDsKey, biz.ListButtonModelToUint32s(buttons)),
		zap.Any(database.ConditionKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(now)),
	)
	return nil
}

func (r *roleRepo) DeleteModel(ctx context.Context, conds ...any) error {
	r.log.Debug(
		"开始删除角色模型",
		zap.Any(database.ConditionKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	now := time.Now()
	if err := database.DBDelete(ctx, r.gormDB, &biz.RoleModel{}, conds...); err != nil {
		r.log.Error(
			"删除角色模型失败",
			zap.Error(err),
			zap.Any(database.ConditionKey, conds),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return err
	}
	r.log.Debug(
		"删除角色模型成功",
		zap.Any(database.ConditionKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(now)),
	)
	return nil
}

func (r *roleRepo) FindModel(
	ctx context.Context,
	preloads []string,
	conds ...any,
) (*biz.RoleModel, error) {
	r.log.Debug(
		"开始查询角色模型",
		zap.Strings(database.PreloadKey, preloads),
		zap.Any(database.ConditionKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	now := time.Now()
	var m biz.RoleModel
	if err := database.DBFind(ctx, r.gormDB, preloads, &m, conds...); err != nil {
		r.log.Error(
			"查询角色模型失败",
			zap.Error(err),
			zap.Strings(database.PreloadKey, preloads),
			zap.Any(database.ConditionKey, conds),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return nil, err
	}
	r.log.Debug(
		"查询角色模型成功",
		zap.Object(database.ModelKey, &m),
		zap.Strings(database.PreloadKey, preloads),
		zap.Any(database.ConditionKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(now)),
	)
	return &m, nil
}

func (r *roleRepo) ListModel(
	ctx context.Context,
	qp database.QueryParams,
) (int64, *[]biz.RoleModel, error) {
	r.log.Debug(
		"开始查询角色模型列表",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	now := time.Now()
	var ms []biz.RoleModel
	count, err := database.DBList(ctx, r.gormDB, &biz.RoleModel{}, &ms, qp)
	if err != nil {
		r.log.Error(
			"查询角色列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return 0, nil, err
	}
	r.log.Debug(
		"查询角色模型列表成功",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(now)),
	)
	return count, &ms, nil
}

func (r *roleRepo) AddGroupPolicy(
	ctx context.Context,
	role *biz.RoleModel,
) error {
	// 检查参数
	if role == nil {
		return goerrors.New("AddGroupPolicy操作失败: 角色模型不能为空")
	}

	r.log.Debug(
		"AddGroupPolicy: 传入参数",
		zap.Object(database.ModelKey, role),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m := *role

	// 检查必要字段
	if m.ID == 0 {
		return goerrors.New("AddGroupPolicy操作失败: 角色ID不能为0")
	}

	sub := auth.RoleToSubject(m.ID)

	r.log.Debug(
		"开始添加角色与权限的关联策略",
		zap.Object(database.ModelKey, role),
		zap.Uint32s(biz.PermissionIDsKey, biz.ListPermissionModelToUint32s(&m.Permissions)),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	permStartTime := time.Now()
	// 批量处理权限
	for i, o := range m.Permissions {
		// 检查权限模型的有效性
		if o.ID == 0 {
			r.log.Warn(
				"跳过无效权限",
				zap.Object(database.ModelKey, role),
				zap.Int("permission_index", i),
				zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			)
			continue
		}

		obj := auth.PermissionToSubject(o.ID)
		if err := auth.AddGroupPolicy(ctx, r.enforcer, sub, obj); err != nil {
			r.log.Error(
				"添加角色与权限的关联策略失败",
				zap.Error(err),
				zap.Object(database.ModelKey, role),
				zap.String(auth.GroupSubKey, sub),
				zap.String(auth.GroupObjKey, obj),
				zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
				zap.Duration(log.DurationKey, time.Since(permStartTime)),
			)
			return err
		}
	}
	r.log.Debug(
		"添加角色与权限的关联策略成功",
		zap.Object(database.ModelKey, role),
		zap.Uint32s(biz.PermissionIDsKey, biz.ListPermissionModelToUint32s(&m.Permissions)),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(permStartTime)),
	)

	r.log.Debug(
		"开始添加角色与菜单的关联策略",
		zap.Object(database.ModelKey, role),
		zap.Uint32s(biz.MenuIDsKey, biz.ListMenuModelToUint32s(&m.Menus)),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	menuStartTime := time.Now()
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
		if err := auth.AddGroupPolicy(ctx, r.enforcer, sub, obj); err != nil {
			r.log.Error(
				"添加角色与菜单的关联策略失败",
				zap.Error(err),
				zap.Object(database.ModelKey, role),
				zap.String(auth.GroupSubKey, sub),
				zap.String(auth.GroupObjKey, obj),
				zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
				zap.Duration(log.DurationKey, time.Since(menuStartTime)),
			)
			return err
		}
	}
	r.log.Debug(
		"添加角色与菜单的关联策略成功",
		zap.Object(database.ModelKey, role),
		zap.Uint32s(biz.MenuIDsKey, biz.ListMenuModelToUint32s(&m.Menus)),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(menuStartTime)),
	)

	r.log.Debug(
		"开始添加角色与按钮的关联策略",
		zap.Object(database.ModelKey, role),
		zap.Uint32s(biz.ButtonIDsKey, biz.ListButtonModelToUint32s(&m.Buttons)),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	buttonStartTime := time.Now()
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
		if err := auth.AddGroupPolicy(ctx, r.enforcer, sub, obj); err != nil {
			r.log.Error(
				"添加角色与按钮的关联策略失败",
				zap.Error(err),
				zap.Object(database.ModelKey, role),
				zap.String(auth.GroupSubKey, sub),
				zap.String(auth.GroupObjKey, obj),
				zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
				zap.Duration(log.DurationKey, time.Since(buttonStartTime)),
			)
			return err
		}
	}
	r.log.Debug(
		"添加角色与按钮的关联策略成功",
		zap.Object(database.ModelKey, role),
		zap.Uint32s(biz.ButtonIDsKey, biz.ListButtonModelToUint32s(&m.Buttons)),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(buttonStartTime)),
	)
	return nil
}

func (r *roleRepo) RemoveGroupPolicy(
	ctx context.Context,
	role *biz.RoleModel,
) error {
	// 检查参数
	if role == nil {
		return goerrors.New("RemoveGroupPolicy操作失败: 角色模型不能为空")
	}

	r.log.Debug(
		"RemoveGroupPolicy: 传入参数",
		zap.Object(database.ModelKey, role),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m := *role
	// 检查必要字段
	if m.ID == 0 {
		return goerrors.New("RemoveGroupPolicy操作失败: 角色ID不能为0")
	}

	sub := auth.RoleToSubject(m.ID)
	r.log.Debug(
		"开始删除该角色作为子级的组策略",
		zap.Object(database.ModelKey, role),
		zap.String(auth.GroupSubKey, sub),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	rmSubStartTime := time.Now()

	// 删除该角色作为子级的策略（被其他策略继承）
	if err := auth.RemoveGroupPolicy(ctx, r.enforcer, 0, sub); err != nil {
		r.log.Error(
			"删除角色作为子级策略失败(该策略继承自其他策略)",
			zap.Error(err),
			zap.Object(database.ModelKey, role),
			zap.String(auth.GroupSubKey, sub),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(rmSubStartTime)),
		)
		return err
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
