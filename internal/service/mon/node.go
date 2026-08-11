package mon

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"

	monmodel "gin-artweb/internal/model/mon"
	monrepo "gin-artweb/internal/repo/mon"
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

type MonNodeService struct {
	log      *zap.Logger
	nodeRepo *monrepo.MonNodeRepo
}

func NewMonNodeService(
	log *zap.Logger,
	nodeRepo *monrepo.MonNodeRepo,
) *MonNodeService {
	return &MonNodeService{
		log:      log,
		nodeRepo: nodeRepo,
	}
}

func (s *MonNodeService) CreateMonNode(
	ctx context.Context,
	dto monmodel.MonNodeUpsertDTO,
) (*monmodel.MonNodeModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"创建mon节点:开始执行",
		zap.Object("mon_node_dto", &dto),
	)

	m := dto.ToModel()
	if err := s.nodeRepo.CreateModel(ctx, &m); err != nil {
		log.Error(
			"创建mon节点:创建数据库模型失败",
			zap.Error(err),
			zap.Object("mon_node_model", &m),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, nil)
	}

	created, rErr := s.FindMonNodeByID(ctx, []string{"Host", "Package", "Jdk"}, m.ID)
	if rErr != nil {
		log.Error(
			"创建mon节点:查询创建后的mon节点失败",
			zap.Error(rErr),
			zap.Object("mon_node_model", &m),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, rErr
	}

	if err := s.OutportMonData(ctx, *created); err != nil {
		log.Error(
			"创建mon节点:导出节点文件失败",
			zap.Error(err),
			zap.Object("mon_node_model", created),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, err
	}

	log.Info(
		"创建mon节点:执行成功",
		zap.Uint32("mon_node_id", created.ID),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return created, nil
}

func (s *MonNodeService) UpdateMonNodeByID(
	ctx context.Context,
	nodeID uint32,
	dto monmodel.MonNodeUpsertDTO,
) (*monmodel.MonNodeModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"更新mon节点:开始执行",
		zap.Uint32("mon_node_id", nodeID),
		zap.Object("mon_node_dto", &dto),
	)

	updateData := dto.ToUpdateMap()
	if err := s.nodeRepo.UpdateModel(ctx, updateData, "id = ?", nodeID); err != nil {
		log.Error(
			"更新mon节点:更新数据库模型失败",
			zap.Error(err),
			zap.Uint32("mon_node_id", nodeID),
			zap.Any("update_data", updateData),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, nil)
	}

	m, rErr := s.FindMonNodeByID(ctx, []string{"Host", "Package", "Jdk"}, nodeID)
	if rErr != nil {
		log.Error(
			"更新mon节点:查询更新后的mon节点失败",
			zap.Error(rErr),
			zap.Uint32("mon_node_id", nodeID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, rErr
	}

	if rErr := s.OutportMonData(ctx, *m); rErr != nil {
		log.Error(
			"更新mon节点:导出节点文件失败",
			zap.Error(rErr),
			zap.Uint32("mon_node_id", nodeID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, rErr
	}

	log.Info(
		"更新mon节点:执行成功",
		zap.Uint32("mon_node_id", nodeID),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return m, nil
}

func (s *MonNodeService) DeleteMonNodeByID(
	ctx context.Context,
	nodeID uint32,
) *errors.Error {
	startTime := time.Now()
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"删除mon节点:开始执行",
		zap.Uint32("mon_node_id", nodeID),
	)

	if err := s.nodeRepo.DeleteModel(ctx, nodeID); err != nil {
		log.Error(
			"删除mon节点:删除数据库模型失败",
			zap.Error(err),
			zap.Uint32("mon_node_id", nodeID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.NewGormError(err, map[string]any{"id": nodeID})
	}

	monNodeBin := GetMonNodeBinDir(nodeID)
	if err := fileutil.RemoveAll(ctx, monNodeBin); err != nil {
		log.Error(
			"删除mon节点:删除节点程序包文件失败, 请手动删除",
			zap.Error(err),
			zap.String("outport_path", monNodeBin),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.ErrDeleteCacheFileFailed.WithCause(err)
	}

	monNodeConf := GetMonNodeConfDir(nodeID)
	if err := fileutil.RemoveAll(ctx, monNodeConf); err != nil {
		log.Error(
			"删除mon节点:删除节点配置文件失败, 请手动删除",
			zap.Error(err),
			zap.String("outport_path", monNodeConf),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.ErrDeleteCacheFileFailed.WithCause(err)
	}

	log.Info(
		"删除mon节点:执行成功",
		zap.Uint32("mon_node_id", nodeID),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

func (s *MonNodeService) FindMonNodeByID(
	ctx context.Context,
	preloads []string,
	nodeID uint32,
) (*monmodel.MonNodeModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	m, err := s.nodeRepo.GetModel(ctx, preloads, nodeID)
	if err != nil {
		log.Error(
			"查询mon节点:查询数据库模型失败",
			zap.Error(err),
			zap.Uint32("mon_node_id", nodeID),
			zap.Strings("preloads", preloads),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, map[string]any{"id": nodeID})
	}

	log.Debug(
		"查询mon节点:查询到的数据库模型详情",
		zap.Object("mon_node_model", m),
	)
	return m, nil
}

func (s *MonNodeService) ListMonNode(
	ctx context.Context,
	dto monmodel.ListMonNodeDTO,
) (int, int, int64, []monmodel.MonNodeModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return 0, 0, 0, nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Debug(
		"查询mon节点列表:参数详情",
		zap.Object("list_mon_node_dto", &dto),
	)

	page, size := dto.StandardModelQuery.GetPageParam()
	limit, offset := common.Page2LimitOffset(page, size)
	qp := database.QueryParams{
		Preloads: []string{"Host"},
		OrderBy:  []string{"id DESC"},
		Limit:    limit,
		Offset:   offset,
		Query:    dto.ToQueryMap(),
	}

	count, err := s.nodeRepo.CountModel(ctx, qp.Query)
	if err != nil {
		log.Error(
			"查询mon节点列表:查询数据库模型总数失败",
			zap.Error(err),
			zap.Any("query", qp.Query),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return page, size, 0, nil, errors.NewGormError(err, nil)
	}

	if count == 0 {
		log.Warn(
			"查询mon节点列表:数据库模型总数为0",
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return page, size, 0, nil, nil
	}

	ms, err := s.nodeRepo.ListModel(ctx, qp)
	if err != nil {
		log.Error(
			"查询mon节点列表:查询数据库模型失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return page, size, 0, nil, errors.NewGormError(err, nil)
	}
	return page, size, count, ms, nil
}

func (s *MonNodeService) OutportMonData(
	ctx context.Context,
	m monmodel.MonNodeModel,
) *errors.Error {
	startTime := time.Now()
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"解压mon程序包并初始化配置文件:开始执行",
		zap.Object("mon_node_model", &m),
	)

	monNodeBin := GetMonNodeBinDir(m.ID)
	if _, err := os.Stat(monNodeBin); !os.IsNotExist(err) {
		if err := os.RemoveAll(monNodeBin); err != nil {
			log.Error(
				"解压mon程序包并初始化配置文件:清理原mon节点程序包文件夹失败",
				zap.Error(err),
				zap.String("path", monNodeBin),
				zap.Duration("total_duration", time.Since(startTime)),
			)
			return errors.ErrDeleteCacheFileFailed.WithCause(err)
		}
	}

	tmpDir, mErr := os.MkdirTemp("/tmp", "mon-")
	if mErr != nil {
		log.Error(
			"解压mon程序包并初始化配置文件:创建临时文件夹失败",
			zap.Error(mErr),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.FromError(mErr)
	}

	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			log.Error(
				"解压mon程序包并初始化配置文件:删除临时文件夹失败",
				zap.Error(err),
				zap.String("path", tmpDir),
				zap.Duration("total_duration", time.Since(startTime)),
			)
		}
	}()

	monPkgPath := resocvs.GetPackageStoragePath(m.Package.StorageFilename)
	monUnTarDirName, valiErr := archive.ValidateSingleDirTarGz(monPkgPath)
	if valiErr != nil {
		log.Error(
			"解压mon程序包并初始化配置文件:mon程序包校验失败",
			zap.Error(valiErr),
			zap.String("path", monPkgPath),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.ErrValidationFailed.WithCause(valiErr)
	}

	if err := archive.UntarGz(monPkgPath, tmpDir, archive.WithContext(ctx)); err != nil {
		log.Error(
			"解压mon程序包并初始化配置文件:解压mon程序包失败",
			zap.Error(err),
			zap.Uint32("mon_id", m.ID),
			zap.String("src_path", monPkgPath),
			zap.String("dest_path", tmpDir),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.ErrUnZIPFailed.WithCause(err)
	}

	monTmpDir := filepath.Join(tmpDir, monUnTarDirName)
	if err := fileutil.CopyDir(ctx, monTmpDir, monNodeBin, true); err != nil {
		log.Error(
			"解压mon程序包并初始化配置文件:复制mon程序包解压目录失败",
			zap.Error(err),
			zap.String("src_path", monTmpDir),
			zap.String("dst_path", monNodeBin),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.FromError(err)
	}

	// 处理配置文件
	monNodeConf := GetMonNodeConfDir(m.ID)
	if _, err := os.Stat(monNodeConf); os.IsNotExist(err) {
		monNodeBinConf := filepath.Join(monNodeBin, "conf")
		if err := fileutil.CopyDir(ctx, monNodeBinConf, monNodeConf, true); err != nil {
			log.Error(
				"解压mon程序包并初始化配置文件:复制mon配置文件失败",
				zap.Error(err),
				zap.String("src_path", monNodeBinConf),
				zap.String("dst_path", monNodeConf),
				zap.Duration("total_duration", time.Since(startTime)),
			)
			return errors.FromError(err)
		}
	}

	monConfPath := filepath.Join(monNodeConf, "mon.yaml")
	monNodeVars := monmodel.MonNodeModelToNodeVars(m)
	if _, err := serializer.WriteYAML(monConfPath, monNodeVars); err != nil {
		log.Error(
			"解压mon程序包并初始化配置文件:写入文件失败",
			zap.Error(err),
			zap.String("path", monConfPath),
			zap.Object("mon_node_vars", &monNodeVars),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.ErrExportCacheFileFailed.WithCause(err)
	}
	return nil
}

func GetMonNodeBinDir(monID uint32) string {
	return filepath.Join(config.StorageDir, "mon", "bin", fmt.Sprintf("%d", monID))
}

func GetMonNodeConfDir(monID uint32) string {
	return filepath.Join(config.StorageDir, "mon", "config", fmt.Sprintf("%d", monID))
}
