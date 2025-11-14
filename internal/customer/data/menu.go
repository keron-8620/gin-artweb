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

type menuRepo struct {
	log    *zap.Logger
	gormDB *gorm.DB
	cache  *auth.AuthEnforcer
}

func NewMenuRepo(
	log *zap.Logger,
	gormDB *gorm.DB,
	cache *auth.AuthEnforcer,
) biz.MenuRepo {
	return &menuRepo{
		log:    log,
		gormDB: gormDB,
		cache:  cache,
	}
}

func (r *menuRepo) CreateModel(ctx context.Context, m *biz.MenuModel) error {
	r.log.Debug(
		"开始创建菜单模型",
		zap.Object(database.ModelKey, m),
	)
	now := time.Now()
	m.CreatedAt = now
	m.UpdatedAt = now
	if err := database.DBCreate(ctx, r.gormDB, &biz.MenuModel{}, m); err != nil {
		r.log.Error(
			"创建菜单模型失败",
			zap.Error(err),
			zap.Object(database.ModelKey, m),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return err
	}
	r.log.Debug(
		"创建菜单模型成功",
		zap.Object(database.ModelKey, m),
		zap.Duration(log.DurationKey, time.Since(now)),
	)
	return nil
}

func (r *menuRepo) UpdateModel(
	ctx context.Context,
	data map[string]any,
	perms []*biz.PermissionModel,
	conds ...any,
) error {
	r.log.Debug(
		"开始更新菜单模型",
		zap.Any(database.UpdateDataKey, data),
		zap.Uint32s(biz.PermissionIDsKey, database.ListModelToIDs(perms)),
		zap.Any(database.ConditionKey, conds),
	)
	now := time.Now()
	upmap := make(map[string]any, 1)
	if len(perms) > 0 {
		upmap["Permissions"] = perms
	}
	if err := database.DBUpdate(ctx, r.gormDB, &biz.MenuModel{}, data, upmap, conds...); err != nil {
		r.log.Error(
			"更新菜单模型失败",
			zap.Error(err),
			zap.Any(database.UpdateDataKey, data),
			zap.Uint32s(biz.PermissionIDsKey, database.ListModelToIDs(perms)),
			zap.Any(database.ConditionKey, conds),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return err
	}
	r.log.Debug(
		"更新菜单模型成功",
		zap.Any(database.UpdateDataKey, data),
		zap.Any(database.ConditionKey, conds),
		zap.Uint32s(biz.PermissionIDsKey, database.ListModelToIDs(perms)),
		zap.Duration(log.DurationKey, time.Since(now)),
	)
	return nil
}

func (r *menuRepo) DeleteModel(ctx context.Context, conds ...any) error {
	r.log.Debug("开始删除菜单模型", zap.Any(database.ConditionKey, conds))
	now := time.Now()
	if err := database.DBDelete(ctx, r.gormDB, &biz.MenuModel{}, conds...); err != nil {
		r.log.Error(
			"删除菜单模型失败",
			zap.Error(err),
			zap.Any(database.ConditionKey, conds),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return err
	}
	r.log.Debug(
		"删除菜单模型成功",
		zap.Any(database.ConditionKey, conds),
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
	)
	now := time.Now()
	var m biz.MenuModel
	if err := database.DBFind(ctx, r.gormDB, preloads, &m, conds...); err != nil {
		r.log.Error(
			"查询菜单模型失败",
			zap.Error(err),
			zap.Strings(database.PreloadKey, preloads),
			zap.Any(database.ConditionKey, conds),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return nil, err
	}
	r.log.Debug(
		"查询菜单模型成功",
		zap.Object(database.ModelKey, &m),
		zap.Strings(database.PreloadKey, preloads),
		zap.Any(database.ConditionKey, conds),
		zap.Duration(log.DurationKey, time.Since(now)),
	)
	return &m, nil
}

func (r *menuRepo) ListModel(
	ctx context.Context,
	qp database.QueryParams,
) (int64, []*biz.MenuModel, error) {
	r.log.Debug(
		"开始查询菜单模型列表",
		zap.Object(database.QueryParamsKey, &qp),
	)
	now := time.Now()
	var ms []*biz.MenuModel
	count, err := database.DBList(ctx, r.gormDB, &biz.MenuModel{}, &ms, qp)
	if err != nil {
		r.log.Error(
			"查询菜单模型列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return 0, nil, err
	}
	r.log.Debug(
		"查询菜单模型列表成功",
		zap.Object(database.QueryParamsKey, &qp),
		zap.Duration(log.DurationKey, time.Since(now)),
	)
	return count, ms, nil
}

func (r *menuRepo) AddGroupPolicy(
	ctx context.Context,
	m biz.MenuModel,
) error {
	if err := errors.CheckContext(ctx); err != nil {
		return err
	}
	sub := auth.MenuToSubject(m.ID)

	// 处理父级关系
	if m.Parent != nil {
		obj := auth.MenuToSubject(*m.ParentID)
		r.log.Debug(
			"开始添加菜单与父级菜单的继承关系策略",
			zap.String(auth.GroupSubKey, sub),
			zap.String(auth.GroupObjKey, obj),
		)
		now := time.Now()
		if err := r.cache.AddGroupPolicy(sub, obj); err != nil {
			r.log.Error(
				"添加菜单与父级菜单的继承关系策略失败",
				zap.Error(err),
				zap.String(auth.GroupSubKey, sub),
				zap.String(auth.GroupObjKey, obj),
				zap.Duration(log.DurationKey, time.Since(now)),
			)
			return err
		}
		r.log.Debug(
			"添加菜单与父级菜单的继承关系策略成功",
			zap.String(auth.GroupSubKey, sub),
			zap.String(auth.GroupObjKey, obj),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
	}

	r.log.Debug(
		"开始添加菜单与权限的关联策略",
		zap.Object(database.ModelKey, &m),
		zap.Uint32s(biz.PermissionIDsKey, database.ListModelToIDs(m.Permissions)),
	)
	now := time.Now()
	// 批量处理权限
	for _, o := range m.Permissions {
		obj := auth.PermissionToSubject(o.ID)
		if err := r.cache.AddGroupPolicy(sub, obj); err != nil {
			r.log.Error(
				"添加菜单与权限的关联策略失败",
				zap.Error(err),
				zap.String(auth.GroupSubKey, sub),
				zap.String(auth.GroupObjKey, obj),
				zap.Duration(log.DurationKey, time.Since(now)),
			)
			return err
		}
	}
	r.log.Debug(
		"添加菜单与权限的关联策略成功",
		zap.Object(database.ModelKey, &m),
		zap.Uint32s(biz.PermissionIDsKey, database.ListModelToIDs(m.Permissions)),
		zap.Duration(log.DurationKey, time.Since(now)),
	)
	return nil
}

func (r *menuRepo) RemoveGroupPolicy(
	ctx context.Context,
	m biz.MenuModel,
	removeInherited bool,
) error {
	if err := errors.CheckContext(ctx); err != nil {
		return err
	}
	sub := auth.MenuToSubject(m.ID)
	r.log.Debug(
		"开始删除该菜单作为子级的组策略",
		zap.Object(database.ModelKey, &m),
		zap.String(auth.GroupSubKey, sub),
	)
	rmSubStartTime := time.Now()
	if err := r.cache.RemoveGroupPolicy(0, sub); err != nil {
		r.log.Error(
			"删除该菜单作为子级的组策略失败",
			zap.Error(err),
			zap.Object(database.ModelKey, &m),
			zap.String(auth.GroupSubKey, sub),
			zap.Duration(log.DurationKey, time.Since(rmSubStartTime)),
		)
		return err
	}
	r.log.Debug(
		"删除该菜单作为子级的组策略成功",
		zap.Object(database.ModelKey, &m),
		zap.String(auth.GroupSubKey, sub),
		zap.Duration(log.DurationKey, time.Since(rmSubStartTime)),
	)
	if removeInherited {
		r.log.Debug(
			"开始删除该菜单作为父级的组策略",
			zap.Object(database.ModelKey, &m),
			zap.String(auth.GroupObjKey, sub),
		)
		rmObjStartTime := time.Now()
		if err := r.cache.RemoveGroupPolicy(1, sub); err != nil {
			r.log.Error(
				"删除该菜单作为父级的组策略失败",
				zap.Error(err),
				zap.Object(database.ModelKey, &m),
				zap.String(auth.GroupObjKey, sub),
				zap.Duration(log.DurationKey, time.Since(rmObjStartTime)),
			)
			return err
		}
		r.log.Debug(
			"删除该菜单作为父级的组策略成功",
			zap.Object(database.ModelKey, &m),
			zap.String(auth.GroupObjKey, sub),
			zap.Duration(log.DurationKey, time.Since(rmObjStartTime)),
		)
	}
	return nil
}
