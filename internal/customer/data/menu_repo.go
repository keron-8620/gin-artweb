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

type menuRepo struct {
	log      *zap.Logger
	gormDB   *gorm.DB
	timeouts *config.DBTimeout
	enforcer *casbin.Enforcer
}

func NewMenuRepo(
	log *zap.Logger,
	gormDB *gorm.DB,
	timeouts *config.DBTimeout,
	enforcer *casbin.Enforcer,
) biz.MenuRepo {
	return &menuRepo{
		log:      log,
		gormDB:   gormDB,
		timeouts: timeouts,
		enforcer: enforcer,
	}
}

func (r *menuRepo) CreateModel(
	ctx context.Context,
	m *biz.MenuModel,
	perms *[]biz.PermissionModel,
) error {
	// 检查参数
	if m == nil {
		return goerrors.New("创建菜单模型失败: 菜单模型不能为空")
	}

	r.log.Debug(
		"开始创建菜单模型",
		zap.Object(database.ModelKey, m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	now := time.Now()
	m.CreatedAt = now
	m.UpdatedAt = now

	upmap := make(map[string]any, 1)
	if perms != nil {
		if len(*perms) > 0 {
			upmap["Permissions"] = *perms
		} else {
			upmap["Permissions"] = []biz.PermissionModel{}
		}
	}

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ListTimeout)
	defer cancel()
	if err := database.DBCreate(dbCtx, r.gormDB, &biz.MenuModel{}, m, upmap); err != nil {
		r.log.Error(
			"创建菜单模型失败",
			zap.Error(err),
			zap.Object(database.ModelKey, m),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return err
	}

	r.log.Debug(
		"创建菜单模型成功",
		zap.Object(database.ModelKey, m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(now)),
	)
	return nil
}

func (r *menuRepo) UpdateModel(
	ctx context.Context,
	data map[string]any,
	perms *[]biz.PermissionModel,
	conds ...any,
) error {
	r.log.Debug(
		"开始更新菜单模型",
		zap.Any(database.UpdateDataKey, data),
		zap.Uint32s(biz.PermissionIDsKey, biz.ListPermissionModelToUint32s(perms)),
		zap.Any(database.ConditionKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	now := time.Now()
	upmap := make(map[string]any, 1)
	if perms != nil {
		if len(*perms) > 0 {
			upmap["Permissions"] = *perms
		} else {
			upmap["Permissions"] = []biz.PermissionModel{}
		}
	}

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ListTimeout)
	defer cancel()
	if err := database.DBUpdate(dbCtx, r.gormDB, &biz.MenuModel{}, data, upmap, conds...); err != nil {
		r.log.Error(
			"更新菜单模型失败",
			zap.Error(err),
			zap.Any(database.UpdateDataKey, data),
			zap.Uint32s(biz.PermissionIDsKey, biz.ListPermissionModelToUint32s(perms)),
			zap.Any(database.ConditionKey, conds),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return err
	}

	r.log.Debug(
		"更新菜单模型成功",
		zap.Any(database.UpdateDataKey, data),
		zap.Any(database.ConditionKey, conds),
		zap.Uint32s(biz.PermissionIDsKey, biz.ListPermissionModelToUint32s(perms)),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(now)),
	)
	return nil
}

func (r *menuRepo) DeleteModel(ctx context.Context, conds ...any) error {
	r.log.Debug(
		"开始删除菜单模型",
		zap.Any(database.ConditionKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	now := time.Now()
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ListTimeout)
	defer cancel()
	if err := database.DBDelete(dbCtx, r.gormDB, &biz.MenuModel{}, conds...); err != nil {
		r.log.Error(
			"删除菜单模型失败",
			zap.Error(err),
			zap.Any(database.ConditionKey, conds),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return err
	}

	r.log.Debug(
		"删除菜单模型成功",
		zap.Any(database.ConditionKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(now)),
	)
	return nil
}

func (r *menuRepo) FindModel(
	ctx context.Context,
	preloads []string,
	conds ...any,
) (*biz.MenuModel, error) {
	r.log.Debug(
		"开始查询菜单模型",
		zap.Strings(database.PreloadKey, preloads),
		zap.Any(database.ConditionKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	now := time.Now()
	var m biz.MenuModel
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ListTimeout)
	defer cancel()
	if err := database.DBFind(dbCtx, r.gormDB, preloads, &m, conds...); err != nil {
		r.log.Error(
			"查询菜单模型失败",
			zap.Error(err),
			zap.Strings(database.PreloadKey, preloads),
			zap.Any(database.ConditionKey, conds),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return nil, err
	}

	r.log.Debug(
		"查询菜单模型成功",
		zap.Object(database.ModelKey, &m),
		zap.Strings(database.PreloadKey, preloads),
		zap.Any(database.ConditionKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(now)),
	)
	return &m, nil
}

func (r *menuRepo) ListModel(
	ctx context.Context,
	qp database.QueryParams,
) (int64, *[]biz.MenuModel, error) {
	r.log.Debug(
		"开始查询菜单模型列表",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	now := time.Now()
	var ms []biz.MenuModel
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ListTimeout)
	defer cancel()
	count, err := database.DBList(dbCtx, r.gormDB, &biz.MenuModel{}, &ms, qp)
	if err != nil {
		r.log.Error(
			"查询菜单模型列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return 0, nil, err
	}

	r.log.Debug(
		"查询菜单模型列表成功",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(now)),
	)
	return count, &ms, nil
}

func (r *menuRepo) AddGroupPolicy(
	ctx context.Context,
	menu *biz.MenuModel,
) error {
	// 检查上下文
	if err := ctxutil.CheckContext(ctx); err != nil {
		return err
	}

	// 检查参数
	if menu == nil {
		return goerrors.New("AddGroupPolicy操作失败: 菜单模型不能为空")
	}

	r.log.Debug(
		"AddGroupPolicy: 传入参数",
		zap.Object(database.ModelKey, menu),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m := *menu
	// 检查必要字段
	if m.ID == 0 {
		return goerrors.New("AddGroupPolicy操作失败: 菜单ID不能为0")
	}

	sub := auth.MenuToSubject(m.ID)

	// 处理父级关系
	if m.ParentID != nil {
		obj := auth.MenuToSubject(*m.ParentID)
		r.log.Debug(
			"开始添加菜单与父级菜单的继承关系策略",
			zap.Object(database.ModelKey, menu),
			zap.String(auth.GroupSubKey, sub),
			zap.String(auth.GroupObjKey, obj),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		menuStartTime := time.Now()
		if err := auth.AddGroupPolicy(ctx, r.enforcer, sub, obj); err != nil {
			r.log.Error(
				"添加菜单与父级菜单的继承关系策略失败",
				zap.Error(err),
				zap.Object(database.ModelKey, menu),
				zap.String(auth.GroupSubKey, sub),
				zap.String(auth.GroupObjKey, obj),
				zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
				zap.Duration(log.DurationKey, time.Since(menuStartTime)),
			)
			return err
		}
		r.log.Debug(
			"添加菜单与父级菜单的继承关系策略成功",
			zap.Object(database.ModelKey, menu),
			zap.String(auth.GroupSubKey, sub),
			zap.String(auth.GroupObjKey, obj),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(menuStartTime)),
		)
	}

	r.log.Debug(
		"开始添加菜单与权限的关联策略",
		zap.Object(database.ModelKey, menu),
		zap.Uint32s(biz.PermissionIDsKey, biz.ListPermissionModelToUint32s(&m.Permissions)),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	permStartTime := time.Now()
	// 批量处理权限
	for _, o := range m.Permissions {
		// 检查权限模型的有效性
		if o.ID == 0 {
			r.log.Warn(
				"跳过无效权限",
				zap.Object(database.ModelKey, menu),
				zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			)
			continue
		}

		obj := auth.PermissionToSubject(o.ID)
		if err := auth.AddGroupPolicy(ctx, r.enforcer, sub, obj); err != nil {
			r.log.Error(
				"添加菜单与权限的关联策略失败",
				zap.Error(err),
				zap.Object(database.ModelKey, menu),
				zap.String(auth.GroupSubKey, sub),
				zap.String(auth.GroupObjKey, obj),
				zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
				zap.Duration(log.DurationKey, time.Since(permStartTime)),
			)
			return err
		}
	}
	r.log.Debug(
		"添加菜单与权限的关联策略成功",
		zap.Object(database.ModelKey, menu),
		zap.Uint32s(biz.PermissionIDsKey, biz.ListPermissionModelToUint32s(&m.Permissions)),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(permStartTime)),
	)
	return nil
}

func (r *menuRepo) RemoveGroupPolicy(
	ctx context.Context,
	menu *biz.MenuModel,
	removeInherited bool,
) error {
	// 检查上下文
	if err := ctxutil.CheckContext(ctx); err != nil {
		return err
	}

	// 检查参数
	if menu == nil {
		return goerrors.New("RemoveGroupPolicy操作失败: 菜单模型不能为空")
	}

	r.log.Debug(
		"RemoveGroupPolicy: 传入参数",
		zap.Object(database.ModelKey, menu),
		zap.Bool("removeInherited", removeInherited),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m := *menu
	// 检查必要字段
	if m.ID == 0 {
		return goerrors.New("RemoveGroupPolicy操作失败: 菜单ID不能为0")
	}

	sub := auth.MenuToSubject(m.ID)
	r.log.Debug(
		"开始删除该菜单作为子级的组策略",
		zap.Object(database.ModelKey, &m),
		zap.String(auth.GroupSubKey, sub),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	rmSubStartTime := time.Now()
	if err := auth.RemoveGroupPolicy(ctx, r.enforcer, 0, sub); err != nil {
		r.log.Error(
			"删除该菜单作为子级的组策略失败",
			zap.Error(err),
			zap.Object(database.ModelKey, menu),
			zap.String(auth.GroupSubKey, sub),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(rmSubStartTime)),
		)
		return err
	}
	r.log.Debug(
		"删除该菜单作为子级的组策略成功",
		zap.Object(database.ModelKey, menu),
		zap.String(auth.GroupSubKey, sub),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(rmSubStartTime)),
	)

	if removeInherited {
		r.log.Debug(
			"开始删除该菜单作为父级的组策略",
			zap.Object(database.ModelKey, menu),
			zap.String(auth.GroupObjKey, sub),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		rmObjStartTime := time.Now()
		if err := auth.RemoveGroupPolicy(ctx, r.enforcer, 1, sub); err != nil {
			r.log.Error(
				"删除该菜单作为父级的组策略失败",
				zap.Error(err),
				zap.Object(database.ModelKey, menu),
				zap.String(auth.GroupObjKey, sub),
				zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
				zap.Duration(log.DurationKey, time.Since(rmObjStartTime)),
			)
			return err
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
