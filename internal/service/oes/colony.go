package oes

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"

	oesmodel "gin-artweb/internal/model/oes"
	oesrepo "gin-artweb/internal/repo/oes"
	resosvc "gin-artweb/internal/service/resource"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
	"gin-artweb/pkg/archive"
	"gin-artweb/pkg/fileutil"
	"gin-artweb/pkg/serializer"
)

type OesColonyService struct {
	log        *zap.Logger
	colonyRepo *oesrepo.OesColonyRepo
	cronSvc    *OesCronService
}

func NewOesColonyService(
	log *zap.Logger,
	colonyRepo *oesrepo.OesColonyRepo,
	cronSvc *OesCronService,
) *OesColonyService {
	return &OesColonyService{
		log:        log,
		colonyRepo: colonyRepo,
		cronSvc:    cronSvc,
	}
}

func (s *OesColonyService) CreateOesColony(
	ctx context.Context,
	dto oesmodel.OesColonyUpsertDTO,
) (*oesmodel.OesColonyModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info("创建oes集群：开始执行")

	log.Debug(
		"创建oes集群：入参详情",
		zap.Object("oes_colony_dto", &dto),
	)

	m := oesmodel.OesColonyModel{
		SystemType:    dto.SystemType,
		ColonyNum:     dto.ColonyNum,
		ExtractedName: dto.ExtractedName,
		IsEnable:      dto.IsEnable,
		PackageID:     dto.PackageID,
		XCounterID:    dto.XCounterID,
		MonNodeID:     dto.MonNodeID,
	}
	createStepStart := time.Now()
	log.Debug(
		"创建oes集群：开始创建数据库模型",
		zap.Object("oes_colony", &m),
	)
	if err := s.colonyRepo.CreateModel(ctx, &m); err != nil {
		log.Error(
			"创建oes集群：创建数据库模型失败",
			zap.Error(err),
			zap.Object("oes_colony", &m),
			zap.Duration("create_step_duration", time.Since(createStepStart)),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, nil)
	}
	createStepDuration := time.Since(createStepStart)
	log.Debug(
		"创建oes集群：创建数据库模型成功",
		zap.Uint32("oes_colony_id", m.ID),
		zap.Duration("create_step_duration", createStepDuration),
	)

	// 查询oes集群关联数据
	queryStepStart := time.Now()
	log.Debug(
		"创建oes集群：开始查询关联数据",
		zap.Uint32("oes_colony_id", m.ID),
	)
	nm, rErr := s.FindOesColonyByID(ctx, []string{"Package", "XCounter", "MonNode"}, m.ID)
	queryStepDuration := time.Since(queryStepStart)
	if rErr != nil {
		log.Error(
			"创建oes集群：查询关联数据失败",
			zap.Error(rErr),
			zap.Uint32("oes_colony_id", m.ID),
			zap.Duration("query_step_duration", queryStepDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, rErr
	}
	log.Debug(
		"创建oes集群：查询关联数据成功",
		zap.Uint32("oes_colony_id", m.ID),
		zap.Duration("query_step_duration", queryStepDuration),
	)

	// 导出oes集群缓存数据
	exportStepStart := time.Now()
	log.Debug(
		"创建oes集群：开始导出缓存数据",
		zap.Uint32("oes_colony_id", m.ID),
	)
	if err := s.OutportOesColonyData(ctx, nm); err != nil {
		log.Error(
			"创建oes集群：导出缓存数据失败",
			zap.Error(err),
			zap.Uint32("oes_colony_id", m.ID),
			zap.Duration("export_step_duration", time.Since(exportStepStart)),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, err
	}
	exportStepDuration := time.Since(exportStepStart)
	log.Debug(
		"创建oes集群：导出缓存数据成功",
		zap.Uint32("oes_colony_id", m.ID),
		zap.Duration("export_step_duration", exportStepDuration),
	)

	// 初始化mds集群定时任务
	initCronStepStart := time.Now()
	log.Debug(
		"创建oes集群：开始初始化oes集群定时任务",
		zap.Uint32("oes_colony_id", m.ID),
	)
	rErr = s.cronSvc.CreateCornByColony(ctx, nm)
	initCronStepDuration := time.Since(initCronStepStart)
	if rErr != nil {
		log.Error(
			"创建oes集群：初始化oes集群定时任务失败",
			zap.Error(rErr),
			zap.Uint32("oes_colony_id", m.ID),
			zap.Duration("init_cron_step_duration", initCronStepDuration),
		)
		return nil, rErr
	}
	log.Debug(
		"创建oes集群：初始化oes集群定时任务成功",
		zap.Uint32("oes_colony_id", m.ID),
		zap.Duration("init_cron_step_duration", initCronStepDuration),
	)

	log.Info(
		"创建oes集群：执行成功",
		zap.Uint32("oes_colony_id", m.ID),
		zap.Duration("create_step_duration", createStepDuration),
		zap.Duration("query_step_duration", queryStepDuration),
		zap.Duration("export_step_duration", exportStepDuration),
		zap.Duration("init_cron_step_duration", initCronStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nm, nil
}

func (s *OesColonyService) UpdateOesColonyByID(
	ctx context.Context,
	oesColonyID uint32,
	dto oesmodel.OesColonyUpsertDTO,
) (*oesmodel.OesColonyModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"更新oes集群：开始执行",
		zap.Uint32("oes_colony_id", oesColonyID),
		zap.Object("oes_colony_dto", &dto),
	)

	findOldStepStart := time.Now()
	om, rErr := s.FindOesColonyByID(ctx, nil, oesColonyID)
	findOldStepDuration := time.Since(findOldStepStart)
	if rErr != nil {
		log.Error(
			"更新oes集群：查询更新前oes集群数据失败",
			zap.Error(rErr),
			zap.Uint32("oes_colony_id", oesColonyID),
			zap.Duration("find_old_step_duration", findOldStepDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, rErr
	}
	log.Debug(
		"更新oes集群：查询更新前oes集群数据成功",
		zap.Uint32("oes_colony_id", oesColonyID),
		zap.Duration("find_old_step_duration", findOldStepDuration),
	)

	updateStepStart := time.Now()
	updateData := dto.ToUpdateMap()
	log.Debug(
		"更新oes集群：开始更新数据库模型",
		zap.Uint32("oes_colony_id", oesColonyID),
		zap.Any("update_data", updateData),
	)
	if err := s.colonyRepo.UpdateModel(ctx, updateData, "id = ?", oesColonyID); err != nil {
		log.Error(
			"更新oes集群：更新数据库模型失败",
			zap.Error(err),
			zap.Uint32("oes_colony_id", oesColonyID),
			zap.Any("update_data", updateData),
			zap.Duration("update_step_duration", time.Since(updateStepStart)),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, updateData)
	}
	updateStepDuration := time.Since(updateStepStart)
	log.Debug(
		"更新oes集群：更新数据库模型成功",
		zap.Uint32("oes_colony_id", oesColonyID),
		zap.Duration("update_step_duration", updateStepDuration),
	)

	// 查询关联数据
	queryStepStart := time.Now()
	log.Debug(
		"更新oes集群：开始查询关联数据",
		zap.Uint32("oes_colony_id", oesColonyID),
	)
	nm, rErr := s.FindOesColonyByID(ctx, []string{"Package", "XCounter", "MonNode"}, oesColonyID)
	queryStepDuration := time.Since(queryStepStart)
	if rErr != nil {
		log.Error(
			"更新oes集群：查询关联数据失败",
			zap.Error(rErr),
			zap.Uint32("oes_colony_id", oesColonyID),
			zap.Duration("query_step_duration", queryStepDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, rErr
	}
	log.Debug(
		"更新oes集群：查询关联数据成功",
		zap.Uint32("oes_colony_id", oesColonyID),
		zap.Duration("query_step_duration", queryStepDuration),
	)

	// 导出数据库缓存数据
	exportStepStart := time.Now()
	log.Debug(
		"更新oes集群：开始导出缓存数据",
		zap.Uint32("oes_colony_id", oesColonyID),
	)
	if err := s.OutportOesColonyData(ctx, nm); err != nil {
		log.Error(
			"更新oes集群：导出缓存数据失败",
			zap.Error(err),
			zap.Uint32("oes_colony_id", oesColonyID),
			zap.Duration("export_step_duration", time.Since(exportStepStart)),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, err
	}
	exportStepDuration := time.Since(exportStepStart)
	log.Debug(
		"更新oes集群：导出缓存数据成功",
		zap.Uint32("oes_colony_id", oesColonyID),
		zap.Duration("export_step_duration", exportStepDuration),
	)

	if om.ColonyNum != nm.ColonyNum || om.IsEnable != nm.IsEnable {
		clearStepStart := time.Now()
		log.Debug(
			"更新oes集群：开始清理计划任务",
			zap.Uint32("oes_colony_id", oesColonyID),
		)
		if err := s.cronSvc.DeleteCornByColonyID(ctx, oesColonyID); err != nil {
			log.Error(
				"更新oes集群：清理计划任务失败",
				zap.Error(err),
				zap.Uint32("oes_colony_id", oesColonyID),
				zap.Duration("clear_step_duration", time.Since(clearStepStart)),
			)
			return nil, err
		}
		clearStepDuration := time.Since(clearStepStart)
		log.Debug(
			"更新oes集群：清理计划任务成功",
			zap.Uint32("oes_colony_id", oesColonyID),
			zap.Duration("clear_step_duration", clearStepDuration),
		)

		initStepStart := time.Now()
		log.Debug(
			"更新oes集群：开始初始化计划任务",
			zap.Uint32("oes_colony_id", oesColonyID),
		)
		if err := s.cronSvc.CreateCornByColony(ctx, nm); err != nil {
			log.Error(
				"更新oes集群：初始化计划任务失败",
				zap.Error(err),
				zap.Uint32("oes_colony_id", oesColonyID),
				zap.Duration("init_step_duration", time.Since(initStepStart)),
			)
			return nil, err
		}
		initStepDuration := time.Since(initStepStart)
		log.Debug(
			"更新oes集群：初始化计划任务成功",
			zap.Uint32("oes_colony_id", oesColonyID),
			zap.Duration("init_step_duration", initStepDuration),
		)
	}

	log.Info(
		"更新oes集群：执行成功",
		zap.Uint32("oes_colony_id", oesColonyID),
		zap.Duration("update_step_duration", updateStepDuration),
		zap.Duration("query_step_duration", queryStepDuration),
		zap.Duration("export_step_duration", exportStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nm, nil
}

func (s *OesColonyService) DeleteOesColonyByID(
	ctx context.Context,
	oesColonyID uint32,
) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"删除oes集群：开始执行",
		zap.Uint32("oes_colony_id", oesColonyID),
	)

	deleteStepStart := time.Now()
	log.Debug(
		"删除oes集群：开始删除数据库模型",
		zap.Uint32("oes_colony_id", oesColonyID),
	)
	if err := s.colonyRepo.DeleteModel(ctx, oesColonyID); err != nil {
		log.Error(
			"删除oes集群：删除数据库模型失败",
			zap.Error(err),
			zap.Uint32("oes_colony_id", oesColonyID),
			zap.Duration("delete_step_duration", time.Since(deleteStepStart)),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.NewGormError(err, map[string]any{"id": oesColonyID})
	}
	deleteStepDuration := time.Since(deleteStepStart)
	log.Debug(
		"删除oes集群：删除数据库模型成功",
		zap.Uint32("oes_colony_id", oesColonyID),
		zap.Duration("delete_step_duration", deleteStepDuration),
	)

	clearStepStart := time.Now()
	log.Debug(
		"删除oes集群：开始清理计划任务",
		zap.Uint32("oes_colony_id", oesColonyID),
	)
	if err := s.cronSvc.DeleteCornByColonyID(ctx, oesColonyID); err != nil {
		log.Error(
			"删除oes集群：清理计划任务失败",
			zap.Error(err),
			zap.Uint32("oes_colony_id", oesColonyID),
			zap.Duration("clear_step_duration", time.Since(clearStepStart)),
		)
		return err
	}
	clearStepDuration := time.Since(clearStepStart)
	log.Debug(
		"删除oes集群：清理计划任务成功",
		zap.Uint32("oes_colony_id", oesColonyID),
		zap.Duration("clear_step_duration", clearStepDuration),
	)

	log.Info(
		"删除oes集群：执行成功",
		zap.Uint32("oes_colony_id", oesColonyID),
		zap.Duration("delete_step_duration", deleteStepDuration),
		zap.Duration("clear_step_duration", clearStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

func (s *OesColonyService) FindOesColonyByID(
	ctx context.Context,
	preloads []string,
	oesColonyID uint32,
) (*oesmodel.OesColonyModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"查询oes集群：开始执行",
		zap.Strings("preloads", preloads),
		zap.Uint32("oes_colony_id", oesColonyID),
	)

	m, err := s.colonyRepo.GetModel(ctx, preloads, oesColonyID)
	if err != nil {
		log.Error(
			"查询oes集群：查询数据库模型失败",
			zap.Error(err),
			zap.Uint32("oes_colony_id", oesColonyID),
			zap.Strings("preloads", preloads),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, map[string]any{"id": oesColonyID})
	}
	log.Debug(
		"查询oes集群：查询到的数据库模型详情",
		zap.Object("oes_colony_model", m),
	)

	log.Info(
		"查询oes集群：执行成功",
		zap.Uint32("oes_colony_id", oesColonyID),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return m, nil
}

func (s *OesColonyService) ListOesColony(
	ctx context.Context,
	page, size int,
	dto oesmodel.ListOesColonyDTO,
) (int64, []oesmodel.OesColonyModel, *errors.Error) {
	if ctx.Err() != nil {
		return 0, nil, errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"查询oes集群列表：开始执行",
		zap.Object("dto", &dto),
	)

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
		"查询oes集群列表：开始查询数据库模型总数",
		zap.Object("query_params", &qp),
	)
	count, err := s.colonyRepo.CountModel(ctx, qp.Query)
	countStepDuration := time.Since(countStepStart)
	if err != nil {
		log.Error(
			"查询oes集群列表：查询数据库模型总数失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("count_step_duration", countStepDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}
	log.Debug(
		"查询oes集群列表：查询数据库模型总数成功",
		zap.Int64("total_count", count),
		zap.Duration("count_step_duration", countStepDuration),
	)
	if count == 0 {
		log.Warn(
			"查询oes集群列表：数据库模型总数为0",
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return count, nil, nil
	}

	listStepStart := time.Now()
	log.Debug(
		"查询oes集群列表：开始查询数据库模型列表",
		zap.Object("query_params", &qp),
	)
	ms, err := s.colonyRepo.ListModel(ctx, qp)
	listStepDuration := time.Since(listStepStart)
	if err != nil {
		log.Error(
			"查询oes集群列表：查询数据库模型列表失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("list_step_duration", listStepDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}
	log.Debug(
		"查询oes集群列表：查询数据库模型列表成功",
		zap.Int("colony_count", len(ms)),
		zap.Duration("list_step_duration", listStepDuration),
	)

	log.Info(
		"查询oes集群列表：执行成功",
		zap.Duration("count_step_duration", countStepDuration),
		zap.Duration("list_step_duration", listStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return count, ms, nil
}

func (s *OesColonyService) OutportOesColonyData(
	ctx context.Context,
	m *oesmodel.OesColonyModel,
) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"解压oes程序包并初始化集群配置文件：开始执行",
		zap.Object("oes_colony", m),
	)

	colonyBinDir := GetOesColonyBinDir(m.ColonyNum)
	colonyConfDir := GetOesColonyConfigDir(m.ColonyNum)

	// 清理原oes集群配置文件
	cleanStepStart := time.Now()
	if _, err := os.Stat(colonyBinDir); !os.IsNotExist(err) {
		log.Debug(
			"解压oes程序包并初始化集群配置文件：开始清理原配置文件",
			zap.String("path", colonyBinDir),
		)
		if err := os.RemoveAll(colonyBinDir); err != nil {
			log.Error(
				"解压oes程序包并初始化集群配置文件：清理原oes集群配置文件失败",
				zap.Error(err),
				zap.String("path", colonyBinDir),
				zap.Duration("clean_step_duration", time.Since(cleanStepStart)),
				zap.Duration("total_duration", time.Since(startTime)),
			)
			return errors.FromError(err)
		}
		log.Debug(
			"解压oes程序包并初始化集群配置文件：清理原配置文件成功",
			zap.String("path", colonyBinDir),
			zap.Duration("clean_step_duration", time.Since(cleanStepStart)),
		)
	}

	// 创建临时目录
	tmpDir, mErr := os.MkdirTemp("/tmp", "oes-")
	if mErr != nil {
		log.Error(
			"解压oes程序包并初始化集群配置文件：创建oes程序包解压的tmp文件夹失败",
			zap.Error(mErr),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.FromError(mErr)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			log.Error(
				"解压oes程序包并初始化集群配置文件：删除oes程序包解压的tmp文件夹失败",
				zap.Error(err),
				zap.String("path", tmpDir),
			)
		}
	}()

	// 处理oes程序包
	oesStepStart := time.Now()
	oesPkgPath := resosvc.GetPackageStoragePath(m.Package.StorageFilename)
	log.Debug(
		"解压oes程序包并初始化集群配置文件：开始处理oes程序包",
		zap.String("pkg_path", oesPkgPath),
	)
	oesUnTarDirName, valiErr := archive.ValidateSingleDirTarGz(oesPkgPath)
	if valiErr != nil {
		log.Error(
			"解压oes程序包并初始化集群配置文件：oes程序包校验失败",
			zap.Error(valiErr),
			zap.String("path", oesPkgPath),
			zap.Duration("oes_step_duration", time.Since(oesStepStart)),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.ErrZIPFileIsNotValid.WithCause(valiErr)
	}

	if err := archive.UntarGz(oesPkgPath, tmpDir, archive.WithContext(ctx)); err != nil {
		log.Error(
			"解压oes程序包并初始化集群配置文件：解压oes程序包失败",
			zap.Error(err),
			zap.String("src_path", oesPkgPath),
			zap.String("dst_path", colonyBinDir),
			zap.Duration("oes_step_duration", time.Since(oesStepStart)),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.ErrUnZIPFailed.WithCause(err)
	}

	oesTmpDir := filepath.Join(tmpDir, oesUnTarDirName)
	if err := fileutil.CopyDir(ctx, oesTmpDir, colonyBinDir, true); err != nil {
		log.Error(
			"解压oes程序包并初始化集群配置文件：复制oes程序包解压目录失败",
			zap.Error(err),
			zap.String("src_path", oesTmpDir),
			zap.String("dst_path", colonyBinDir),
			zap.Duration("oes_step_duration", time.Since(oesStepStart)),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.FromError(err)
	}
	oesStepDuration := time.Since(oesStepStart)
	log.Debug(
		"解压oes程序包并初始化集群配置文件：处理oes程序包成功",
		zap.String("pkg_path", oesPkgPath),
		zap.Duration("oes_step_duration", oesStepDuration),
	)

	// 处理xcounter程序包
	xcterStepStart := time.Now()
	xcterPkgPath := resosvc.GetPackageStoragePath(m.XCounter.StorageFilename)
	log.Debug(
		"解压oes程序包并初始化集群配置文件：开始处理xcounter程序包",
		zap.String("pkg_path", xcterPkgPath),
	)
	xcterUnTarDirName, valiErr := archive.ValidateSingleDirTarGz(xcterPkgPath)
	if valiErr != nil {
		log.Error(
			"解压oes程序包并初始化集群配置文件：xcounter程序包校验失败",
			zap.Error(valiErr),
			zap.String("path", xcterPkgPath),
			zap.Duration("xcter_step_duration", time.Since(xcterStepStart)),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.ErrZIPFileIsNotValid.WithCause(valiErr)
	}
	if err := archive.UntarGz(xcterPkgPath, tmpDir); err != nil {
		log.Error(
			"解压oes程序包并初始化集群配置文件：解压xcounter程序包失败",
			zap.Error(err),
			zap.String("src_path", xcterPkgPath),
			zap.String("dst_path", tmpDir),
			zap.Duration("xcter_step_duration", time.Since(xcterStepStart)),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.ErrUnZIPFailed.WithCause(err)
	}
	xcterTmpDir := filepath.Join(tmpDir, xcterUnTarDirName, "bin")
	oesBinDir := filepath.Join(colonyBinDir, "bin")
	if err := fileutil.CopyDir(ctx, xcterTmpDir, oesBinDir, true); err != nil {
		log.Error(
			"解压oes程序包并初始化集群配置文件：复制xcounter程序包解压目录失败",
			zap.Error(err),
			zap.String("src_path", xcterTmpDir),
			zap.String("dst_path", oesBinDir),
			zap.Duration("xcter_step_duration", time.Since(xcterStepStart)),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.FromError(err)
	}
	xcterStepDuration := time.Since(xcterStepStart)
	log.Debug(
		"解压oes程序包并初始化集群配置文件：处理xcounter程序包成功",
		zap.String("pkg_path", xcterPkgPath),
		zap.Duration("xcter_step_duration", xcterStepDuration),
	)

	// 处理配置文件
	confStepStart := time.Now()
	colonyConfAll := filepath.Join(colonyConfDir, "all")
	log.Debug(
		"解压oes程序包并初始化集群配置文件：开始处理配置文件",
		zap.String("conf_dir", colonyConfAll),
	)
	if _, err := os.Stat(colonyConfAll); os.IsNotExist(err) {
		colonyBinConf := filepath.Join(colonyBinDir, "conf")
		if err := fileutil.CopyDir(ctx, colonyBinConf, colonyConfAll, true); err != nil {
			log.Error(
				"解压oes程序包并初始化集群配置文件：复制oes集群配置文件失败",
				zap.Error(err),
				zap.String("src_path", colonyBinConf),
				zap.String("dst_path", colonyConfAll),
				zap.Duration("conf_step_duration", time.Since(confStepStart)),
				zap.Duration("total_duration", time.Since(startTime)),
			)
			return errors.FromError(err)
		}
		srcPath := filepath.Join(config.ConfigDir, fmt.Sprintf("automatic_oes_%s.yaml", m.SystemType))
		dstPath := filepath.Join(colonyConfAll, "automatic.yaml")
		if err := fileutil.CopyFile(ctx, srcPath, dstPath); err != nil {
			log.Error(
				"解压oes程序包并初始化集群配置文件：复制oes的automatic配置文件失败",
				zap.Error(err),
				zap.String("src_path", srcPath),
				zap.String("dst_path", dstPath),
				zap.Duration("conf_step_duration", time.Since(confStepStart)),
				zap.Duration("total_duration", time.Since(startTime)),
			)
			return errors.FromError(err)
		}
	}

	// 导出oes集群配置变量文件
	oesVars := oesmodel.OesColonyVars{
		ID:        m.ID,
		ColonyNum: m.ColonyNum,
		PkgName:   m.ExtractedName,
		PackageID: m.PackageID,
		Version:   m.Package.Version,
		MonNodeID: m.MonNodeID,
		IsEnable:  m.IsEnable,
	}
	oesColonyConf := filepath.Join(colonyConfAll, "colony.yaml")
	if _, err := serializer.WriteYAML(oesColonyConf, oesVars); err != nil {
		log.Error(
			"解压oes程序包并初始化集群配置文件：导出oes集群配置变量文件失败",
			zap.Error(err),
			zap.String("path", oesColonyConf),
			zap.Object("oes_colony_vars", &oesVars),
			zap.Duration("conf_step_duration", time.Since(confStepStart)),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.ErrExportCacheFileFailed.WithCause(err)
	}
	confStepDuration := time.Since(confStepStart)
	log.Debug(
		"解压oes程序包并初始化集群配置文件：处理配置文件成功",
		zap.String("conf_dir", colonyConfAll),
		zap.Duration("conf_step_duration", confStepDuration),
	)

	log.Info(
		"解压oes程序包并初始化集群配置文件：执行成功",
		zap.String("path", oesColonyConf),
		zap.Object("oes_colony_vars", &oesVars),
		zap.Duration("oes_step_duration", oesStepDuration),
		zap.Duration("xcter_step_duration", xcterStepDuration),
		zap.Duration("conf_step_duration", confStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

func GetOesColonyBinDir(colonyNum string) string {
	return filepath.Join(config.StorageDir, "oes", "bin", colonyNum)
}

func GetOesColonyConfigDir(colonyNum string) string {
	return filepath.Join(config.StorageDir, "oes", "config", colonyNum)
}
