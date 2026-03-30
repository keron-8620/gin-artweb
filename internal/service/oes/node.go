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
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info("创建oes节点：开始执行")

	log.Debug(
		"创建oes节点：入参详情",
		zap.Object("oes_node_dto", &dto),
	)

	m := oesmodel.OesNodeModel{
		NodeRole:    dto.NodeRole,
		IsEnable:    dto.IsEnable,
		OesColonyID: dto.OesColonyID,
		HostID:      dto.HostID,
	}

	createStepStart := time.Now()
	log.Debug(
		"创建oes节点：开始创建数据库模型",
		zap.Object("oes_node", &m),
	)
	if err := s.nodeRepo.CreateModel(ctx, &m); err != nil {
		log.Error(
			"创建oes节点：创建数据库模型失败",
			zap.Error(err),
			zap.Object("oes_node", &m),
			zap.Duration("create_step_duration", time.Since(createStepStart)),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, nil)
	}
	createStepDuration := time.Since(createStepStart)
	log.Debug(
		"创建oes节点：创建数据库模型成功",
		zap.Uint32("oes_node_id", m.ID),
		zap.Duration("create_step_duration", createStepDuration),
	)

	// 查询oes节点关联数据
	queryStepStart := time.Now()
	log.Debug(
		"创建oes节点：开始查询关联数据",
		zap.Uint32("oes_node_id", m.ID),
	)
	nm, rErr := s.FindOesNodeByID(ctx, []string{"OesColony", "Host"}, m.ID)
	queryStepDuration := time.Since(queryStepStart)
	if rErr != nil {
		log.Error(
			"创建oes节点：查询关联数据失败",
			zap.Error(rErr),
			zap.Uint32("oes_node_id", m.ID),
			zap.Duration("query_step_duration", queryStepDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, rErr
	}
	log.Debug(
		"创建oes节点：查询关联数据成功",
		zap.Uint32("oes_node_id", m.ID),
		zap.Duration("query_step_duration", queryStepDuration),
	)

	// 导出oes节点缓存数据
	exportStepStart := time.Now()
	log.Debug(
		"创建oes节点：开始导出缓存数据",
		zap.Uint32("oes_node_id", m.ID),
	)
	if err := s.OutPortOesNodeData(ctx, nm); err != nil {
		log.Error(
			"创建oes节点：导出缓存数据失败",
			zap.Error(err),
			zap.Uint32("oes_node_id", m.ID),
			zap.Duration("export_step_duration", time.Since(exportStepStart)),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, err
	}
	exportStepDuration := time.Since(exportStepStart)
	log.Debug(
		"创建oes节点：导出缓存数据成功",
		zap.Uint32("oes_node_id", m.ID),
		zap.Duration("export_step_duration", exportStepDuration),
	)

	log.Info(
		"创建oes节点：执行成功",
		zap.Uint32("oes_node_id", m.ID),
		zap.Duration("create_step_duration", createStepDuration),
		zap.Duration("query_step_duration", queryStepDuration),
		zap.Duration("export_step_duration", exportStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nm, nil
}

func (s *OesNodeService) UpdateOesNodeByID(
	ctx context.Context,
	oesNodeID uint32,
	dto oesmodel.OesNodeUpsertDTO,
) (*oesmodel.OesNodeModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"更新oes节点：开始执行",
		zap.Uint32("oes_node_id", oesNodeID),
	)

	log.Debug(
		"更新oes节点：入参详情",
		zap.Object("oes_node_dto", &dto),
	)

	updateData := dto.ToUpdateMap()
	log.Debug(
		"更新oes节点：转换为数据库更新参数",
		zap.Uint32("oes_node_id", oesNodeID),
		zap.Any("update_data", updateData),
	)

	updateStepStart := time.Now()
	log.Debug(
		"更新oes节点：开始更新数据库模型",
		zap.Uint32("oes_node_id", oesNodeID),
		zap.Any("update_data", updateData),
	)
	if err := s.nodeRepo.UpdateModel(ctx, updateData, "id = ?", oesNodeID); err != nil {
		log.Error(
			"更新oes节点：更新数据库模型失败",
			zap.Error(err),
			zap.Uint32("oes_node_id", oesNodeID),
			zap.Any("update_data", updateData),
			zap.Duration("update_step_duration", time.Since(updateStepStart)),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, updateData)
	}
	updateStepDuration := time.Since(updateStepStart)
	log.Debug(
		"更新oes节点：更新数据库模型成功",
		zap.Uint32("oes_node_id", oesNodeID),
		zap.Duration("update_step_duration", updateStepDuration),
	)

	// 查询oes节点关联数据
	queryStepStart := time.Now()
	log.Debug(
		"更新oes节点：开始查询关联数据",
		zap.Uint32("oes_node_id", oesNodeID),
	)
	m, rErr := s.FindOesNodeByID(ctx, []string{"OesColony", "Host"}, oesNodeID)
	queryStepDuration := time.Since(queryStepStart)
	if rErr != nil {
		log.Error(
			"更新oes节点：查询关联数据失败",
			zap.Error(rErr),
			zap.Uint32("oes_node_id", oesNodeID),
			zap.Duration("query_step_duration", queryStepDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, rErr
	}
	log.Debug(
		"更新oes节点：查询关联数据成功",
		zap.Uint32("oes_node_id", oesNodeID),
		zap.Duration("query_step_duration", queryStepDuration),
	)

	// 导出oes节点缓存数据
	exportStepStart := time.Now()
	log.Debug(
		"更新oes节点：开始导出缓存数据",
		zap.Uint32("oes_node_id", oesNodeID),
	)
	if err := s.OutPortOesNodeData(ctx, m); err != nil {
		log.Error(
			"更新oes节点：导出缓存数据失败",
			zap.Error(err),
			zap.Uint32("oes_node_id", oesNodeID),
			zap.Duration("export_step_duration", time.Since(exportStepStart)),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, err
	}
	exportStepDuration := time.Since(exportStepStart)
	log.Debug(
		"更新oes节点：导出缓存数据成功",
		zap.Uint32("oes_node_id", oesNodeID),
		zap.Duration("export_step_duration", exportStepDuration),
	)

	log.Info(
		"更新oes节点：执行成功",
		zap.Uint32("oes_node_id", oesNodeID),
		zap.Duration("update_step_duration", updateStepDuration),
		zap.Duration("query_step_duration", queryStepDuration),
		zap.Duration("export_step_duration", exportStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
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

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"删除oes节点：开始执行",
		zap.Uint32("oes_node_id", oesNodeID),
	)

	deleteStepStart := time.Now()
	log.Debug(
		"删除oes节点：开始删除数据库模型",
		zap.Uint32("oes_node_id", oesNodeID),
	)
	if err := s.nodeRepo.DeleteModel(ctx, oesNodeID); err != nil {
		log.Error(
			"删除oes节点：删除数据库模型失败",
			zap.Error(err),
			zap.Uint32("oes_node_id", oesNodeID),
			zap.Duration("delete_step_duration", time.Since(deleteStepStart)),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.NewGormError(err, map[string]any{"id": oesNodeID})
	}
	deleteStepDuration := time.Since(deleteStepStart)
	log.Debug(
		"删除oes节点：删除数据库模型成功",
		zap.Uint32("oes_node_id", oesNodeID),
		zap.Duration("delete_step_duration", deleteStepDuration),
	)

	log.Info(
		"删除oes节点：执行成功",
		zap.Uint32("oes_node_id", oesNodeID),
		zap.Duration("delete_step_duration", deleteStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
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

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"查询oes节点：开始执行",
		zap.Strings("preloads", preloads),
		zap.Uint32("oes_node_id", oesNodeID),
	)

	m, err := s.nodeRepo.GetModel(ctx, preloads, oesNodeID)
	if err != nil {
		log.Error(
			"查询oes节点：查询数据库模型失败",
			zap.Error(err),
			zap.Uint32("oes_node_id", oesNodeID),
			zap.Strings("preloads", preloads),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, map[string]any{"id": oesNodeID})
	}
	log.Debug(
		"查询oes节点：查询到的数据库模型详情",
		zap.Object("oes_node_model", m),
	)

	log.Info(
		"查询oes节点：执行成功",
		zap.Uint32("oes_node_id", oesNodeID),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return m, nil
}

func (s *OesNodeService) ListOesNode(
	ctx context.Context,
	page, size int,
	dto oesmodel.ListOesNodeDTO,
) (int64, []oesmodel.OesNodeModel, *errors.Error) {
	if ctx.Err() != nil {
		return 0, nil, errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info("查询oes节点列表：开始执行")

	log.Debug(
		"查询oes节点列表：入参详情",
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
	log.Debug(
		"查询oes节点列表：查询参数",
		zap.Object("query_params", &qp),
	)

	countStepStart := time.Now()
	log.Debug(
		"查询oes节点列表：开始查询数据库模型总数",
		zap.Object("query_params", &qp),
	)
	count, err := s.nodeRepo.CountModel(ctx, qp.Query)
	countStepDuration := time.Since(countStepStart)
	if err != nil {
		log.Error(
			"查询oes节点列表：查询数据库模型总数失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("count_step_duration", countStepDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}
	log.Debug(
		"查询oes节点列表：查询数据库模型总数成功",
		zap.Int64("total_count", count),
		zap.Duration("count_step_duration", countStepDuration),
	)
	if count == 0 {
		log.Warn(
			"查询oes节点列表：数据库模型总数为0",
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return count, nil, nil
	}
	listStepStart := time.Now()
	log.Debug(
		"查询oes节点列表：开始查询数据库模型列表",
		zap.Object("query_params", &qp),
	)
	ms, err := s.nodeRepo.ListModel(ctx, qp)
	listStepDuration := time.Since(listStepStart)
	if err != nil {
		log.Error(
			"查询oes节点列表：查询数据库模型列表失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("list_step_duration", listStepDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}
	log.Debug(
		"查询oes节点列表：查询数据库模型列表成功",
		zap.Int("node_count", len(ms)),
		zap.Duration("list_step_duration", listStepDuration),
	)

	log.Info(
		"查询oes节点列表：执行成功",
		zap.Duration("count_step_duration", countStepDuration),
		zap.Duration("list_step_duration", listStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return count, ms, nil
}

func (s *OesNodeService) OutPortOesNodeData(
	ctx context.Context,
	m *oesmodel.OesNodeModel,
) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"导出oes节点变量文件：开始执行",
		zap.Object("oes_node", m),
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
	confDir := GetOesColonyConfigDir(m.OesColony.ColonyNum)
	oesColonyConf := filepath.Join(confDir, specdir, "node.yaml")

	exportStepStart := time.Now()
	log.Debug(
		"导出oes节点变量文件：开始写入文件",
		zap.String("path", oesColonyConf),
		zap.Object("oes_node_vars", &oesVars),
	)
	if _, err := serializer.WriteYAML(oesColonyConf, oesVars); err != nil {
		log.Error(
			"导出oes节点变量文件：写入文件失败",
			zap.Error(err),
			zap.Uint32("oes_node_id", m.ID),
			zap.String("colony_num", m.OesColony.ColonyNum),
			zap.String("path", oesColonyConf),
			zap.Object("oes_node_vars", &oesVars),
			zap.Duration("export_step_duration", time.Since(exportStepStart)),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.ErrExportCacheFileFailed.WithCause(err)
	}
	exportStepDuration := time.Since(exportStepStart)
	log.Debug(
		"导出oes节点变量文件：写入文件成功",
		zap.String("path", oesColonyConf),
		zap.Duration("export_step_duration", exportStepDuration),
	)

	log.Info(
		"导出oes节点变量文件：执行成功",
		zap.String("path", oesColonyConf),
		zap.Object("oes_node_vars", &oesVars),
		zap.Duration("export_step_duration", exportStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}
