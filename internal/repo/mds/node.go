package mds

import (
	"context"
	"time"

	"emperror.dev/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"

	mdsmodel "gin-artweb/internal/model/mds"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/database"
)

type MdsNodeRepo struct {
	log      *zap.Logger
	gormDB   *gorm.DB
	timeouts *config.DBTimeout
}

func NewMdsNodeRepo(
	log *zap.Logger,
	gormDB *gorm.DB,
	timeouts *config.DBTimeout,
) *MdsNodeRepo {
	return &MdsNodeRepo{
		log:      log,
		gormDB:   gormDB,
		timeouts: timeouts,
	}
}

func (r *MdsNodeRepo) CreateModel(ctx context.Context, m *mdsmodel.MdsNodeModel) error {
	log := ctxutil.NewLogger(r.log, ctx)

	// 检查参数
	if m == nil {
		err := errors.New("创建mds节点：模型不能为空")
		log.Error(
			"创建mds节点：模型不能为空",
			zap.Error(err),
		)
		return err
	}
	log.Debug(
		"创建mds节点：开始执行",
		zap.Object("model", m),
	)
	now := time.Now()
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	if err := database.DBCreate(dbCtx, r.gormDB, &mdsmodel.MdsNodeModel{}, m, nil); err != nil {
		log.Error(
			"创建mds节点：数据库操作失败",
			zap.Error(err),
			zap.Object("node_model", m),
			zap.Duration("create_duration", time.Since(now)),
		)
		return errors.WrapIf(err, "创建mds节点：数据库操作失败")
	}
	log.Debug(
		"创建mds节点：执行成功",
		zap.Object("node_model", m),
		zap.Duration("create_duration", time.Since(now)),
	)
	return nil
}

func (r *MdsNodeRepo) UpdateModel(ctx context.Context, data map[string]any, conds ...any) error {
	log := ctxutil.NewLogger(r.log, ctx)

	// 检查参数
	if len(data) == 0 {
		err := errors.New("更新mds节点：更新数据不能为空")
		log.Error(
			"更新mds节点：更新数据不能为空",
			zap.Error(err),
			zap.Any("update_data", data),
			zap.Any("conds", conds),
		)
		return err
	}

	log.Debug(
		"更新mds节点：开始执行",
		zap.Any("update_data", data),
		zap.Any("conds", conds),
	)
	startTime := time.Now()
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	if err := database.DBUpdate(dbCtx, r.gormDB, &mdsmodel.MdsNodeModel{}, data, nil, conds...); err != nil {
		log.Error(
			"更新mds节点：数据库操作失败",
			zap.Error(err),
			zap.Any("update_data", data),
			zap.Any("conds", conds),
			zap.Duration("update_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "更新mds节点：数据库操作失败")
	}
	log.Debug(
		"更新mds节点：执行成功",
		zap.Any("update_data", data),
		zap.Any("conds", conds),
		zap.Duration("update_duration", time.Since(startTime)),
	)
	return nil
}

func (r *MdsNodeRepo) DeleteModel(ctx context.Context, conds ...any) error {
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"删除mds节点：开始执行",
		zap.Any("conds", conds),
	)
	startTime := time.Now()
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	if err := database.DBDelete(dbCtx, r.gormDB, &mdsmodel.MdsNodeModel{}, conds...); err != nil {
		log.Error(
			"删除mds节点：数据库操作失败",
			zap.Error(err),
			zap.Any("conds", conds),
			zap.Duration("delete_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "删除mds节点：数据库操作失败")
	}
	log.Debug(
		"删除mds节点：执行成功",
		zap.Any("conds", conds),
		zap.Duration("delete_duration", time.Since(startTime)),
	)
	return nil
}

func (r *MdsNodeRepo) GetModel(
	ctx context.Context,
	preloads []string,
	conds ...any,
) (*mdsmodel.MdsNodeModel, error) {
	log := ctxutil.NewLogger(r.log, ctx)
	log.Debug(
		"查询mds节点：开始执行",
		zap.Any("conds", conds),
	)
	startTime := time.Now()
	var m mdsmodel.MdsNodeModel
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()
	if err := database.DBGet(dbCtx, r.gormDB, preloads, &m, conds...); err != nil {
		log.Error(
			"查询mds节点：数据库操作失败",
			zap.Error(err),
			zap.Any("conds", conds),
			zap.Duration("get_duration", time.Since(startTime)),
		)
		return nil, errors.WrapIf(err, "查询mds节点：数据库操作失败")
	}
	log.Debug(
		"查询mds节点：执行成功",
		zap.Object("node_model", &m),
		zap.Any("conds", conds),
		zap.Duration("get_duration", time.Since(startTime)),
	)
	return &m, nil
}

func (r *MdsNodeRepo) ListModel(
	ctx context.Context,
	qp database.QueryParams,
) ([]mdsmodel.MdsNodeModel, error) {
	log := ctxutil.NewLogger(r.log, ctx)
	log.Debug(
		"查询mds节点列表：开始执行",
		zap.Object("query_params", &qp),
	)
	startTime := time.Now()
	var ms []mdsmodel.MdsNodeModel
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ListTimeout)
	defer cancel()
	if err := database.DBList(dbCtx, r.gormDB, &mdsmodel.MdsNodeModel{}, &ms, qp); err != nil {
		log.Error(
			"查询mds节点列表：数据库操作失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("list_duration", time.Since(startTime)),
		)
		return nil, errors.WrapIf(err, "查询mds节点列表：数据库操作失败")
	}
	log.Debug(
		"查询mds节点列表：执行成功",
		zap.Object("query_params", &qp),
		zap.Duration("list_duration", time.Since(startTime)),
	)
	return ms, nil
}

func (r *MdsNodeRepo) CountModel(
	ctx context.Context,
	query map[string]any,
) (int64, error) {
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"查询mds节点总数：开始执行",
		zap.Any("query", query),
	)
	now := time.Now()
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()
	count, err := database.DBCount(dbCtx, r.gormDB, &mdsmodel.MdsNodeModel{}, query)
	if err != nil {
		log.Error(
			"查询mds节点总数：数据库查询失败",
			zap.Error(err),
			zap.Any("query", query),
			zap.Duration("count_duration", time.Since(now)),
		)
		return 0, errors.WrapIf(err, "查询mds节点总数：数据库查询失败")
	}
	log.Debug(
		"查询mds节点总数：执行成功",
		zap.Any("query", query),
		zap.Int64("count", count),
		zap.Duration("count_duration", time.Since(now)),
	)
	return count, nil
}
