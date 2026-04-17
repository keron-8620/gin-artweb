package oes

import (
	"context"
	"io"
	"os"
	"time"

	"emperror.dev/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"

	oesmodel "gin-artweb/internal/model/oes"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/database"
	"gin-artweb/pkg/fileutil"
)

type OesColonyRepo struct {
	log           *zap.Logger
	gormDB        *gorm.DB
	timeouts      *config.DBTimeout
	slowThreshold *config.DBSlowThreshold
}

func NewOesColonyRepo(
	log *zap.Logger,
	gormDB *gorm.DB,
	timeouts *config.DBTimeout,
	slowThreshold *config.DBSlowThreshold,
) *OesColonyRepo {
	return &OesColonyRepo{
		log:           log,
		gormDB:        gormDB,
		timeouts:      timeouts,
		slowThreshold: slowThreshold,
	}
}

func (r *OesColonyRepo) CreateModel(
	ctx context.Context,
	m *oesmodel.OesColonyModel,
) error {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	// 检查参数
	if m == nil {
		err := errors.New("创建oes集群:模型不能为空")
		log.Error(
			"创建oes集群:模型不能为空",
			zap.Error(err),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return err
	}

	m.CreatedAt = startTime
	m.UpdatedAt = startTime

	log.Debug(
		"创建oes集群:模型详情",
		zap.Object("oes_colony_model", m),
	)

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()

	createStartTime := time.Now()
	err := database.DBCreate(dbCtx, r.gormDB, &oesmodel.OesColonyModel{}, m, nil)
	createDuration := time.Since(createStartTime)
	if err != nil {
		log.Error(
			"创建oes集群:数据库操作失败",
			zap.Error(err),
			zap.Object("oes_colony_model", m),
			zap.Duration("create_duration", createDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "创建oes集群失败")
	}

	if createDuration > r.slowThreshold.WriteSlow {
		log.Warn("创建oes集群模型:数据库创建耗时超过慢查询阈值，可能影响性能",
			zap.Uint32("oes_colony_id", m.ID),
			zap.Duration("create_duration", createDuration),
			zap.Duration("threshold", r.slowThreshold.WriteSlow),
		)
	}
	return nil
}

func (r *OesColonyRepo) UpdateModel(
	ctx context.Context,
	updateData map[string]any,
	conds ...any,
) error {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	if len(updateData) == 0 {
		err := errors.New("更新oes集群:更新数据为空")
		log.Error(
			"更新oes集群:更新数据为空",
			zap.Error(err),
			zap.Any("update_data", updateData),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return err
	}

	updateData["updated_at"] = startTime

	log.Debug(
		"更新oes集群:更新数据",
		zap.Any("update_data", updateData),
		zap.Any("conds", conds),
	)

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()

	updateStartTime := time.Now()
	err := database.DBUpdate(dbCtx, r.gormDB, &oesmodel.OesColonyModel{}, updateData, nil, conds...)
	updateDuration := time.Since(updateStartTime)
	if err != nil {
		log.Error(
			"更新oes集群:数据库操作失败",
			zap.Error(err),
			zap.Any("update_data", updateData),
			zap.Any("conds", conds),
			zap.Duration("update_duration", updateDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "更新oes集群失败")
	}

	if updateDuration > r.slowThreshold.WriteSlow {
		log.Warn("更新oes集群:数据库更新耗时超过慢查询阈值，可能影响性能",
			zap.Duration("update_duration", updateDuration),
			zap.Duration("threshold", r.slowThreshold.WriteSlow),
		)
	}
	return nil
}

func (r *OesColonyRepo) DeleteModel(
	ctx context.Context,
	conds ...any,
) error {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"删除oes集群:删除条件",
		zap.Any("conds", conds),
	)

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()

	deleteStartTime := time.Now()
	err := database.DBDelete(dbCtx, r.gormDB, &oesmodel.OesColonyModel{}, conds...)
	deleteOesColonyDuration := time.Since(deleteStartTime)
	if err != nil {
		log.Error(
			"删除oes集群:数据库操作失败",
			zap.Error(err),
			zap.Any("conds", conds),
			zap.Duration("delete_duration", deleteOesColonyDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "删除oes集群失败")
	}

	if deleteOesColonyDuration > r.slowThreshold.WriteSlow {
		log.Warn("删除oes集群:数据库删除耗时超过慢查询阈值，可能影响性能",
			zap.Duration("delete_duration", deleteOesColonyDuration),
			zap.Duration("threshold", r.slowThreshold.WriteSlow),
		)
	}
	return nil
}

func (r *OesColonyRepo) GetModel(
	ctx context.Context,
	preloads []string,
	conds ...any,
) (*oesmodel.OesColonyModel, error) {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"查询oes集群:查询条件",
		zap.Any("conds", conds),
	)

	var m oesmodel.OesColonyModel

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()

	getStartTime := time.Now()
	err := database.DBGet(dbCtx, r.gormDB, preloads, &m, conds...)
	getOesColonyDuration := time.Since(getStartTime)
	if err != nil {
		log.Error(
			"查询oes集群:数据库操作失败",
			zap.Error(err),
			zap.Any("conds", conds),
			zap.Duration("get_duration", getOesColonyDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.WrapIf(err, "查询oes集群失败")
	}

	log.Debug(
		"查询oes集群:查询到的模型详情",
		zap.Object("oes_colony_model", &m),
	)

	if getOesColonyDuration > r.slowThreshold.ReadSlow {
		log.Warn("查询oes集群:数据库查询耗时超过慢查询阈值，可能影响性能",
			zap.Duration("get_duration", getOesColonyDuration),
			zap.Duration("threshold", r.slowThreshold.ReadSlow),
		)
	}
	return &m, nil
}

func (r *OesColonyRepo) ListModel(
	ctx context.Context,
	qp database.QueryParams,
) ([]oesmodel.OesColonyModel, error) {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"查询oes集群列表:入参详情",
		zap.Object("query_params", &qp),
	)

	var ms []oesmodel.OesColonyModel

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ListTimeout)
	defer cancel()

	listStartTime := time.Now()
	err := database.DBList(dbCtx, r.gormDB, &oesmodel.OesColonyModel{}, &ms, qp)
	listDuration := time.Since(listStartTime)
	if err != nil {
		log.Error(
			"查询oes集群列表:数据库操作失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("list_duration", listDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.WrapIf(err, "查询oes集群列表失败")
	}

	log.Debug(
		"查询oes集群列表:查询到的模型列表",
		zap.Uint32s("colony_ids", oesmodel.ListOesColonyModelToUint32s(ms)),
	)

	if listDuration > r.slowThreshold.ListSlow {
		log.Warn("查询oes集群列表:数据库查询耗时超过慢查询阈值，可能影响性能",
			zap.Duration("list_duration", listDuration),
			zap.Duration("threshold", r.slowThreshold.ListSlow),
		)
	}
	return ms, nil
}

func (r *OesColonyRepo) CountModel(
	ctx context.Context,
	query map[string]any,
) (int64, error) {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"查询oes集群总数:查询条件",
		zap.Any("query", query),
	)

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()

	countStartTime := time.Now()
	count, err := database.DBCount(dbCtx, r.gormDB, &oesmodel.OesColonyModel{}, query)
	countDuration := time.Since(countStartTime)
	if err != nil {
		log.Error(
			"查询oes集群总数:数据库操作失败",
			zap.Error(err),
			zap.Any("query", query),
			zap.Duration("count_duration", countDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, errors.WrapIf(err, "查询oes集群总数失败")
	}

	log.Debug(
		"查询oes集群总数:查询到的记录数",
		zap.Int64("count", count),
	)

	if countDuration > r.slowThreshold.ReadSlow {
		log.Warn("查询oes集群总数:数据库查询耗时超过慢查询阈值，可能影响性能",
			zap.Duration("count_duration", countDuration),
			zap.Duration("threshold", r.slowThreshold.ReadSlow),
		)
	}
	return count, nil
}

func (r *OesColonyRepo) SaveConfigFile(
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

func (r *OesColonyRepo) RemoveConfigFile(
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

func (r *OesColonyRepo) MoveConfigFile(
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
