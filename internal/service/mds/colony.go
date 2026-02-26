package biz

import (
	"context"
	"os"
	"path/filepath"

	"go.uber.org/zap"

	mdsmodel "gin-artweb/internal/model/mds"
	mdsrepo "gin-artweb/internal/repository/mds"
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
}

func NewMdsColonyService(
	log *zap.Logger,
	colonyRepo *mdsrepo.MdsColonyRepo,
) *MdsColonyService {
	return &MdsColonyService{
		log:        log,
		colonyRepo: colonyRepo,
	}
}

func (s *MdsColonyService) CreateMdsColony(
	ctx context.Context,
	m mdsmodel.MdsColonyModel,
) (*mdsmodel.MdsColonyModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始创建mds集群",
		zap.Object(database.ModelKey, &m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	if err := s.colonyRepo.CreateModel(ctx, &m); err != nil {
		s.log.Error(
			"创建mds集群失败",
			zap.Error(err),
			zap.Object(database.ModelKey, &m),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, nil)
	}

	// 查询mds集群关联数据
	nm, rErr := s.FindMdsColonyByID(ctx, []string{"Package", "MonNode"}, m.ID)
	if rErr != nil {
		return nil, rErr
	}

	// 导出mds集群缓存数据
	if err := s.OutportMdsColonyData(ctx, nm); err != nil {
		return nil, err
	}
	s.log.Info(
		"创建mds集群成功",
		zap.Object(database.ModelKey, nm),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return nm, nil
}

func (s *MdsColonyService) UpdateMdsColonyByID(
	ctx context.Context,
	mdsColonyID uint32,
	data map[string]any,
) (*mdsmodel.MdsColonyModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始更新mds集群",
		zap.Uint32("mds_colony_id", mdsColonyID),
		zap.Any(database.UpdateDataKey, data),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	data["id"] = mdsColonyID
	if err := s.colonyRepo.UpdateModel(ctx, data, "id = ?", mdsColonyID); err != nil {
		s.log.Error(
			"更新mds集群失败",
			zap.Error(err),
			zap.Uint32("mds_colony_id", mdsColonyID),
			zap.Any(database.UpdateDataKey, data),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, data)
	}

	// 查询mds集群关联数据
	m, rErr := s.FindMdsColonyByID(ctx, []string{"Package", "MonNode"}, mdsColonyID)
	if rErr != nil {
		return nil, rErr
	}

	// 导出mds集群缓存数据
	if err := s.OutportMdsColonyData(ctx, m); err != nil {
		return nil, err
	}

	s.log.Info(
		"更新mds集群成功",
		zap.Uint32("mds_colony_id", mdsColonyID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return m, nil
}

func (s *MdsColonyService) DeleteMdsColonyByID(
	ctx context.Context,
	mdsColonyID uint32,
) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始删除mds集群",
		zap.Uint32("mds_colony_id", mdsColonyID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	if err := s.colonyRepo.DeleteModel(ctx, mdsColonyID); err != nil {
		s.log.Error(
			"删除mds集群失败",
			zap.Error(err),
			zap.Uint32("mds_colony_id", mdsColonyID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.NewGormError(err, map[string]any{"id": mdsColonyID})
	}

	s.log.Info(
		"删除mds集群成功",
		zap.Uint32("mds_colony_id", mdsColonyID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
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

	s.log.Info(
		"开始查询mds集群",
		zap.Strings(database.PreloadKey, preloads),
		zap.Uint32("mds_colony_id", mdsColonyID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, err := s.colonyRepo.GetModel(ctx, preloads, mdsColonyID)
	if err != nil {
		s.log.Error(
			"查询mds集群失败",
			zap.Error(err),
			zap.Uint32("mds_colony_id", mdsColonyID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, map[string]any{"id": mdsColonyID})
	}

	s.log.Info(
		"查询mds集群成功",
		zap.Uint32("mds_colony_id", mdsColonyID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return m, nil
}

func (s *MdsColonyService) ListMdsColony(
	ctx context.Context,
	qp database.QueryParams,
) (int64, *[]mdsmodel.MdsColonyModel, *errors.Error) {
	if ctx.Err() != nil {
		return 0, nil, errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始查询角色列表",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	count, ms, err := s.colonyRepo.ListModel(ctx, qp)
	if err != nil {
		s.log.Error(
			"查询mds集群列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}

	s.log.Info(
		"查询mds集群列表成功",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
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

	s.log.Info(
		"开始解压mds程序包并初始化集群配置文件",
		zap.Object(database.ModelKey, m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	colonyBinDir := common.GetMdsColonyBinDir(m.ColonyNum)
	colonyConfDir := common.GetMdsColonyConfigDir(m.ColonyNum)

	if _, err := os.Stat(colonyBinDir); !os.IsNotExist(err) {
		if err := os.RemoveAll(colonyBinDir); err != nil {
			s.log.Error(
				"清理原mds集群配置文件失败",
				zap.Error(err),
				zap.String("path", colonyBinDir),
			)
			return errors.ErrDeleteCacheFileFailed.WithCause(err)
		}
	}

	tmpDir, mErr := os.MkdirTemp("/tmp", "mds-")
	if mErr != nil {
		s.log.Error(
			"创建mds程序包解压的tmp文件夹失败",
			zap.Error(mErr),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.FromError(mErr)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			s.log.Error(
				"删除mds程序包解压的tmp文件夹失败",
				zap.Error(err),
				zap.String("path", tmpDir),
			)
		}
	}()

	mdsPkgPath := common.GetPackageStoragePath(m.Package.StorageFilename)
	mdsUnTarDirName, valiErr := archive.ValidateSingleDirTarGz(mdsPkgPath)
	if valiErr != nil {
		s.log.Error(
			"mds程序包校验失败",
			zap.Error(valiErr),
			zap.String("path", mdsPkgPath),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.ErrValidationFailed.WithCause(valiErr)
	}
	if err := archive.UntarGz(mdsPkgPath, tmpDir, archive.WithContext(ctx)); err != nil {
		s.log.Error(
			"解压mds程序包失败",
			zap.Error(err),
			zap.Uint32("mds_colony_id", m.ID),
			zap.String("pkg_name", m.ExtractedName),
			zap.String("path", m.Package.StorageFilename),
			zap.String("dest", colonyBinDir),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.ErrUnZIPFailed.WithCause(err).WithField("pkg_name", m.ExtractedName)
	}

	mdsTmpDir := filepath.Join(tmpDir, mdsUnTarDirName)
	if err := fileutil.CopyDir(ctx, mdsTmpDir, colonyBinDir, true); err != nil {
		s.log.Error(
			"复制mds程序包解压目录失败",
			zap.Error(err),
			zap.String("src_path", mdsTmpDir),
			zap.String("dst_path", colonyBinDir),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.FromError(err)
	}

	colonyConfAll := filepath.Join(colonyConfDir, "all")
	if _, err := os.Stat(colonyConfAll); os.IsNotExist(err) {
		colonyBinConf := filepath.Join(colonyBinDir, "conf")
		if err := fileutil.CopyDir(ctx, colonyBinConf, colonyConfAll, true); err != nil {
			s.log.Error(
				"复制mds集群配置文件失败",
				zap.Error(err),
				zap.String("src_path", colonyBinConf),
				zap.String("dst_path", colonyConfAll),
				zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			)
			return errors.FromError(err)
		}
		srcPath := filepath.Join(config.ConfigDir, "automatic_mds.yaml")
		dstPath := filepath.Join(colonyConfAll, "automatic.yaml")
		if err := fileutil.CopyFile(ctx, srcPath, dstPath); err != nil {
			s.log.Error(
				"复制mds的automatic配置文件失败",
				zap.Error(err),
				zap.String("src_path", colonyConfDir),
				zap.String("dst_path", dstPath),
				zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
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
		s.log.Error(
			"导出mds集群配置变量文件失败",
			zap.Error(err),
			zap.String("path", mdsColonyConf),
			zap.Object("mds_colony_vars", &mdsVars),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.ErrExportCacheFileFailed.WithCause(err)
	}

	s.log.Info(
		"解压mds程序包并初始化集群配置文件成功",
		zap.String("path", mdsColonyConf),
		zap.Object("mds_colony_vars", &mdsVars),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return nil
}

// func (s *MdsColonyUsecase) InitCrontab(ctx context.Context, mdsColonyID uint32) *errors.Error {
// 	if ctx.Err() != nil {
// 		return errors.FromError(ctx.Err())
// 	}

// 	s.log.Info(
// 		"初始化mds集群crontab",
// 		zap.Uint32("mds_colony_id", mdsColonyID),
// 		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
// 	)
// 	m, rErr := s.FindMdsColonyByID(ctx, nil, mdsColonyID)
// 	if rErr != nil {
// 		return rErr
// 	}
// 	return nil
// }
