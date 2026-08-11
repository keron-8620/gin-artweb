package mds

import (
	"context"
	"path/filepath"
	"time"

	"go.uber.org/zap"

	mdsmodel "gin-artweb/internal/model/mds"
	mdsrepo "gin-artweb/internal/repo/mds"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
	"gin-artweb/pkg/fileutil"
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
	dto mdsmodel.MdsNodeUpsertDTO,
) (*mdsmodel.MdsNodeModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"创建mds节点:开始执行",
		zap.Object("mds_node_dto", &dto),
	)

	m := dto.ToModel()
	if err := s.nodeRepo.CreateModel(ctx, &m); err != nil {
		log.Error(
			"创建mds节点:创建数据库模型失败",
			zap.Error(err),
			zap.Object("mds_node_model", &m),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, nil)
	}

	// 查询mds节点关联数据
	node, rErr := s.FindMdsNodeByID(ctx, []string{"MdsColony", "Host"}, m.ID)
	if rErr != nil {
		log.Error(
			"创建mds节点:查询mds节点关联数据失败",
			zap.Error(rErr),
			zap.Uint32("mds_node_id", node.ID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, rErr
	}

	// 导出mds节点缓存数据
	if err := s.OutPortMdsNodeData(ctx, node); err != nil {
		log.Error(
			"创建mds节点:导出mds节点缓存数据失败",
			zap.Error(err),
			zap.Uint32("mds_node_id", node.ID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, err
	}

	log.Info(
		"创建mds节点:执行成功",
		zap.Uint32("mds_node_id", node.ID),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return node, nil
}

func (s *MdsNodeService) UpdateMdsNodeByID(
	ctx context.Context,
	mdsNodeID uint32,
	dto mdsmodel.MdsNodeUpsertDTO,
) (*mdsmodel.MdsNodeModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"更新mds节点:开始执行",
		zap.Uint32("mds_node_id", mdsNodeID),
		zap.Object("mds_node_dto", &dto),
	)

	updateData := dto.ToUpdateMap()
	if err := s.nodeRepo.UpdateModel(ctx, updateData, "id = ?", mdsNodeID); err != nil {
		log.Error(
			"更新mds节点:更新数据库模型失败",
			zap.Error(err),
			zap.Uint32("mds_node_id", mdsNodeID),
			zap.Any("update_data", updateData),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, updateData)
	}

	// 查询mds节点关联数据
	m, rErr := s.FindMdsNodeByID(ctx, []string{"MdsColony", "Host"}, mdsNodeID)
	if rErr != nil {
		log.Error(
			"更新mds节点:查询mds节点关联数据失败",
			zap.Error(rErr),
			zap.Uint32("mds_node_id", mdsNodeID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, rErr
	}

	// 导出mds节点缓存数据
	if err := s.OutPortMdsNodeData(ctx, m); err != nil {
		log.Error(
			"更新mds节点:导出mds节点缓存数据失败",
			zap.Error(err),
			zap.Uint32("mds_node_id", mdsNodeID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, err
	}

	log.Info(
		"更新mds节点:执行成功",
		zap.Uint32("mds_node_id", mdsNodeID),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return m, nil
}

func (s *MdsNodeService) DeleteMdsNodeByID(
	ctx context.Context,
	mdsNodeID uint32,
) *errors.Error {
	startTime := time.Now()
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"删除mds节点:开始执行",
		zap.Uint32("mds_node_id", mdsNodeID),
	)

	m, err := s.nodeRepo.GetModel(ctx, []string{"MdsColony"}, mdsNodeID)
	if err != nil {
		log.Error(
			"查询mds节点:查询数据库模型失败",
			zap.Error(err),
			zap.Uint32("mds_node_id", mdsNodeID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.NewGormError(err, map[string]any{"id": mdsNodeID})
	}

	if err := s.nodeRepo.DeleteModel(ctx, mdsNodeID); err != nil {
		log.Error(
			"删除mds节点:删除数据库模型失败",
			zap.Error(err),
			zap.Uint32("mds_node_id", mdsNodeID),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.NewGormError(err, map[string]any{"id": mdsNodeID})
	}

	specdir := RoleToSpecdir(m.NodeRole)
	oesNodeConf := filepath.Join(GetMdsColonyConfigDir(m.MdsColony.ColonyNum), specdir)
	if err := fileutil.Remove(ctx, oesNodeConf); err != nil {
		log.Error(
			"删除mds节点:删除节点文件失败, 请手动删除",
			zap.Error(err),
			zap.String("outport_path", oesNodeConf),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.ErrDeleteCacheFileFailed.WithCause(err)
	}

	log.Info(
		"删除mds节点:执行成功",
		zap.Uint32("mds_node_id", mdsNodeID),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

func (s *MdsNodeService) FindMdsNodeByID(
	ctx context.Context,
	preloads []string,
	mdsNodeID uint32,
) (*mdsmodel.MdsNodeModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	m, err := s.nodeRepo.GetModel(ctx, preloads, mdsNodeID)
	if err != nil {
		log.Error(
			"查询mds节点:查询数据库模型失败",
			zap.Error(err),
			zap.Uint32("mds_node_id", mdsNodeID),
			zap.Strings("preloads", preloads),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, map[string]any{"id": mdsNodeID})
	}

	log.Debug(
		"查询mds节点:查询到的数据库模型详情",
		zap.Object("mds_node_model", m),
	)
	return m, nil
}

func (s *MdsNodeService) ListMdsNode(
	ctx context.Context,
	dto *mdsmodel.ListMdsNodeDTO,
) (int, int, int64, []mdsmodel.MdsNodeModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return 0, 0, 0, nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Debug(
		"查询mds节点列表:入参详情",
		zap.Object("mds_node_dto", dto),
	)

	page, size := dto.StandardModelQuery.GetPageParam()
	limit, offset := common.Page2LimitOffset(page, size)
	qp := database.QueryParams{
		Preloads: []string{"Host", "MdsColony"},
		OrderBy:  []string{"id DESC"},
		Limit:    limit,
		Offset:   offset,
		Query:    dto.ToQueryMap(),
	}

	count, err := s.nodeRepo.CountModel(ctx, qp.Query)
	if err != nil {
		log.Error(
			"查询mds节点列表:查询数据库模型总数失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return page, size, 0, nil, errors.NewGormError(err, nil)
	}

	if count == 0 {
		log.Warn(
			"查询mds节点列表:数据库模型总数为0",
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return page, size, 0, nil, nil
	}

	ms, err := s.nodeRepo.ListModel(ctx, qp)
	if err != nil {
		log.Error(
			"查询mds节点列表:查询数据库模型失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return page, size, 0, nil, errors.NewGormError(err, nil)
	}
	return page, size, count, ms, nil
}

func (s *MdsNodeService) OutPortMdsNodeData(
	ctx context.Context,
	m *mdsmodel.MdsNodeModel,
) *errors.Error {
	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"导出mds节点变量文件:开始执行",
		zap.Object("mds_node_model", m),
	)

	specdir := RoleToSpecdir(m.NodeRole)
	mdsNodeConfFile := filepath.Join(GetMdsColonyConfigDir(m.MdsColony.ColonyNum), specdir, "node.yaml")

	mdsVars := mdsmodel.MdsNodeVars{
		ID:       m.ID,
		NodeRole: m.NodeRole,
		Specdir:  specdir,
		HostID:   m.HostID,
		IsEnable: m.IsEnable,
	}

	if _, err := serializer.WriteYAML(mdsNodeConfFile, mdsVars); err != nil {
		log.Error(
			"导出mds节点变量文件:写入文件失败",
			zap.Error(err),
			zap.Uint32("mds_node_id", m.ID),
			zap.String("colony_num", m.MdsColony.ColonyNum),
			zap.String("path", mdsNodeConfFile),
			zap.Object("mds_colony_vars", &mdsVars),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.ErrExportCacheFileFailed.WithCause(err)
	}

	log.Info(
		"导出mds节点变量文件:执行成功",
		zap.String("path", mdsNodeConfFile),
		zap.Object("mds_colony_vars", &mdsVars),
		zap.Duration("total_duration", time.Since(startTime)),
	)
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
