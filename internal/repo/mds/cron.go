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

type MdsCronRepo struct {
	log      *zap.Logger
	gormDB   *gorm.DB
	timeouts *config.DBTimeout
}

func NewMdsCronRepo(
	log *zap.Logger,
	gormDB *gorm.DB,
	timeouts *config.DBTimeout,
) *MdsCronRepo {
	return &MdsCronRepo{
		log:      log,
		gormDB:   gormDB,
		timeouts: timeouts,
	}
}

func (r *MdsCronRepo) CreateModel(
	ctx context.Context,
	m *mdsmodel.MdsCronModel,
) error {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	// 检查参数
	if m == nil {
		err := errors.New("创建mds计划任务:模型不能为空")
		log.Error(
			"创建mds计划任务:模型不能为空",
			zap.Error(err),
		)
		return err
	}
	log.Debug(
		"创建mds计划任务:开始执行",
		zap.Object("model", m),
	)
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	createMdsCronStartTime := time.Now()
	err := database.DBCreate(dbCtx, r.gormDB, &mdsmodel.MdsCronModel{}, m, nil)
	createMdsCronDuration := time.Since(createMdsCronStartTime)
	if err != nil {
		log.Error(
			"创建mds计划任务:数据库操作失败",
			zap.Error(err),
			zap.Object("cron_model", m),
			zap.Duration("create_mds_cron_duration", createMdsCronDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "创建mds计划任务:数据库操作失败")
	}
	log.Debug(
		"创建mds计划任务:执行成功",
		zap.Object("cron_model", m),
		zap.Duration("create_mds_cron_duration", createMdsCronDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

func (r *MdsCronRepo) UpdateModel(
	ctx context.Context,
	data map[string]any,
	conds ...any,
) error {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	// 检查参数
	if len(data) == 0 {
		err := errors.New("更新mds计划任务:更新数据不能为空")
		log.Error(
			"更新mds计划任务:更新数据不能为空",
			zap.Error(err),
			zap.Any("update_data", data),
			zap.Any("conds", conds),
		)
		return err
	}

	log.Debug(
		"更新mds计划任务:开始执行",
		zap.Any("update_data", data),
		zap.Any("conds", conds),
	)
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	updateMdsCronStartTime := time.Now()
	err := database.DBUpdate(dbCtx, r.gormDB, &mdsmodel.MdsCronModel{}, data, nil, conds...)
	updateMdsCronDuration := time.Since(updateMdsCronStartTime)
	if err != nil {
		log.Error(
			"更新mds计划任务:数据库操作失败",
			zap.Error(err),
			zap.Any("update_data", data),
			zap.Any("conds", conds),
			zap.Duration("update_mds_cron_duration", updateMdsCronDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "更新mds计划任务:数据库操作失败")
	}
	log.Debug(
		"更新mds计划任务:执行成功",
		zap.Any("update_data", data),
		zap.Any("conds", conds),
		zap.Duration("update_mds_cron_duration", updateMdsCronDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

func (r *MdsCronRepo) DeleteModel(
	ctx context.Context,
	conds ...any,
) error {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"删除mds计划任务:开始执行",
		zap.Any("conds", conds),
	)
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	deleteMdsCronStartTime := time.Now()
	err := database.DBDelete(dbCtx, r.gormDB, &mdsmodel.MdsCronModel{}, conds...)
	deleteMdsCronDuration := time.Since(deleteMdsCronStartTime)
	if err != nil {
		log.Error(
			"删除mds计划任务:数据库操作失败",
			zap.Error(err),
			zap.Any("conds", conds),
			zap.Duration("delete_mds_cron_duration", deleteMdsCronDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "删除mds计划任务:数据库操作失败")
	}
	log.Debug(
		"删除mds计划任务:执行成功",
		zap.Any("conds", conds),
		zap.Duration("delete_mds_cron_duration", deleteMdsCronDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

func (r *MdsCronRepo) GetModel(
	ctx context.Context,
	preloads []string,
	conds ...any,
) (*mdsmodel.MdsCronModel, error) {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)
	log.Debug(
		"查询mds计划任务:开始执行",
		zap.Any("conds", conds),
	)
	var m mdsmodel.MdsCronModel
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()
	getMdsCronStartTime := time.Now()
	err := database.DBGet(dbCtx, r.gormDB, preloads, &m, conds...)
	getMdsCronDuration := time.Since(getMdsCronStartTime)
	if err != nil {
		log.Error(
			"查询mds计划任务:数据库操作失败",
			zap.Error(err),
			zap.Any("conds", conds),
			zap.Duration("get_mds_cron_duration", getMdsCronDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.WrapIf(err, "查询mds计划任务:数据库操作失败")
	}
	log.Debug(
		"查询mds计划任务:执行成功",
		zap.Object("cron_model", &m),
		zap.Any("conds", conds),
		zap.Duration("get_mds_cron_duration", getMdsCronDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return &m, nil
}

func (r *MdsCronRepo) ListModel(
	ctx context.Context,
	qp database.QueryParams,
) ([]mdsmodel.MdsCronModel, error) {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)
	log.Debug(
		"查询mds计划任务列表:开始执行",
		zap.Object("query_params", &qp),
	)
	var ms []mdsmodel.MdsCronModel
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ListTimeout)
	defer cancel()
	listMdsCronStartTime := time.Now()
	err := database.DBList(dbCtx, r.gormDB, &mdsmodel.MdsCronModel{}, &ms, qp)
	listMdsCronDuration := time.Since(listMdsCronStartTime)
	if err != nil {
		log.Error(
			"查询mds计划任务列表:数据库操作失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("list_mds_cron_duration", listMdsCronDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.WrapIf(err, "查询mds计划任务列表:数据库操作失败")
	}
	log.Debug(
		"查询mds计划任务列表:执行成功",
		zap.Object("query_params", &qp),
		zap.Duration("list_mds_cron_duration", listMdsCronDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return ms, nil
}

func (r *MdsCronRepo) CountModel(
	ctx context.Context,
	query map[string]any,
) (int64, error) {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"查询mds计划任务总数:开始执行",
		zap.Any("query", query),
	)
	countMdsCronStartTime := time.Now()
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()
	countMdsCronStartTime = time.Now()
	count, err := database.DBCount(dbCtx, r.gormDB, &mdsmodel.MdsCronModel{}, query)
	countMdsCronDuration := time.Since(countMdsCronStartTime)
	if err != nil {
		log.Error(
			"查询mds计划任务总数:数据库查询失败",
			zap.Error(err),
			zap.Any("query", query),
			zap.Duration("count_mds_cron_duration", countMdsCronDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, errors.WrapIf(err, "查询mds计划任务总数:数据库查询失败")
	}
	log.Debug(
		"查询mds计划任务总数:执行成功",
		zap.Any("query", query),
		zap.Int64("count", count),
		zap.Duration("count_mds_cron_duration", countMdsCronDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return count, nil
}
