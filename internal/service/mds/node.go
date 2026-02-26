package biz

import (
	"context"
	"path/filepath"

	"go.uber.org/zap"

	mdsmodel "gin-artweb/internal/model/mds"
	mdsrepo "gin-artweb/internal/repository/mds"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
	"gin-artweb/pkg/serializer"
)

type MdsNodeService struct {
	log      *zap.Logger
	nodeRepo *mdsrepo.MdsNodeRepo
}

func NewMdsNodeService(
	log *zap.Logger,
	nodeRepo *mdsrepo.MdsNodeRepo,
) *MdsNodeService {
	return &MdsNodeService{
		log:      log,
		nodeRepo: nodeRepo,
	}
}

func (s *MdsNodeService) CreateMdsNode(
	ctx context.Context,
	m mdsmodel.MdsNodeModel,
) (*mdsmodel.MdsNodeModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始创建mds节点",
		zap.Object(database.ModelKey, &m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	if err := s.nodeRepo.CreateModel(ctx, &m); err != nil {
		s.log.Error(
			"创建mds节点失败",
			zap.Error(err),
			zap.Object(database.ModelKey, &m),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, nil)
	}

	// 查询mds节点关联数据
	nm, rErr := s.FindMdsNodeByID(ctx, []string{"MdsColony", "Host"}, m.ID)
	if rErr != nil {
		return nil, rErr
	}

	// 导出mds节点缓存数据
	if err := s.OutPortMdsNodeData(ctx, nm); err != nil {
		return nil, err
	}

	s.log.Info(
		"创建mds节点成功",
		zap.Object(database.ModelKey, nm),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return nm, nil
}

func (s *MdsNodeService) UpdateMdsNodeByID(
	ctx context.Context,
	mdsNodeID uint32,
	data map[string]any,
) (*mdsmodel.MdsNodeModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始更新mds节点",
		zap.Uint32("mds_node_id", mdsNodeID),
		zap.Any(database.UpdateDataKey, data),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	data["id"] = mdsNodeID
	if err := s.nodeRepo.UpdateModel(ctx, data, "id = ?", mdsNodeID); err != nil {
		s.log.Error(
			"更新mds节点失败",
			zap.Error(err),
			zap.Uint32("mds_node_id", mdsNodeID),
			zap.Any(database.UpdateDataKey, data),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, data)
	}

	// 查询mds节点关联数据
	m, rErr := s.FindMdsNodeByID(ctx, []string{"MdsColony", "Host"}, mdsNodeID)
	if rErr != nil {
		return nil, rErr
	}

	// 导出mds节点缓存数据
	if err := s.OutPortMdsNodeData(ctx, m); err != nil {
		return nil, err
	}

	s.log.Info(
		"更新mds节点成功",
		zap.Uint32("mds_node_id", mdsNodeID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return m, nil
}

func (s *MdsNodeService) DeleteMdsNodeByID(
	ctx context.Context,
	mdsNodeID uint32,
) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始删除mds节点",
		zap.Uint32("mds_node_id", mdsNodeID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	if err := s.nodeRepo.DeleteModel(ctx, mdsNodeID); err != nil {
		s.log.Error(
			"删除mds节点失败",
			zap.Error(err),
			zap.Uint32("mds_node_id", mdsNodeID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.NewGormError(err, map[string]any{"id": mdsNodeID})
	}

	s.log.Info(
		"删除mds节点成功",
		zap.Uint32("mds_node_id", mdsNodeID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return nil
}

func (s *MdsNodeService) FindMdsNodeByID(
	ctx context.Context,
	preloads []string,
	mdsNodeID uint32,
) (*mdsmodel.MdsNodeModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始查询mds节点",
		zap.Strings(database.PreloadKey, preloads),
		zap.Uint32("mds_node_id", mdsNodeID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, err := s.nodeRepo.GetModel(ctx, preloads, mdsNodeID)
	if err != nil {
		s.log.Error(
			"查询mds节点失败",
			zap.Error(err),
			zap.Uint32("mds_node_id", mdsNodeID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, map[string]any{"id": mdsNodeID})
	}

	s.log.Info(
		"查询mds节点成功",
		zap.Uint32("mds_node_id", mdsNodeID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return m, nil
}

func (s *MdsNodeService) ListMdsNode(
	ctx context.Context,
	qp database.QueryParams,
) (int64, *[]mdsmodel.MdsNodeModel, *errors.Error) {
	if ctx.Err() != nil {
		return 0, nil, errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始查询mds节点列表",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	count, ms, err := s.nodeRepo.ListModel(ctx, qp)
	if err != nil {
		s.log.Error(
			"查询mds节点列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}

	s.log.Info(
		"查询mds节点列表成功",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return count, ms, nil
}

func (s *MdsNodeService) OutPortMdsNodeData(ctx context.Context, m *mdsmodel.MdsNodeModel) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始导出mds节点变量文件",
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
	mdsVars := mdsmodel.MdsNodeVars{
		ID:       m.ID,
		NodeRole: m.NodeRole,
		Specdir:  specdir,
		HostID:   m.HostID,
		IsEnable: m.IsEnable,
	}

	confDir := common.GetMdsColonyConfigDir(m.MdsColony.ColonyNum)
	mdsColonyConf := filepath.Join(confDir, specdir, "node.yaml")
	if _, err := serializer.WriteYAML(mdsColonyConf, mdsVars); err != nil {
		s.log.Error(
			"导出mds节点变量文件失败",
			zap.Error(err),
			zap.Uint32("mds_node_id", m.ID),
			zap.String("colony_num", m.MdsColony.ColonyNum),
			zap.String("path", mdsColonyConf),
			zap.Object("mds_node_vars", &mdsVars),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.ErrExportCacheFileFailed.WithCause(err)
	}

	s.log.Info(
		"导出mds节点变量文件失败",
		zap.String("path", mdsColonyConf),
		zap.Object("mds_colony_vars", &mdsVars),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return nil
}
