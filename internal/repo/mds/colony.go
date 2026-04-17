package mds

import (
	"context"
	"io"
	"os"
	"time"

	"emperror.dev/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"

	mdsmodel "gin-artweb/internal/model/mds"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/database"
	"gin-artweb/pkg/fileutil"
)

type MdsColonyRepo struct {
	log           *zap.Logger
	gormDB        *gorm.DB
	timeouts      *config.DBTimeout
	slowThreshold *config.DBSlowThreshold
}

func NewMdsColonyRepo(
	log *zap.Logger,
	gormDB *gorm.DB,
	timeouts *config.DBTimeout,
	slowThreshold *config.DBSlowThreshold,
) *MdsColonyRepo {
	return &MdsColonyRepo{
		log:           log,
		gormDB:        gormDB,
		timeouts:      timeouts,
		slowThreshold: slowThreshold,
	}
}

func (r *MdsColonyRepo) CreateModel(
	ctx context.Context,
	m *mdsmodel.MdsColonyModel,
) error {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	// 检查参数
	if m == nil {
		err := errors.New("创建mds集群:模型不能为空")
		log.Error(
			"创建mds集群:模型不能为空",
			zap.Error(err),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return err
	}

	m.CreatedAt = startTime
	m.UpdatedAt = startTime

	log.Debug(
		"创建mds集群:模型详情",
		zap.Object("mds_colony_model", m),
	)

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()

	createStartTime := time.Now()
	err := database.DBCreate(dbCtx, r.gormDB, &mdsmodel.MdsColonyModel{}, m, nil)
	createDuration := time.Since(createStartTime)
	if err != nil {
		log.Error(
			"创建mds集群:数据库操作失败",
			zap.Error(err),
			zap.Object("mds_colony_model", m),
			zap.Duration("create_duration", createDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "创建mds集群:数据库操作失败")
	}

	if createDuration > r.slowThreshold.WriteSlow {
		log.Warn("创建mds集群模型:数据库创建耗时超过慢查询阈值，可能影响性能",
			zap.Uint32("colony_id", m.ID),
			zap.Duration("create_duration", createDuration),
			zap.Duration("threshold", r.slowThreshold.WriteSlow),
		)
	}
	return nil
}

func (r *MdsColonyRepo) UpdateModel(
	ctx context.Context,
	updateData map[string]any,
	conds ...any,
) error {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	if len(updateData) == 0 {
		err := errors.New("更新mds集群:更新数据为空")
		log.Error(
			"更新mds集群:更新数据为空",
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
	err := database.DBUpdate(dbCtx, r.gormDB, &mdsmodel.MdsColonyModel{}, updateData, nil, conds...)
	updateDuration := time.Since(updateStartTime)
	if err != nil {
		log.Error(
			"更新mds集群:数据库操作失败",
			zap.Error(err),
			zap.Any("conds", conds),
			zap.Any("update_data", updateData),
			zap.Duration("update_duration", updateDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "更新mds集群:数据库操作失败")
	}

	if updateDuration > r.slowThreshold.WriteSlow {
		log.Warn("更新mds集群:数据库更新耗时超过慢查询阈值，可能影响性能",
			zap.Duration("update_duration", updateDuration),
			zap.Duration("threshold", r.slowThreshold.WriteSlow),
		)
	}
	return nil
}

func (r *MdsColonyRepo) DeleteModel(
	ctx context.Context,
	conds ...any,
) error {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"删除mds集群:删除条件",
		zap.Any("conds", conds),
	)

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()

	deleteStartTime := time.Now()
	err := database.DBDelete(dbCtx, r.gormDB, &mdsmodel.MdsColonyModel{}, conds...)
	deleteDuration := time.Since(deleteStartTime)
	if err != nil {
		log.Error(
			"删除mds集群:数据库操作失败",
			zap.Error(err),
			zap.Any("conds", conds),
			zap.Duration("delete_duration", deleteDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "删除mds集群:数据库操作失败")
	}

	if deleteDuration > r.slowThreshold.WriteSlow {
		log.Warn("删除mds集群:数据库删除耗时超过慢查询阈值，可能影响性能",
			zap.Duration("delete_duration", deleteDuration),
			zap.Duration("threshold", r.slowThreshold.WriteSlow),
		)
	}
	return nil
}

func (r *MdsColonyRepo) GetModel(
	ctx context.Context,
	preloads []string,
	conds ...any,
) (*mdsmodel.MdsColonyModel, error) {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"查询mds集群:查询条件",
		zap.Any("conds", conds),
	)

	var m mdsmodel.MdsColonyModel

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()

	getStartTime := time.Now()
	err := database.DBGet(dbCtx, r.gormDB, preloads, &m, conds...)
	getDuration := time.Since(getStartTime)
	if err != nil {
		log.Error(
			"查询mds集群:数据库操作失败",
			zap.Error(err),
			zap.Any("conds", conds),
			zap.Duration("get_duration", getDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.WrapIf(err, "查询mds集群:数据库操作失败")
	}

	log.Debug(
		"查询mds集群:查询到的模型详情",
		zap.Object("mds_colony_model", &m),
	)

	if getDuration > r.slowThreshold.ReadSlow {
		log.Warn("查询mds集群:数据库查询耗时超过慢查询阈值，可能影响性能",
			zap.Duration("get_duration", getDuration),
			zap.Duration("threshold", r.slowThreshold.ReadSlow),
		)
	}
	return &m, nil
}

func (r *MdsColonyRepo) ListModel(
	ctx context.Context,
	qp database.QueryParams,
) ([]mdsmodel.MdsColonyModel, error) {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"查询mds集群列表:入参详情",
		zap.Object("query_params", &qp),
	)

	var ms []mdsmodel.MdsColonyModel

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ListTimeout)
	defer cancel()

	listStartTime := time.Now()
	err := database.DBList(dbCtx, r.gormDB, &mdsmodel.MdsColonyModel{}, &ms, qp)
	listDuration := time.Since(listStartTime)
	if err != nil {
		log.Error(
			"查询mds集群列表:数据库操作失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("list_duration", listDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.WrapIf(err, "查询mds集群列表:数据库操作失败")
	}

	log.Debug(
		"查询mds集群列表:查询到的模型列表",
		zap.Uint32s("colony_ids", mdsmodel.ListMdsColonyModelToUint32s(ms)),
	)

	if listDuration > r.slowThreshold.ListSlow {
		log.Warn("查询mds集群列表:数据库查询耗时超过慢查询阈值，可能影响性能",
			zap.Duration("list_duration", listDuration),
			zap.Duration("threshold", r.slowThreshold.ListSlow),
		)
	}
	return ms, nil
}

func (r *MdsColonyRepo) CountModel(
	ctx context.Context,
	query map[string]any,
) (int64, error) {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"查询mds集群总数:查询条件",
		zap.Any("query", query),
	)

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()

	countStartTime := time.Now()
	count, err := database.DBCount(dbCtx, r.gormDB, &mdsmodel.MdsColonyModel{}, query)
	countDuration := time.Since(countStartTime)
	if err != nil {
		log.Error(
			"查询mds集群总数:数据库查询失败",
			zap.Error(err),
			zap.Any("query", query),
			zap.Duration("count_duration", countDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, errors.WrapIf(err, "查询mds集群总数:数据库查询失败")
	}

	log.Debug(
		"查询mds集群总数:查询到的记录数",
		zap.Int64("count", count),
	)

	if countDuration > r.slowThreshold.ReadSlow {
		log.Warn("查询mds集群总数:数据库查询耗时超过慢查询阈值，可能影响性能",
			zap.Duration("count_duration", countDuration),
			zap.Duration("threshold", r.slowThreshold.ReadSlow),
		)
	}
	return count, nil
}

func (r *MdsColonyRepo) SaveConfigFile(
	ctx context.Context,
	fileReader io.Reader,
	confPath string,
	overwrite bool,
) error {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"保存配置文件:入参详情",
		zap.String("conf_path", confPath),
		zap.Bool("overwrite", overwrite),
	)

	saveStartTime := time.Now()
	err := fileutil.WriteReaderToFile(ctx, fileReader, confPath, os.FileMode(0o644), overwrite)
	saveDuration := time.Since(saveStartTime)
	if err != nil {
		log.Error(
			"保存配置文件:文件写入失败",
			zap.Error(err),
			zap.String("conf_path", confPath),
			zap.Duration("save_duration", saveDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "保存配置文件:文件写入失败")
	}
	return nil
}

func (r *MdsColonyRepo) RemoveConfigFile(
	ctx context.Context,
	confPath string,
) error {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"删除配置文件:入参详情",
		zap.String("conf_path", confPath),
	)

	// 删除文件
	removeStartTime := time.Now()
	err := fileutil.Remove(ctx, confPath)
	removeDuration := time.Since(removeStartTime)
	if err != nil {
		log.Error(
			"删除配置文件:文件删除失败",
			zap.Error(err),
			zap.String("conf_path", confPath),
			zap.Duration("remove_duration", removeDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "删除配置文件:文件删除失败")
	}
	return nil
}

func (r *MdsColonyRepo) MoveConfigFile(
	ctx context.Context,
	oldConfPath string,
	newConfPath string,
) error {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"移动配置文件:入参详情",
		zap.String("old_conf_path", oldConfPath),
		zap.String("new_conf_path", newConfPath),
	)

	// 移动文件
	moveStartTime := time.Now()
	err := fileutil.Move(ctx, oldConfPath, newConfPath)
	moveDuration := time.Since(moveStartTime)
	if err != nil {
		log.Error(
			"移动配置文件:文件移动失败",
			zap.Error(err),
			zap.String("old_conf_path", oldConfPath),
			zap.String("new_conf_path", newConfPath),
			zap.Duration("move_duration", moveDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "移动配置文件:文件移动失败")
	}
	return nil
}
