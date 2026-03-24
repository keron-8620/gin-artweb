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

type OesCronRepo struct {
	log      *zap.Logger
	gormDB   *gorm.DB
	timeouts *config.DBTimeout
}

func NewOesCronRepo(
	log *zap.Logger,
	gormDB *gorm.DB,
	timeouts *config.DBTimeout,
) *OesCronRepo {
	return &OesCronRepo{
		log:      log,
		gormDB:   gormDB,
		timeouts: timeouts,
	}
}

func (r *OesCronRepo) CreateModel(ctx context.Context, m *oesmodel.OesCronModel) error {
	log := ctxutil.NewLogger(r.log, ctx)

	// 检查参数
	if m == nil {
		err := errors.New("创建oes计划任务：模型不能为空")
		log.Error(
			"创建oes计划任务：模型不能为空",
			zap.Error(err),
		)
		return err
	}
	log.Debug(
		"创建oes计划任务：开始执行",
		zap.Object("node_model", m),
	)
	now := time.Now()
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	if err := database.DBCreate(dbCtx, r.gormDB, &oesmodel.OesCronModel{}, m, nil); err != nil {
		log.Error(
			"创建oes计划任务：数据库操作失败",
			zap.Error(err),
			zap.Object("node_model", m),
			zap.Duration("create_duration", time.Since(now)),
		)
		return errors.WrapIf(err, "创建oes计划任务失败")
	}
	log.Debug(
		"创建oes计划任务：执行成功",
		zap.Object("node_model", m),
		zap.Duration("create_duration", time.Since(now)),
	)
	return nil
}

func (r *OesCronRepo) UpdateModel(ctx context.Context, data map[string]any, conds ...any) error {
	log := ctxutil.NewLogger(r.log, ctx)

	// 检查参数
	if len(data) == 0 {
		err := errors.New("更新oes计划任务：更新数据为空")
		log.Error(
			"更新oes计划任务：更新数据为空",
			zap.Error(err),
			zap.Any("conds", conds),
		)
		return err
	}
	log.Debug(
		"更新oes计划任务：开始执行",
		zap.Any("update_data", data),
		zap.Any("conds", conds),
	)
	startTime := time.Now()
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	if err := database.DBUpdate(dbCtx, r.gormDB, &oesmodel.OesCronModel{}, data, nil, conds...); err != nil {
		log.Error(
			"更新oes计划任务：数据库操作失败",
			zap.Error(err),
			zap.Any("update_data", data),
			zap.Any("conds", conds),
			zap.Duration("update_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "更新oes计划任务失败")
	}
	log.Debug(
		"更新oes计划任务：执行成功",
		zap.Any("update_data", data),
		zap.Any("conds", conds),
		zap.Duration("update_duration", time.Since(startTime)),
	)
	return nil
}

func (r *OesCronRepo) DeleteModel(ctx context.Context, conds ...any) error {
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"删除oes计划任务：开始执行",
		zap.Any("conds", conds),
	)
	startTime := time.Now()
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	if err := database.DBDelete(dbCtx, r.gormDB, &oesmodel.OesCronModel{}, conds...); err != nil {
		log.Error(
			"删除oes计划任务：数据库操作失败",
			zap.Error(err),
			zap.Any("conds", conds),
			zap.Duration("delete_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "删除oes计划任务失败")
	}
	log.Debug(
		"删除oes计划任务：执行成功",
		zap.Any("conds", conds),
		zap.Duration("delete_duration", time.Since(startTime)),
	)
	return nil
}

func (r *OesCronRepo) GetModel(
	ctx context.Context,
	preloads []string,
	conds ...any,
) (*oesmodel.OesCronModel, error) {
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"查询oes计划任务：开始执行",
		zap.Any("conds", conds),
	)
	startTime := time.Now()
	var m oesmodel.OesCronModel
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()
	if err := database.DBGet(dbCtx, r.gormDB, preloads, &m, conds...); err != nil {
		log.Error(
			"查询oes计划任务失败",
			zap.Error(err),
			zap.Any("conds", conds),
			zap.Duration("get_duration", time.Since(startTime)),
		)
		return nil, errors.WrapIf(err, "查询oes计划任务失败")
	}
	log.Debug(
		"查询oes计划任务：执行成功",
		zap.Object("node_model", &m),
		zap.Any("conds", conds),
		zap.Duration("get_duration", time.Since(startTime)),
	)
	return &m, nil
}

func (r *OesCronRepo) ListModel(
	ctx context.Context,
	qp database.QueryParams,
) ([]oesmodel.OesCronModel, error) {
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"查询oes计划任务列表：开始执行",
		zap.Object("query_params", &qp),
	)
	startTime := time.Now()
	var ms []oesmodel.OesCronModel
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ListTimeout)
	defer cancel()
	if err := database.DBList(dbCtx, r.gormDB, &oesmodel.OesCronModel{}, &ms, qp); err != nil {
		log.Error(
			"查询oes计划任务列表：数据库操作失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("list_duration", time.Since(startTime)),
		)
		return nil, errors.WrapIf(err, "查询oes计划任务列表失败")
	}
	log.Debug(
		"查询oes计划任务列表：执行成功",
		zap.Object("query_params", &qp),
		zap.Duration("list_duration", time.Since(startTime)),
	)
	return ms, nil
}

func (r *OesCronRepo) CountModel(
	ctx context.Context,
	query map[string]any,
) (int64, error) {
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"查询oes计划任务总数：开始执行",
		zap.Any("query", query),
	)
	startTime := time.Now()
	var count int64
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()
	count, err := database.DBCount(dbCtx, r.gormDB, &oesmodel.OesCronModel{}, query)
	if err != nil {
		log.Error(
			"查询oes计划任务总数：数据库操作失败",
			zap.Error(err),
			zap.Any("query", query),
			zap.Duration("count_duration", time.Since(startTime)),
		)
		return 0, errors.WrapIf(err, "查询oes计划任务总数失败")
	}
	log.Debug(
		"查询oes计划任务总数：执行成功",
		zap.Any("query", query),
		zap.Int64("count", count),
		zap.Duration("count_duration", time.Since(startTime)),
	)
	return count, nil
}
