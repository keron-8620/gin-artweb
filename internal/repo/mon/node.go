package mon

import (
	"context"
	"time"

	"emperror.dev/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"

	monmodel "gin-artweb/internal/model/mon"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/database"
)

type MonNodeRepo struct {
	log      *zap.Logger
	gormDB   *gorm.DB
	timeouts *config.DBTimeout
}

func NewMonNodeRepo(
	log *zap.Logger,
	gormDB *gorm.DB,
	timeouts *config.DBTimeout,
) *MonNodeRepo {
	return &MonNodeRepo{
		log:      log,
		gormDB:   gormDB,
		timeouts: timeouts,
	}
}

func (r *MonNodeRepo) CreateModel(
	ctx context.Context,
	m *monmodel.MonNodeModel,
) error {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	// 检查参数
	if m == nil {
		err := errors.New("创建mon模型:模型不能为空")
		log.Error(
			"创建mon模型:模型不能为空",
			zap.Error(err),
		)
		return err
	}

	log.Debug(
		"创建mon模型:开始执行",
		zap.Object("mon_model", m),
	)

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	createMonNodeStartTime := time.Now()
	err := database.DBCreate(dbCtx, r.gormDB, &monmodel.MonNodeModel{}, m, nil)
	createMonNodeDuration := time.Since(createMonNodeStartTime)
	if err != nil {
		log.Error(
			"创建mon模型:数据库操作失败",
			zap.Error(err),
			zap.Object("mon_model", m),
			zap.Duration("create_duration", createMonNodeDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "创建mon模型:数据库操作失败")
	}
	log.Debug(
		"创建mon模型:执行成功",
		zap.Object("mon_model", m),
		zap.Duration("create_duration", createMonNodeDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

func (r *MonNodeRepo) UpdateModel(
	ctx context.Context,
	data map[string]any,
	conds ...any,
) error {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	// 检查参数
	if len(data) == 0 {
		err := errors.New("更新mon模型:更新数据为空")
		log.Error(
			"更新mon模型:更新数据为空",
			zap.Error(err),
			zap.Any("update_data", data),
			zap.Any("conds", conds),
		)
		return err
	}

	log.Debug(
		"更新mon模型:开始执行",
		zap.Any("update_data", data),
		zap.Any("conds", conds),
	)
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	updateMonNodeStartTime := time.Now()
	err := database.DBUpdate(dbCtx, r.gormDB, &monmodel.MonNodeModel{}, data, nil, conds...)
	updateMonNodeDuration := time.Since(updateMonNodeStartTime)
	if err != nil {
		log.Error(
			"更新mon模型:数据库操作失败",
			zap.Error(err),
			zap.Any("update_data", data),
			zap.Any("conds", conds),
			zap.Duration("update_duration", updateMonNodeDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "更新mon模型:数据库操作失败")
	}
	log.Debug(
		"更新mon模型:执行成功",
		zap.Any("update_data", data),
		zap.Any("conds", conds),
		zap.Duration("update_duration", updateMonNodeDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

func (r *MonNodeRepo) DeleteModel(
	ctx context.Context,
	conds ...any,
) error {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"删除mon模型:开始执行",
		zap.Any("conds", conds),
	)
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	deleteMonNodeStartTime := time.Now()
	err := database.DBDelete(dbCtx, r.gormDB, &monmodel.MonNodeModel{}, conds...)
	deleteMonNodeDuration := time.Since(deleteMonNodeStartTime)
	if err != nil {
		log.Error(
			"删除mon模型:数据库操作失败",
			zap.Error(err),
			zap.Any("conds", conds),
			zap.Duration("delete_duration", deleteMonNodeDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "删除mon模型:数据库操作失败")
	}
	log.Debug(
		"删除mon模型:执行成功",
		zap.Any("conds", conds),
		zap.Duration("delete_duration", deleteMonNodeDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

func (r *MonNodeRepo) GetModel(
	ctx context.Context,
	preloads []string,
	conds ...any,
) (*monmodel.MonNodeModel, error) {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"查询mon模型:开始执行",
		zap.Any("conds", conds),
	)
	var m monmodel.MonNodeModel
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()
	getMonNodeStartTime := time.Now()
	err := database.DBGet(dbCtx, r.gormDB, preloads, &m, conds...)
	getMonNodeDuration := time.Since(getMonNodeStartTime)
	if err != nil {
		log.Error(
			"查询mon模型:数据库操作失败",
			zap.Error(err),
			zap.Any("conds", conds),
			zap.Duration("get_duration", getMonNodeDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.WrapIf(err, "查询mon模型:数据库操作失败")
	}
	log.Debug(
		"查询mon模型:执行成功",
		zap.Object("mon_model", &m),
		zap.Any("conds", conds),
		zap.Duration("get_duration", getMonNodeDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return &m, nil
}

func (r *MonNodeRepo) ListModel(
	ctx context.Context,
	qp database.QueryParams,
) ([]monmodel.MonNodeModel, error) {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"查询mon模型列表:开始执行",
		zap.Object("query_params", &qp),
	)
	var ms []monmodel.MonNodeModel
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ListTimeout)
	defer cancel()
	listMonNodeStartTime := time.Now()
	err := database.DBList(dbCtx, r.gormDB, &monmodel.MonNodeModel{}, &ms, qp)
	listMonNodeDuration := time.Since(listMonNodeStartTime)
	if err != nil {
		log.Error(
			"查询mon模型列表:数据库操作失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("list_duration", listMonNodeDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.WrapIf(err, "查询mon模型列表:数据库操作失败")
	}
	log.Debug(
		"查询mon模型列表:执行成功",
		zap.Object("query_params", &qp),
		zap.Duration("list_duration", listMonNodeDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return ms, nil
}

func (r *MonNodeRepo) CountModel(
	ctx context.Context,
	query map[string]any,
) (int64, error) {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"查询mon模型总数:开始执行",
		zap.Any("query", query),
	)

	// 开启数据库事务
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()
	countMonNodeStartTime := time.Now()
	count, err := database.DBCount(dbCtx, r.gormDB, &monmodel.MonNodeModel{}, query)
	countMonNodeDuration := time.Since(countMonNodeStartTime)
	if err != nil {
		log.Error(
			"查询mon模型总数:数据库查询失败",
			zap.Error(err),
			zap.Any("query", query),
			zap.Duration("count_duration", countMonNodeDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, errors.WrapIf(err, "查询mon模型总数:数据库查询失败")
	}
	log.Debug(
		"查询mon模型总数:执行成功",
		zap.Any("query", query),
		zap.Int64("count", count),
		zap.Duration("count_duration", countMonNodeDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return count, nil
}
