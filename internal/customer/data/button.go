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
			zap.Object("model", m),
			zap.Error(err),
		)
		return err
	}
	return nil
}

func (r *buttonRepo) UpdateModel(ctx context.Context, data map[string]any, upmap map[string]any, conds ...any) error {
	if err := database.DBUpdate(ctx, r.gormDB, &biz.ButtonModel{}, data, upmap, conds...); err != nil {
		r.log.Error(
			"更新按钮模型失败",
			zap.Any("data", data),
			zap.Any("conditions", conds),
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
			zap.Any("conditions", conds),
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
			zap.Strings("preloads", preloads),
			zap.Any("conditions", conds),
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
			zap.Object("query_params", &qp),
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
	sub := buttonModelToSub(m)
	menuObj := menuModelToSub(m.Menu)
	if err := r.cache.AddGroupPolicy(sub, menuObj); err != nil {
		r.log.Error(
			"添加按钮权限失败",
			zap.String(auth.SubKey, sub),
			zap.String(auth.ObjKey, menuObj),
			zap.Uint32("button_id", m.Id),
			zap.Uint32("menu_id", m.MenuId),
			zap.Error(err),
		)
		return err
	}

	// 批量处理权限
	for _, o := range m.Permissions {
		obj := permissionModelToSub(o)
		if err := r.cache.AddGroupPolicy(sub, obj); err != nil {
			r.log.Error(
				"添加按钮与菜单的继承关系策略失败",
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
	sub := buttonModelToSub(m)

	// 删除该按钮作为父级的策略（被其他菜单或权限继承）
	if err := r.cache.RemoveGroupPolicy(1, sub); err != nil {
		r.log.Error(
			"删除按钮作为子级策略失败(该策略继承自其他策略)",
			zap.String("sub", sub),
			zap.Error(err),
		)
		return err
	}
	// 删除该按钮作为子级的策略（从其他菜单或权限继承）
	if removeInherited {
		if err := r.cache.RemoveGroupPolicy(0, sub); err != nil {
			r.log.Error(
				"删除按钮作为子级策略失败(该策略被其他策略继承)",
				zap.String(auth.ObjKey, sub),
				zap.Error(err),
			)
			return err
		}
	}
	return nil
}

func buttonModelToSub(m biz.ButtonModel) string {
	return fmt.Sprintf("button_%d", m.Id)
}
