package data

import (
	"context"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"gin-artweb/internal/customer/biz"
	"gin-artweb/pkg/auth"
	"gin-artweb/pkg/database"
	"gin-artweb/pkg/errors"
	"gin-artweb/pkg/log"
)

const (
	ButtonSubjectFormat = "button_%d"
)

type buttonRepo struct {
	log    *zap.Logger
	gormDB *gorm.DB
	cache  *auth.AuthEnforcer
}

func NewButtonRepo(
	log *zap.Logger,
	gormDB *gorm.DB,
	cache *auth.AuthEnforcer,
) biz.ButtonRepo {
	return &buttonRepo{
		log:    log,
		gormDB: gormDB,
		cache:  cache,
	}
}

func (r *buttonRepo) CreateModel(ctx context.Context, m *biz.ButtonModel) error {
	r.log.Debug(
		"开始创建按钮模型",
		zap.Object(database.ModelKey, m),
	)
	now := time.Now()
	m.CreatedAt = now
	m.UpdatedAt = now
	if err := database.DBCreate(ctx, r.gormDB, &biz.ButtonModel{}, m); err != nil {
		r.log.Error(
			"创建按钮模型失败",
			zap.Error(err),
			zap.Object(database.ModelKey, m),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return err
	}
	r.log.Debug(
		"创建按钮模型成功",
		zap.Object(database.ModelKey, m),
		zap.Duration(log.DurationKey, time.Since(now)),
	)
	return nil
}

func (r *buttonRepo) UpdateModel(
	ctx context.Context,
	data map[string]any,
	perms []biz.PermissionModel,
	conds ...any,
) error {
	r.log.Debug(
		"开始更新按钮模型",
		zap.Any(database.UpdateDataKey, data),
		zap.Any(database.ConditionKey, conds),
		zap.Uint32s(biz.PermissionIDsKey, biz.ListPermissionModelToUint32s(perms)),
	)
	now := time.Now()
	upmap := make(map[string]any, 1)
	if len(perms) > 0 {
		upmap["Permissions"] = perms
	}
	if err := database.DBUpdate(ctx, r.gormDB, &biz.ButtonModel{}, data, upmap, conds...); err != nil {
		r.log.Error(
			"更新按钮模型失败",
			zap.Error(err),
			zap.Any(database.UpdateDataKey, data),
			zap.Uint32s(biz.PermissionIDsKey, biz.ListPermissionModelToUint32s(perms)),
			zap.Any(database.ConditionKey, conds),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return err
	}
	r.log.Debug(
		"更新按钮模型成功",
		zap.Any(database.UpdateDataKey, data),
		zap.Any(database.ConditionKey, conds),
		zap.Uint32s(biz.PermissionIDsKey, biz.ListPermissionModelToUint32s(perms)),
		zap.Duration(log.DurationKey, time.Since(now)),
	)
	return nil
}

func (r *buttonRepo) DeleteModel(ctx context.Context, conds ...any) error {
	r.log.Debug("开始删除按钮模型", zap.Any(database.ConditionKey, conds))
	now := time.Now()
	if err := database.DBDelete(ctx, r.gormDB, &biz.ButtonModel{}, conds...); err != nil {
		r.log.Error(
			"删除按钮模型失败",
			zap.Error(err),
			zap.Any(database.ConditionKey, conds),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return err
	}
	r.log.Debug(
		"删除按钮模型成功",
		zap.Any(database.ConditionKey, conds),
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
	)
	now := time.Now()
	var m biz.ButtonModel
	if err := database.DBFind(ctx, r.gormDB, preloads, &m, conds...); err != nil {
		r.log.Error(
			"查询按钮模型失败",
			zap.Error(err),
			zap.Strings(database.PreloadKey, preloads),
			zap.Any(database.ConditionKey, conds),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return nil, err
	}
	r.log.Debug(
		"查询按钮模型成功",
		zap.Object(database.ModelKey, &m),
		zap.Strings(database.PreloadKey, preloads),
		zap.Any(database.ConditionKey, conds),
		zap.Duration(log.DurationKey, time.Since(now)),
	)
	return &m, nil
}

func (r *buttonRepo) ListModel(
	ctx context.Context,
	qp database.QueryParams,
) (int64, []biz.ButtonModel, error) {
	r.log.Debug(
		"开始查询按钮模型列表",
		zap.Object(database.QueryParamsKey, &qp),
	)
	now := time.Now()
	var ms []biz.ButtonModel
	count, err := database.DBList(ctx, r.gormDB, &biz.ButtonModel{}, &ms, qp)
	if err != nil {
		r.log.Error(
			"查询按钮列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return 0, nil, err
	}
	r.log.Debug(
		"查询按钮模型列表成功",
		zap.Object(database.QueryParamsKey, &qp),
		zap.Duration(log.DurationKey, time.Since(now)),
	)
	return count, ms, nil
}

func (r *buttonRepo) AddGroupPolicy(
	ctx context.Context,
	m biz.ButtonModel,
) error {
	if err := errors.CheckContext(ctx); err != nil {
		return err
	}
	sub := auth.ButtonToSubject(m.ID)
	menuObj := auth.MenuToSubject(m.MenuID)
	r.log.Debug(
		"开始添加按钮与父级菜单的继承关系策略",
		zap.String(auth.GroupSubKey, sub),
		zap.String(auth.GroupObjKey, menuObj),
	)
	if err := r.cache.AddGroupPolicy(sub, menuObj); err != nil {
		r.log.Error(
			"添加按钮与父级菜单的继承关系策略失败",
			zap.Error(err),
			zap.String(auth.GroupSubKey, sub),
			zap.String(auth.GroupObjKey, menuObj),
		)
		return err
	}

	// 批量处理权限
	for _, o := range m.Permissions {
		obj := auth.PermissionToSubject(o.ID)
		if err := r.cache.AddGroupPolicy(sub, obj); err != nil {
			r.log.Error(
				"添加按钮与菜单的继承关系策略失败",
				zap.String(auth.GroupSubKey, sub),
				zap.String(auth.GroupObjKey, obj),
				zap.Uint32("menu_id", m.ID),
				zap.Uint32("permission_id", o.ID),
				zap.Error(err),
			)
			return err
		}
	}
	return nil
}

func (r *buttonRepo) RemoveGroupPolicy(
	ctx context.Context,
	m biz.ButtonModel,
	removeInherited bool,
) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	sub := auth.ButtonToSubject(m.ID)

	// 删除该按钮作为父级的策略（被其他菜单或权限继承）
	if err := r.cache.RemoveGroupPolicy(0, sub); err != nil {
		r.log.Error(
			"删除按钮作为子级策略失败(该策略继承自其他策略)",
			zap.String(auth.GroupSubKey, sub),
			zap.Error(err),
		)
		return err
	}
	// 删除该按钮作为子级的策略（从其他菜单或权限继承）
	if removeInherited {
		if err := r.cache.RemoveGroupPolicy(1, sub); err != nil {
			r.log.Error(
				"删除按钮作为父级策略失败(该策略被其他策略继承)",
				zap.String(auth.GroupObjKey, sub),
				zap.Error(err),
			)
			return err
		}
	}
	return nil
}
