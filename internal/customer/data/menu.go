package data

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/gorm"

	"gin-artweb/internal/customer/biz"
	"gin-artweb/pkg/auth"
	"gin-artweb/pkg/database"
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
	now := time.Now()
	m.CreatedAt = now
	m.UpdatedAt = now
	if err := database.DBCreate(ctx, r.gormDB, &biz.MenuModel{}, m); err != nil {
		r.log.Error(
			"新增菜单模型失败",
			zap.Object(database.ModelKey, m),
			zap.Error(err),
		)
		return err
	}
	return nil
}

func (r *menuRepo) UpdateModel(
	ctx context.Context,
	data map[string]any,
	perms []biz.PermissionModel,
	conds ...any,
) error {
	upmap := make(map[string]any, 1)
	if len(perms) > 0 {
		upmap["Permissions"] = perms
	}
	if err := database.DBUpdate(ctx, r.gormDB, &biz.MenuModel{}, data, upmap, conds...); err != nil {
		r.log.Error(
			"更新菜单模型失败",
			zap.Any(database.UpdateDataKey, data),
			zap.Objects("permissions", func() []zapcore.ObjectMarshaler {
				objs := make([]zapcore.ObjectMarshaler, len(perms))
				for i := range perms {
					objs[i] = &perms[i]
				}
				return objs
			}()),
			zap.Any(database.ConditionKey, conds),
			zap.Error(err),
		)
		return err
	}
	return nil
}

func (r *menuRepo) DeleteModel(ctx context.Context, conds ...any) error {
	if err := database.DBDelete(ctx, r.gormDB, &biz.MenuModel{}, conds...); err != nil {
		r.log.Error(
			"删除菜单模型失败",
			zap.Any(database.ConditionKey, conds),
			zap.Error(err),
		)
		return err
	}
	return nil
}

func (r *menuRepo) FindModel(
	ctx context.Context,
	preloads []string,
	conds ...any,
) (*biz.MenuModel, error) {
	var m biz.MenuModel
	if err := database.DBFind(ctx, r.gormDB, preloads, &m, conds...); err != nil {
		r.log.Error(
			"查询菜单模型失败",
			zap.Strings(database.PreloadKey, preloads),
			zap.Any(database.ConditionKey, conds),
			zap.Error(err),
		)
		return nil, err
	}
	return &m, nil
}

func (r *menuRepo) ListModel(
	ctx context.Context,
	qp database.QueryParams,
) (int64, []biz.MenuModel, error) {
	var ms []biz.MenuModel
	count, err := database.DBList(ctx, r.gormDB, &biz.MenuModel{}, &ms, qp)
	if err != nil {
		r.log.Error(
			"查询菜单列表失败",
			zap.Object(database.QueryParamsKey, &qp),
			zap.Error(err),
		)
		return 0, nil, err
	}
	return count, ms, nil
}

func (r *menuRepo) AddGroupPolicy(
	ctx context.Context,
	m biz.MenuModel,
) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	sub := menuModelToSubject(m)

	// 处理父级关系
	if m.Parent != nil {
		obj := menuModelToSubject(*m.Parent)
		if err := r.cache.AddGroupPolicy(sub, obj); err != nil {
			r.log.Error(
				"添加菜单与父级菜单的继承关系策略失败",
				zap.String(auth.GroupSubKey, sub),
				zap.String(auth.GroupObjKey, obj),
				zap.Uint32("menu_id", m.ID),
				zap.Uint32("parent_menu_id", *m.ParentID),
				zap.Error(err),
			)
			return err
		}
	}

	// 批量处理权限
	for _, o := range m.Permissions {
		obj := permissionModelToSubject(o)
		if err := r.cache.AddGroupPolicy(sub, obj); err != nil {
			r.log.Error(
				"添加菜单与权限的关联策略失败",
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

func (r *menuRepo) RemoveGroupPolicy(
	ctx context.Context,
	m biz.MenuModel,
	removeInherited bool,
) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	sub := menuModelToSubject(m)
	if err := r.cache.RemoveGroupPolicy(0, sub); err != nil {
		r.log.Error(
			"删除菜单作为子级策略失败(该策略继承自其他策略)",
			zap.String(auth.GroupSubKey, sub),
			zap.Error(err),
		)
		return err
	}
	if removeInherited {
		if err := r.cache.RemoveGroupPolicy(1, sub); err != nil {
			r.log.Error(
				"删除菜单作为父级策略失败(该策略被其他策略继承)",
				zap.String(auth.GroupObjKey, sub),
				zap.Error(err),
			)
			return err
		}
	}
	return nil
}

func menuModelToSubject(m biz.MenuModel) string {
	return fmt.Sprintf(auth.MenuSubjectFormat, m.ID)
}
