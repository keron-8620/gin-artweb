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

type roleRepo struct {
	log    *zap.Logger
	gormDB *gorm.DB
	cache  *auth.AuthEnforcer
}

func NewRoleRepo(
	log *zap.Logger,
	gormDB *gorm.DB,
	cache *auth.AuthEnforcer,
) biz.RoleRepo {
	return &roleRepo{
		log:    log,
		gormDB: gormDB,
		cache:  cache,
	}
}

func (r *roleRepo) CreateModel(ctx context.Context, m *biz.RoleModel) error {
	now := time.Now()
	m.CreatedAt = now
	m.UpdatedAt = now
	if err := database.DBCreate(ctx, r.gormDB, &biz.RoleModel{}, m); err != nil {
		r.log.Error(
			"新增角色模型失败",
			zap.Object("model", m),
			zap.Error(err),
		)
		return err
	}
	return nil
}

func (r *roleRepo) UpdateModel(ctx context.Context, data map[string]any, upmap map[string]any, conds ...any) error {
	if err := database.DBUpdate(ctx, r.gormDB, &biz.RoleModel{}, data, upmap, conds...); err != nil {
		r.log.Error(
			"更新角色模型失败",
			zap.Any("data", data),
			zap.Any("conditions", conds),
			zap.Error(err),
		)
		return err
	}
	return nil
}

func (r *roleRepo) DeleteModel(ctx context.Context, conds ...any) error {
	if err := database.DBDelete(ctx, r.gormDB, &biz.RoleModel{}, conds...); err != nil {
		r.log.Error(
			"删除角色模型失败",
			zap.Any("conditions", conds),
			zap.Error(err),
		)
		return err
	}
	return nil
}

func (r *roleRepo) FindModel(
	ctx context.Context,
	preloads []string,
	conds ...any,
) (*biz.RoleModel, error) {
	var m biz.RoleModel
	if err := database.DBFind(ctx, r.gormDB, preloads, &m, conds...); err != nil {
		r.log.Error(
			"查询角色模型失败",
			zap.Strings("preloads", preloads),
			zap.Any("conditions", conds),
			zap.Error(err),
		)
		return nil, err
	}
	return &m, nil
}

func (r *roleRepo) ListModel(
	ctx context.Context,
	qp database.QueryParams,
) (int64, []biz.RoleModel, error) {
	var ms []biz.RoleModel
	count, err := database.DBList(ctx, r.gormDB, &biz.RoleModel{}, &ms, qp)
	if err != nil {
		r.log.Error(
			"查询角色列表失败",
			zap.Object("query_params", &qp),
			zap.Error(err),
		)
		return 0, nil, err
	}
	return count, ms, nil
}

func (r *roleRepo) RoleModelToSub(m biz.RoleModel) string {
	return fmt.Sprintf("role_%d", m.Id)
}

func (r *roleRepo) AddGroupPolicy(
	ctx context.Context,
	m biz.RoleModel,
) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	sub := r.RoleModelToSub(m)

	// 批量处理权限
	for _, o := range m.Permissions {
		obj := permissionModelToSub(o)
		if err := r.cache.AddGroupPolicy(sub, obj); err != nil {
			r.log.Error(
				"添加角色与权限的关联策略失败",
				zap.String(auth.SubKey, sub),
				zap.String(auth.ObjKey, obj),
				zap.Uint32("role_id", m.Id),
				zap.Uint32("permission_id", o.Id),
				zap.Error(err),
			)
			return err
		}
	}

	// 批量处理菜单
	for _, o := range m.Menus {
		obj := menuModelToSub(o)
		if err := r.cache.AddGroupPolicy(sub, obj); err != nil {
			r.log.Error(
				"添加角色与菜单的关联策略失败",
				zap.String(auth.SubKey, sub),
				zap.String(auth.ObjKey, obj),
				zap.Uint32("role_id", m.Id),
				zap.Uint32("menu_id", o.Id),
				zap.Error(err),
			)
			return err
		}
	}

	// 批量处理按钮
	for _, o := range m.Buttons {
		obj := buttonModelToSub(o)
		if err := r.cache.AddGroupPolicy(sub, obj); err != nil {
			r.log.Error(
				"添加角色与按钮的关联策略失败",
				zap.String(auth.SubKey, sub),
				zap.String(auth.ObjKey, obj),
				zap.Uint32("role_id", m.Id),
				zap.Uint32("button_id", o.Id),
				zap.Error(err),
			)
			return err
		}
	}
	return nil
}

func (r *roleRepo) RemoveGroupPolicy(
	ctx context.Context,
	m biz.RoleModel,
) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	sub := r.RoleModelToSub(m)

	// 删除角色作为子级的策略（从其他菜单或权限继承）
	if err := r.cache.RemoveGroupPolicy(0, sub); err != nil {
		r.log.Error(
			"删除角色作为子级策略失败(该策略继承自其他策略)",
			zap.String(auth.ObjKey, sub),
			zap.Error(err),
		)
		return err
	}
	return nil
}
