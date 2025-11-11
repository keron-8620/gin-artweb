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
	now := time.Now()
	m.CreatedAt = now
	m.UpdatedAt = now
	if err := database.DBCreate(ctx, r.gormDB, &biz.ButtonModel{}, m); err != nil {
		r.log.Error(
			"新增按钮模型失败",
			zap.Object(database.ModelKey, m),
			zap.Error(err),
		)
		return err
	}
	return nil
}

func (r *buttonRepo) UpdateModel(
	ctx context.Context,
	data map[string]any,
	perms []biz.PermissionModel,
	conds ...any,
) error {
	upmap := make(map[string]any, 1)
	if len(perms) > 0 {
		upmap["Permissions"] = perms
	}
	if err := database.DBUpdate(ctx, r.gormDB, &biz.ButtonModel{}, data, upmap, conds...); err != nil {
		r.log.Error(
			"更新按钮模型失败",
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

func (r *buttonRepo) DeleteModel(ctx context.Context, conds ...any) error {
	if err := database.DBDelete(ctx, r.gormDB, &biz.ButtonModel{}, conds...); err != nil {
		r.log.Error(
			"删除按钮模型失败",
			zap.Any(database.ConditionKey, conds),
			zap.Error(err),
		)
		return err
	}
	return nil
}

func (r *buttonRepo) FindModel(
	ctx context.Context,
	preloads []string,
	conds ...any,
) (*biz.ButtonModel, error) {
	var m biz.ButtonModel
	if err := database.DBFind(ctx, r.gormDB, preloads, &m, conds...); err != nil {
		r.log.Error(
			"查询按钮模型失败",
			zap.Strings(database.PreloadKey, preloads),
			zap.Any(database.ConditionKey, conds),
			zap.Error(err),
		)
		return nil, err
	}
	return &m, nil
}

func (r *buttonRepo) ListModel(
	ctx context.Context,
	qp database.QueryParams,
) (int64, []biz.ButtonModel, error) {
	var ms []biz.ButtonModel
	count, err := database.DBList(ctx, r.gormDB, &biz.ButtonModel{}, &ms, qp)
	if err != nil {
		r.log.Error(
			"查询按钮列表失败",
			zap.Object(database.QueryParamsKey, &qp),
			zap.Error(err),
		)
		return 0, nil, err
	}
	return count, ms, nil
}

func (r *buttonRepo) AddGroupPolicy(
	ctx context.Context,
	m biz.ButtonModel,
) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	sub := buttonModelToSubject(m)
	menuObj := menuModelToSubject(m.Menu)
	if err := r.cache.AddGroupPolicy(sub, menuObj); err != nil {
		r.log.Error(
			"添加按钮权限失败",
			zap.String(auth.GroupSubKey, sub),
			zap.String(auth.GroupObjKey, menuObj),
			zap.Uint32("button_id", m.ID),
			zap.Uint32("menu_id", m.MenuID),
			zap.Error(err),
		)
		return err
	}

	// 批量处理权限
	for _, o := range m.Permissions {
		obj := permissionModelToSubject(o)
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
	sub := buttonModelToSubject(m)

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

func buttonModelToSubject(m biz.ButtonModel) string {
	return fmt.Sprintf(auth.ButtonSubjectFormat, m.ID)
}
