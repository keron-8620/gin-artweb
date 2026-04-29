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
	log           *zap.Logger
	gormDB        *gorm.DB
	timeouts      *config.DBTimeout
	slowThreshold *config.DBSlowThreshold
}

func NewMdsNodeRepo(
	log *zap.Logger,
	gormDB *gorm.DB,
	timeouts *config.DBTimeout,
	slowThreshold *config.DBSlowThreshold,
) *MdsNodeRepo {
	return &MdsNodeRepo{
		log:           log,
		gormDB:        gormDB,
		timeouts:      timeouts,
		slowThreshold: slowThreshold,
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
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return err
	}

	m.CreatedAt = startTime
	m.UpdatedAt = startTime

	log.Debug(
		"创建mds节点:模型详情",
		zap.Object("mds_node_model", m),
	)

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()

	createStartTime := time.Now()
	err := database.DBCreate(dbCtx, r.gormDB, &mdsmodel.MdsNodeModel{}, m)
	createDuration := time.Since(createStartTime)
	if err != nil {
		log.Error(
			"创建mds节点:数据库操作失败",
			zap.Error(err),
			zap.Object("mds_node_model", m),
			zap.Duration("create_duration", createDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "创建mds节点:数据库操作失败")
	}

	if createDuration > r.slowThreshold.WriteSlow {
		log.Warn("创建mds节点模型:数据库创建耗时超过慢查询阈值，可能影响性能",
			zap.Uint32("node_id", m.ID),
			zap.Duration("create_duration", createDuration),
			zap.Duration("threshold", r.slowThreshold.WriteSlow),
		)
	}
	return nil
}

func (r *MdsNodeRepo) UpdateModel(
	ctx context.Context,
	updateData map[string]any,
	conds ...any,
) error {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	if len(updateData) == 0 {
		err := errors.New("更新mds节点:更新数据为空")
		log.Error(
			"更新mds节点:更新数据为空",
			zap.Error(err),
			zap.Any("update_data", updateData),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return err
	}

	updateData["updated_at"] = startTime

	log.Debug(
		"更新mds集群:更新数据",
		zap.Any("update_data", updateData),
		zap.Any("conds", conds),
	)

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()

	updateStartTime := time.Now()
	err := database.DBUpdateTx(dbCtx, r.gormDB, &mdsmodel.MdsNodeModel{}, updateData, conds...)
	updateDuration := time.Since(updateStartTime)
	if err != nil {
		log.Error(
			"更新mds节点:数据库操作失败",
			zap.Error(err),
			zap.Any("update_data", updateData),
			zap.Any("conds", conds),
			zap.Duration("update_duration", updateDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "更新mds节点:数据库操作失败")
	}

	if updateDuration > r.slowThreshold.WriteSlow {
		log.Warn("更新mds节点:数据库更新耗时超过慢查询阈值，可能影响性能",
			zap.Duration("update_duration", updateDuration),
			zap.Duration("threshold", r.slowThreshold.WriteSlow),
		)
	}
	return nil
}

func (r *MdsNodeRepo) DeleteModel(
	ctx context.Context,
	conds ...any,
) error {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"删除mds节点:删除条件",
		zap.Any("conds", conds),
	)

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()

	deleteStartTime := time.Now()
	err := database.DBDeleteTx(dbCtx, r.gormDB, &mdsmodel.MdsNodeModel{}, conds...)
	deleteDuration := time.Since(deleteStartTime)
	if err != nil {
		log.Error(
			"删除mds节点:数据库操作失败",
			zap.Error(err),
			zap.Any("conds", conds),
			zap.Duration("delete_duration", deleteDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "删除mds节点:数据库操作失败")
	}

	if deleteDuration > r.slowThreshold.WriteSlow {
		log.Warn("删除mds节点:数据库删除耗时超过慢查询阈值，可能影响性能",
			zap.Duration("delete_duration", deleteDuration),
			zap.Duration("threshold", r.slowThreshold.WriteSlow),
		)
	}
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
		"查询mds节点:查询条件",
		zap.Any("conds", conds),
	)

	var m mdsmodel.MdsNodeModel

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()

	getStartTime := time.Now()
	err := database.DBGet(dbCtx, r.gormDB, preloads, &m, conds...)
	getDuration := time.Since(getStartTime)
	if err != nil {
		log.Error(
			"查询mds节点:数据库操作失败",
			zap.Error(err),
			zap.Any("conds", conds),
			zap.Duration("get_duration", getDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.WrapIf(err, "查询mds节点:数据库操作失败")
	}

	log.Debug(
		"查询mds节点:查询到的模型详情",
		zap.Object("mds_node_model", &m),
	)

	if getDuration > r.slowThreshold.ReadSlow {
		log.Warn("查询mds节点:数据库查询耗时超过慢查询阈值，可能影响性能",
			zap.Duration("get_duration", getDuration),
			zap.Duration("threshold", r.slowThreshold.ReadSlow),
		)
	}
	return &m, nil
}

func (r *MdsNodeRepo) ListModel(
	ctx context.Context,
	qp database.QueryParams,
) ([]mdsmodel.MdsNodeModel, error) {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"查询mds节点列表:入参详情",
		zap.Object("query_params", &qp),
	)

	var ms []mdsmodel.MdsNodeModel

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ListTimeout)
	defer cancel()

	listStartTime := time.Now()
	err := database.DBList(dbCtx, r.gormDB, &mdsmodel.MdsNodeModel{}, &ms, qp)
	listDuration := time.Since(listStartTime)
	if err != nil {
		log.Error(
			"查询mds节点列表:数据库操作失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("list_duration", listDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.WrapIf(err, "查询mds节点列表:数据库操作失败")
	}

	log.Debug(
		"查询mds节点列表:查询到的模型列表",
		zap.Uint32s("node_ids", mdsmodel.ListMdsNodeModelToUint32s(ms)),
	)

	if listDuration > r.slowThreshold.ListSlow {
		log.Warn("查询mds节点列表:数据库查询耗时超过慢查询阈值，可能影响性能",
			zap.Duration("list_duration", listDuration),
			zap.Duration("threshold", r.slowThreshold.ListSlow),
		)
	}
	return ms, nil
}

func (r *MdsNodeRepo) CountModel(
	ctx context.Context,
	query map[string]any,
) (int64, error) {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"查询mds节点总数:查询条件",
		zap.Any("query", query),
	)

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()

	countStartTime := time.Now()
	count, err := database.DBCount(dbCtx, r.gormDB, &mdsmodel.MdsNodeModel{}, query)
	countDuration := time.Since(countStartTime)
	if err != nil {
		log.Error(
			"查询mds节点总数:数据库查询失败",
			zap.Error(err),
			zap.Any("query", query),
			zap.Duration("count_duration", countDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, errors.WrapIf(err, "查询mds节点总数:数据库查询失败")
	}

	log.Debug(
		"查询mds节点总数:查询到的记录数",
		zap.Int64("count", count),
	)

	if countDuration > r.slowThreshold.ReadSlow {
		log.Warn("查询mds节点总数:数据库查询耗时超过慢查询阈值，可能影响性能",
			zap.Duration("count_duration", countDuration),
			zap.Duration("threshold", r.slowThreshold.ReadSlow),
		)
	}
	return count, nil
}
