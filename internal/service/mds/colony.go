package mds

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"

	jobmodel "gin-artweb/internal/model/job"
	mdsmodel "gin-artweb/internal/model/mds"
	mdsrepo "gin-artweb/internal/repo/mds"
	resocvs "gin-artweb/internal/service/resource"
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
	startTime := time.Now()
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"创建mds集群:开始执行",
		zap.Object("mds_colony_dto", dto),
	)

	count, err := s.colonyRepo.CountModel(ctx, map[string]any{"colony_num": dto.ColonyNum})
	if err != nil {
		log.Error(
			"创建mds集群:查询mds集群是否存在失败",
			zap.Error(err),
			zap.String("colony_num", dto.ColonyNum),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, nil)
	}
	if count > 0 {
		log.Error(
			"创建mds集群:mds集群已存在",
			zap.String("colony_num", dto.ColonyNum),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.ErrDuplicatedKey.WithField("colony_num", dto.ColonyNum)
	}

	m := dto.ToModel()
	if err := s.colonyRepo.CreateModel(ctx, &m); err != nil {
		log.Error(
			"创建mds集群:创建数据库模型失败",
			zap.Error(err),
			zap.Object("mds_colony_model", &m),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, nil)
	}

	colony, rErr := s.FindMdsColonyByID(ctx, []string{"Package", "MonNode"}, m.ID)
	if rErr != nil {
		log.Error(
			"创建mds集群:查询创建后的mds集群关联数据失败",
			zap.Error(rErr),
			zap.Uint32("mds_colony_id", colony.ID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, rErr
	}

	if rErr := s.cronSvc.CreateCornByColony(ctx, *colony); rErr != nil {
		s.log.Error(
			"创建mds集群:初始化mds集群定时任务失败",
			zap.Error(rErr),
			zap.Uint32("mds_colony_id", colony.ID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, rErr
	}

	// 导出mds集群缓存数据
	if err := s.OutportMdsColonyData(ctx, colony); err != nil {
		log.Error(
			"创建mds集群:导出mds集群缓存数据失败",
			zap.Error(err),
			zap.Uint32("mds_colony_id", colony.ID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, err
	}

	log.Info(
		"创建mds集群:执行成功",
		zap.Uint32("mds_colony_id", colony.ID),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	return colony, nil
}

func (s *MdsColonyService) UpdateMdsColonyByID(
	ctx context.Context,
	mdsColonyID uint32,
	dto mdsmodel.MdsColonyUpsertDTO,
) (*mdsmodel.MdsColonyModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"更新mds集群:开始执行",
		zap.Uint32("mds_colony_id", mdsColonyID),
		zap.Object("mds_colony_dto", &dto),
	)

	om, rErr := s.FindMdsColonyByID(ctx, nil, mdsColonyID)
	if rErr != nil {
		log.Error(
			"更新mds集群:查询更新前mds集群关联数据失败",
			zap.Error(rErr),
			zap.Uint32("mds_colony_id", mdsColonyID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, rErr
	}

	updateData := dto.ToUpdateMap()
	if err := s.colonyRepo.UpdateModel(ctx, updateData, "id = ?", mdsColonyID); err != nil {
		log.Error(
			"更新mds集群:更新数据库模型失败",
			zap.Error(err),
			zap.Uint32("mds_colony_id", mdsColonyID),
			zap.Any("update_data", updateData),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, updateData)
	}

	// 查询mds集群关联数据
	nm, rErr := s.FindMdsColonyByID(ctx, []string{"Package", "MonNode"}, mdsColonyID)
	if rErr != nil {
		log.Error(
			"更新mds集群:查询mds集群关联数据失败",
			zap.Error(rErr),
			zap.Uint32("mds_colony_id", mdsColonyID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, rErr
	}

	if om.ColonyNum != nm.ColonyNum || om.IsEnable != nm.IsEnable {
		if err := s.cronSvc.DeleteCornByColonyID(ctx, mdsColonyID); err != nil {
			log.Error(
				"更新mds集群:清理计划任务失败",
				zap.Error(err),
				zap.Uint32("mds_colony_id", mdsColonyID),
				zap.Duration("total_duration", time.Since(startTime)),
			)
			return nil, err
		}

		if err := s.cronSvc.CreateCornByColony(ctx, *nm); err != nil {
			log.Error(
				"更新mds集群:初始化计划任务失败",
				zap.Error(err),
				zap.Uint32("mds_colony_id", mdsColonyID),
				zap.Duration("total_duration", time.Since(startTime)),
			)
			return nil, err
		}
	}

	// 导出mds集群缓存数据
	if err := s.OutportMdsColonyData(ctx, nm); err != nil {
		log.Error(
			"更新mds集群:导出mds集群缓存数据失败",
			zap.Error(err),
			zap.Uint32("mds_colony_id", mdsColonyID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, err
	}

	log.Info(
		"更新mds集群:执行成功",
		zap.Uint32("mds_colony_id", mdsColonyID),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nm, nil
}

func (s *MdsColonyService) DeleteMdsColonyByID(
	ctx context.Context,
	mdsColonyID uint32,
) *errors.Error {
	startTime := time.Now()
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"删除mds集群:开始执行",
		zap.Uint32("mds_colony_id", mdsColonyID),
	)

	colony, rErr := s.FindMdsColonyByID(ctx, nil, mdsColonyID)
	if rErr != nil {
		log.Error(
			"删除mds集群:删除前查询mds集群数据失败",
			zap.Error(rErr),
			zap.Uint32("mds_colony_id", colony.ID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return rErr
	}

	if err := s.cronSvc.DeleteCornByColonyID(ctx, mdsColonyID); err != nil {
		log.Error(
			"删除mds集群:清理计划任务失败",
			zap.Error(err),
			zap.Uint32("mds_colony_id", mdsColonyID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return err
	}

	if err := s.colonyRepo.DeleteModel(ctx, "id = ?", mdsColonyID); err != nil {
		log.Error(
			"删除mds集群:删除数据库模型失败",
			zap.Error(err),
			zap.Uint32("mds_colony_id", mdsColonyID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.NewGormError(err, map[string]any{"id": mdsColonyID})
	}

	confPath := GetMdsColonyConfigDir(colony.ColonyNum)
	if _, statErr := os.Stat(confPath); statErr == nil {
		savePath := GetMdsColonyBackupDir(colony.ColonyNum)
		if err := fileutil.RemoveAll(ctx, savePath); err != nil {
			log.Error(
				"删除mds集群:清理原备份文件失败",
				zap.String("clear_path", savePath),
				zap.Duration("total_duration", time.Since(startTime)),
			)
			return errors.ErrUnknown.WithCause(err).WithField("clear_path", savePath)
		}

		if err := s.colonyRepo.MoveConfigFile(ctx, confPath, savePath); err != nil {
			log.Error(
				"删除mds集群:迁移mds原配置文件失败",
				zap.Error(err),
				zap.String("coluny_num", colony.ColonyNum),
				zap.String("src_path", confPath),
				zap.String("dst_path", savePath),
				zap.Duration("total_duration", time.Since(startTime)),
			)
			return errors.ErrUnknown.WithCause(err).WithFields(map[string]any{
				"coluny_num": colony.ColonyNum,
				"src_path":   confPath,
				"dst_path":   savePath,
			})
		}
	}

	log.Info(
		"删除mds集群:执行成功",
		zap.Uint32("mds_colony_id", mdsColonyID),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

func (s *MdsColonyService) FindMdsColonyByID(
	ctx context.Context,
	preloads []string,
	mdsColonyID uint32,
) (*mdsmodel.MdsColonyModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	m, err := s.colonyRepo.GetModel(ctx, preloads, mdsColonyID)
	if err != nil {
		log.Error(
			"查询mds集群:查询数据库模型失败",
			zap.Error(err),
			zap.Uint32("mds_colony_id", mdsColonyID),
			zap.Strings("preloads", preloads),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, map[string]any{"id": mdsColonyID})
	}

	log.Debug(
		"查询mds集群:查询到的数据库模型详情",
		zap.Object("mds_colony_model", m),
	)
	return m, nil
}

func (s *MdsColonyService) ListMdsColony(
	ctx context.Context,
	page, size int,
	dto mdsmodel.ListMdsColonyDTO,
) (int64, []mdsmodel.MdsColonyModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return 0, nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Debug(
		"查询mds集群列表:入参详情",
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

	count, err := s.colonyRepo.CountModel(ctx, qp.Query)
	if err != nil {
		log.Error(
			"查询mds集群列表:查询数据库模型总数失败",
			zap.Error(err),
			zap.Any("query", qp.Query),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}

	if count == 0 {
		log.Warn(
			"查询mds集群列表:数据库模型总数为0",
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, nil, nil
	}

	ms, err := s.colonyRepo.ListModel(ctx, qp)
	if err != nil {
		log.Error(
			"查询mds集群列表:查询数据库模型失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}
	return count, ms, nil
}

func (s *MdsColonyService) ListMdsSchedules(
	ctx context.Context,
	mdsColonyID uint32,
) ([]jobmodel.ScheduleModel, *errors.Error) {
	return s.cronSvc.ListCornByColonyID(ctx, mdsColonyID)
}

func (s *MdsColonyService) OutportMdsColonyData(
	ctx context.Context,
	m *mdsmodel.MdsColonyModel,
) *errors.Error {
	startTime := time.Now()
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"导出mds集群数据:开始执行",
		zap.Object("mds_colony_model", m),
	)

	colonyBinDir := GetMdsColonyBinDir(m.ColonyNum)
	colonyConfDir := GetMdsColonyConfigDir(m.ColonyNum)

	if _, err := os.Stat(colonyBinDir); !os.IsNotExist(err) {
		if err := os.RemoveAll(colonyBinDir); err != nil {
			log.Error(
				"导出mds集群数据:清理原mds集群配置文件失败",
				zap.Error(err),
				zap.String("path", colonyBinDir),
				zap.Duration("total_duration", time.Since(startTime)),
			)
			return errors.ErrDeleteCacheFileFailed.WithCause(err)
		}
	}

	tmpDir, mErr := os.MkdirTemp("/tmp", "mds-")
	if mErr != nil {
		log.Error(
			"导出mds集群数据:创建临时文件夹失败",
			zap.Error(mErr),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.FromError(mErr)
	}

	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			log.Error(
				"导出mds集群数据:删除临时文件夹失败",
				zap.Error(err),
				zap.String("path", tmpDir),
				zap.Duration("total_duration", time.Since(startTime)),
			)
		}
	}()

	mdsPkgPath := resocvs.GetPackageStoragePath(m.Package.StorageFilename)
	mdsUnTarDirName, valiErr := archive.ValidateSingleDirTarGz(mdsPkgPath)
	if valiErr != nil {
		log.Error(
			"导出mds集群数据:mds程序包校验失败",
			zap.Error(valiErr),
			zap.String("path", mdsPkgPath),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.ErrValidationFailed.WithCause(valiErr)
	}

	if err := archive.UntarGz(mdsPkgPath, tmpDir, archive.WithContext(ctx)); err != nil {
		log.Error(
			"导出mds集群数据:解压mds程序包失败",
			zap.Error(err),
			zap.Uint32("mds_colony_id", m.ID),
			zap.String("pkg_name", m.ExtractedName),
			zap.String("src_path", mdsPkgPath),
			zap.String("dest_path", tmpDir),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.ErrUnZIPFailed.WithCause(err).WithField("pkg_name", m.ExtractedName)
	}

	mdsTmpDir := filepath.Join(tmpDir, mdsUnTarDirName)
	if err := fileutil.CopyDir(ctx, mdsTmpDir, colonyBinDir, true); err != nil {
		log.Error(
			"导出mds集群数据:复制mds程序包解压目录失败",
			zap.Error(err),
			zap.String("src_path", mdsTmpDir),
			zap.String("dst_path", colonyBinDir),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.FromError(err)
	}

	colonyConfAll := filepath.Join(colonyConfDir, "all")
	if _, err := os.Stat(colonyConfAll); os.IsNotExist(err) {
		colonyBinConf := filepath.Join(colonyBinDir, "conf")
		if err := fileutil.CopyDir(ctx, colonyBinConf, colonyConfAll, true); err != nil {
			log.Error(
				"导出mds集群数据:复制mds集群配置文件失败",
				zap.Error(err),
				zap.String("src_path", colonyBinConf),
				zap.String("dst_path", colonyConfAll),
				zap.Duration("total_duration", time.Since(startTime)),
			)
			return errors.FromError(err)
		}
		srcPath := filepath.Join(config.ConfigDir, "automatic_mds.yaml")
		dstPath := filepath.Join(colonyConfAll, "automatic.yaml")
		if err := fileutil.CopyFile(ctx, srcPath, dstPath); err != nil {
			log.Error(
				"导出mds集群数据:复制mds的automatic配置文件失败",
				zap.Error(err),
				zap.String("src_path", srcPath),
				zap.String("dst_path", dstPath),
				zap.Duration("total_duration", time.Since(startTime)),
			)
			return errors.FromError(err)
		}
	}

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
	if _, err := serializer.WriteYAML(mdsColonyConf, mdsVars); err != nil {
		log.Error(
			"导出mds集群数据:导出mds集群配置变量文件失败",
			zap.Error(err),
			zap.String("path", mdsColonyConf),
			zap.Object("mds_colony_vars", &mdsVars),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.ErrExportCacheFileFailed.WithCause(err)
	}

	log.Info(
		"导出mds集群数据:执行成功",
		zap.Uint32("mds_colony_id", m.ID),
		zap.String("colony_num", m.ColonyNum),
		zap.String("config_path", mdsColonyConf),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

func GetMdsColonyBinDir(colonyNum string) string {
	return filepath.Join(config.StorageDir, "mds", "bin", colonyNum)
}

func GetMdsColonyConfigDir(colonyNum string) string {
	return filepath.Join(config.StorageDir, "mds", "config", colonyNum)
}

func GetMdsColonyBackupDir(colonyNum string) string {
	return filepath.Join(config.StorageDir, "mds", "config", "backup", colonyNum)
}
