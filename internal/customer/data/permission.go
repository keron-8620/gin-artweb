package data

import (
	"context"
	"strconv"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"gitee.com/keion8620/go-dango-gin/internal/customer/biz"
	"gitee.com/keion8620/go-dango-gin/pkg/auth"
	"gitee.com/keion8620/go-dango-gin/pkg/database"
)

type permissionRepo struct {
	log    *zap.Logger
	gormDB *gorm.DB
	cache  *auth.AuthEnforcer
}

func NewPermissionRepo(
	log *zap.Logger,
	gormDB *gorm.DB,
	cache *auth.AuthEnforcer,
) biz.PermissionRepo {
	return &permissionRepo{
		log:    log,
		gormDB: gormDB,
		cache:  cache,
	}
}

func (r *permissionRepo) CreateModel(ctx context.Context, m *biz.PermissionModel) error {
	now := time.Now()
	m.CreatedAt = now
	m.UpdatedAt = now
	if err := database.DBCreate(ctx, r.gormDB, &biz.PermissionModel{}, m); err != nil {
		r.log.Error(
			"failed to create permission model",
			zap.Object("model", m),
			zap.Error(err),
		)
		return err
	}
	return nil
}

func (r *permissionRepo) UpdateModel(ctx context.Context, data map[string]any, conds ...any) error {
	if err := database.DBUpdate(ctx, r.gormDB, &biz.PermissionModel{}, data, nil, conds...); err != nil {
		r.log.Error(
			"failed to update permission model",
			zap.Any("data", data),
			zap.Any("conditions", conds),
			zap.Error(err),
		)
		return err
	}
	return nil
}

func (r *permissionRepo) DeleteModel(ctx context.Context, conds ...any) error {
	if err := database.DBDelete(ctx, r.gormDB, &biz.PermissionModel{}, conds...); err != nil {
		r.log.Error(
			"failed to delete permission model",
			zap.Any("conditions", conds),
			zap.Error(err),
		)
		return err
	}
	return nil
}

func (r *permissionRepo) FindModel(
	ctx context.Context,
	preloads []string,
	conds ...any,
) (*biz.PermissionModel, error) {
	var m biz.PermissionModel
	if err := database.DBFind(ctx, r.gormDB, preloads, &m, conds...); err != nil {
		r.log.Error(
			"failed to find permission model",
			zap.Strings("preloads", preloads),
			zap.Any("conditions", conds),
			zap.Error(err),
		)
		return nil, err
	}
	return &m, nil
}

func (r *permissionRepo) ListModel(
	ctx context.Context,
	qp database.QueryParams,
) (int64, []biz.PermissionModel, error) {
	var ms []biz.PermissionModel
	count, err := database.DBList(ctx, r.gormDB, &biz.PermissionModel{}, &ms, qp)
	if err != nil {
		r.log.Error(
			"failed to list permission model",
			zap.Object("query_params", &qp),
			zap.Error(err),
		)
		return 0, nil, err
	}
	return count, ms, nil
}

func (r *permissionRepo) AddPolicy(
	ctx context.Context,
	m biz.PermissionModel,
) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	sub := permissionModelToSub(m)
	return r.cache.AddPolicy(sub, m.HttpUrl, m.Method)
}

func (r *permissionRepo) RemovePolicy(
	ctx context.Context,
	m biz.PermissionModel,
) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	sub := permissionModelToSub(m)
	if err := r.cache.RemovePolicy(sub, m.HttpUrl, m.Method); err != nil {
		return err
	}
	if err := r.cache.RemoveGroupPolicy(1, sub); err != nil {
		return err
	}
	return nil
}

func permissionModelToSub(m biz.PermissionModel) string {
	return strconv.FormatUint(uint64(m.Id), 10)
}
