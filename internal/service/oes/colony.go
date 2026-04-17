package oes

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"

	jobmodel "gin-artweb/internal/model/job"
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
	startTime := time.Now()
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Debug(
		"创建oes集群:入参详情",
		zap.Object("oes_colony_dto", &dto),
	)

	m := dto.ToModel()
	if err := s.colonyRepo.CreateModel(ctx, &m); err != nil {
		log.Error(
			"创建oes集群:创建数据库模型失败",
			zap.Error(err),
			zap.Object("oes_colony", &m),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, nil)
	}

	if err := s.cronSvc.CreateCornByColony(ctx, &m); err != nil {
		log.Error(
			"创建oes集群:初始化oes集群定时任务失败",
			zap.Error(err),
			zap.Uint32("oes_colony_id", m.ID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, err
	}

	if err := s.OutportOesColonyData(ctx, &m); err != nil {
		log.Error(
			"创建oes集群:导出缓存数据失败",
			zap.Error(err),
			zap.Uint32("oes_colony_id", m.ID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, err
	}

	log.Info(
		"创建oes集群:执行成功",
		zap.Uint32("oes_colony_id", m.ID),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return &m, nil
}

func (s *OesColonyService) UpdateOesColonyByID(
	ctx context.Context,
	oesColonyID uint32,
	dto oesmodel.OesColonyUpsertDTO,
) (*oesmodel.OesColonyModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"更新oes集群:开始执行",
		zap.Uint32("oes_colony_id", oesColonyID),
		zap.Object("oes_colony_dto", &dto),
	)

	om, rErr := s.FindOesColonyByID(ctx, nil, oesColonyID)
	if rErr != nil {
		log.Error(
			"更新oes集群:查询更新前oes集群数据失败",
			zap.Error(rErr),
			zap.Uint32("oes_colony_id", oesColonyID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, rErr
	}

	updateData := dto.ToUpdateMap()
	if err := s.colonyRepo.UpdateModel(ctx, updateData, "id = ?", oesColonyID); err != nil {
		log.Error(
			"更新oes集群:更新数据库模型失败",
			zap.Error(err),
			zap.Uint32("oes_colony_id", oesColonyID),
			zap.Any("update_data", updateData),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, updateData)
	}

	nm, rErr := s.FindOesColonyByID(ctx, []string{"Package", "XCounter", "MonNode"}, oesColonyID)
	if rErr != nil {
		log.Error(
			"更新oes集群:查询关联数据失败",
			zap.Error(rErr),
			zap.Uint32("oes_colony_id", oesColonyID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, rErr
	}

	if om.ColonyNum != nm.ColonyNum || om.IsEnable != nm.IsEnable {
		if err := s.cronSvc.DeleteCornByColonyID(ctx, oesColonyID); err != nil {
			log.Error(
				"更新oes集群:清理计划任务失败",
				zap.Error(err),
				zap.Uint32("oes_colony_id", oesColonyID),
				zap.Duration("total_duration", time.Since(startTime)),
			)
			return nil, err
		}

		if err := s.cronSvc.CreateCornByColony(ctx, nm); err != nil {
			log.Error(
				"更新oes集群:初始化计划任务失败",
				zap.Error(err),
				zap.Uint32("oes_colony_id", oesColonyID),
				zap.Duration("total_duration", time.Since(startTime)),
			)
			return nil, err
		}
	}

	if err := s.OutportOesColonyData(ctx, nm); err != nil {
		log.Error(
			"更新oes集群:导出缓存数据失败",
			zap.Error(err),
			zap.Uint32("oes_colony_id", oesColonyID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, err
	}

	log.Info(
		"更新oes集群:执行成功",
		zap.Uint32("oes_colony_id", oesColonyID),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nm, nil
}

func (s *OesColonyService) DeleteOesColonyByID(
	ctx context.Context,
	oesColonyID uint32,
) *errors.Error {
	startTime := time.Now()
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"删除oes集群:开始执行",
		zap.Uint32("oes_colony_id", oesColonyID),
	)

	if err := s.colonyRepo.DeleteModel(ctx, oesColonyID); err != nil {
		log.Error(
			"删除oes集群:删除数据库模型失败",
			zap.Error(err),
			zap.Uint32("oes_colony_id", oesColonyID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.NewGormError(err, map[string]any{"id": oesColonyID})
	}

	if err := s.cronSvc.DeleteCornByColonyID(ctx, oesColonyID); err != nil {
		log.Error(
			"删除oes集群:清理计划任务失败",
			zap.Error(err),
			zap.Uint32("oes_colony_id", oesColonyID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return err
	}

	log.Info(
		"删除oes集群:执行成功",
		zap.Uint32("oes_colony_id", oesColonyID),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

func (s *OesColonyService) FindOesColonyByID(
	ctx context.Context,
	preloads []string,
	oesColonyID uint32,
) (*oesmodel.OesColonyModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	m, err := s.colonyRepo.GetModel(ctx, preloads, oesColonyID)
	if err != nil {
		log.Error(
			"查询oes集群:查询数据库模型失败",
			zap.Error(err),
			zap.Uint32("oes_colony_id", oesColonyID),
			zap.Strings("preloads", preloads),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, map[string]any{"id": oesColonyID})
	}

	log.Debug(
		"查询oes集群:查询到的数据库模型详情",
		zap.Object("oes_colony_model", m),
	)
	return m, nil
}

func (s *OesColonyService) ListOesColony(
	ctx context.Context,
	page, size int,
	dto oesmodel.ListOesColonyDTO,
) (int64, []oesmodel.OesColonyModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return 0, nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Debug(
		"查询mds集群列表:入参详情",
		zap.Int("page", page),
		zap.Int("size", size),
		zap.Object("oes_colony_dto", &dto),
	)

	limit, offset := common.Page2LimitOffset(page, size)
	qp := database.QueryParams{
		Preloads: []string{"Package", "XCounter", "MonNode"},
		OrderBy:  []string{"id DESC"},
		Limit:    limit,
		Offset:   offset,
		Query:    dto.ToQueryMap(),
	}

	count, err := s.colonyRepo.CountModel(ctx, qp.Query)
	if err != nil {
		log.Error(
			"查询oes集群列表:查询数据库模型总数失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}

	if count == 0 {
		log.Warn(
			"查询oes集群列表:数据库模型总数为0",
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, nil, nil
	}

	ms, err := s.colonyRepo.ListModel(ctx, qp)
	if err != nil {
		log.Error(
			"查询oes集群列表:查询数据库模型列表失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}
	return count, ms, nil
}

func (s *OesColonyService) ListOesSchedules(
	ctx context.Context,
	oesColonyID uint32,
) ([]jobmodel.ScheduleModel, *errors.Error) {
	return s.cronSvc.ListCornByColonyID(ctx, oesColonyID)
}

func (s *OesColonyService) OutportOesColonyData(
	ctx context.Context,
	m *oesmodel.OesColonyModel,
) *errors.Error {
	startTime := time.Now()
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"解压oes程序包并初始化集群配置文件:开始执行",
		zap.Object("oes_colony", m),
	)

	colonyBinDir := GetOesColonyBinDir(m.ColonyNum)
	colonyConfDir := GetOesColonyConfigDir(m.ColonyNum)

	// 清理原oes集群配置文件
	if _, err := os.Stat(colonyBinDir); !os.IsNotExist(err) {
		if err := os.RemoveAll(colonyBinDir); err != nil {
			log.Error(
				"解压oes程序包并初始化集群配置文件:清理原oes集群配置文件失败",
				zap.Error(err),
				zap.String("path", colonyBinDir),
				zap.Duration("total_duration", time.Since(startTime)),
			)
			return errors.FromError(err)
		}
	}

	// 创建临时目录
	tmpDir, mErr := os.MkdirTemp("/tmp", "oes-")
	if mErr != nil {
		log.Error(
			"解压oes程序包并初始化集群配置文件:创建oes程序包解压的tmp文件夹失败",
			zap.Error(mErr),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.FromError(mErr)
	}

	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			log.Error(
				"解压oes程序包并初始化集群配置文件:删除oes程序包解压的tmp文件夹失败",
				zap.Error(err),
				zap.String("path", tmpDir),
			)
		}
	}()

	// 处理oes程序包
	oesPkgPath := resosvc.GetPackageStoragePath(m.Package.StorageFilename)
	oesUnTarDirName, valiErr := archive.ValidateSingleDirTarGz(oesPkgPath)
	if valiErr != nil {
		log.Error(
			"解压oes程序包并初始化集群配置文件:oes程序包校验失败",
			zap.Error(valiErr),
			zap.String("path", oesPkgPath),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.ErrZIPFileIsNotValid.WithCause(valiErr)
	}

	if err := archive.UntarGz(oesPkgPath, tmpDir, archive.WithContext(ctx)); err != nil {
		log.Error(
			"解压oes程序包并初始化集群配置文件:解压oes程序包失败",
			zap.Error(err),
			zap.String("src_path", oesPkgPath),
			zap.String("dst_path", colonyBinDir),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.ErrUnZIPFailed.WithCause(err)
	}

	oesTmpDir := filepath.Join(tmpDir, oesUnTarDirName)
	if err := fileutil.CopyDir(ctx, oesTmpDir, colonyBinDir, true); err != nil {
		log.Error(
			"解压oes程序包并初始化集群配置文件:复制oes程序包解压目录失败",
			zap.Error(err),
			zap.String("src_path", oesTmpDir),
			zap.String("dst_path", colonyBinDir),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.FromError(err)
	}

	// 处理xcounter程序包
	xcterPkgPath := resosvc.GetPackageStoragePath(m.XCounter.StorageFilename)
	xcterUnTarDirName, valiErr := archive.ValidateSingleDirTarGz(xcterPkgPath)
	if valiErr != nil {
		log.Error(
			"解压oes程序包并初始化集群配置文件:xcounter程序包校验失败",
			zap.Error(valiErr),
			zap.String("path", xcterPkgPath),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.ErrZIPFileIsNotValid.WithCause(valiErr)
	}
	if err := archive.UntarGz(xcterPkgPath, tmpDir); err != nil {
		log.Error(
			"解压oes程序包并初始化集群配置文件:解压xcounter程序包失败",
			zap.Error(err),
			zap.String("src_path", xcterPkgPath),
			zap.String("dst_path", tmpDir),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.ErrUnZIPFailed.WithCause(err)
	}
	xcterTmpDir := filepath.Join(tmpDir, xcterUnTarDirName, "bin")
	oesBinDir := filepath.Join(colonyBinDir, "bin")
	if err := fileutil.CopyDir(ctx, xcterTmpDir, oesBinDir, true); err != nil {
		log.Error(
			"解压oes程序包并初始化集群配置文件:复制xcounter程序包解压目录失败",
			zap.Error(err),
			zap.String("src_path", xcterTmpDir),
			zap.String("dst_path", oesBinDir),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.FromError(err)
	}

	// 处理配置文件
	colonyConfAll := filepath.Join(colonyConfDir, "all")
	if _, err := os.Stat(colonyConfAll); os.IsNotExist(err) {
		colonyBinConf := filepath.Join(colonyBinDir, "conf")
		if err := fileutil.CopyDir(ctx, colonyBinConf, colonyConfAll, true); err != nil {
			log.Error(
				"解压oes程序包并初始化集群配置文件:复制oes集群配置文件失败",
				zap.Error(err),
				zap.String("src_path", colonyBinConf),
				zap.String("dst_path", colonyConfAll),
				zap.Duration("total_duration", time.Since(startTime)),
			)
			return errors.FromError(err)
		}
		srcPath := filepath.Join(config.ConfigDir, fmt.Sprintf("automatic_oes_%s.yaml", m.SystemType))
		dstPath := filepath.Join(colonyConfAll, "automatic.yaml")
		if err := fileutil.CopyFile(ctx, srcPath, dstPath); err != nil {
			log.Error(
				"解压oes程序包并初始化集群配置文件:复制oes的automatic配置文件失败",
				zap.Error(err),
				zap.String("src_path", srcPath),
				zap.String("dst_path", dstPath),
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
			"解压oes程序包并初始化集群配置文件:导出oes集群配置变量文件失败",
			zap.Error(err),
			zap.String("path", oesColonyConf),
			zap.Object("oes_colony_vars", &oesVars),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.ErrExportCacheFileFailed.WithCause(err)
	}

	log.Info(
		"解压oes程序包并初始化集群配置文件:执行成功",
		zap.String("path", oesColonyConf),
		zap.Object("oes_colony_vars", &oesVars),
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
