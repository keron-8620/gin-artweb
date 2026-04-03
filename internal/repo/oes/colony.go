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
	log      *zap.Logger
	gormDB   *gorm.DB
	timeouts *config.DBTimeout
}

func NewOesColonyRepo(
	log *zap.Logger,
	gormDB *gorm.DB,
	timeouts *config.DBTimeout,
) *OesColonyRepo {
	return &OesColonyRepo{
		log:      log,
		gormDB:   gormDB,
		timeouts: timeouts,
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
		)
		return err
	}
	log.Debug(
		"创建oes集群:开始执行",
		zap.Object("colony_model", m),
	)

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	createOesColonyStartTime := time.Now()
	err := database.DBCreate(dbCtx, r.gormDB, &oesmodel.OesColonyModel{}, m, nil)
	createOesColonyDuration := time.Since(createOesColonyStartTime)
	if err != nil {
		log.Error(
			"创建oes集群:数据库操作失败",
			zap.Error(err),
			zap.Object("colony_model", m),
			zap.Duration("create_oes_colony_duration", createOesColonyDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "创建oes集群失败")
	}
	log.Debug(
		"创建oes集群:执行成功",
		zap.Object("colony_model", m),
		zap.Duration("create_oes_colony_duration", createOesColonyDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

func (r *OesColonyRepo) UpdateModel(
	ctx context.Context,
	data map[string]any,
	conds ...any,
) error {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	// 检查参数
	if len(data) == 0 {
		err := errors.New("更新oes集群:更新数据不能为空")
		log.Error(
			"更新oes集群:更新数据不能为空",
			zap.Error(err),
			zap.Any("update_data", data),
			zap.Any("conds", conds),
		)
		return err
	}

	log.Debug(
		"更新oes集群:开始执行",
		zap.Any("update_data", data),
		zap.Any("conds", conds),
	)

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	updateOesColonyStartTime := time.Now()
	err := database.DBUpdate(dbCtx, r.gormDB, &oesmodel.OesColonyModel{}, data, nil, conds...)
	updateOesColonyDuration := time.Since(updateOesColonyStartTime)
	if err != nil {
		log.Error(
			"更新oes集群:数据库操作失败",
			zap.Error(err),
			zap.Any("update_data", data),
			zap.Any("conds", conds),
			zap.Duration("update_oes_colony_duration", updateOesColonyDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "更新oes集群失败")
	}
	log.Debug(
		"更新oes集群:执行成功",
		zap.Any("update_data", data),
		zap.Any("conds", conds),
		zap.Duration("update_oes_colony_duration", updateOesColonyDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

func (r *OesColonyRepo) DeleteModel(
	ctx context.Context,
	conds ...any,
) error {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"删除oes集群:开始执行",
		zap.Any("conds", conds),
	)

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	deleteOesColonyStartTime := time.Now()
	err := database.DBDelete(dbCtx, r.gormDB, &oesmodel.OesColonyModel{}, conds...)
	deleteOesColonyDuration := time.Since(deleteOesColonyStartTime)
	if err != nil {
		log.Error(
			"删除oes集群:数据库操作失败",
			zap.Error(err),
			zap.Any("conds", conds),
			zap.Duration("delete_oes_colony_duration", deleteOesColonyDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "删除oes集群失败")
	}
	log.Debug(
		"删除oes集群:执行成功",
		zap.Any("conds", conds),
		zap.Duration("delete_oes_colony_duration", deleteOesColonyDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
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
		"查询oes集群:开始执行",
		zap.Any("conds", conds),
	)
	var m oesmodel.OesColonyModel
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()
	getOesColonyStartTime := time.Now()
	err := database.DBGet(dbCtx, r.gormDB, preloads, &m, conds...)
	getOesColonyDuration := time.Since(getOesColonyStartTime)
	if err != nil {
		log.Error(
			"查询oes集群:数据库操作失败",
			zap.Error(err),
			zap.Any("conds", conds),
			zap.Duration("get_oes_colony_duration", getOesColonyDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.WrapIf(err, "查询oes集群失败")
	}
	log.Debug(
		"查询oes集群:执行成功",
		zap.Object("colony_model", &m),
		zap.Any("conds", conds),
		zap.Duration("get_oes_colony_duration", getOesColonyDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return &m, nil
}

func (r *OesColonyRepo) ListModel(
	ctx context.Context,
	qp database.QueryParams,
) ([]oesmodel.OesColonyModel, error) {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"查询oes集群列表:开始执行",
		zap.Object("query_params", &qp),
	)
	var ms []oesmodel.OesColonyModel
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ListTimeout)
	defer cancel()
	listOesColonyStartTime := time.Now()
	err := database.DBList(dbCtx, r.gormDB, &oesmodel.OesColonyModel{}, &ms, qp)
	listOesColonyDuration := time.Since(listOesColonyStartTime)
	if err != nil {
		log.Error(
			"查询oes集群列表:数据库操作失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("list_oes_colony_duration", listOesColonyDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.WrapIf(err, "查询oes集群列表失败")
	}
	log.Debug(
		"查询oes集群列表:执行成功",
		zap.Object("query_params", &qp),
		zap.Duration("list_oes_colony_duration", listOesColonyDuration),
		zap.Duration("total_duration", time.Since(listOesColonyStartTime)),
	)
	return ms, nil
}

func (r *OesColonyRepo) CountModel(
	ctx context.Context,
	query map[string]any,
) (int64, error) {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"查询oes集群总数:开始执行",
		zap.Any("query", query),
	)
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()
	countOesColonyStartTime := time.Now()
	count, err := database.DBCount(dbCtx, r.gormDB, &oesmodel.OesColonyModel{}, query)
	countOesColonyDuration := time.Since(countOesColonyStartTime)
	if err != nil {
		log.Error(
			"查询oes集群总数:数据库操作失败",
			zap.Error(err),
			zap.Any("query", query),
			zap.Duration("count_oes_colony_duration", countOesColonyDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, errors.WrapIf(err, "查询oes集群总数失败")
	}
	log.Debug(
		"查询oes集群总数:执行成功",
		zap.Any("query", query),
		zap.Int64("count", count),
		zap.Duration("count_oes_colony_duration", countOesColonyDuration),
		zap.Duration("total_duration", time.Since(countOesColonyStartTime)),
	)
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
		"保存oes集群配置文件:开始执行",
		zap.String("conf_path", confPath),
		zap.Bool("overwrite", overwrite),
	)
	saveScriptStartTime := time.Now()
	err := fileutil.WriteReaderToFile(ctx, fileReader, confPath, os.FileMode(0o750), overwrite)
	saveScriptDuration := time.Since(saveScriptStartTime)
	if err != nil {
		log.Error(
			"保存oes集群配置文件:文件写入失败",
			zap.Error(err),
			zap.String("conf_path", confPath),
			zap.Duration("save_conf_duration", saveScriptDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "保存oes集群配置文件:文件写入失败")
	}
	log.Debug(
		"保存oes集群配置文件:执行成功",
		zap.String("conf_path", confPath),
		zap.Duration("save_conf_duration", saveScriptDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

func (r *OesColonyRepo) RemoveConfigFile(
	ctx context.Context,
	pkgPath string,
) error {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"删除oes集群配置文件:开始执行",
		zap.String("pkg_path", pkgPath),
	)

	// 检查文件是否存在
	if _, err := os.Stat(pkgPath); err != nil {
		if os.IsNotExist(err) {
			log.Warn(
				"删除oes集群配置文件:配置文件不存在",
				zap.Error(err),
				zap.String("pkg_path", pkgPath),
				zap.Duration("total_duration", time.Since(startTime)),
			)
			return nil
		}
		log.Error(
			"删除oes集群配置文件:检查配置文件失败",
			zap.Error(err),
			zap.String("pkg_path", pkgPath),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "删除oes集群配置文件:检查配置文件失败")
	}
	// 删除文件
	deleteOesColonyStartTime := time.Now()
	err := os.Remove(pkgPath)
	deleteOesColonyDuration := time.Since(deleteOesColonyStartTime)
	if err != nil {
		log.Error(
			"删除oes集群配置文件:文件删除失败",
			zap.Error(err),
			zap.String("pkg_path", pkgPath),
			zap.Duration("delete_oes_colony_duration", deleteOesColonyDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "删除oes集群配置文件:文件删除失败")
	}
	log.Debug(
		"删除oes集群配置文件:执行成功",
		zap.String("pkg_path", pkgPath),
		zap.Duration("delete_oes_colony_duration", deleteOesColonyDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}
