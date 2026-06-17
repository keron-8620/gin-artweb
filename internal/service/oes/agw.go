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

type AgwService struct {
	log     *zap.Logger
	agwRepo *oesrepo.AgwRepo
}

func NewAgwService(
	log *zap.Logger,
	agwRepo *oesrepo.AgwRepo,
) *AgwService {
	return &AgwService{
		log:     log,
		agwRepo: agwRepo,
	}
}

func (s *AgwService) CreateAgw(
	ctx context.Context,
	dto oesmodel.AgwUpsertDTO,
) (*oesmodel.AgwModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"创建agw节点:开始执行",
		zap.Object("agw_dto", &dto),
	)

	m := dto.ToModel()
	if err := s.agwRepo.CreateModel(ctx, &m); err != nil {
		log.Error(
			"创建agw节点:创建数据库模型失败",
			zap.Error(err),
			zap.Object("agw_model", &m),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, nil)
	}

	if err := s.OutportAgwData(ctx, &m); err != nil {
		log.Error(
			"创建agw节点:导出节点文件失败",
			zap.Error(err),
			zap.Object("agw_model", &m),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, err
	}

	log.Info(
		"创建agw节点:执行成功",
		zap.Uint32("agw_id", m.ID),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return s.FindAgwByID(ctx, []string{"Host", "Package"}, m.ID)
}

func (s *AgwService) UpdateAgwByID(
	ctx context.Context,
	agwID uint32,
	dto oesmodel.AgwUpsertDTO,
) (*oesmodel.AgwModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"更新agw节点:开始执行",
		zap.Uint32("agw_id", agwID),
		zap.Object("agw_dto", &dto),
	)

	updateData := dto.ToUpdateMap()
	if err := s.agwRepo.UpdateModel(ctx, updateData, "id = ?", agwID); err != nil {
		log.Error(
			"更新agw节点:更新数据库模型失败",
			zap.Error(err),
			zap.Uint32("agw_id", agwID),
			zap.Any("update_data", updateData),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, nil)
	}

	m, rErr := s.FindAgwByID(ctx, []string{"Host", "Package"}, agwID)
	if rErr != nil {
		log.Error(
			"更新agw节点:查询更新后的agw节点失败",
			zap.Error(rErr),
			zap.Uint32("agw_id", agwID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, rErr
	}

	if rErr := s.OutportAgwData(ctx, m); rErr != nil {
		log.Error(
			"更新agw节点:导出节点文件失败",
			zap.Error(rErr),
			zap.Uint32("agw_id", agwID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, rErr
	}

	log.Info(
		"更新agw节点:执行成功",
		zap.Uint32("agw_id", agwID),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return m, nil
}

func (s *AgwService) DeleteAgwByID(
	ctx context.Context,
	agwID uint32,
) *errors.Error {
	startTime := time.Now()
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"删除agw节点:开始执行",
		zap.Uint32("agw_id", agwID),
	)

	if err := s.agwRepo.DeleteModel(ctx, agwID); err != nil {
		log.Error(
			"删除agw节点:删除数据库模型失败",
			zap.Error(err),
			zap.Uint32("agw_id", agwID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.NewGormError(err, map[string]any{"id": agwID})
	}

	log.Info(
		"删除agw节点:执行成功",
		zap.Uint32("agw_id", agwID),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

func (s *AgwService) FindAgwByID(
	ctx context.Context,
	preloads []string,
	agwID uint32,
) (*oesmodel.AgwModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	m, err := s.agwRepo.GetModel(ctx, preloads, agwID)
	if err != nil {
		log.Error(
			"查询agw节点:查询数据库模型失败",
			zap.Error(err),
			zap.Uint32("agw_id", agwID),
			zap.Strings("preloads", preloads),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, map[string]any{"id": agwID})
	}

	log.Debug(
		"查询agw节点:查询到的数据库模型详情",
		zap.Object("agw_model", m),
	)
	return m, nil
}

func (s *AgwService) ListAgw(
	ctx context.Context,
	page, size int,
	dto oesmodel.ListAgwDTO,
) (int64, []oesmodel.AgwModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return 0, nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Debug(
		"查询agw节点列表:参数详情",
		zap.Int("page", page),
		zap.Int("size", size),
		zap.Object("list_agw_dto", &dto),
	)

	limit, offset := common.Page2LimitOffset(page, size)
	qp := database.QueryParams{
		Preloads: []string{"Host", "Package"},
		OrderBy:  []string{"id DESC"},
		Limit:    limit,
		Offset:   offset,
		Query:    dto.ToQueryMap(),
	}

	count, err := s.agwRepo.CountModel(ctx, qp.Query)
	if err != nil {
		log.Error(
			"查询agw节点列表:查询数据库模型总数失败",
			zap.Error(err),
			zap.Any("query", qp.Query),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}

	if count == 0 {
		log.Warn(
			"查询agw节点列表:数据库模型总数为0",
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, nil, nil
	}

	ms, err := s.agwRepo.ListModel(ctx, qp)
	if err != nil {
		log.Error(
			"查询agw节点列表:查询数据库模型失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}
	return count, ms, nil
}

func (s *AgwService) OutportAgwData(
	ctx context.Context,
	m *oesmodel.AgwModel,
) *errors.Error {
	startTime := time.Now()
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"解压agw程序包并初始化集群配置文件:开始执行",
		zap.Object("agw", m),
	)

	agwBinDir := GetAgwBinDir(m.ID)
	agwConfDir := GetAgwConfigDir(m.ID)

	// 清理原agw集群配置文件
	if _, err := os.Stat(agwBinDir); !os.IsNotExist(err) {
		if err := os.RemoveAll(agwBinDir); err != nil {
			log.Error(
				"解压agw程序包并初始化集群配置文件:清理原agw集群配置文件失败",
				zap.Error(err),
				zap.String("path", agwBinDir),
				zap.Duration("total_duration", time.Since(startTime)),
			)
			return errors.FromError(err)
		}
	}

	// 创建临时目录
	tmpDir, mErr := os.MkdirTemp("/tmp", "agw-")
	if mErr != nil {
		log.Error(
			"解压agw程序包并初始化集群配置文件:创建agw程序包解压的tmp文件夹失败",
			zap.Error(mErr),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.FromError(mErr)
	}

	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			log.Error(
				"解压agw程序包并初始化集群配置文件:删除agw程序包解压的tmp文件夹失败",
				zap.Error(err),
				zap.String("path", tmpDir),
			)
		}
	}()

	// 处理agw程序包
	agwPkgPath := resosvc.GetPackageStoragePath(m.Package.StorageFilename)
	agwUnTarDirName, valiErr := archive.ValidateSingleDirTarGz(agwPkgPath)
	if valiErr != nil {
		log.Error(
			"解压agw程序包并初始化集群配置文件:agw程序包校验失败",
			zap.Error(valiErr),
			zap.String("path", agwPkgPath),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.ErrZIPFileIsNotValid.WithCause(valiErr)
	}

	if err := archive.UntarGz(agwPkgPath, tmpDir, archive.WithContext(ctx)); err != nil {
		log.Error(
			"解压agw程序包并初始化集群配置文件:解压agw程序包失败",
			zap.Error(err),
			zap.String("src_path", agwPkgPath),
			zap.String("dst_path", agwBinDir),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.ErrUnZIPFailed.WithCause(err)
	}

	agwTmpDir := filepath.Join(tmpDir, agwUnTarDirName)
	if err := fileutil.CopyDir(ctx, agwTmpDir, agwBinDir, true); err != nil {
		log.Error(
			"解压agw程序包并初始化集群配置文件:复制agw程序包解压目录失败",
			zap.Error(err),
			zap.String("src_path", agwTmpDir),
			zap.String("dst_path", agwBinDir),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.FromError(err)
	}

	// 处理配置文件
	if _, err := os.Stat(agwConfDir); os.IsNotExist(err) {
		colonyBinConf := filepath.Join(agwBinDir, "conf")
		if err := fileutil.CopyDir(ctx, colonyBinConf, agwConfDir, true); err != nil {
			log.Error(
				"解压agw程序包并初始化集群配置文件:复制agw集群配置文件失败",
				zap.Error(err),
				zap.String("src_path", colonyBinConf),
				zap.String("dst_path", agwConfDir),
				zap.Duration("total_duration", time.Since(startTime)),
			)
			return errors.FromError(err)
		}
	}

	// 导出agw集群配置变量文件
	agwVars := oesmodel.OesAgwVars{
		ID:         m.ID,
		Name:       m.Name,
		DeployPath: m.DeployPath,
		HostID:     m.HostID,
		PackageID:  m.PackageID,
	}
	agwConfPath := filepath.Join(agwConfDir, "agw.yaml")
	if _, err := serializer.WriteYAML(agwConfPath, agwVars); err != nil {
		log.Error(
			"解压agw程序包并初始化集群配置文件:导出agw集群配置变量文件失败",
			zap.Error(err),
			zap.String("path", agwConfPath),
			zap.Object("agw_vars", &agwVars),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.ErrExportCacheFileFailed.WithCause(err)
	}

	log.Info(
		"解压agw程序包并初始化集群配置文件:执行成功",
		zap.String("path", agwConfPath),
		zap.Object("agw_vars", &agwVars),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

func GetAgwBinDir(agwID uint32) string {
	return filepath.Join(config.StorageDir, "agw", "bin", fmt.Sprintf("%d", agwID))
}

func GetAgwConfigDir(agwID uint32) string {
	return filepath.Join(config.StorageDir, "agw", "config", fmt.Sprintf("%d", agwID))
}
