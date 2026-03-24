package oes

import (
	"context"
	"time"

	"emperror.dev/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"

	oesmodel "gin-artweb/internal/model/oes"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/database"
)

type OesNodeRepo struct {
	log      *zap.Logger
	gormDB   *gorm.DB
	timeouts *config.DBTimeout
}

func NewOesNodeRepo(
	log *zap.Logger,
	gormDB *gorm.DB,
	timeouts *config.DBTimeout,
) *OesNodeRepo {
	return &OesNodeRepo{
		log:      log,
		gormDB:   gormDB,
		timeouts: timeouts,
	}
}

func (r *OesNodeRepo) CreateModel(ctx context.Context, m *oesmodel.OesNodeModel) error {
	log := ctxutil.NewLogger(r.log, ctx)

	// 检查参数
	if m == nil {
		err := errors.New("创建oes节点：模型不能为空")
		log.Error(
			"创建oes节点：模型不能为空",
			zap.Error(err),
		)
		return err
	}
	log.Debug(
		"创建oes节点：开始执行",
		zap.Object("node_model", m),
	)
	now := time.Now()
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	if err := database.DBCreate(dbCtx, r.gormDB, &oesmodel.OesNodeModel{}, m, nil); err != nil {
		log.Error(
			"创建oes节点：数据库操作失败",
			zap.Error(err),
			zap.Object("node_model", m),
			zap.Duration("create_duration", time.Since(now)),
		)
		return errors.WrapIf(err, "创建oes节点失败")
	}
	log.Debug(
		"创建oes节点：执行成功",
		zap.Object("node_model", m),
		zap.Duration("create_duration", time.Since(now)),
	)
	return nil
}

func (r *OesNodeRepo) UpdateModel(ctx context.Context, data map[string]any, conds ...any) error {
	log := ctxutil.NewLogger(r.log, ctx)

	// 检查参数
	if len(data) == 0 {
		err := errors.New("更新oes节点：更新数据为空")
		log.Error(
			"更新oes节点：更新数据为空",
			zap.Error(err),
			zap.Any("conds", conds),
		)
		return err
	}
	log.Debug(
		"更新oes节点：开始执行",
		zap.Any("update_data", data),
		zap.Any("conds", conds),
	)
	startTime := time.Now()
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	if err := database.DBUpdate(dbCtx, r.gormDB, &oesmodel.OesNodeModel{}, data, nil, conds...); err != nil {
		log.Error(
			"更新oes节点：数据库操作失败",
			zap.Error(err),
			zap.Any("update_data", data),
			zap.Any("conds", conds),
			zap.Duration("update_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "更新oes节点失败")
	}
	log.Debug(
		"更新oes节点：执行成功",
		zap.Any("update_data", data),
		zap.Any("conds", conds),
		zap.Duration("update_duration", time.Since(startTime)),
	)
	return nil
}

func (r *OesNodeRepo) DeleteModel(ctx context.Context, conds ...any) error {
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"删除oes节点：开始执行",
		zap.Any("conds", conds),
	)
	startTime := time.Now()
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	if err := database.DBDelete(dbCtx, r.gormDB, &oesmodel.OesNodeModel{}, conds...); err != nil {
		log.Error(
			"删除oes节点：数据库操作失败",
			zap.Error(err),
			zap.Any("conds", conds),
			zap.Duration("delete_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "删除oes节点失败")
	}
	log.Debug(
		"删除oes节点：执行成功",
		zap.Any("conds", conds),
		zap.Duration("delete_duration", time.Since(startTime)),
	)
	return nil
}

func (r *OesNodeRepo) GetModel(
	ctx context.Context,
	preloads []string,
	conds ...any,
) (*oesmodel.OesNodeModel, error) {
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"查询oes节点：开始执行",
		zap.Any("conds", conds),
	)
	startTime := time.Now()
	var m oesmodel.OesNodeModel
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()
	if err := database.DBGet(dbCtx, r.gormDB, preloads, &m, conds...); err != nil {
		log.Error(
			"查询oes节点失败",
			zap.Error(err),
			zap.Any("conds", conds),
			zap.Duration("get_duration", time.Since(startTime)),
		)
		return nil, errors.WrapIf(err, "查询oes节点失败")
	}
	log.Debug(
		"查询oes节点：执行成功",
		zap.Object("node_model", &m),
		zap.Any("conds", conds),
		zap.Duration("get_duration", time.Since(startTime)),
	)
	return &m, nil
}

func (r *OesNodeRepo) ListModel(
	ctx context.Context,
	qp database.QueryParams,
) ([]oesmodel.OesNodeModel, error) {
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"查询oes节点列表：开始执行",
		zap.Object("query_params", &qp),
	)
	startTime := time.Now()
	var ms []oesmodel.OesNodeModel
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ListTimeout)
	defer cancel()
	if err := database.DBList(dbCtx, r.gormDB, &oesmodel.OesNodeModel{}, &ms, qp); err != nil {
		log.Error(
			"查询oes节点列表：数据库操作失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("list_duration", time.Since(startTime)),
		)
		return nil, errors.WrapIf(err, "查询oes节点列表失败")
	}
	log.Debug(
		"查询oes节点列表：执行成功",
		zap.Object("query_params", &qp),
		zap.Duration("list_duration", time.Since(startTime)),
	)
	return ms, nil
}

func (r *OesNodeRepo) CountModel(
	ctx context.Context,
	query map[string]any,
) (int64, error) {
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"查询oes节点总数：开始执行",
		zap.Any("query", query),
	)
	startTime := time.Now()
	var count int64
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()
	count, err := database.DBCount(dbCtx, r.gormDB, &oesmodel.OesNodeModel{}, query)
	if err != nil {
		log.Error(
			"查询oes节点总数：数据库操作失败",
			zap.Error(err),
			zap.Any("query", query),
			zap.Duration("count_duration", time.Since(startTime)),
		)
		return 0, errors.WrapIf(err, "查询oes节点总数失败")
	}
	log.Debug(
		"查询oes节点总数：执行成功",
		zap.Any("query", query),
		zap.Int64("count", count),
		zap.Duration("count_duration", time.Since(startTime)),
	)
	return count, nil
}
