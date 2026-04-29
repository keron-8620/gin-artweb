package mon

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"go.uber.org/zap"

	monmodel "gin-artweb/internal/model/mon"
	monrepo "gin-artweb/internal/repo/mon"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
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

	outputPath := GetMonNodeExportPath(m.ID)
	if err := s.ExportMonNode(ctx, m, outputPath); err != nil {
		log.Error(
			"创建mon节点:导出节点文件失败",
			zap.Error(err),
			zap.Object("mon_node_model", &m),
			zap.String("output_path", outputPath),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, err
	}

	log.Info(
		"创建mon节点:执行成功",
		zap.Uint32("mon_node_id", m.ID),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return s.FindMonNodeByID(ctx, []string{"Host"}, m.ID)
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

	m, rErr := s.FindMonNodeByID(ctx, []string{"Host"}, nodeID)
	if rErr != nil {
		log.Error(
			"更新mon节点:查询更新后的mon节点失败",
			zap.Error(rErr),
			zap.Uint32("mon_node_id", nodeID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, rErr
	}

	outputPath := GetMonNodeExportPath(nodeID)
	if rErr := s.ExportMonNode(ctx, *m, outputPath); rErr != nil {
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

	outportPath := GetMonNodeExportPath(nodeID)
	if err := fileutil.Remove(ctx, outportPath); err != nil {
		log.Error(
			"删除mon节点:删除节点文件失败, 请手动删除",
			zap.Error(err),
			zap.String("outport_path", outportPath),
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
	page, size int,
	dto monmodel.ListMonNodeDTO,
) (int64, []monmodel.MonNodeModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return 0, nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Debug(
		"查询mon节点列表:参数详情",
		zap.Int("page", page),
		zap.Int("size", size),
		zap.Object("list_mon_node_dto", &dto),
	)

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
		return 0, nil, errors.NewGormError(err, nil)
	}

	if count == 0 {
		log.Warn(
			"查询mon节点列表:数据库模型总数为0",
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, nil, nil
	}

	ms, err := s.nodeRepo.ListModel(ctx, qp)
	if err != nil {
		log.Error(
			"查询mon节点列表:查询数据库模型失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}
	return count, ms, nil
}

func (s *MonNodeService) ExportMonNode(
	ctx context.Context,
	m monmodel.MonNodeModel,
	outputPath string,
) *errors.Error {
	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Debug(
		"导出mon节点文件:入参详情",
		zap.Object("mon_node_model", &m),
		zap.String("output_path", outputPath),
	)

	monNodeVars := monmodel.MonNodeModelToNodeVars(m)
	if _, err := serializer.WriteYAML(outputPath, monNodeVars); err != nil {
		log.Error(
			"导出mon节点文件:写入文件失败",
			zap.Error(err),
			zap.String("path", outputPath),
			zap.Object("mon_node_vars", &monNodeVars),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.ErrExportCacheFileFailed.WithCause(err)
	}
	return nil
}

func GetMonNodeExportPath(pk uint32) string {
	return filepath.Join(config.StorageDir, "mon", "config", fmt.Sprintf("%d", pk), "mon.yaml")
}
