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
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info("创建mds节点：开始执行")

	log.Debug(
		"创建mds节点：入参详情",
		zap.Object("mds_node_dto", &dto),
	)

	m := mdsmodel.MdsNodeModel{
		NodeRole:    dto.NodeRole,
		IsEnable:    dto.IsEnable,
		MdsColonyID: dto.MdsColonyID,
		HostID:      dto.HostID,
	}

	createStepStart := time.Now()
	log.Debug(
		"创建mds节点：开始创建数据库模型",
		zap.Object("mds_node_model", &m),
	)
	if err := s.nodeRepo.CreateModel(ctx, &m); err != nil {
		log.Error(
			"创建mds节点：创建数据库模型失败",
			zap.Error(err),
			zap.Object("mds_node_model", &m),
			zap.Duration("create_step_duration", time.Since(createStepStart)),
		)
		return nil, errors.NewGormError(err, nil)
	}
	createStepDuration := time.Since(createStepStart)
	log.Debug(
		"创建mds节点：创建数据库模型成功",
		zap.Uint32("mds_node_id", m.ID),
		zap.Duration("create_step_duration", createStepDuration),
	)

	queryStepStart := time.Now()
	log.Debug(
		"创建mds节点：开始查询mds节点关联数据",
		zap.Uint32("mds_node_id", m.ID),
		zap.Strings("preloads", []string{"MdsColony", "Host"}),
	)
	// 查询mds节点关联数据
	nm, rErr := s.FindMdsNodeByID(ctx, []string{"MdsColony", "Host"}, m.ID)
	if rErr != nil {
		log.Error(
			"创建mds节点：查询mds节点关联数据失败",
			zap.Error(rErr),
			zap.Uint32("mds_node_id", m.ID),
			zap.Duration("query_step_duration", time.Since(queryStepStart)),
		)
		return nil, rErr
	}
	queryStepDuration := time.Since(queryStepStart)
	log.Debug(
		"创建mds节点：查询mds节点关联数据成功",
		zap.Uint32("mds_node_id", m.ID),
		zap.Duration("query_step_duration", queryStepDuration),
	)

	exportStepStart := time.Now()
	log.Debug(
		"创建mds节点：开始导出mds节点缓存数据",
		zap.Uint32("mds_node_id", m.ID),
	)
	// 导出mds节点缓存数据
	if err := s.OutPortMdsNodeData(ctx, nm); err != nil {
		log.Error(
			"创建mds节点：导出mds节点缓存数据失败",
			zap.Error(err),
			zap.Uint32("mds_node_id", m.ID),
			zap.Duration("export_step_duration", time.Since(exportStepStart)),
		)
		return nil, err
	}
	exportStepDuration := time.Since(exportStepStart)
	log.Debug(
		"创建mds节点：导出mds节点缓存数据成功",
		zap.Uint32("mds_node_id", m.ID),
		zap.Duration("export_step_duration", exportStepDuration),
	)

	log.Info(
		"创建mds节点：执行成功",
		zap.Uint32("mds_node_id", m.ID),
		zap.Duration("create_step_duration", createStepDuration),
		zap.Duration("query_step_duration", queryStepDuration),
		zap.Duration("export_step_duration", exportStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nm, nil
}

func (s *MdsNodeService) UpdateMdsNodeByID(
	ctx context.Context,
	mdsNodeID uint32,
	dto mdsmodel.MdsNodeUpsertDTO,
) (*mdsmodel.MdsNodeModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"更新mds节点：开始执行",
		zap.Uint32("mds_node_id", mdsNodeID),
	)

	log.Debug(
		"更新mds节点：入参详情",
		zap.Uint32("mds_node_id", mdsNodeID),
		zap.Object("mds_node_dto", &dto),
	)

	updateData := dto.ToUpdateMap()
	log.Debug(
		"更新mds节点：转换为数据库更新参数",
		zap.Uint32("mds_node_id", mdsNodeID),
		zap.Any("update_data", updateData),
	)

	updateStepStart := time.Now()
	log.Debug(
		"更新mds节点：开始更新数据库模型",
		zap.Uint32("mds_node_id", mdsNodeID),
		zap.Any("update_data", updateData),
	)
	if err := s.nodeRepo.UpdateModel(ctx, updateData, "id = ?", mdsNodeID); err != nil {
		log.Error(
			"更新mds节点：更新数据库模型失败",
			zap.Error(err),
			zap.Uint32("mds_node_id", mdsNodeID),
			zap.Any("update_data", updateData),
			zap.Duration("update_step_duration", time.Since(updateStepStart)),
		)
		return nil, errors.NewGormError(err, updateData)
	}
	updateStepDuration := time.Since(updateStepStart)
	log.Debug(
		"更新mds节点：更新数据库模型成功",
		zap.Uint32("mds_node_id", mdsNodeID),
		zap.Duration("update_step_duration", updateStepDuration),
	)

	queryStepStart := time.Now()
	log.Debug(
		"更新mds节点：开始查询mds节点关联数据",
		zap.Uint32("mds_node_id", mdsNodeID),
		zap.Strings("preloads", []string{"MdsColony", "Host"}),
	)
	// 查询mds节点关联数据
	m, rErr := s.FindMdsNodeByID(ctx, []string{"MdsColony", "Host"}, mdsNodeID)
	if rErr != nil {
		log.Error(
			"更新mds节点：查询mds节点关联数据失败",
			zap.Error(rErr),
			zap.Uint32("mds_node_id", mdsNodeID),
			zap.Duration("query_step_duration", time.Since(queryStepStart)),
		)
		return nil, rErr
	}
	queryStepDuration := time.Since(queryStepStart)
	log.Debug(
		"更新mds节点：查询mds节点关联数据成功",
		zap.Uint32("mds_node_id", mdsNodeID),
		zap.Duration("query_step_duration", queryStepDuration),
	)

	exportStepStart := time.Now()
	log.Debug(
		"更新mds节点：开始导出mds节点缓存数据",
		zap.Uint32("mds_node_id", mdsNodeID),
	)
	// 导出mds节点缓存数据
	if err := s.OutPortMdsNodeData(ctx, m); err != nil {
		log.Error(
			"更新mds节点：导出mds节点缓存数据失败",
			zap.Error(err),
			zap.Uint32("mds_node_id", mdsNodeID),
			zap.Duration("export_step_duration", time.Since(exportStepStart)),
		)
		return nil, err
	}
	exportStepDuration := time.Since(exportStepStart)
	log.Debug(
		"更新mds节点：导出mds节点缓存数据成功",
		zap.Uint32("mds_node_id", mdsNodeID),
		zap.Duration("export_step_duration", exportStepDuration),
	)

	log.Info(
		"更新mds节点：执行成功",
		zap.Uint32("mds_node_id", mdsNodeID),
		zap.Duration("update_step_duration", updateStepDuration),
		zap.Duration("query_step_duration", queryStepDuration),
		zap.Duration("export_step_duration", exportStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
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

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"删除mds节点：开始执行",
		zap.Uint32("mds_node_id", mdsNodeID),
	)

	deleteStepStart := time.Now()
	log.Debug(
		"删除mds节点：开始删除数据库模型",
		zap.Uint32("mds_node_id", mdsNodeID),
	)
	if err := s.nodeRepo.DeleteModel(ctx, mdsNodeID); err != nil {
		log.Error(
			"删除mds节点：删除数据库模型失败",
			zap.Error(err),
			zap.Uint32("mds_node_id", mdsNodeID),
			zap.Duration("delete_step_duration", time.Since(deleteStepStart)),
		)
		return errors.NewGormError(err, map[string]any{"id": mdsNodeID})
	}
	deleteStepDuration := time.Since(deleteStepStart)
	log.Debug(
		"删除mds节点：删除数据库模型成功",
		zap.Uint32("mds_node_id", mdsNodeID),
		zap.Duration("delete_step_duration", deleteStepDuration),
	)

	log.Info(
		"删除mds节点：执行成功",
		zap.Uint32("mds_node_id", mdsNodeID),
		zap.Duration("delete_step_duration", deleteStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
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

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"查询mds节点：开始执行",
		zap.Uint32("mds_node_id", mdsNodeID),
		zap.Strings("preloads", preloads),
	)

	queryStepStart := time.Now()
	log.Debug(
		"查询mds节点：开始查询数据库模型",
		zap.Uint32("mds_node_id", mdsNodeID),
		zap.Strings("preloads", preloads),
	)
	m, err := s.nodeRepo.GetModel(ctx, preloads, mdsNodeID)
	if err != nil {
		log.Error(
			"查询mds节点：查询数据库模型失败",
			zap.Error(err),
			zap.Uint32("mds_node_id", mdsNodeID),
			zap.Strings("preloads", preloads),
			zap.Duration("query_step_duration", time.Since(queryStepStart)),
		)
		return nil, errors.NewGormError(err, map[string]any{"id": mdsNodeID})
	}
	queryStepDuration := time.Since(queryStepStart)
	log.Debug(
		"查询mds节点：查询到的数据库模型详情",
		zap.Object("mds_node_model", m),
	)

	log.Info(
		"查询mds节点：执行成功",
		zap.Uint32("mds_node_id", mdsNodeID),
		zap.Duration("query_step_duration", queryStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return m, nil
}

func (s *MdsNodeService) ListMdsNode(
	ctx context.Context,
	page, size int,
	dto *mdsmodel.ListMdsNodeDTO,
) (int64, []mdsmodel.MdsNodeModel, *errors.Error) {
	if ctx.Err() != nil {
		return 0, nil, errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info("查询mds节点列表：开始执行")

	log.Debug(
		"查询mds节点列表：入参详情",
		zap.Object("mds_node_dto", dto),
	)

	limit, offset := common.Page2LimitOffset(page, size)
	qp := database.QueryParams{
		Preloads: []string{"Host", "MdsColony"},
		OrderBy:  []string{"id DESC"},
		Limit:    limit,
		Offset:   offset,
		Query:    dto.ToQueryMap(),
	}
	log.Debug(
		"查询mds节点列表：查询参数",
		zap.Object("query_params", &qp),
	)

	countStepStart := time.Now()
	log.Debug(
		"查询mds节点列表：开始查询数据库模型总数",
		zap.Object("query_params", &qp),
	)

	count, err := s.nodeRepo.CountModel(ctx, qp.Query)
	countStepDuration := time.Since(countStepStart)
	if err != nil {
		log.Error(
			"查询mds节点列表：查询数据库模型总数失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("count_step_duration", countStepDuration),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}
	log.Debug(
		"查询mds节点列表：查询数据库模型总数成功",
		zap.Int64("total_count", count),
		zap.Duration("count_step_duration", countStepDuration),
	)
	if count == 0 {
		log.Warn(
			"查询mds节点列表：数据库模型总数为0",
			zap.Duration("total_duration", time.Since(countStepStart)),
		)
		return count, nil, nil
	}

	listStepStart := time.Now()
	log.Debug(
		"查询mds节点列表：开始查询数据库模型",
		zap.Object("query_params", &qp),
	)
	ms, err := s.nodeRepo.ListModel(ctx, qp)
	listStepDuration := time.Since(listStepStart)
	if err != nil {
		log.Error(
			"查询mds节点列表：查询数据库模型失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("list_step_duration", listStepDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}
	log.Debug(
		"查询mds节点列表：查询数据库模型成功",
		zap.Int64("total_count", count),
		zap.Duration("list_step_duration", listStepDuration),
	)

	log.Info(
		"查询mds节点列表：执行成功",
		zap.Object("query_params", &qp),
		zap.Duration("list_step_duration", listStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return count, ms, nil
}

func (s *MdsNodeService) OutPortMdsNodeData(ctx context.Context, m *mdsmodel.MdsNodeModel) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"导出mds节点变量文件：开始执行",
		zap.Object("mds_node_model", m),
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

	exportStepStart := time.Now()
	log.Debug(
		"导出mds节点变量文件：开始写入文件",
		zap.String("path", mdsColonyConf),
		zap.Object("mds_colony_vars", &mdsVars),
	)
	if _, err := serializer.WriteYAML(mdsColonyConf, mdsVars); err != nil {
		log.Error(
			"导出mds节点变量文件：写入文件失败",
			zap.Error(err),
			zap.Uint32("mds_node_id", m.ID),
			zap.String("colony_num", m.MdsColony.ColonyNum),
			zap.String("path", mdsColonyConf),
			zap.Object("mds_colony_vars", &mdsVars),
			zap.Duration("export_step_duration", time.Since(exportStepStart)),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.ErrExportCacheFileFailed.WithCause(err)
	}
	exportStepDuration := time.Since(exportStepStart)
	log.Debug(
		"导出mds节点变量文件：写入文件成功",
		zap.String("path", mdsColonyConf),
		zap.Duration("export_step_duration", exportStepDuration),
	)

	log.Info(
		"导出mds节点变量文件：执行成功",
		zap.String("path", mdsColonyConf),
		zap.Object("mds_colony_vars", &mdsVars),
		zap.Duration("export_step_duration", exportStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}
