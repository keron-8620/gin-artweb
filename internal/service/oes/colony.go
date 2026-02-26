package biz

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"

	oesmodel "gin-artweb/internal/model/oes"
	oesrepo "gin-artweb/internal/repository/oes"
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
}

func NewOesColonyService(
	log *zap.Logger,
	colonyRepo *oesrepo.OesColonyRepo,
) *OesColonyService {
	return &OesColonyService{
		log:        log,
		colonyRepo: colonyRepo,
	}
}

func (s *OesColonyService) CreateOesColony(
	ctx context.Context,
	m oesmodel.OesColonyModel,
) (*oesmodel.OesColonyModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始创建oes集群",
		zap.Object(database.ModelKey, &m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	if err := s.colonyRepo.CreateModel(ctx, &m); err != nil {
		s.log.Error(
			"创建oes集群失败",
			zap.Error(err),
			zap.Object(database.ModelKey, &m),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, nil)
	}

	// 查询oes集群关联数据
	nm, rErr := s.FindOesColonyByID(ctx, []string{"Package", "XCounter", "MonNode"}, m.ID)
	if rErr != nil {
		return nil, rErr
	}

	// 导出oes集群缓存数据
	if err := s.OutportOesColonyData(ctx, nm); err != nil {
		return nil, err
	}

	s.log.Info(
		"创建oes集群成功",
		zap.Object(database.ModelKey, nm),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return nm, nil
}

func (s *OesColonyService) UpdateOesColonyByID(
	ctx context.Context,
	oesColonyID uint32,
	data map[string]any,
) (*oesmodel.OesColonyModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始更新oes集群",
		zap.Uint32("oes_colony_id", oesColonyID),
		zap.Any(database.UpdateDataKey, data),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	data["id"] = oesColonyID
	if err := s.colonyRepo.UpdateModel(ctx, data, "id = ?", oesColonyID); err != nil {
		s.log.Error(
			"更新oes集群失败",
			zap.Error(err),
			zap.Uint32("oes_colony_id", oesColonyID),
			zap.Any(database.UpdateDataKey, data),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, data)
	}

	// 查询关联数据
	m, rErr := s.FindOesColonyByID(ctx, []string{"Package", "XCounter", "MonNode"}, oesColonyID)
	if rErr != nil {
		return nil, rErr
	}

	// 导出数据库缓存数据
	if err := s.OutportOesColonyData(ctx, m); err != nil {
		return nil, err
	}

	s.log.Info(
		"更新oes集群成功",
		zap.Uint32("oes_colony_id", oesColonyID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return m, nil
}

func (s *OesColonyService) DeleteOesColonyByID(
	ctx context.Context,
	oesColonyID uint32,
) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始删除oes集群",
		zap.Uint32("oes_colony_id", oesColonyID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	if err := s.colonyRepo.DeleteModel(ctx, oesColonyID); err != nil {
		s.log.Error(
			"删除oes集群失败",
			zap.Error(err),
			zap.Uint32("oes_colony_id", oesColonyID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.NewGormError(err, map[string]any{"id": oesColonyID})
	}

	s.log.Info(
		"删除oes集群成功",
		zap.Uint32("oes_colony_id", oesColonyID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
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

	s.log.Info(
		"开始查询oes集群",
		zap.Strings(database.PreloadKey, preloads),
		zap.Uint32("oes_colony_id", oesColonyID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, err := s.colonyRepo.GetModel(ctx, preloads, oesColonyID)
	if err != nil {
		s.log.Error(
			"查询oes集群失败",
			zap.Error(err),
			zap.Uint32("oes_colony_id", oesColonyID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, map[string]any{"id": oesColonyID})
	}

	s.log.Info(
		"查询oes集群成功",
		zap.Uint32("oes_colony_id", oesColonyID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return m, nil
}

func (s *OesColonyService) ListOesColony(
	ctx context.Context,
	qp database.QueryParams,
) (int64, *[]oesmodel.OesColonyModel, *errors.Error) {
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
			"查询oes集群列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}

	s.log.Info(
		"查询oes集群列表成功",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
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

	s.log.Info(
		"开始解压oes程序包并初始化集群配置文件",
		zap.Object(database.ModelKey, m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	colonyBinDir := common.GetOesColonyBinDir(m.ColonyNum)
	colonyConfDir := common.GetOesColonyConfigDir(m.ColonyNum)

	if _, err := os.Stat(colonyBinDir); !os.IsNotExist(err) {
		if err := os.RemoveAll(colonyBinDir); err != nil {
			s.log.Error(
				"清理原oes集群配置文件失败",
				zap.Error(err),
				zap.String("path", colonyBinDir),
			)
			return errors.FromError(err)
		}
	}

	tmpDir, mErr := os.MkdirTemp("/tmp", "oes-")
	if mErr != nil {
		s.log.Error(
			"创建oes程序包解压的tmp文件夹失败",
			zap.Error(mErr),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.FromError(mErr)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			s.log.Error(
				"删除oes程序包解压的tmp文件夹失败",
				zap.Error(err),
				zap.String("path", tmpDir),
			)
		}
	}()

	oesPkgPath := common.GetPackageStoragePath(m.Package.StorageFilename)
	oesUnTarDirName, valiErr := archive.ValidateSingleDirTarGz(oesPkgPath)
	if valiErr != nil {
		s.log.Error(
			"oes程序包校验失败",
			zap.Error(valiErr),
			zap.String("path", oesPkgPath),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.ErrZIPFileIsNotValid.WithCause(valiErr)
	}

	if err := archive.UntarGz(oesPkgPath, tmpDir, archive.WithContext(ctx)); err != nil {
		s.log.Error(
			"解压oes程序包失败",
			zap.Error(err),
			zap.String("src_path", oesPkgPath),
			zap.String("dst_path", colonyBinDir),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.ErrUnZIPFailed.WithCause(err)
	}

	oesTmpDir := filepath.Join(tmpDir, oesUnTarDirName)
	if err := fileutil.CopyDir(ctx, oesTmpDir, colonyBinDir, true); err != nil {
		s.log.Error(
			"复制oes程序包解压目录失败",
			zap.Error(err),
			zap.String("src_path", oesTmpDir),
			zap.String("dst_path", colonyBinDir),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.FromError(err)
	}

	xcterPkgPath := common.GetPackageStoragePath(m.XCounter.StorageFilename)
	xcterUnTarDirName, valiErr := archive.ValidateSingleDirTarGz(xcterPkgPath)
	if valiErr != nil {
		s.log.Error(
			"xcounter程序包校验失败",
			zap.Error(valiErr),
			zap.String("path", xcterPkgPath),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.ErrZIPFileIsNotValid.WithCause(valiErr)
	}
	if err := archive.UntarGz(xcterPkgPath, tmpDir); err != nil {
		s.log.Error(
			"压缩xcounter程序包失败",
			zap.Error(err),
			zap.String("src_path", xcterPkgPath),
			zap.String("dst_path", tmpDir),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.ErrUnZIPFailed.WithCause(err)
	}
	xcterTmpDir := filepath.Join(tmpDir, xcterUnTarDirName, "bin")
	oesBinDir := filepath.Join(colonyBinDir, "bin")
	if err := fileutil.CopyDir(ctx, xcterTmpDir, oesBinDir, true); err != nil {
		s.log.Error(
			"复制xcounter程序包解压目录失败",
			zap.Error(err),
			zap.String("src_path", xcterTmpDir),
			zap.String("dst_path", oesBinDir),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.FromError(err)
	}

	colonyConfAll := filepath.Join(colonyConfDir, "all")
	if _, err := os.Stat(colonyConfAll); os.IsNotExist(err) {
		colonyBinConf := filepath.Join(colonyBinDir, "conf")
		if err := fileutil.CopyDir(ctx, colonyBinConf, colonyConfAll, true); err != nil {
			s.log.Error(
				"复制oes集群配置文件失败",
				zap.Error(err),
				zap.String("src_path", colonyBinConf),
				zap.String("dst_path", colonyConfAll),
				zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			)
			return errors.FromError(err)
		}
		srcPath := filepath.Join(config.ConfigDir, fmt.Sprintf("automatic_oes_%s.yaml", m.SystemType))
		dstPath := filepath.Join(colonyConfAll, "automatic.yaml")
		if err := fileutil.CopyFile(ctx, srcPath, dstPath); err != nil {
			s.log.Error(
				"复制oes的automatic配置文件失败",
				zap.Error(err),
				zap.String("src_path", colonyConfDir),
				zap.String("dst_path", dstPath),
				zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			)
			return errors.FromError(err)
		}
	}
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
		s.log.Error(
			"导出oes集群配置变量文件失败",
			zap.Error(err),
			zap.String("path", oesColonyConf),
			zap.Object("oes_colony_vars", &oesVars),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.ErrExportCacheFileFailed.WithCause(err)
	}

	s.log.Info(
		"解压oes程序包并初始化集群配置文件成功",
		zap.String("path", oesColonyConf),
		zap.Object("oes_colony_vars", &oesVars),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return nil
}
