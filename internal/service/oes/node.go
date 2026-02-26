package biz

import (
	"context"
	"path/filepath"

	"go.uber.org/zap"

	oesmodel "gin-artweb/internal/model/oes"
	oesrepo "gin-artweb/internal/repository/oes"
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
	m oesmodel.OesNodeModel,
) (*oesmodel.OesNodeModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始创建oes节点",
		zap.Object(database.ModelKey, &m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	if err := s.nodeRepo.CreateModel(ctx, &m); err != nil {
		s.log.Error(
			"创建oes节点失败",
			zap.Error(err),
			zap.Object(database.ModelKey, &m),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, nil)
	}

	// 查询oes节点关联数据
	nm, rErr := s.FindOesNodeByID(ctx, []string{"OesColony", "Host"}, m.ID)
	if rErr != nil {
		return nil, rErr
	}

	// 导出oes节点缓存数据
	if err := s.OutPortOesNodeData(ctx, nm); err != nil {
		return nil, err
	}

	s.log.Info(
		"创建oes节点成功",
		zap.Object(database.ModelKey, nm),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return nm, nil
}

func (s *OesNodeService) UpdateOesNodeByID(
	ctx context.Context,
	oesNodeID uint32,
	data map[string]any,
) (*oesmodel.OesNodeModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始更新oes节点",
		zap.Uint32("oes_node_id", oesNodeID),
		zap.Any(database.UpdateDataKey, data),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	data["id"] = oesNodeID
	if err := s.nodeRepo.UpdateModel(ctx, data, "id = ?", oesNodeID); err != nil {
		s.log.Error(
			"更新oes节点失败",
			zap.Error(err),
			zap.Uint32("oes_node_id", oesNodeID),
			zap.Any(database.UpdateDataKey, data),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, data)
	}

	// 查询oes节点关联数据
	m, rErr := s.FindOesNodeByID(ctx, []string{"OesColony", "Host"}, oesNodeID)
	if rErr != nil {
		return nil, rErr
	}

	// 导出oes节点缓存数据
	if err := s.OutPortOesNodeData(ctx, m); err != nil {
		return nil, err
	}

	s.log.Info(
		"更新oes节点成功",
		zap.Uint32("oes_node_id", oesNodeID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return m, nil
}

func (s *OesNodeService) DeleteOesNodeByID(
	ctx context.Context,
	oesNodeID uint32,
) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始删除oes节点",
		zap.Uint32("oes_node_id", oesNodeID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	if err := s.nodeRepo.DeleteModel(ctx, oesNodeID); err != nil {
		s.log.Error(
			"删除oes节点失败",
			zap.Error(err),
			zap.Uint32("oes_node_id", oesNodeID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.NewGormError(err, map[string]any{"id": oesNodeID})
	}

	s.log.Info(
		"删除oes节点成功",
		zap.Uint32("oes_node_id", oesNodeID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return nil
}

func (s *OesNodeService) FindOesNodeByID(
	ctx context.Context,
	preloads []string,
	oesNodeID uint32,
) (*oesmodel.OesNodeModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始查询oes节点",
		zap.Strings(database.PreloadKey, preloads),
		zap.Uint32("oes_node_id", oesNodeID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, err := s.nodeRepo.GetModel(ctx, preloads, oesNodeID)
	if err != nil {
		s.log.Error(
			"查询oes节点失败",
			zap.Error(err),
			zap.Uint32("oes_node_id", oesNodeID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, map[string]any{"id": oesNodeID})
	}

	s.log.Info(
		"查询oes节点成功",
		zap.Uint32("oes_node_id", oesNodeID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return m, nil
}

func (s *OesNodeService) ListOesNode(
	ctx context.Context,
	qp database.QueryParams,
) (int64, *[]oesmodel.OesNodeModel, *errors.Error) {
	if ctx.Err() != nil {
		return 0, nil, errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始查询oes节点列表",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	count, ms, err := s.nodeRepo.ListModel(ctx, qp)
	if err != nil {
		s.log.Error(
			"查询oes节点列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}

	s.log.Info(
		"查询oes节点列表成功",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return count, ms, nil
}

func (s *OesNodeService) OutPortOesNodeData(ctx context.Context, m *oesmodel.OesNodeModel) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始导出oes节点变量文件",
		zap.Object(database.ModelKey, m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	var specdir string
	switch m.NodeRole {
	case "master":
		specdir = "host_01"
	case "follow":
		specdir = "host_02"
	default:
		specdir = "host_03"
	}
	oesVars := oesmodel.OesNodeVars{
		ID:       m.ID,
		NodeRole: m.NodeRole,
		Specdir:  specdir,
		HostID:   m.HostID,
		IsEnable: m.IsEnable,
	}
	confDir := common.GetOesColonyConfigDir(m.OesColony.ColonyNum)
	oesColonyConf := filepath.Join(confDir, specdir, "node.yaml")
	if _, err := serializer.WriteYAML(oesColonyConf, oesVars); err != nil {
		s.log.Error(
			"导出oes节点变量文件失败",
			zap.Error(err),
			zap.Uint32("oes_node_id", m.ID),
			zap.String("colony_num", m.OesColony.ColonyNum),
			zap.String("path", oesColonyConf),
			zap.Object("oes_node_vars", &oesVars),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.ErrExportCacheFileFailed.WithCause(err)
	}

	s.log.Info(
		"导出oes节点变量文件失败",
		zap.String("path", oesColonyConf),
		zap.Object("oes_colony_vars", &oesVars),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return nil
}
