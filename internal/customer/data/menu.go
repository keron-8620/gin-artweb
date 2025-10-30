package data

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
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
			zap.Object("model", m),
			zap.Error(err),
		)
		return err
	}
	return nil
}

func (r *menuRepo) UpdateModel(ctx context.Context, data map[string]any, upmap map[string]any, conds ...any) error {
	if err := database.DBUpdate(ctx, r.gormDB, &biz.MenuModel{}, data, upmap, conds...); err != nil {
		r.log.Error(
			"更新菜单模型失败",
			zap.Any("data", data),
			zap.Any("conditions", conds),
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
			zap.Any("conditions", conds),
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
			zap.Strings("preloads", preloads),
			zap.Any("conditions", conds),
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
			zap.Object("query_params", &qp),
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
	sub := menuModelToSub(m)

	// 处理父级关系
	if m.Parent != nil {
		obj := menuModelToSub(*m.Parent)
		if err := r.cache.AddGroupPolicy(sub, obj); err != nil {
			r.log.Error(
				"添加菜单与父级菜单的继承关系策略失败",
				zap.String(auth.SubKey, sub),
				zap.String(auth.ObjKey, obj),
				zap.Uint32("menu_id", m.Id),
				zap.Uint32("parent_menu_id", *m.ParentId),
				zap.Error(err),
			)
			return err
		}
	}

	// 批量处理权限
	for _, o := range m.Permissions {
		obj := permissionModelToSub(o)
		if err := r.cache.AddGroupPolicy(sub, obj); err != nil {
			r.log.Error(
				"添加菜单与权限的关联策略失败",
				zap.String(auth.SubKey, sub),
				zap.String(auth.ObjKey, obj),
				zap.Uint32("menu_id", m.Id),
				zap.Uint32("permission_id", o.Id),
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
	sub := menuModelToSub(m)

	// 删除该菜单作为父级的策略（被其他菜单或权限继承）
	if err := r.cache.RemoveGroupPolicy(1, sub); err != nil {
		r.log.Error(
			"删除菜单作为子级策略失败(该策略继承自其他策略)",
			zap.String("sub", sub),
			zap.Error(err),
		)
		return err
	}
	// 删除该菜单作为子级的策略（从其他菜单或权限继承）
	if removeInherited {
		if err := r.cache.RemoveGroupPolicy(0, sub); err != nil {
			r.log.Error(
				"删除菜单作为父级策略失败(该策略被其他策略继承)",
				zap.String(auth.ObjKey, sub),
				zap.Error(err),
			)
			return err
		}
	}
	return nil
}

func menuModelToSub(m biz.MenuModel) string {
	return fmt.Sprintf("menu_%d", m.Id)
}
