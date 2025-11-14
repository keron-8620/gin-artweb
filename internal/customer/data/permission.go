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
	r.log.Debug(
		"开始创建权限模型",
		zap.Object(database.ModelKey, m),
	)
	now := time.Now()
	m.CreatedAt = now
	m.UpdatedAt = now
	if err := database.DBCreate(ctx, r.gormDB, &biz.PermissionModel{}, m); err != nil {
		r.log.Error(
			"创建权限模型失败",
			zap.Error(err),
			zap.Object(database.ModelKey, m),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return err
	}
	r.log.Debug(
		"创建权限模型成功",
		zap.Object(database.ModelKey, m),
		zap.Duration(log.DurationKey, time.Since(now)),
	)
	return nil
}

func (r *permissionRepo) UpdateModel(ctx context.Context, data map[string]any, conds ...any) error {
	r.log.Debug(
		"开始更新权限模型",
		zap.Any(database.UpdateDataKey, data),
		zap.Any(database.ConditionKey, conds),
	)
	startTime := time.Now()
	if err := database.DBUpdate(ctx, r.gormDB, &biz.PermissionModel{}, data, nil, conds...); err != nil {
		r.log.Error(
			"更新权限模型失败",
			zap.Error(err),
			zap.Any(database.UpdateDataKey, data),
			zap.Any(database.ConditionKey, conds),
			zap.Duration(log.DurationKey, time.Since(startTime)),
		)
		return err
	}
	r.log.Debug(
		"更新权限模型成功",
		zap.Any(database.UpdateDataKey, data),
		zap.Any(database.ConditionKey, conds),
		zap.Duration(log.DurationKey, time.Since(startTime)),
	)
	return nil
}

func (r *permissionRepo) DeleteModel(ctx context.Context, conds ...any) error {
	r.log.Debug("开始删除权限模型", zap.Any(database.ConditionKey, conds))
	startTime := time.Now()
	if err := database.DBDelete(ctx, r.gormDB, &biz.PermissionModel{}, conds...); err != nil {
		r.log.Error(
			"删除权限模型失败",
			zap.Error(err),
			zap.Any(database.ConditionKey, conds),
			zap.Duration(log.DurationKey, time.Since(startTime)),
		)
		return err
	}
	r.log.Debug(
		"删除权限模型成功",
		zap.Any(database.ConditionKey, conds),
		zap.Duration(log.DurationKey, time.Since(startTime)),
	)
	return nil
}

func (r *permissionRepo) FindModel(
	ctx context.Context,
	preloads []string,
	conds ...any,
) (*biz.PermissionModel, error) {
	r.log.Debug(
		"开始查询权限模型",
		zap.Strings(database.PreloadKey, preloads),
		zap.Any(database.ConditionKey, conds),
	)
	startTime := time.Now()
	var m biz.PermissionModel
	if err := database.DBFind(ctx, r.gormDB, preloads, &m, conds...); err != nil {
		r.log.Error(
			"查询权限模型失败",
			zap.Error(err),
			zap.Strings(database.PreloadKey, preloads),
			zap.Any(database.ConditionKey, conds),
			zap.Duration(log.DurationKey, time.Since(startTime)),
		)
		return nil, err
	}
	r.log.Debug(
		"查询权限模型成功",
		zap.Object(database.ModelKey, &m),
		zap.Strings(database.PreloadKey, preloads),
		zap.Any(database.ConditionKey, conds),
		zap.Duration(log.DurationKey, time.Since(startTime)),
	)
	return &m, nil
}

func (r *permissionRepo) ListModel(
	ctx context.Context,
	qp database.QueryParams,
) (int64, *[]biz.PermissionModel, error) {
	r.log.Debug(
		"开始查询权限模型列表",
		zap.Object(database.QueryParamsKey, &qp),
	)
	startTime := time.Now()
	var ms []biz.PermissionModel
	count, err := database.DBList(ctx, r.gormDB, &biz.PermissionModel{}, &ms, qp)
	if err != nil {
		r.log.Error(
			"查询权限模型列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.Duration(log.DurationKey, time.Since(startTime)),
		)
		return 0, nil, err
	}
	r.log.Debug(
		"查询权限模型列表成功",
		zap.Object(database.QueryParamsKey, &qp),
		zap.Duration(log.DurationKey, time.Since(startTime)),
	)
	return count, &ms, nil
}

func (r *permissionRepo) AddPolicy(
	ctx context.Context,
	m biz.PermissionModel,
) error {
	if err := errors.CheckContext(ctx); err != nil {
		return err
	}
	r.log.Debug(
		"开始添加权限策略",
		zap.Object(database.ModelKey, &m),
	)
	startTime := time.Now()
	sub := auth.PermissionToSubject(m.ID)
	if err := r.cache.AddPolicy(sub, m.URL, m.Method); err != nil {
		r.log.Error(
			"添加权限策略失败",
			zap.Error(err),
			zap.String(auth.SubKey, sub),
			zap.String(auth.ObjKey, m.URL),
			zap.String(auth.ActKey, m.Method),
			zap.Duration(log.DurationKey, time.Since(startTime)),
		)
		return err
	}
	r.log.Debug(
		"添加权限策略成功",
		zap.Object(database.ModelKey, &m),
		zap.String(auth.SubKey, sub),
		zap.Duration(log.DurationKey, time.Since(startTime)),
	)
	return nil
}

func (r *permissionRepo) RemovePolicy(
	ctx context.Context,
	m biz.PermissionModel,
	removeInherited bool,
) error {
	if err := errors.CheckContext(ctx); err != nil {
		return err
	}
	r.log.Debug(
		"开始删除权限策略",
		zap.Object(database.ModelKey, &m),
	)
	startTime := time.Now()
	sub := auth.PermissionToSubject(m.ID)
	if err := r.cache.RemovePolicy(sub, m.URL, m.Method); err != nil {
		r.log.Error(
			"删除权限策略失败",
			zap.Error(err),
			zap.Object(database.ModelKey, &m),
			zap.String(auth.SubKey, sub),
			zap.String(auth.ObjKey, m.URL),
			zap.String(auth.ActKey, m.Method),
			zap.Duration(log.DurationKey, time.Since(startTime)),
		)
		return err
	}
	r.log.Debug(
		"删除权限策略成功",
		zap.Object(database.ModelKey, &m),
		zap.String(auth.SubKey, sub),
		zap.Duration(log.DurationKey, time.Since(startTime)),
	)
	if removeInherited {
		r.log.Debug(
			"开始删除继承该权限的组策略",
			zap.Object(database.ModelKey, &m),
			zap.Uint32(auth.GroupObjKey, 1),
			zap.String(auth.GroupObjKey, sub),
		)
		if err := r.cache.RemoveGroupPolicy(1, sub); err != nil {
			r.log.Error(
				"删除继承该权限的组策略失败",
				zap.Error(err),
				zap.Object(database.ModelKey, &m),
				zap.String(auth.GroupObjKey, sub),
			)
			return err
		}
		r.log.Debug(
			"删除继承该权限的组策略成功",
			zap.Object(database.ModelKey, &m),
			zap.Uint32(auth.GroupObjKey, 1),
			zap.String(auth.GroupObjKey, sub),
		)
	}
	return nil
}
