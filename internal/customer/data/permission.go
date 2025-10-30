package data

import (
	"context"
	"strconv"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"gin-artweb/internal/customer/biz"
	"gin-artweb/pkg/auth"
	"gin-artweb/pkg/database"
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
			"新增权限模型失败",
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
			"更新权限模型失败",
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
			"删除权限模型失败",
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
			"查询权限模型失败",
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
			"查询权限列表失败",
			zap.Object("query_params", &qp),
			zap.Error(err),
		)
		return 0, nil, err
	}
	return count, ms, nil
}

func (r *permissionRepo) ListModelByIds(
	ctx context.Context,
	ids []uint32,
) ([]biz.PermissionModel, error) {
	if len(ids) == 0 {
		return []biz.PermissionModel{}, nil
	}
	qp := database.NewPksQueryParams(ids)
	_, ms, err := r.ListModel(ctx, qp)
	return ms, err
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
	if err := r.cache.AddPolicy(sub, m.Url, m.Method); err != nil {
		r.log.Error(
			"添加权限策略失败",
			zap.String(auth.SubKey, sub),
			zap.String(auth.ObjKey, m.Url),
			zap.String(auth.ActKey, m.Method),
			zap.Error(err),
		)
		return err
	}
	return nil
}

func (r *permissionRepo) RemovePolicy(
	ctx context.Context,
	m biz.PermissionModel,
	removeInherited bool,
) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	sub := permissionModelToSub(m)
	if err := r.cache.RemovePolicy(sub, m.Url, m.Method); err != nil {
		r.log.Error(
			"删除权限策略失败",
			zap.String(auth.SubKey, sub),
			zap.String(auth.ObjKey, m.Url),
			zap.String(auth.ActKey, m.Method),
			zap.Error(err),
		)
		return err
	}
	if removeInherited {
		if err := r.cache.RemoveGroupPolicy(1, sub); err != nil {
			r.log.Error(
				"删除权限作为父级策略失败(该策略被其他策略继承)",
				zap.String(auth.ObjKey, sub),
				zap.Error(err),
			)
			return err
		}
	}
	return nil
}

func permissionModelToSub(m biz.PermissionModel) string {
	return strconv.FormatUint(uint64(m.Id), 10)
}
