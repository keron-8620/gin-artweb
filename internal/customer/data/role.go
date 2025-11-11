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
			zap.Object(database.ModelKey, m),
			zap.Error(err),
		)
		return err
	}
	return nil
}

func (r *roleRepo) UpdateModel(
	ctx context.Context,
	data map[string]any,
	perms []biz.PermissionModel,
	menus []biz.MenuModel,
	buttons []biz.ButtonModel,
	conds ...any,
) error {
	upmap := make(map[string]any, 3)
	if len(perms) > 0 {
		upmap["Permissions"] = perms
	}
	if len(menus) > 0 {
		upmap["Menus"] = menus
	}
	if len(buttons) > 0 {
		upmap["Buttons"] = buttons
	}
	if err := database.DBUpdate(ctx, r.gormDB, &biz.RoleModel{}, data, upmap, conds...); err != nil {
		r.log.Error(
			"更新角色模型失败",
			zap.Any(database.UpdateDataKey, data),
			zap.Objects("permissions", func() []zapcore.ObjectMarshaler {
				pobjs := make([]zapcore.ObjectMarshaler, len(perms))
				for i := range perms {
					pobjs[i] = &perms[i]
				}
				return pobjs
			}()),
			zap.Objects("menus", func() []zapcore.ObjectMarshaler {
				mobjs := make([]zapcore.ObjectMarshaler, len(menus))
				for i := range menus {
					mobjs[i] = &menus[i]
				}
				return mobjs
			}()),
			zap.Objects("buttons", func() []zapcore.ObjectMarshaler {
				bobjs := make([]zapcore.ObjectMarshaler, len(buttons))
				for i := range buttons {
					bobjs[i] = &buttons[i]
				}
				return bobjs
			}()),
			zap.Any(database.ConditionKey, conds),
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
			zap.Any(database.ConditionKey, conds),
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
			zap.Strings(database.PreloadKey, preloads),
			zap.Any(database.ConditionKey, conds),
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
			zap.Object(database.QueryParamsKey, &qp),
			zap.Error(err),
		)
		return 0, nil, err
	}
	return count, ms, nil
}

func (r *roleRepo) RoleModelToSub(m biz.RoleModel) string {
	return fmt.Sprintf(auth.RoleSubjectFormat, m.ID)
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
		obj := permissionModelToSubject(o)
		if err := r.cache.AddGroupPolicy(sub, obj); err != nil {
			r.log.Error(
				"添加角色与权限的关联策略失败",
				zap.String(auth.GroupSubKey, sub),
				zap.String(auth.GroupObjKey, obj),
				zap.Uint32("role_id", m.ID),
				zap.Uint32("permission_id", o.ID),
				zap.Error(err),
			)
			return err
		}
	}

	// 批量处理菜单
	for _, o := range m.Menus {
		obj := menuModelToSubject(o)
		if err := r.cache.AddGroupPolicy(sub, obj); err != nil {
			r.log.Error(
				"添加角色与菜单的关联策略失败",
				zap.String(auth.GroupSubKey, sub),
				zap.String(auth.GroupObjKey, obj),
				zap.Uint32("role_id", m.ID),
				zap.Uint32("menu_id", o.ID),
				zap.Error(err),
			)
			return err
		}
	}

	// 批量处理按钮
	for _, o := range m.Buttons {
		obj := buttonModelToSubject(o)
		if err := r.cache.AddGroupPolicy(sub, obj); err != nil {
			r.log.Error(
				"添加角色与按钮的关联策略失败",
				zap.String(auth.GroupSubKey, sub),
				zap.String(auth.GroupObjKey, obj),
				zap.Uint32("role_id", m.ID),
				zap.Uint32("button_id", o.ID),
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
	if err := r.cache.RemoveGroupPolicy(1, sub); err != nil {
		r.log.Error(
			"删除角色作为子级策略失败(该策略继承自其他策略)",
			zap.String(auth.GroupObjKey, sub),
			zap.Error(err),
		)
		return err
	}
	return nil
}
