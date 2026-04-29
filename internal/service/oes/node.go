package oes

import (
	"context"
	"path/filepath"
	"time"

	"go.uber.org/zap"

	oesmodel "gin-artweb/internal/model/oes"
	oesrepo "gin-artweb/internal/repo/oes"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
	"gin-artweb/pkg/serializer"
)

type OesNodeService struct {
	log      *zap.Logger
	nodeRepo *oesrepo.OesNodeRepo
}

func NewOesNodeService(
	log *zap.Logger,
	nodeRepo *oesrepo.OesNodeRepo,
) *OesNodeService {
	return &OesNodeService{
		log:      log,
		nodeRepo: nodeRepo,
	}
}

func (s *OesNodeService) CreateOesNode(
	ctx context.Context,
	dto oesmodel.OesNodeUpsertDTO,
) (*oesmodel.OesNodeModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"创建oes节点:开始执行",
		zap.Object("oes_node_dto", &dto),
	)

	m := dto.ToModel()
	if err := s.nodeRepo.CreateModel(ctx, &m); err != nil {
		log.Error(
			"创建oes节点:创建数据库模型失败",
			zap.Error(err),
			zap.Object("oes_node", &m),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, nil)
	}

	// 导出oes节点缓存数据
	if err := s.OutPortOesNodeData(ctx, &m); err != nil {
		log.Error(
			"创建oes节点:导出缓存数据失败",
			zap.Error(err),
			zap.Uint32("oes_node_id", m.ID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, err
	}

	log.Info(
		"创建oes节点:执行成功",
		zap.Uint32("oes_node_id", m.ID),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return &m, nil
}

func (s *OesNodeService) UpdateOesNodeByID(
	ctx context.Context,
	oesNodeID uint32,
	dto oesmodel.OesNodeUpsertDTO,
) (*oesmodel.OesNodeModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"更新oes节点:开始执行",
		zap.Uint32("oes_node_id", oesNodeID),
		zap.Object("oes_node_dto", &dto),
	)

	updateData := dto.ToUpdateMap()
	if err := s.nodeRepo.UpdateModel(ctx, updateData, "id = ?", oesNodeID); err != nil {
		log.Error(
			"更新oes节点:更新数据库模型失败",
			zap.Error(err),
			zap.Uint32("oes_node_id", oesNodeID),
			zap.Any("update_data", updateData),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, updateData)
	}

	// 查询oes节点关联数据
	m, rErr := s.FindOesNodeByID(ctx, []string{"OesColony", "Host"}, oesNodeID)
	if rErr != nil {
		log.Error(
			"更新oes节点:查询关联数据失败",
			zap.Error(rErr),
			zap.Uint32("oes_node_id", oesNodeID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, rErr
	}

	// 导出oes节点缓存数据
	if err := s.OutPortOesNodeData(ctx, m); err != nil {
		log.Error(
			"更新oes节点:导出缓存数据失败",
			zap.Error(err),
			zap.Uint32("oes_node_id", oesNodeID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, err
	}

	log.Info(
		"更新oes节点:执行成功",
		zap.Uint32("oes_node_id", oesNodeID),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return m, nil
}

func (s *OesNodeService) DeleteOesNodeByID(
	ctx context.Context,
	oesNodeID uint32,
) *errors.Error {
	startTime := time.Now()
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"删除oes节点:开始执行",
		zap.Uint32("oes_node_id", oesNodeID),
	)

	if err := s.nodeRepo.DeleteModel(ctx, oesNodeID); err != nil {
		log.Error(
			"删除oes节点:删除数据库模型失败",
			zap.Error(err),
			zap.Uint32("oes_node_id", oesNodeID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.NewGormError(err, map[string]any{"id": oesNodeID})
	}

	log.Info(
		"删除oes节点:执行成功",
		zap.Uint32("oes_node_id", oesNodeID),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

func (s *OesNodeService) FindOesNodeByID(
	ctx context.Context,
	preloads []string,
	oesNodeID uint32,
) (*oesmodel.OesNodeModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	m, err := s.nodeRepo.GetModel(ctx, preloads, oesNodeID)
	if err != nil {
		log.Error(
			"查询oes节点:查询数据库模型失败",
			zap.Error(err),
			zap.Uint32("oes_node_id", oesNodeID),
			zap.Strings("preloads", preloads),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, map[string]any{"id": oesNodeID})
	}

	log.Debug(
		"查询oes节点:查询到的数据库模型详情",
		zap.Object("oes_node_model", m),
	)
	return m, nil
}

func (s *OesNodeService) ListOesNode(
	ctx context.Context,
	page, size int,
	dto oesmodel.ListOesNodeDTO,
) (int64, []oesmodel.OesNodeModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return 0, nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Debug(
		"查询oes节点列表:入参详情",
		zap.Int("page", page),
		zap.Int("size", size),
		zap.Object("oes_node_dto", &dto),
	)

	limit, offset := common.Page2LimitOffset(page, size)
	qp := database.QueryParams{
		Preloads: []string{"Host", "OesColony"},
		OrderBy:  []string{"id DESC"},
		Limit:    limit,
		Offset:   offset,
		Query:    dto.ToQueryMap(),
	}

	count, err := s.nodeRepo.CountModel(ctx, qp.Query)
	if err != nil {
		log.Error(
			"查询oes节点列表:查询数据库模型总数失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}

	if count == 0 {
		log.Warn(
			"查询oes节点列表:数据库模型总数为0",
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, nil, nil
	}

	ms, err := s.nodeRepo.ListModel(ctx, qp)
	if err != nil {
		log.Error(
			"查询oes节点列表:查询数据库模型列表失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}
	return count, ms, nil
}

func (s *OesNodeService) OutPortOesNodeData(
	ctx context.Context,
	m *oesmodel.OesNodeModel,
) *errors.Error {
	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Debug(
		"导出oes节点变量文件:开始执行",
		zap.Object("oes_node", m),
	)

	specdir := RoleToSpecdir(m.NodeRole)
	oesColonyConf := GetOesNodeConfigPath(m.OesColony.ColonyNum, specdir)

	oesVars := oesmodel.OesNodeVars{
		ID:       m.ID,
		NodeRole: m.NodeRole,
		Specdir:  specdir,
		HostID:   m.HostID,
		IsEnable: m.IsEnable,
	}

	if _, err := serializer.WriteYAML(oesColonyConf, oesVars); err != nil {
		log.Error(
			"导出oes节点变量文件:写入文件失败",
			zap.Error(err),
			zap.Uint32("oes_node_id", m.ID),
			zap.String("colony_num", m.OesColony.ColonyNum),
			zap.String("path", oesColonyConf),
			zap.Object("oes_node_vars", &oesVars),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.ErrExportCacheFileFailed.WithCause(err)
	}
	return nil
}

func RoleToSpecdir(role string) string {
	switch role {
	case "master":
		return "host_01"
	case "follow":
		return "host_02"
	default:
		return "host_03"
	}
}

func GetOesNodeConfigPath(colonyNum, specdir string) string {
	return filepath.Join(GetOesColonyConfigDir(colonyNum), specdir, "node.yaml")
}
