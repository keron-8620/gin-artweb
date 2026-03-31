package mds

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"time"

	"emperror.dev/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"

	mdsmodel "gin-artweb/internal/model/mds"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/database"
)

type MdsColonyRepo struct {
	log      *zap.Logger
	gormDB   *gorm.DB
	timeouts *config.DBTimeout
}

func NewMdsColonyRepo(
	log *zap.Logger,
	gormDB *gorm.DB,
	timeouts *config.DBTimeout,
) *MdsColonyRepo {
	return &MdsColonyRepo{
		log:      log,
		gormDB:   gormDB,
		timeouts: timeouts,
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
		err := errors.New("创建mds集群：模型为空")
		log.Error(
			"创建mds集群：模型为空",
			zap.Error(err),
		)
		return err
	}
	log.Debug(
		"创建mds集群：开始执行",
		zap.Object("colony_model", m),
	)

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	createMdsColonyStartTime := time.Now()
	err := database.DBCreate(dbCtx, r.gormDB, &mdsmodel.MdsColonyModel{}, m, nil)
	createMdsColonyDuration := time.Since(createMdsColonyStartTime)
	if err != nil {
		log.Error(
			"创建mds集群：数据库操作失败",
			zap.Error(err),
			zap.Object("colony_model", m),
			zap.Duration("create_duration", createMdsColonyDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "创建mds集群：数据库操作失败")
	}
	log.Debug(
		"创建mds集群：执行成功",
		zap.Object("colony_model", m),
		zap.Duration("create_duration", createMdsColonyDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

func (r *MdsColonyRepo) UpdateModel(
	ctx context.Context,
	data map[string]any,
	conds ...any,
) error {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	// 检查参数
	if len(data) == 0 {
		err := errors.New("更新mds集群：更新数据为空")
		log.Error(
			"更新mds集群：更新数据为空",
			zap.Error(err),
			zap.Any("update_data", data),
			zap.Any("conds", conds),
		)
		return err
	}

	log.Debug(
		"更新mds集群：开始执行",
		zap.Any("update_data", data),
		zap.Any("conds", conds),
	)
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	updateMdsColonyStartTime := time.Now()
	err := database.DBUpdate(dbCtx, r.gormDB, &mdsmodel.MdsColonyModel{}, data, nil, conds...)
	updateMdsColonyDuration := time.Since(updateMdsColonyStartTime)
	if err != nil {
		log.Error(
			"更新mds集群：数据库操作失败",
			zap.Error(err),
			zap.Any("update_data", data),
			zap.Any("conds", conds),
			zap.Duration("update_duration", updateMdsColonyDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "更新mds集群：数据库操作失败")
	}
	log.Debug(
		"更新mds集群：执行成功",
		zap.Any("update_data", data),
		zap.Any("conds", conds),
		zap.Duration("update_duration", updateMdsColonyDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

func (r *MdsColonyRepo) DeleteModel(
	ctx context.Context,
	conds ...any,
) error {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"删除mds集群：开始执行",
		zap.Any("conds", conds),
	)
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	deleteMdsColonyStartTime := time.Now()
	err := database.DBDelete(dbCtx, r.gormDB, &mdsmodel.MdsColonyModel{}, conds...)
	deleteMdsColonyDuration := time.Since(deleteMdsColonyStartTime)
	if err != nil {
		log.Error(
			"删除mds集群：数据库操作失败",
			zap.Error(err),
			zap.Any("conds", conds),
			zap.Duration("delete_duration", deleteMdsColonyDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "删除mds集群：数据库操作失败")
	}
	log.Debug(
		"删除mds集群：执行成功",
		zap.Any("conds", conds),
		zap.Duration("delete_duration", deleteMdsColonyDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
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
		"查询mds集群：开始执行",
		zap.Any("conds", conds),
	)
	var m mdsmodel.MdsColonyModel
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()
	getMdsColonyStartTime := time.Now()
	err := database.DBGet(dbCtx, r.gormDB, preloads, &m, conds...)
	getMdsColonyDuration := time.Since(getMdsColonyStartTime)
	if err != nil {
		log.Error(
			"查询mds集群：数据库操作失败",
			zap.Error(err),
			zap.Any("conds", conds),
			zap.Duration("get_duration", getMdsColonyDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.WrapIf(err, "查询mds集群：数据库操作失败")
	}
	log.Debug(
		"查询mds集群：执行成功",
		zap.Object("colony_model", &m),
		zap.Any("conds", conds),
		zap.Duration("get_duration", getMdsColonyDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return &m, nil
}

func (r *MdsColonyRepo) ListModel(
	ctx context.Context,
	qp database.QueryParams,
) ([]mdsmodel.MdsColonyModel, error) {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"查询mds集群列表：开始执行",
		zap.Object("query_params", &qp),
	)
	var ms []mdsmodel.MdsColonyModel
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ListTimeout)
	defer cancel()
	listMdsColonyStartTime := time.Now()
	err := database.DBList(dbCtx, r.gormDB, &mdsmodel.MdsColonyModel{}, &ms, qp)
	listMdsColonyDuration := time.Since(listMdsColonyStartTime)
	if err != nil {
		log.Error(
			"查询mds集群列表：数据库操作失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("list_duration", listMdsColonyDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.WrapIf(err, "查询mds集群列表：数据库操作失败")
	}
	log.Debug(
		"查询mds集群列表：执行成功",
		zap.Object("query_params", &qp),
		zap.Duration("list_duration", listMdsColonyDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return ms, nil
}

func (r *MdsColonyRepo) CountModel(
	ctx context.Context,
	query map[string]any,
) (int64, error) {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"查询mds集群总数：开始执行",
		zap.Any("query", query),
	)
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()
	countMdsColonyStartTime := time.Now()
	count, err := database.DBCount(dbCtx, r.gormDB, &mdsmodel.MdsColonyModel{}, query)
	countMdsColonyDuration := time.Since(countMdsColonyStartTime)
	if err != nil {
		log.Error(
			"查询mds集群总数：数据库查询失败",
			zap.Error(err),
			zap.Any("query", query),
			zap.Duration("count_duration", countMdsColonyDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, errors.WrapIf(err, "查询mds集群总数：数据库查询失败")
	}
	log.Debug(
		"查询mds集群总数：执行成功",
		zap.Any("query", query),
		zap.Int64("count", count),
		zap.Duration("count_duration", countMdsColonyDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return count, nil
}

func (r *MdsColonyRepo) SaveConfigFile(
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
			err := errors.New("保存mds集群配置文件：配置文件已存在")
			log.Error(
				"保存mds集群配置文件：配置文件已存在",
				zap.Error(err),
				zap.String("pkg_path", pkgPath),
				zap.Duration("total_duration", time.Since(startTime)),
			)
			return err
		}
	}

	log.Debug(
		"保存mds集群配置文件：开始执行",
		zap.String("pkg_path", pkgPath),
		zap.Bool("overwrite", overwrite),
	)

	mode := os.FileMode(0o644)
	dir := filepath.Dir(pkgPath)
	if err := os.MkdirAll(dir, mode); err != nil {
		log.Error(
			"保存mds集群配置文件：创建配置文件目录失败",
			zap.Error(err),
			zap.String("pkg_dir", dir),
		)
		return errors.WrapIf(err, "保存mds集群配置文件：创建配置文件目录失败")
	}
	if err := os.Chmod(dir, mode); err != nil {
		// 忽略权限设置错误，因为目录可能已经存在并且权限正确
		log.Debug(
			"保存mds集群配置文件：设置配置文件目录权限失败（忽略）",
			zap.Error(err),
			zap.String("pkg_dir", dir),
		)
	}

	// 创建文件并写入内容
	file, err := os.OpenFile(pkgPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		log.Error(
			"保存mds集群配置文件：创建配置文件失败",
			zap.Error(err),
			zap.String("pkg_path", pkgPath),
		)
		return errors.WrapIf(err, "保存mds集群配置文件：创建配置文件失败")
	}
	defer file.Close()

	if _, err = io.Copy(file, fileReader); err != nil {
		log.Error(
			"保存mds集群配置文件：写入配置文件失败",
			zap.Error(err),
			zap.String("save_path", pkgPath),
		)
		return errors.WrapIf(err, "保存mds集群配置文件：写入配置文件失败")
	}
	log.Debug(
		"保存mds集群配置文件：执行成功",
		zap.String("pkg_path", pkgPath),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

func (r *MdsColonyRepo) RemoveConfigFile(
	ctx context.Context,
	pkgPath string,
) error {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"删除mds集群配置文件：开始执行",
		zap.String("pkg_path", pkgPath),
	)

	// 检查文件是否存在
	if _, err := os.Stat(pkgPath); err != nil {
		if os.IsNotExist(err) {
			log.Warn(
				"删除mds集群配置文件：配置文件不存在",
				zap.Error(err),
				zap.String("pkg_path", pkgPath),
				zap.Duration("total_duration", time.Since(startTime)),
			)
			return nil
		}
		log.Error(
			"删除mds集群配置文件：检查配置文件失败",
			zap.Error(err),
			zap.String("pkg_path", pkgPath),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "删除mds集群配置文件：检查配置文件失败")
	}
	// 删除文件
	removeMdsColonyStartTime := time.Now()
	err := os.Remove(pkgPath)
	removeMdsColonyDuration := time.Since(removeMdsColonyStartTime)
	if err != nil {
		log.Error(
			"删除mds集群配置文件：文件删除失败",
			zap.Error(err),
			zap.String("pkg_path", pkgPath),
			zap.Duration("remove_duration", removeMdsColonyDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "删除mds集群配置文件：文件删除失败")
	}
	log.Debug(
		"删除mds集群配置文件：执行成功",
		zap.String("pkg_path", pkgPath),
		zap.Duration("remove_duration", removeMdsColonyDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}
