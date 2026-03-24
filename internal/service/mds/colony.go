package mds

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"

	mdsmodel "gin-artweb/internal/model/mds"
	mdsrepo "gin-artweb/internal/repo/mds"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
	"gin-artweb/pkg/archive"
	"gin-artweb/pkg/fileutil"
	"gin-artweb/pkg/serializer"
)

type MdsColonyService struct {
	log        *zap.Logger
	colonyRepo *mdsrepo.MdsColonyRepo
	cronSvc    *MdsCronService
}

func NewMdsColonyService(
	log *zap.Logger,
	colonyRepo *mdsrepo.MdsColonyRepo,
	cronSvc *MdsCronService,
) *MdsColonyService {
	return &MdsColonyService{
		log:        log,
		colonyRepo: colonyRepo,
		cronSvc:    cronSvc,
	}
}

func (s *MdsColonyService) CreateMdsColony(
	ctx context.Context,
	dto *mdsmodel.MdsColonyUpsertDTO,
) (*mdsmodel.MdsColonyModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info("创建mds集群：开始执行")

	log.Debug(
		"创建mds集群：入参详情",
		zap.Object("mds_colony_dto", dto),
	)

	m := mdsmodel.MdsColonyModel{}

	createStepStart := time.Now()
	log.Debug(
		"创建mds集群：开始创建数据库模型",
		zap.Object("mds_colony_model", &m),
	)
	if err := s.colonyRepo.CreateModel(ctx, &m); err != nil {
		log.Error(
			"创建mds集群：创建数据库模型失败",
			zap.Error(err),
			zap.Object("mds_colony_model", &m),
			zap.Duration("create_step_duration", time.Since(createStepStart)),
		)
		return nil, errors.NewGormError(err, nil)
	}
	createStepDuration := time.Since(createStepStart)
	log.Debug(
		"创建mds集群：创建数据库模型成功",
		zap.Uint32("mds_colony_id", m.ID),
		zap.Duration("create_step_duration", createStepDuration),
	)

	queryStepStart := time.Now()
	log.Debug(
		"创建mds集群：开始查询mds集群关联数据",
		zap.Uint32("mds_colony_id", m.ID),
		zap.Strings("preloads", []string{"Package", "MonNode"}),
	)
	// 查询mds集群关联数据
	nm, rErr := s.FindMdsColonyByID(ctx, []string{"Package", "MonNode"}, m.ID)
	if rErr != nil {
		log.Error(
			"创建mds集群：查询mds集群关联数据失败",
			zap.Error(rErr),
			zap.Uint32("mds_colony_id", m.ID),
			zap.Duration("query_step_duration", time.Since(queryStepStart)),
		)
		return nil, rErr
	}
	queryStepDuration := time.Since(queryStepStart)
	log.Debug(
		"创建mds集群：查询mds集群关联数据成功",
		zap.Uint32("mds_colony_id", m.ID),
		zap.Duration("query_step_duration", queryStepDuration),
	)

	exportStepStart := time.Now()
	log.Debug(
		"创建mds集群：开始导出mds集群缓存数据",
		zap.Uint32("mds_colony_id", m.ID),
	)
	// 导出mds集群缓存数据
	if err := s.OutportMdsColonyData(ctx, nm); err != nil {
		log.Error(
			"创建mds集群：导出mds集群缓存数据失败",
			zap.Error(err),
			zap.Uint32("mds_colony_id", m.ID),
			zap.Duration("export_step_duration", time.Since(exportStepStart)),
		)
		return nil, err
	}
	exportStepDuration := time.Since(exportStepStart)
	log.Debug(
		"创建mds集群：导出mds集群缓存数据成功",
		zap.Uint32("mds_colony_id", m.ID),
		zap.Duration("export_step_duration", exportStepDuration),
	)

	// 初始化mds集群定时任务
	initCronStepStart := time.Now()
	log.Debug(
		"创建mds集群：开始初始化mds集群定时任务",
		zap.Uint32("mds_colony_id", m.ID),
	)
	rErr = s.cronSvc.CreateCornByColony(ctx, nm)
	initCronStepDuration := time.Since(initCronStepStart)
	if rErr != nil {
		log.Error(
			"创建mds集群：初始化mds集群定时任务失败",
			zap.Error(rErr),
			zap.Uint32("mds_colony_id", m.ID),
			zap.Duration("init_cron_step_duration", initCronStepDuration),
		)
		return nil, rErr
	}
	log.Debug(
		"创建mds集群：初始化mds集群定时任务成功",
		zap.Uint32("mds_colony_id", m.ID),
		zap.Duration("init_cron_step_duration", initCronStepDuration),
	)

	log.Info(
		"创建mds集群：执行成功",
		zap.Uint32("mds_colony_id", m.ID),
		zap.Duration("create_step_duration", createStepDuration),
		zap.Duration("query_step_duration", queryStepDuration),
		zap.Duration("export_step_duration", exportStepDuration),
		zap.Duration("init_cron_step_duration", initCronStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	return nm, nil
}

func (s *MdsColonyService) UpdateMdsColonyByID(
	ctx context.Context,
	mdsColonyID uint32,
	dto mdsmodel.MdsColonyUpsertDTO,
) (*mdsmodel.MdsColonyModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"更新mds集群：开始执行",
		zap.Uint32("mds_colony_id", mdsColonyID),
	)

	log.Debug(
		"更新mds集群：入参详情",
		zap.Uint32("mds_colony_id", mdsColonyID),
		zap.Object("mds_colony_dto", &dto),
	)

	findOldStepStart := time.Now()
	om, rErr := s.FindMdsColonyByID(ctx, nil, mdsColonyID)
	if rErr != nil {
		log.Error(
			"更新mds集群：查询更新前mds集群关联数据失败",
			zap.Error(rErr),
			zap.Uint32("mds_colony_id", mdsColonyID),
			zap.Duration("find_old_step_duration", time.Since(findOldStepStart)),
		)
		return nil, rErr
	}
	findOldStepDuration := time.Since(findOldStepStart)
	log.Debug(
		"更新mds集群：查询更新前mds集群关联数据成功",
		zap.Object("mds_colony_model", om),
		zap.Duration("find_old_step_duration", findOldStepDuration),
	)

	updateData := dto.ToUpdateMap()
	log.Debug(
		"更新mds集群：转换为数据库更新参数",
		zap.Any("update_data", updateData),
	)
	updateStepStart := time.Now()
	log.Debug(
		"更新mds集群：开始更新数据库模型",
		zap.Uint32("mds_colony_id", mdsColonyID),
		zap.Any("update_data", updateData),
	)
	if err := s.colonyRepo.UpdateModel(ctx, updateData, "id = ?", mdsColonyID); err != nil {
		log.Error(
			"更新mds集群：更新数据库模型失败",
			zap.Error(err),
			zap.Uint32("mds_colony_id", mdsColonyID),
			zap.Any("update_data", updateData),
			zap.Duration("update_step_duration", time.Since(updateStepStart)),
		)
		return nil, errors.NewGormError(err, updateData)
	}
	updateStepDuration := time.Since(updateStepStart)
	log.Debug(
		"更新mds集群：更新数据库模型成功",
		zap.Uint32("mds_colony_id", mdsColonyID),
		zap.Duration("update_step_duration", updateStepDuration),
	)

	queryStepStart := time.Now()
	log.Debug(
		"更新mds集群：开始查询mds集群关联数据",
		zap.Uint32("mds_colony_id", mdsColonyID),
		zap.Strings("preloads", []string{"Package", "MonNode"}),
	)
	// 查询mds集群关联数据
	nm, rErr := s.FindMdsColonyByID(ctx, []string{"Package", "MonNode"}, mdsColonyID)
	if rErr != nil {
		log.Error(
			"更新mds集群：查询mds集群关联数据失败",
			zap.Error(rErr),
			zap.Uint32("mds_colony_id", mdsColonyID),
			zap.Duration("query_step_duration", time.Since(queryStepStart)),
		)
		return nil, rErr
	}
	queryStepDuration := time.Since(queryStepStart)
	log.Debug(
		"更新mds集群：查询mds集群关联数据成功",
		zap.Uint32("mds_colony_id", mdsColonyID),
		zap.Duration("query_step_duration", queryStepDuration),
	)
	exportStepStart := time.Now()
	log.Debug(
		"更新mds集群：开始导出mds集群缓存数据",
		zap.Uint32("mds_colony_id", mdsColonyID),
	)
	// 导出mds集群缓存数据
	if err := s.OutportMdsColonyData(ctx, nm); err != nil {
		log.Error(
			"更新mds集群：导出mds集群缓存数据失败",
			zap.Error(err),
			zap.Uint32("mds_colony_id", mdsColonyID),
			zap.Duration("export_step_duration", time.Since(exportStepStart)),
		)
		return nil, err
	}
	exportStepDuration := time.Since(exportStepStart)
	log.Debug(
		"更新mds集群：导出mds集群缓存数据成功",
		zap.Uint32("mds_colony_id", mdsColonyID),
		zap.Duration("export_step_duration", exportStepDuration),
	)

	if om.ColonyNum != nm.ColonyNum {
		clearStepStart := time.Now()
		log.Debug(
			"更新mds集群：开始清理计划任务",
			zap.Uint32("mds_colony_id", mdsColonyID),
		)
		if err := s.cronSvc.DeleteCornByColonyID(ctx, mdsColonyID); err != nil {
			log.Error(
				"更新mds集群：清理计划任务失败",
				zap.Error(err),
				zap.Uint32("mds_colony_id", mdsColonyID),
				zap.Duration("clear_step_duration", time.Since(clearStepStart)),
			)
			return nil, err
		}
		clearStepDuration := time.Since(clearStepStart)
		log.Debug(
			"更新mds集群：清理计划任务成功",
			zap.Uint32("mds_colony_id", mdsColonyID),
			zap.Duration("clear_step_duration", clearStepDuration),
		)

		initStepStart := time.Now()
		log.Debug(
			"更新mds集群：开始初始化计划任务",
			zap.Uint32("mds_colony_id", mdsColonyID),
		)
		if err := s.cronSvc.CreateCornByColony(ctx, nm); err != nil {
			log.Error(
				"更新mds集群：初始化计划任务失败",
				zap.Error(err),
				zap.Uint32("mds_colony_id", mdsColonyID),
				zap.Duration("init_step_duration", time.Since(initStepStart)),
			)
			return nil, err
		}
		initStepDuration := time.Since(initStepStart)
		log.Debug(
			"更新mds集群：初始化计划任务成功",
			zap.Uint32("mds_colony_id", mdsColonyID),
			zap.Duration("init_step_duration", initStepDuration),
		)
	}

	log.Info(
		"更新mds集群：执行成功",
		zap.Uint32("mds_colony_id", mdsColonyID),
		zap.Duration("update_step_duration", updateStepDuration),
		zap.Duration("query_step_duration", queryStepDuration),
		zap.Duration("export_step_duration", exportStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nm, nil
}

func (s *MdsColonyService) DeleteMdsColonyByID(
	ctx context.Context,
	mdsColonyID uint32,
) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"删除mds集群：开始执行",
		zap.Uint32("mds_colony_id", mdsColonyID),
	)

	deleteStepStart := time.Now()
	log.Debug(
		"删除mds集群：开始删除数据库模型",
		zap.Uint32("mds_colony_id", mdsColonyID),
	)
	if err := s.colonyRepo.DeleteModel(ctx, mdsColonyID); err != nil {
		log.Error(
			"删除mds集群：删除数据库模型失败",
			zap.Error(err),
			zap.Uint32("mds_colony_id", mdsColonyID),
			zap.Duration("delete_step_duration", time.Since(deleteStepStart)),
		)
		return errors.NewGormError(err, map[string]any{"id": mdsColonyID})
	}
	deleteStepDuration := time.Since(deleteStepStart)
	log.Debug(
		"删除mds集群：删除数据库模型成功",
		zap.Uint32("mds_colony_id", mdsColonyID),
		zap.Duration("delete_step_duration", deleteStepDuration),
	)

	clearStepStart := time.Now()
	log.Debug(
		"删除mds集群：开始清理计划任务",
		zap.Uint32("mds_colony_id", mdsColonyID),
	)
	if err := s.cronSvc.DeleteCornByColonyID(ctx, mdsColonyID); err != nil {
		log.Error(
			"删除mds集群：清理计划任务失败",
			zap.Error(err),
			zap.Uint32("mds_colony_id", mdsColonyID),
			zap.Duration("clear_step_duration", time.Since(clearStepStart)),
		)
		return err
	}
	clearStepDuration := time.Since(clearStepStart)
	log.Debug(
		"删除mds集群：清理计划任务成功",
		zap.Uint32("mds_colony_id", mdsColonyID),
		zap.Duration("clear_step_duration", clearStepDuration),
	)

	log.Info(
		"删除mds集群：执行成功",
		zap.Uint32("mds_colony_id", mdsColonyID),
		zap.Duration("delete_step_duration", deleteStepDuration),
		zap.Duration("clear_step_duration", clearStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

func (s *MdsColonyService) FindMdsColonyByID(
	ctx context.Context,
	preloads []string,
	mdsColonyID uint32,
) (*mdsmodel.MdsColonyModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"查询mds集群：开始执行",
		zap.Uint32("mds_colony_id", mdsColonyID),
		zap.Strings("preloads", preloads),
	)

	queryStepStart := time.Now()
	log.Debug(
		"查询mds集群：开始查询数据库模型",
		zap.Uint32("mds_colony_id", mdsColonyID),
		zap.Strings("preloads", preloads),
	)
	m, err := s.colonyRepo.GetModel(ctx, preloads, mdsColonyID)
	if err != nil {
		log.Error(
			"查询mds集群：查询数据库模型失败",
			zap.Error(err),
			zap.Uint32("mds_colony_id", mdsColonyID),
			zap.Strings("preloads", preloads),
			zap.Duration("query_step_duration", time.Since(queryStepStart)),
		)
		return nil, errors.NewGormError(err, map[string]any{"id": mdsColonyID})
	}
	queryStepDuration := time.Since(queryStepStart)
	log.Debug(
		"查询mds集群：查询到的数据库模型详情",
		zap.Object("mds_colony_model", m),
	)

	log.Info(
		"查询mds集群：执行成功",
		zap.Uint32("mds_colony_id", mdsColonyID),
		zap.Duration("query_step_duration", queryStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return m, nil
}

func (s *MdsColonyService) ListMdsColony(
	ctx context.Context,
	page, size int,
	dto mdsmodel.ListMdsColonyDTO,
) (int64, []mdsmodel.MdsColonyModel, *errors.Error) {
	if ctx.Err() != nil {
		return 0, nil, errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info("查询mds集群列表：开始执行")

	log.Debug(
		"查询mds集群列表：入参详情",
		zap.Int("page", page),
		zap.Int("size", size),
		zap.Object("mds_colony_dto", &dto),
	)

	limit, offset := common.Page2LimitOffset(page, size)
	qp := database.QueryParams{
		Preloads: []string{"Package", "MonNode"},
		OrderBy:  []string{"id DESC"},
		Limit:    limit,
		Offset:   offset,
		Query:    dto.ToQueryMap(),
	}
	log.Debug(
		"查询mds集群列表：查询参数",
		zap.Object("query_params", &qp),
	)

	countStepStart := time.Now()
	log.Debug(
		"查询mds集群列表：开始查询数据库模型总数",
		zap.Any("query", qp.Query),
	)
	count, err := s.colonyRepo.CountModel(ctx, qp.Query)
	if err != nil {
		log.Error(
			"查询mds集群列表：查询数据库模型总数失败",
			zap.Error(err),
			zap.Any("query", qp.Query),
			zap.Duration("count_step_duration", time.Since(countStepStart)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}
	countStepDuration := time.Since(countStepStart)
	log.Debug(
		"查询mds集群列表：查询数据库模型总数成功",
		zap.Any("query", qp.Query),
		zap.Duration("count_step_duration", countStepDuration),
	)
	if count == 0 {
		return count, nil, nil
	}

	listStepStart := time.Now()
	log.Debug(
		"查询mds集群列表：开始查询数据库模型",
		zap.Any("query", qp.Query),
	)
	ms, err := s.colonyRepo.ListModel(ctx, qp)
	if err != nil {
		log.Error(
			"查询mds集群列表：查询数据库模型失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("list_step_duration", time.Since(listStepStart)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}
	listStepDuration := time.Since(listStepStart)
	log.Info(
		"查询mds集群列表：执行成功",
		zap.Int64("total_count", count),
		zap.Duration("count_step_duration", countStepDuration),
		zap.Duration("list_step_duration", listStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return count, ms, nil
}

func (s *MdsColonyService) OutportMdsColonyData(
	ctx context.Context,
	m *mdsmodel.MdsColonyModel,
) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"导出mds集群数据：开始执行",
		zap.Uint32("mds_colony_id", m.ID),
		zap.String("colony_num", m.ColonyNum),
	)

	log.Debug(
		"导出mds集群数据：入参详情",
		zap.Object("mds_colony_model", m),
	)

	colonyBinDir := common.GetMdsColonyBinDir(m.ColonyNum)
	colonyConfDir := common.GetMdsColonyConfigDir(m.ColonyNum)

	cleanStepStart := time.Now()
	log.Debug(
		"导出mds集群数据：开始清理原mds集群配置文件",
		zap.String("path", colonyBinDir),
	)
	if _, err := os.Stat(colonyBinDir); !os.IsNotExist(err) {
		if err := os.RemoveAll(colonyBinDir); err != nil {
			log.Error(
				"导出mds集群数据：清理原mds集群配置文件失败",
				zap.Error(err),
				zap.String("path", colonyBinDir),
				zap.Duration("clean_step_duration", time.Since(cleanStepStart)),
			)
			return errors.ErrDeleteCacheFileFailed.WithCause(err)
		}
	}
	cleanStepDuration := time.Since(cleanStepStart)
	log.Debug(
		"导出mds集群数据：清理原mds集群配置文件成功",
		zap.String("path", colonyBinDir),
		zap.Duration("clean_step_duration", cleanStepDuration),
	)

	tempStepStart := time.Now()
	log.Debug(
		"导出mds集群数据：开始创建临时文件夹",
	)
	tmpDir, mErr := os.MkdirTemp("/tmp", "mds-")
	if mErr != nil {
		log.Error(
			"导出mds集群数据：创建临时文件夹失败",
			zap.Error(mErr),
			zap.Duration("temp_step_duration", time.Since(tempStepStart)),
		)
		return errors.FromError(mErr)
	}
	tempStepDuration := time.Since(tempStepStart)
	log.Debug(
		"导出mds集群数据：创建临时文件夹成功",
		zap.String("path", tmpDir),
		zap.Duration("temp_step_duration", tempStepDuration),
	)
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			log.Error(
				"导出mds集群数据：删除临时文件夹失败",
				zap.Error(err),
				zap.String("path", tmpDir),
			)
		}
	}()

	validateStepStart := time.Now()
	mdsPkgPath := common.GetPackageStoragePath(m.Package.StorageFilename)
	log.Debug(
		"导出mds集群数据：开始校验mds程序包",
		zap.String("path", mdsPkgPath),
	)
	mdsUnTarDirName, valiErr := archive.ValidateSingleDirTarGz(mdsPkgPath)
	if valiErr != nil {
		log.Error(
			"导出mds集群数据：mds程序包校验失败",
			zap.Error(valiErr),
			zap.String("path", mdsPkgPath),
			zap.Duration("validate_step_duration", time.Since(validateStepStart)),
		)
		return errors.ErrValidationFailed.WithCause(valiErr)
	}
	validateStepDuration := time.Since(validateStepStart)
	log.Debug(
		"导出mds集群数据：mds程序包校验成功",
		zap.String("path", mdsPkgPath),
		zap.String("un_tar_dir_name", mdsUnTarDirName),
		zap.Duration("validate_step_duration", validateStepDuration),
	)

	extractStepStart := time.Now()
	log.Debug(
		"导出mds集群数据：开始解压mds程序包",
		zap.String("src_path", mdsPkgPath),
		zap.String("dest_path", tmpDir),
	)
	if err := archive.UntarGz(mdsPkgPath, tmpDir, archive.WithContext(ctx)); err != nil {
		log.Error(
			"导出mds集群数据：解压mds程序包失败",
			zap.Error(err),
			zap.Uint32("mds_colony_id", m.ID),
			zap.String("pkg_name", m.ExtractedName),
			zap.String("src_path", mdsPkgPath),
			zap.String("dest_path", tmpDir),
			zap.Duration("extract_step_duration", time.Since(extractStepStart)),
		)
		return errors.ErrUnZIPFailed.WithCause(err).WithField("pkg_name", m.ExtractedName)
	}
	extractStepDuration := time.Since(extractStepStart)
	log.Debug(
		"导出mds集群数据：解压mds程序包成功",
		zap.String("src_path", mdsPkgPath),
		zap.String("dest_path", tmpDir),
		zap.Duration("extract_step_duration", extractStepDuration),
	)

	copyStepStart := time.Now()
	mdsTmpDir := filepath.Join(tmpDir, mdsUnTarDirName)
	log.Debug(
		"导出mds集群数据：开始复制mds程序包解压目录",
		zap.String("src_path", mdsTmpDir),
		zap.String("dst_path", colonyBinDir),
	)
	if err := fileutil.CopyDir(ctx, mdsTmpDir, colonyBinDir, true); err != nil {
		log.Error(
			"导出mds集群数据：复制mds程序包解压目录失败",
			zap.Error(err),
			zap.String("src_path", mdsTmpDir),
			zap.String("dst_path", colonyBinDir),
			zap.Duration("copy_step_duration", time.Since(copyStepStart)),
		)
		return errors.FromError(err)
	}
	copyStepDuration := time.Since(copyStepStart)
	log.Debug(
		"导出mds集群数据：复制mds程序包解压目录成功",
		zap.String("src_path", mdsTmpDir),
		zap.String("dst_path", colonyBinDir),
		zap.Duration("copy_step_duration", copyStepDuration),
	)

	configStepStart := time.Now()
	colonyConfAll := filepath.Join(colonyConfDir, "all")
	log.Debug(
		"导出mds集群数据：开始处理mds集群配置文件",
		zap.String("path", colonyConfAll),
	)
	if _, err := os.Stat(colonyConfAll); os.IsNotExist(err) {
		colonyBinConf := filepath.Join(colonyBinDir, "conf")
		if err := fileutil.CopyDir(ctx, colonyBinConf, colonyConfAll, true); err != nil {
			log.Error(
				"导出mds集群数据：复制mds集群配置文件失败",
				zap.Error(err),
				zap.String("src_path", colonyBinConf),
				zap.String("dst_path", colonyConfAll),
				zap.Duration("config_step_duration", time.Since(configStepStart)),
			)
			return errors.FromError(err)
		}
		srcPath := filepath.Join(config.ConfigDir, "automatic_mds.yaml")
		dstPath := filepath.Join(colonyConfAll, "automatic.yaml")
		if err := fileutil.CopyFile(ctx, srcPath, dstPath); err != nil {
			log.Error(
				"导出mds集群数据：复制mds的automatic配置文件失败",
				zap.Error(err),
				zap.String("src_path", srcPath),
				zap.String("dst_path", dstPath),
				zap.Duration("config_step_duration", time.Since(configStepStart)),
			)
			return errors.FromError(err)
		}
	}
	configStepDuration := time.Since(configStepStart)
	log.Debug(
		"导出mds集群数据：处理mds集群配置文件成功",
		zap.String("path", colonyConfAll),
		zap.Duration("config_step_duration", configStepDuration),
	)

	exportStepStart := time.Now()
	mdsVars := mdsmodel.MdsColonyVars{
		ID:        m.ID,
		ColonyNum: m.ColonyNum,
		PkgName:   m.ExtractedName,
		PackageID: m.PackageID,
		Version:   m.Package.Version,
		MonNodeID: m.MonNodeID,
		IsEnable:  m.IsEnable,
	}
	mdsColonyConf := filepath.Join(colonyConfAll, "colony.yaml")
	log.Debug(
		"导出mds集群数据：开始导出mds集群配置变量文件",
		zap.String("path", mdsColonyConf),
		zap.Object("mds_colony_vars", &mdsVars),
	)
	if _, err := serializer.WriteYAML(mdsColonyConf, mdsVars); err != nil {
		log.Error(
			"导出mds集群数据：导出mds集群配置变量文件失败",
			zap.Error(err),
			zap.String("path", mdsColonyConf),
			zap.Object("mds_colony_vars", &mdsVars),
			zap.Duration("export_step_duration", time.Since(exportStepStart)),
		)
		return errors.ErrExportCacheFileFailed.WithCause(err)
	}
	exportStepDuration := time.Since(exportStepStart)
	log.Debug(
		"导出mds集群数据：导出mds集群配置变量文件成功",
		zap.String("path", mdsColonyConf),
		zap.Object("mds_colony_vars", &mdsVars),
		zap.Duration("export_step_duration", exportStepDuration),
	)

	log.Info(
		"导出mds集群数据：执行成功",
		zap.Uint32("mds_colony_id", m.ID),
		zap.String("colony_num", m.ColonyNum),
		zap.String("config_path", mdsColonyConf),
		zap.Duration("clean_step_duration", cleanStepDuration),
		zap.Duration("temp_step_duration", tempStepDuration),
		zap.Duration("validate_step_duration", validateStepDuration),
		zap.Duration("extract_step_duration", extractStepDuration),
		zap.Duration("copy_step_duration", copyStepDuration),
		zap.Duration("config_step_duration", configStepDuration),
		zap.Duration("export_step_duration", exportStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}
