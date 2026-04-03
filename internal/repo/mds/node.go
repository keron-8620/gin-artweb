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

func (r *MdsNodeRepo) CreateModel(
	ctx context.Context,
	m *mdsmodel.MdsNodeModel,
) error {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	// 检查参数
	if m == nil {
		err := errors.New("创建mds节点:模型不能为空")
		log.Error(
			"创建mds节点:模型不能为空",
			zap.Error(err),
		)
		return err
	}
	log.Debug(
		"创建mds节点:开始执行",
		zap.Object("model", m),
	)

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	createMdsNodeStartTime := time.Now()
	err := database.DBCreate(dbCtx, r.gormDB, &mdsmodel.MdsNodeModel{}, m, nil)
	createMdsNodeDuration := time.Since(createMdsNodeStartTime)
	if err != nil {
		log.Error(
			"创建mds节点:数据库操作失败",
			zap.Error(err),
			zap.Object("node_model", m),
			zap.Duration("create_mds_node_duration", createMdsNodeDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "创建mds节点:数据库操作失败")
	}
	log.Debug(
		"创建mds节点:执行成功",
		zap.Object("node_model", m),
		zap.Duration("create_mds_node_duration", createMdsNodeDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

func (r *MdsNodeRepo) UpdateModel(
	ctx context.Context,
	data map[string]any,
	conds ...any,
) error {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	// 检查参数
	if len(data) == 0 {
		err := errors.New("更新mds节点:更新数据不能为空")
		log.Error(
			"更新mds节点:更新数据不能为空",
			zap.Error(err),
			zap.Any("update_data", data),
			zap.Any("conds", conds),
		)
		return err
	}

	log.Debug(
		"更新mds节点:开始执行",
		zap.Any("update_data", data),
		zap.Any("conds", conds),
	)

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	updateMdsNodeStartTime := time.Now()
	err := database.DBUpdate(dbCtx, r.gormDB, &mdsmodel.MdsNodeModel{}, data, nil, conds...)
	updateMdsNodeDuration := time.Since(updateMdsNodeStartTime)
	if err != nil {
		log.Error(
			"更新mds节点:数据库操作失败",
			zap.Error(err),
			zap.Any("update_data", data),
			zap.Any("conds", conds),
			zap.Duration("update_mds_node_duration", updateMdsNodeDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "更新mds节点:数据库操作失败")
	}
	log.Debug(
		"更新mds节点:执行成功",
		zap.Any("update_data", data),
		zap.Any("conds", conds),
		zap.Duration("update_mds_node_duration", updateMdsNodeDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

func (r *MdsNodeRepo) DeleteModel(
	ctx context.Context,
	conds ...any,
) error {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"删除mds节点:开始执行",
		zap.Any("conds", conds),
	)

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	deleteMdsNodeStartTime := time.Now()
	err := database.DBDelete(dbCtx, r.gormDB, &mdsmodel.MdsNodeModel{}, conds...)
	deleteMdsNodeDuration := time.Since(deleteMdsNodeStartTime)
	if err != nil {
		log.Error(
			"删除mds节点:数据库操作失败",
			zap.Error(err),
			zap.Any("conds", conds),
			zap.Duration("delete_mds_node_duration", deleteMdsNodeDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "删除mds节点:数据库操作失败")
	}
	log.Debug(
		"删除mds节点:执行成功",
		zap.Any("conds", conds),
		zap.Duration("delete_mds_node_duration", deleteMdsNodeDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

func (r *MdsNodeRepo) GetModel(
	ctx context.Context,
	preloads []string,
	conds ...any,
) (*mdsmodel.MdsNodeModel, error) {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)
	log.Debug(
		"查询mds节点:开始执行",
		zap.Any("conds", conds),
	)
	var m mdsmodel.MdsNodeModel
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()
	getMdsNodeStartTime := time.Now()
	err := database.DBGet(dbCtx, r.gormDB, preloads, &m, conds...)
	getMdsNodeDuration := time.Since(getMdsNodeStartTime)
	if err != nil {
		log.Error(
			"查询mds节点:数据库操作失败",
			zap.Error(err),
			zap.Any("conds", conds),
			zap.Duration("get_mds_node_duration", getMdsNodeDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.WrapIf(err, "查询mds节点:数据库操作失败")
	}
	log.Debug(
		"查询mds节点:执行成功",
		zap.Object("node_model", &m),
		zap.Any("conds", conds),
		zap.Duration("get_mds_node_duration", getMdsNodeDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return &m, nil
}

func (r *MdsNodeRepo) ListModel(
	ctx context.Context,
	qp database.QueryParams,
) ([]mdsmodel.MdsNodeModel, error) {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)
	log.Debug(
		"查询mds节点列表:开始执行",
		zap.Object("query_params", &qp),
	)
	var ms []mdsmodel.MdsNodeModel
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ListTimeout)
	defer cancel()
	listMdsNodeStartTime := time.Now()
	err := database.DBList(dbCtx, r.gormDB, &mdsmodel.MdsNodeModel{}, &ms, qp)
	listMdsNodeDuration := time.Since(listMdsNodeStartTime)
	if err != nil {
		log.Error(
			"查询mds节点列表:数据库操作失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("list_mds_node_duration", listMdsNodeDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.WrapIf(err, "查询mds节点列表:数据库操作失败")
	}
	log.Debug(
		"查询mds节点列表:执行成功",
		zap.Object("query_params", &qp),
		zap.Duration("list_mds_node_duration", listMdsNodeDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return ms, nil
}

func (r *MdsNodeRepo) CountModel(
	ctx context.Context,
	query map[string]any,
) (int64, error) {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"查询mds节点总数:开始执行",
		zap.Any("query", query),
	)

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()
	countMdsNodeStartTime := time.Now()
	count, err := database.DBCount(dbCtx, r.gormDB, &mdsmodel.MdsNodeModel{}, query)
	countMdsNodeDuration := time.Since(countMdsNodeStartTime)
	if err != nil {
		log.Error(
			"查询mds节点总数:数据库查询失败",
			zap.Error(err),
			zap.Any("query", query),
			zap.Duration("count_mds_node_duration", countMdsNodeDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, errors.WrapIf(err, "查询mds节点总数:数据库查询失败")
	}
	log.Debug(
		"查询mds节点总数:执行成功",
		zap.Any("query", query),
		zap.Int64("count", count),
		zap.Duration("count_mds_node_duration", countMdsNodeDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return count, nil
}
