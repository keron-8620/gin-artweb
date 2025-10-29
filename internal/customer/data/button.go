package data

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"gitee.com/keion8620/go-dango-gin/internal/customer/biz"
	"gitee.com/keion8620/go-dango-gin/pkg/auth"
	"gitee.com/keion8620/go-dango-gin/pkg/database"
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
			"failed to create button model",
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
			"failed to update button model",
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
			"failed to delete button model",
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
			"failed to find button model",
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
			"failed to list button model",
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
		return err
	}

	// 批量处理权限
	for _, o := range m.Permissions {
		obj := permissionModelToSub(o)
		if err := r.cache.AddGroupPolicy(sub, obj); err != nil {
			return err
		}
	}
	return nil
}

func (r *buttonRepo) RemoveGroupPolicy(
	ctx context.Context,
	m biz.ButtonModel,
) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	sub := buttonModelToSub(m)

	// 删除该按钮作为父级的策略（被其他菜单或权限继承）
	if err := r.cache.RemoveGroupPolicy(1, sub); err != nil {
		return err
	}
	// 删除该按钮作为子级的策略（从其他菜单或权限继承）
	if err := r.cache.RemoveGroupPolicy(0, sub); err != nil {
		return err
	}
	return nil
}

func buttonModelToSub(m biz.ButtonModel) string {
	return fmt.Sprintf("button_%d", m.Id)
}
