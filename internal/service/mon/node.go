package biz

import (
	"context"
	"os"

	"go.uber.org/zap"

	monmodel "gin-artweb/internal/model/mon"
	monrepo "gin-artweb/internal/repository/mon"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
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
	m monmodel.MonNodeModel,
) (*monmodel.MonNodeModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始创建mon节点",
		zap.Object(database.ModelKey, &m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	if err := s.nodeRepo.CreateModel(ctx, &m); err != nil {
		s.log.Error(
			"创建mon节点失败",
			zap.Error(err),
			zap.Object(database.ModelKey, &m),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, nil)
	}

	if err := s.ExportMonNode(ctx, m); err != nil {
		return nil, err
	}

	s.log.Info(
		"创建mon节点成功",
		zap.Object(database.ModelKey, &m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return s.FindMonNodeByID(ctx, []string{"Host"}, m.ID)
}

func (s *MonNodeService) UpdateMonNodeByID(
	ctx context.Context,
	nodeID uint32,
	data map[string]any,
) (*monmodel.MonNodeModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始更新mon节点",
		zap.Uint32("mon_node_id", nodeID),
		zap.Any(database.UpdateDataKey, data),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	if err := s.nodeRepo.UpdateModel(ctx, data, "id = ?", nodeID); err != nil {
		s.log.Error(
			"更新mon节点失败",
			zap.Error(err),
			zap.Uint32("mon_node_id", nodeID),
			zap.Any(database.UpdateDataKey, data),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, data)
	}

	m, rErr := s.FindMonNodeByID(ctx, []string{"Host"}, nodeID)
	if rErr != nil {
		return nil, rErr
	}

	if rErr := s.ExportMonNode(ctx, *m); rErr != nil {
		return nil, rErr
	}

	s.log.Info(
		"更新mon节点成功",
		zap.Uint32("mon_node_id", nodeID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return m, nil
}

func (s *MonNodeService) DeleteMonNodeByID(
	ctx context.Context,
	nodeID uint32,
) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始删除mon",
		zap.Uint32("mon_node_id", nodeID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	if err := s.nodeRepo.DeleteModel(ctx, nodeID); err != nil {
		s.log.Error(
			"删除mon失败",
			zap.Error(err),
			zap.Uint32("mon_node_id", nodeID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.NewGormError(err, map[string]any{"id": nodeID})
	}

	path := common.GetMonNodeExportPath(nodeID)
	if err := os.RemoveAll(path); err != nil && !os.IsNotExist(err) {
		s.log.Error(
			"删除mon节点文件失败",
			zap.Error(err),
			zap.String("path", path),
			zap.Uint32("mon_node_id", nodeID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.ErrDeleteCacheFileFailed.WithCause(err)
	}

	s.log.Info(
		"mon删除成功",
		zap.Uint32("mon_node_id", nodeID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return nil
}

func (s *MonNodeService) FindMonNodeByID(
	ctx context.Context,
	preloads []string,
	nodeID uint32,
) (*monmodel.MonNodeModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始查询mon",
		zap.Uint32("mon_node_id", nodeID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, err := s.nodeRepo.GetModel(ctx, preloads, nodeID)
	if err != nil {
		s.log.Error(
			"查询mon失败",
			zap.Error(err),
			zap.Uint32("mon_node_id", nodeID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, map[string]any{"id": nodeID})
	}

	s.log.Info(
		"查询mon成功",
		zap.Uint32("mon_node_id", nodeID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return m, nil
}

func (s *MonNodeService) ListMonNode(
	ctx context.Context,
	qp database.QueryParams,
) (int64, *[]monmodel.MonNodeModel, *errors.Error) {
	if ctx.Err() != nil {
		return 0, nil, errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始查询mon列表",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	count, ms, err := s.nodeRepo.ListModel(ctx, qp)
	if err != nil {
		s.log.Error(
			"查询mon列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}

	s.log.Info(
		"查询mon列表成功",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return count, ms, nil
}

func (s *MonNodeService) ExportMonNode(ctx context.Context, m monmodel.MonNodeModel) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始导出mon节点文件",
		zap.Object(database.ModelKey, &m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	monNode := monmodel.MonNodeVars{
		ID:          m.ID,
		Name:        m.Name,
		DeployPath:  m.DeployPath,
		OutportPath: m.OutportPath,
		JavaHome:    m.JavaHome,
		URL:         m.URL,
		HostID:      m.HostID,
	}

	path := common.GetMonNodeExportPath(m.ID)
	if _, err := serializer.WriteYAML(path, monNode); err != nil {
		s.log.Error(
			"导出mon节点文件失败",
			zap.Error(err),
			zap.String("path", path),
			zap.Object("mon_node", &monNode),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.ErrExportCacheFileFailed.WithCause(err)
	}

	s.log.Info(
		"导出mon节点文件成功",
		zap.String("path", path),
		zap.Object("mon_node", &monNode),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return nil
}
