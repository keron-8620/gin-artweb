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

const (
	ButtonSubjectFormat = "button_%d"
)

type buttonRepo struct {
	log      *zap.Logger
	gormDB   *gorm.DB
	timeouts *config.DBTimeout
	enforcer *casbin.Enforcer
}

func NewButtonRepo(
	log *zap.Logger,
	gormDB *gorm.DB,
	timeouts *config.DBTimeout,
	enforcer *casbin.Enforcer,
) biz.ButtonRepo {
	return &buttonRepo{
		log:      log,
		gormDB:   gormDB,
		timeouts: timeouts,
		enforcer: enforcer,
	}
}

func (r *buttonRepo) CreateModel(
	ctx context.Context,
	m *biz.ButtonModel,
	perms *[]biz.PermissionModel,
) error {
	// 检查参数
	if m == nil {
		return goerrors.New("创建按钮模型失败: 按钮模型不能为空")
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
	if perms != nil {
		if len(*perms) > 0 {
			upmap["Permissions"] = *perms
		}
	}

	if err := database.DBCreate(ctx, r.gormDB, &biz.ButtonModel{}, m, upmap); err != nil {
		r.log.Error(
			"创建按钮模型失败",
			zap.Error(err),
			zap.Object(database.ModelKey, m),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return err
	}

	r.log.Debug(
		"创建按钮模型成功",
		zap.Object(database.ModelKey, m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(now)),
	)
	return nil
}

func (r *buttonRepo) UpdateModel(
	ctx context.Context,
	data map[string]any,
	perms *[]biz.PermissionModel,
	conds ...any,
) error {
	r.log.Debug(
		"开始更新按钮模型",
		zap.Any(database.UpdateDataKey, data),
		zap.Any(database.ConditionKey, conds),
		zap.Uint32s(biz.PermissionIDsKey, biz.ListPermissionModelToUint32s(perms)),
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

	if err := database.DBUpdate(ctx, r.gormDB, &biz.ButtonModel{}, data, upmap, conds...); err != nil {
		r.log.Error(
			"更新按钮模型失败",
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
		"更新按钮模型成功",
		zap.Any(database.UpdateDataKey, data),
		zap.Uint32s(biz.PermissionIDsKey, biz.ListPermissionModelToUint32s(perms)),
		zap.Any(database.ConditionKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(now)),
	)
	return nil
}

func (r *buttonRepo) DeleteModel(ctx context.Context, conds ...any) error {
	r.log.Debug(
		"开始删除按钮模型",
		zap.Any(database.ConditionKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	now := time.Now()
	if err := database.DBDelete(ctx, r.gormDB, &biz.ButtonModel{}, conds...); err != nil {
		r.log.Error(
			"删除按钮模型失败",
			zap.Error(err),
			zap.Any(database.ConditionKey, conds),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return err
	}

	r.log.Debug(
		"删除按钮模型成功",
		zap.Any(database.ConditionKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(now)),
	)
	return nil
}

func (r *buttonRepo) FindModel(
	ctx context.Context,
	preloads []string,
	conds ...any,
) (*biz.ButtonModel, error) {
	r.log.Debug(
		"开始查询按钮模型",
		zap.Strings(database.PreloadKey, preloads),
		zap.Any(database.ConditionKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	now := time.Now()
	var m biz.ButtonModel
	if err := database.DBFind(ctx, r.gormDB, preloads, &m, conds...); err != nil {
		r.log.Error(
			"查询按钮模型失败",
			zap.Error(err),
			zap.Strings(database.PreloadKey, preloads),
			zap.Any(database.ConditionKey, conds),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return nil, err
	}

	r.log.Debug(
		"查询按钮模型成功",
		zap.Object(database.ModelKey, &m),
		zap.Strings(database.PreloadKey, preloads),
		zap.Any(database.ConditionKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(now)),
	)
	return &m, nil
}

func (r *buttonRepo) ListModel(
	ctx context.Context,
	qp database.QueryParams,
) (int64, *[]biz.ButtonModel, error) {
	r.log.Debug(
		"开始查询按钮模型列表",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	now := time.Now()
	var ms []biz.ButtonModel
	count, err := database.DBList(ctx, r.gormDB, &biz.ButtonModel{}, &ms, qp)
	if err != nil {
		r.log.Error(
			"查询按钮列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return 0, nil, err
	}

	r.log.Debug(
		"查询按钮模型列表成功",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(now)),
	)
	return count, &ms, nil
}

func (r *buttonRepo) AddGroupPolicy(
	ctx context.Context,
	button *biz.ButtonModel,
) error {
	// 检查上下文
	if err := ctxutil.CheckContext(ctx); err != nil {
		return err
	}

	// 检查参数
	if button == nil {
		return goerrors.New("AddGroupPolicy操作失败: 按钮模型不能为空")
	}

	r.log.Debug(
		"AddGroupPolicy: 传入参数",
		zap.Object(database.ModelKey, button),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m := *button

	// 检查必要字段
	if m.ID == 0 {
		return goerrors.New("AddGroupPolicy操作失败: 按钮ID不能为0")
	}
	if m.MenuID == 0 {
		return goerrors.New("AddGroupPolicy操作失败: 菜单ID不能为0")
	}

	sub := auth.ButtonToSubject(m.ID)
	menuObj := auth.MenuToSubject(m.MenuID)
	r.log.Debug(
		"开始添加按钮与父级菜单的继承关系策略",
		zap.Object(database.ModelKey, button),
		zap.String(auth.GroupSubKey, sub),
		zap.String(auth.GroupObjKey, menuObj),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	menuStartTime := time.Now()
	if err := auth.AddGroupPolicy(ctx, r.enforcer, sub, menuObj); err != nil {
		r.log.Error(
			"添加按钮与父级菜单的继承关系策略失败",
			zap.Error(err),
			zap.Object(database.ModelKey, button),
			zap.String(auth.GroupSubKey, sub),
			zap.String(auth.GroupObjKey, menuObj),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(menuStartTime)),
		)
		return err
	}
	r.log.Debug(
		"添加按钮与父级菜单的继承关系策略成功",
		zap.Object(database.ModelKey, button),
		zap.String(auth.GroupSubKey, sub),
		zap.String(auth.GroupObjKey, menuObj),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(menuStartTime)),
	)

	r.log.Debug(
		"开始添加按钮与权限的关联策略",
		zap.Object(database.ModelKey, button),
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
				zap.Object(database.ModelKey, button),
				zap.Int("permission_index", i),
				zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			)
			continue
		}

		obj := auth.PermissionToSubject(o.ID)
		if err := auth.AddGroupPolicy(ctx, r.enforcer, sub, obj); err != nil {
			r.log.Error(
				"添加按钮与权限的关联策略失败",
				zap.Error(err),
				zap.Object(database.ModelKey, button),
				zap.String(auth.GroupSubKey, sub),
				zap.String(auth.GroupObjKey, obj),
				zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
				zap.Duration(log.DurationKey, time.Since(permStartTime)),
			)
			return err
		}
	}
	r.log.Debug(
		"添加按钮与权限的关联策略成功",
		zap.Object(database.ModelKey, button),
		zap.Uint32s(biz.PermissionIDsKey, biz.ListPermissionModelToUint32s(&m.Permissions)),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(permStartTime)),
	)
	return nil
}

func (r *buttonRepo) RemoveGroupPolicy(
	ctx context.Context,
	button *biz.ButtonModel,
	removeInherited bool,
) error {
	// 检查上下文
	if err := ctxutil.CheckContext(ctx); err != nil {
		return err
	}

	// 检查参数
	if button == nil {
		return goerrors.New("RemoveGroupPolicy操作失败: 按钮模型不能为空")
	}

	r.log.Debug(
		"RemoveGroupPolicy: 传入参数",
		zap.Object(database.ModelKey, button),
		zap.Bool("removeInherited", removeInherited),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m := *button

	// 检查必要字段
	if m.ID == 0 {
		return goerrors.New("RemoveGroupPolicy操作失败: 按钮ID不能为0")
	}

	sub := auth.ButtonToSubject(m.ID)
	r.log.Debug(
		"开始删除该按钮作为子级的组策略",
		zap.Object(database.ModelKey, button),
		zap.String(auth.GroupSubKey, sub),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	rmSubStartTime := time.Now()

	// 删除该按钮作为子级的策略（被其他策略继承）
	if err := auth.RemoveGroupPolicy(ctx, r.enforcer, 0, sub); err != nil {
		r.log.Error(
			"删除按钮作为子级策略失败(该策略继承自其他策略)",
			zap.Error(err),
			zap.Object(database.ModelKey, button),
			zap.String(auth.GroupSubKey, sub),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(rmSubStartTime)),
		)
		return err
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
		if err := auth.RemoveGroupPolicy(ctx, r.enforcer, 1, sub); err != nil {
			r.log.Error(
				"删除按钮作为父级策略失败(该策略被其他策略继承)",
				zap.Error(err),
				zap.Object(database.ModelKey, button),
				zap.String(auth.GroupObjKey, sub),
				zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
				zap.Duration(log.DurationKey, time.Since(rmObjStartTime)),
			)
			return err
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
