package oes

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"time"

	"emperror.dev/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"

	oesmodel "gin-artweb/internal/model/oes"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/database"
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

func (r *OesColonyRepo) CreateModel(ctx context.Context, m *oesmodel.OesColonyModel) error {
	log := ctxutil.NewLogger(r.log, ctx)

	// 检查参数
	if m == nil {
		err := errors.New("创建oes集群：模型不能为空")
		log.Error(
			"创建oes集群：模型不能为空",
			zap.Error(err),
		)
		return err
	}
	log.Debug(
		"创建oes集群：开始执行",
		zap.Object("colony_model", m),
	)
	now := time.Now()
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	if err := database.DBCreate(dbCtx, r.gormDB, &oesmodel.OesColonyModel{}, m, nil); err != nil {
		log.Error(
			"创建oes集群：数据库操作失败",
			zap.Error(err),
			zap.Object("colony_model", m),
			zap.Duration("create_duration", time.Since(now)),
		)
		return errors.WrapIf(err, "创建oes集群失败")
	}
	log.Debug(
		"创建oes集群：执行成功",
		zap.Object("colony_model", m),
		zap.Duration("create_duration", time.Since(now)),
	)
	return nil
}

func (r *OesColonyRepo) UpdateModel(ctx context.Context, data map[string]any, conds ...any) error {
	log := ctxutil.NewLogger(r.log, ctx)

	// 检查参数
	if len(data) == 0 {
		err := errors.New("更新oes集群：更新数据不能为空")
		log.Error(
			"更新oes集群：更新数据不能为空",
			zap.Error(err),
			zap.Any("update_data", data),
			zap.Any("conds", conds),
		)
		return err
	}

	log.Debug(
		"更新oes集群：开始执行",
		zap.Any("update_data", data),
		zap.Any("conds", conds),
	)
	startTime := time.Now()
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	if err := database.DBUpdate(dbCtx, r.gormDB, &oesmodel.OesColonyModel{}, data, nil, conds...); err != nil {
		log.Error(
			"更新oes集群：数据库操作失败",
			zap.Error(err),
			zap.Any("update_data", data),
			zap.Any("conds", conds),
			zap.Duration("update_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "更新oes集群失败")
	}
	log.Debug(
		"更新oes集群：执行成功",
		zap.Any("update_data", data),
		zap.Any("conds", conds),
		zap.Duration("update_duration", time.Since(startTime)),
	)
	return nil
}

func (r *OesColonyRepo) DeleteModel(ctx context.Context, conds ...any) error {
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"删除oes集群：开始执行",
		zap.Any("conds", conds),
	)
	startTime := time.Now()
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	if err := database.DBDelete(dbCtx, r.gormDB, &oesmodel.OesColonyModel{}, conds...); err != nil {
		log.Error(
			"删除oes集群：数据库操作失败",
			zap.Error(err),
			zap.Any("conds", conds),
			zap.Duration("delete_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "删除oes集群失败")
	}
	log.Debug(
		"删除oes集群：执行成功",
		zap.Any("conds", conds),
		zap.Duration("delete_duration", time.Since(startTime)),
	)
	return nil
}

func (r *OesColonyRepo) GetModel(
	ctx context.Context,
	preloads []string,
	conds ...any,
) (*oesmodel.OesColonyModel, error) {
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"查询oes集群：开始执行",
		zap.Any("conds", conds),
	)
	startTime := time.Now()
	var m oesmodel.OesColonyModel
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()
	if err := database.DBGet(dbCtx, r.gormDB, preloads, &m, conds...); err != nil {
		log.Error(
			"查询oes集群失败",
			zap.Error(err),
			zap.Any("conds", conds),
			zap.Duration("get_duration", time.Since(startTime)),
		)
		return nil, errors.WrapIf(err, "查询oes集群失败")
	}
	log.Debug(
		"查询oes集群：执行成功",
		zap.Object("colony_model", &m),
		zap.Any("conds", conds),
		zap.Duration("get_duration", time.Since(startTime)),
	)
	return &m, nil
}

func (r *OesColonyRepo) ListModel(
	ctx context.Context,
	qp database.QueryParams,
) ([]oesmodel.OesColonyModel, error) {
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"查询oes集群列表：开始执行",
		zap.Object("query_params", &qp),
	)
	startTime := time.Now()
	var ms []oesmodel.OesColonyModel
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ListTimeout)
	defer cancel()
	if err := database.DBList(dbCtx, r.gormDB, &oesmodel.OesColonyModel{}, &ms, qp); err != nil {
		log.Error(
			"查询oes集群列表：数据库操作失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("list_duration", time.Since(startTime)),
		)
		return nil, errors.WrapIf(err, "查询oes集群列表失败")
	}
	log.Debug(
		"查询oes集群列表：执行成功",
		zap.Object("query_params", &qp),
		zap.Duration("list_duration", time.Since(startTime)),
	)
	return ms, nil
}

func (r *OesColonyRepo) CountModel(
	ctx context.Context,
	query map[string]any,
) (int64, error) {
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"查询oes集群总数：开始执行",
		zap.Any("query", query),
	)
	startTime := time.Now()
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()
	count, err := database.DBCount(dbCtx, r.gormDB, &oesmodel.OesColonyModel{}, query)
	if err != nil {
		log.Error(
			"查询oes集群总数：数据库操作失败",
			zap.Error(err),
			zap.Any("query", query),
			zap.Duration("count_duration", time.Since(startTime)),
		)
		return 0, errors.WrapIf(err, "查询oes集群总数失败")
	}
	log.Debug(
		"查询oes集群总数：执行成功",
		zap.Any("query", query),
		zap.Int64("count", count),
		zap.Duration("count_duration", time.Since(startTime)),
	)
	return count, nil
}

func (r *OesColonyRepo) SaveConfigFile(
	ctx context.Context,
	fileReader io.Reader,
	pkgPath string,
	overwrite bool,
) error {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	// 检查文件是否已存在且不允许覆盖
	if !overwrite {
		if _, err := os.Stat(pkgPath); err == nil {
			err := errors.New("保存oes集群配置文件：配置文件已存在")
			log.Error(
				"保存oes集群配置文件：配置文件已存在",
				zap.Error(err),
				zap.String("pkg_path", pkgPath),
				zap.Duration("total_duration", time.Since(startTime)),
			)
			return err
		}
	}

	mode := os.FileMode(0o644)
	dir := filepath.Dir(pkgPath)
	if err := os.MkdirAll(dir, mode); err != nil {
		log.Error(
			"保存oes集群配置文件：创建配置文件目录失败",
			zap.Error(err),
			zap.String("pkg_dir", dir),
		)
		return errors.WrapIf(err, "保存oes集群配置文件：创建配置文件目录失败")
	}
	if err := os.Chmod(dir, mode); err != nil {
		log.Error(
			"保存oes集群配置文件：设置配置文件目录权限失败",
			zap.Error(err),
			zap.String("pkg_dir", dir),
		)
		return errors.WrapIf(err, "保存oes集群配置文件：设置配置文件目录权限失败")
	}

	// 创建文件并写入内容
	file, err := os.OpenFile(pkgPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		log.Error(
			"保存oes集群配置文件：创建配置文件失败",
			zap.Error(err),
			zap.String("pkg_path", pkgPath),
		)
		return errors.WrapIf(err, "保存oes集群配置文件：创建配置文件失败")
	}
	defer file.Close()

	if _, err = io.Copy(file, fileReader); err != nil {
		log.Error(
			"保存oes集群配置文件：写入配置文件失败",
			zap.Error(err),
			zap.String("save_path", pkgPath),
		)
		return errors.WrapIf(err, "保存oes集群配置文件：写入配置文件失败")
	}
	return nil
}

func (r *OesColonyRepo) RemoveConfigFile(ctx context.Context, pkgPath string) error {
	log := ctxutil.NewLogger(r.log, ctx)

	// 检查文件是否存在
	if _, err := os.Stat(pkgPath); err != nil {
		if os.IsNotExist(err) {
			log.Warn(
				"删除oes集群配置文件：配置文件不存在",
				zap.Error(err),
				zap.String("pkg_path", pkgPath),
			)
			return nil
		}
		log.Error(
			"删除oes集群配置文件：检查配置文件失败",
			zap.Error(err),
			zap.String("pkg_path", pkgPath),
		)
		return errors.WrapIf(err, "删除oes集群配置文件：检查配置文件失败")
	}
	// 删除文件
	if err := os.Remove(pkgPath); err != nil {
		log.Error(
			"删除oes集群配置文件：文件删除失败",
			zap.Error(err),
			zap.String("pkg_path", pkgPath),
		)
		return errors.WrapIf(err, "删除oes集群配置文件：文件删除失败")
	}
	return nil
}
