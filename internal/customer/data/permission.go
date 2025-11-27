package data

import (
	"context"
	goerrors "errors"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"gin-artweb/internal/customer/biz"
	"gin-artweb/internal/shared/auth"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
	"gin-artweb/internal/shared/log"
)

type permissionRepo struct {
	log      *zap.Logger
	gormDB   *gorm.DB
	timeouts *database.DBTimeout
	cache    *auth.AuthEnforcer
}

func NewPermissionRepo(
	log *zap.Logger,
	gormDB *gorm.DB,
	timeouts *database.DBTimeout,
	cache *auth.AuthEnforcer,
) biz.PermissionRepo {
	return &permissionRepo{
		log:      log,
		gormDB:   gormDB,
		timeouts: timeouts,
		cache:    cache,
	}
}

func (r *permissionRepo) CreateModel(ctx context.Context, m *biz.PermissionModel) error {
	r.log.Debug(
		"开始创建权限模型",
		zap.Object(database.ModelKey, m),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	now := time.Now()
	m.CreatedAt = now
	m.UpdatedAt = now

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	if err := database.DBCreate(dbCtx, r.gormDB, &biz.PermissionModel{}, m, nil); err != nil {
		r.log.Error(
			"创建权限模型失败",
			zap.Error(err),
			zap.Object(database.ModelKey, m),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return err
	}

	r.log.Debug(
		"创建权限模型成功",
		zap.Object(database.ModelKey, m),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(now)),
	)
	return nil
}

func (r *permissionRepo) UpdateModel(ctx context.Context, data map[string]any, conds ...any) error {
	r.log.Debug(
		"开始更新权限模型",
		zap.Any(database.UpdateDataKey, data),
		zap.Any(database.ConditionKey, conds),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	startTime := time.Now()
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	if err := database.DBUpdate(dbCtx, r.gormDB, &biz.PermissionModel{}, data, nil, conds...); err != nil {
		r.log.Error(
			"更新权限模型失败",
			zap.Error(err),
			zap.Any(database.UpdateDataKey, data),
			zap.Any(database.ConditionKey, conds),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(startTime)),
		)
		return err
	}

	r.log.Debug(
		"更新权限模型成功",
		zap.Any(database.UpdateDataKey, data),
		zap.Any(database.ConditionKey, conds),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(startTime)),
	)
	return nil
}

func (r *permissionRepo) DeleteModel(ctx context.Context, conds ...any) error {
	r.log.Debug(
		"开始删除权限模型",
		zap.Any(database.ConditionKey, conds),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	startTime := time.Now()
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	if err := database.DBDelete(dbCtx, r.gormDB, &biz.PermissionModel{}, conds...); err != nil {
		r.log.Error(
			"删除权限模型失败",
			zap.Error(err),
			zap.Any(database.ConditionKey, conds),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(startTime)),
		)
		return err
	}

	r.log.Debug(
		"删除权限模型成功",
		zap.Any(database.ConditionKey, conds),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(startTime)),
	)
	return nil
}

func (r *permissionRepo) FindModel(
	ctx context.Context,
	conds ...any,
) (*biz.PermissionModel, error) {
	r.log.Debug(
		"开始查询权限模型",
		zap.Any(database.ConditionKey, conds),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	startTime := time.Now()
	var m biz.PermissionModel
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()
	if err := database.DBFind(dbCtx, r.gormDB, nil, &m, conds...); err != nil {
		r.log.Error(
			"查询权限模型失败",
			zap.Error(err),
			zap.Any(database.ConditionKey, conds),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(startTime)),
		)
		return nil, err
	}

	r.log.Debug(
		"查询权限模型成功",
		zap.Object(database.ModelKey, &m),
		zap.Any(database.ConditionKey, conds),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
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
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	startTime := time.Now()
	var ms []biz.PermissionModel
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ListTimeout)
	defer cancel()
	count, err := database.DBList(dbCtx, r.gormDB, &biz.PermissionModel{}, &ms, qp)
	if err != nil {
		r.log.Error(
			"查询权限模型列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(startTime)),
		)
		return 0, nil, err
	}
	
	r.log.Debug(
		"查询权限模型列表成功",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(startTime)),
	)
	return count, &ms, nil
}

func (r *permissionRepo) AddPolicy(
	ctx context.Context,
	m biz.PermissionModel,
) error {
	// 检查上下文
	if err := errors.CheckContext(ctx); err != nil {
		return err
	}

	r.log.Debug(
		"AddPolicy: 传入参数",
		zap.Object(database.ModelKey, &m),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	// 检查权限模型的有效性
	if m.ID == 0 {
		return goerrors.New("添加权限策略失败: 权限ID不能为0")
	}
	if m.URL == "" {
		return goerrors.New("添加权限策略失败: URL不能为空")
	}
	if m.Method == "" {
		return goerrors.New("添加权限策略失败: 请求方法不能为空")
	}

	r.log.Debug(
		"开始添加权限策略",
		zap.Object(database.ModelKey, &m),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)
	startTime := time.Now()
	sub := auth.PermissionToSubject(m.ID)
	if err := r.cache.AddPolicy(ctx, sub, m.URL, m.Method); err != nil {
		r.log.Error(
			"添加权限策略失败",
			zap.Error(err),
			zap.String(auth.SubKey, sub),
			zap.String(auth.ObjKey, m.URL),
			zap.String(auth.ActKey, m.Method),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(startTime)),
		)
		return err
	}
	r.log.Debug(
		"添加权限策略成功",
		zap.Object(database.ModelKey, &m),
		zap.String(auth.SubKey, sub),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(startTime)),
	)
	return nil
}

func (r *permissionRepo) RemovePolicy(
	ctx context.Context,
	m biz.PermissionModel,
	removeInherited bool,
) error {
	// 检查上下文
	if err := errors.CheckContext(ctx); err != nil {
		return err
	}

	r.log.Debug(
		"RemoveGroupPolicy: 传入参数",
		zap.Object(database.ModelKey, &m),
		zap.Bool("removeInherited", removeInherited),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	// 检查权限模型的有效性
	if m.ID == 0 {
		return goerrors.New("删除权限策略失败: 权限ID不能为0")
	}
	if m.URL == "" {
		return goerrors.New("删除权限策略失败: URL不能为空")
	}
	if m.Method == "" {
		return goerrors.New("删除权限策略失败: 请求方法不能为空")
	}

	r.log.Debug(
		"开始删除权限策略",
		zap.Object(database.ModelKey, &m),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)
	rmSubStartTime := time.Now()
	sub := auth.PermissionToSubject(m.ID)
	if err := r.cache.RemovePolicy(ctx, sub, m.URL, m.Method); err != nil {
		r.log.Error(
			"删除权限策略失败",
			zap.Error(err),
			zap.Object(database.ModelKey, &m),
			zap.String(auth.SubKey, sub),
			zap.String(auth.ObjKey, m.URL),
			zap.String(auth.ActKey, m.Method),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(rmSubStartTime)),
		)
		return err
	}
	r.log.Debug(
		"删除权限策略成功",
		zap.Object(database.ModelKey, &m),
		zap.String(auth.SubKey, sub),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(rmSubStartTime)),
	)

	// 如果需要删除继承该权限的组策略
	if removeInherited {
		rmObjStartTime := time.Now()
		r.log.Debug(
			"开始删除继承该权限的组策略",
			zap.Object(database.ModelKey, &m),
			zap.String(auth.GroupObjKey, sub),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(rmObjStartTime)),
		)
		if err := r.cache.RemoveGroupPolicy(ctx, 1, sub); err != nil {
			r.log.Error(
				"删除继承该权限的组策略失败",
				zap.Error(err),
				zap.Object(database.ModelKey, &m),
				zap.String(auth.GroupObjKey, sub),
				zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
				zap.Duration(log.DurationKey, time.Since(rmObjStartTime)),
			)
			return err
		}
		r.log.Debug(
			"删除继承该权限的组策略成功",
			zap.Object(database.ModelKey, &m),
			zap.String(auth.GroupObjKey, sub),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(rmObjStartTime)),
		)
	}
	return nil
}
