package mon

import (
	"context"
	"os"
	"time"

	"go.uber.org/zap"

	monmodel "gin-artweb/internal/model/mon"
	monrepo "gin-artweb/internal/repo/mon"
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
	dto monmodel.MonNodeUpsertDTO,
) (*monmodel.MonNodeModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info("创建mon节点：开始执行")

	log.Debug(
		"创建mon节点：入参详情",
		zap.Object("mon_node_dto", &dto),
	)

	m := monmodel.MonNodeModel{
		Name:        dto.Name,
		DeployPath:  dto.DeployPath,
		OutportPath: dto.OutportPath,
		JavaHome:    dto.JavaHome,
		URL:         dto.URL,
		HostID:      dto.HostID,
	}

	createStepStart := time.Now()
	log.Debug(
		"创建mon节点：开始创建数据库模型",
		zap.Object("mon_node_model", &m),
	)
	if err := s.nodeRepo.CreateModel(ctx, &m); err != nil {
		log.Error(
			"创建mon节点：创建数据库模型失败",
			zap.Error(err),
			zap.Object("mon_node_model", &m),
			zap.Duration("create_step_duration", time.Since(createStepStart)),
		)
		return nil, errors.NewGormError(err, nil)
	}
	createStepDuration := time.Since(createStepStart)
	log.Debug(
		"创建mon节点：创建数据库模型成功",
		zap.Object("mon_node_model", &m),
		zap.Duration("create_step_duration", createStepDuration),
	)

	exportStepStart := time.Now()
	log.Debug(
		"创建mon节点：开始导出节点文件",
		zap.Uint32("mon_node_id", m.ID),
	)
	if err := s.ExportMonNode(ctx, m); err != nil {
		log.Error(
			"创建mon节点：导出节点文件失败",
			zap.Error(err),
			zap.Uint32("mon_node_id", m.ID),
			zap.Duration("export_step_duration", time.Since(exportStepStart)),
		)
		return nil, err
	}
	exportStepDuration := time.Since(exportStepStart)
	log.Debug(
		"创建mon节点：导出节点文件成功",
		zap.Uint32("mon_node_id", m.ID),
		zap.Duration("export_step_duration", exportStepDuration),
	)

	log.Info(
		"创建mon节点：执行成功",
		zap.Uint32("mon_node_id", m.ID),
		zap.Duration("create_step_duration", createStepDuration),
		zap.Duration("export_step_duration", exportStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return s.FindMonNodeByID(ctx, []string{"Host"}, m.ID)
}

func (s *MonNodeService) UpdateMonNodeByID(
	ctx context.Context,
	nodeID uint32,
	dto monmodel.MonNodeUpsertDTO,
) (*monmodel.MonNodeModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"更新mon节点：开始执行",
		zap.Uint32("mon_node_id", nodeID),
	)

	log.Debug(
		"更新mon节点：入参详情",
		zap.Uint32("mon_node_id", nodeID),
		zap.Object("mon_node_dto", &dto),
	)

	updateData := dto.ToUpdateMap()
	log.Debug(
		"更新mon节点：转换为数据库更新参数",
		zap.Uint32("mon_node_id", nodeID),
		zap.Any("update_data", updateData),
	)

	updateStepStart := time.Now()
	log.Debug(
		"更新mon节点：开始更新数据库模型",
		zap.Uint32("mon_node_id", nodeID),
		zap.Any("update_data", updateData),
	)
	if err := s.nodeRepo.UpdateModel(ctx, updateData, "id = ?", nodeID); err != nil {
		log.Error(
			"更新mon节点：更新数据库模型失败",
			zap.Error(err),
			zap.Uint32("mon_node_id", nodeID),
			zap.Any("update_data", updateData),
			zap.Duration("update_step_duration", time.Since(updateStepStart)),
		)
		return nil, errors.NewGormError(err, nil)
	}
	updateStepDuration := time.Since(updateStepStart)
	log.Debug(
		"更新mon节点：更新数据库模型成功",
		zap.Uint32("mon_node_id", nodeID),
		zap.Duration("update_step_duration", updateStepDuration),
	)

	preloads := []string{"Host"}
	log.Debug(
		"更新mon节点：开始查询更新后的mon节点",
		zap.Uint32("mon_node_id", nodeID),
		zap.Strings("preloads", preloads),
	)
	m, rErr := s.FindMonNodeByID(ctx, preloads, nodeID)
	if rErr != nil {
		log.Error(
			"更新mon节点：查询更新后的mon节点失败",
			zap.Error(rErr),
			zap.Uint32("mon_node_id", nodeID),
			zap.Strings("preloads", preloads),
		)
		return nil, rErr
	}

	exportStepStart := time.Now()
	log.Debug(
		"更新mon节点：开始导出节点文件",
		zap.Uint32("mon_node_id", nodeID),
	)
	if rErr := s.ExportMonNode(ctx, *m); rErr != nil {
		log.Error(
			"更新mon节点：导出节点文件失败",
			zap.Error(rErr),
			zap.Uint32("mon_node_id", nodeID),
			zap.Duration("export_step_duration", time.Since(exportStepStart)),
		)
		return nil, rErr
	}
	exportStepDuration := time.Since(exportStepStart)
	log.Debug(
		"更新mon节点：导出节点文件成功",
		zap.Uint32("mon_node_id", nodeID),
		zap.Duration("export_step_duration", exportStepDuration),
	)

	log.Info(
		"更新mon节点：执行成功",
		zap.Uint32("mon_node_id", nodeID),
		zap.Duration("update_step_duration", updateStepDuration),
		zap.Duration("export_step_duration", exportStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
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

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"删除mon节点：开始执行",
		zap.Uint32("mon_node_id", nodeID),
	)

	deleteStepStart := time.Now()
	log.Debug(
		"删除mon节点：开始删除数据库模型",
		zap.Uint32("mon_node_id", nodeID),
	)
	if err := s.nodeRepo.DeleteModel(ctx, nodeID); err != nil {
		log.Error(
			"删除mon节点：删除数据库模型失败",
			zap.Error(err),
			zap.Uint32("mon_node_id", nodeID),
			zap.Duration("delete_step_duration", time.Since(deleteStepStart)),
		)
		return errors.NewGormError(err, map[string]any{"id": nodeID})
	}
	deleteStepDuration := time.Since(deleteStepStart)
	log.Debug(
		"删除mon节点：删除数据库模型成功",
		zap.Uint32("mon_node_id", nodeID),
		zap.Duration("delete_step_duration", deleteStepDuration),
	)

	removeStepStart := time.Now()
	path := common.GetMonNodeExportPath(nodeID)
	log.Debug(
		"删除mon节点：开始删除节点文件",
		zap.Uint32("mon_node_id", nodeID),
		zap.String("path", path),
	)
	if err := os.RemoveAll(path); err != nil && !os.IsNotExist(err) {
		log.Error(
			"删除mon节点：删除节点文件失败",
			zap.Error(err),
			zap.String("path", path),
			zap.Uint32("mon_node_id", nodeID),
			zap.Duration("remove_step_duration", time.Since(removeStepStart)),
		)
		return errors.ErrDeleteCacheFileFailed.WithCause(err)
	}
	removeStepDuration := time.Since(removeStepStart)
	log.Debug(
		"删除mon节点：删除节点文件成功",
		zap.Uint32("mon_node_id", nodeID),
		zap.String("path", path),
		zap.Duration("remove_step_duration", removeStepDuration),
	)

	log.Info(
		"删除mon节点：执行成功",
		zap.Uint32("mon_node_id", nodeID),
		zap.Duration("delete_step_duration", deleteStepDuration),
		zap.Duration("remove_step_duration", removeStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
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

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"查询mon节点：开始执行",
		zap.Uint32("mon_node_id", nodeID),
		zap.Strings("preloads", preloads),
	)

	m, err := s.nodeRepo.GetModel(ctx, preloads, nodeID)
	if err != nil {
		log.Error(
			"查询mon节点：查询数据库模型失败",
			zap.Error(err),
			zap.Uint32("mon_node_id", nodeID),
			zap.Strings("preloads", preloads),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, map[string]any{"id": nodeID})
	}
	log.Debug(
		"查询mon节点：查询到的数据库模型详情",
		zap.Object("mon_node_model", m),
	)

	log.Info(
		"查询mon节点：执行成功",
		zap.Uint32("mon_node_id", nodeID),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return m, nil
}

func (s *MonNodeService) ListMonNode(
	ctx context.Context,
	page, size int,
	dto monmodel.ListMonNodeDTO,
) (int64, []monmodel.MonNodeModel, *errors.Error) {
	if ctx.Err() != nil {
		return 0, nil, errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info("查询mon节点列表：开始执行")

	log.Debug(
		"查询mon节点列表：参数详情",
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

	log.Debug(
		"查询mon节点列表：查询数据库模型参数",
		zap.Object("query_params", &qp),
	)

	countStepStart := time.Now()
	log.Debug(
		"查询mon节点列表：开始查询数据库模型总数",
		zap.Any("query", qp.Query),
	)
	count, err := s.nodeRepo.CountModel(ctx, qp.Query)
	if err != nil {
		log.Error(
			"查询mon节点列表：查询数据库模型总数失败",
			zap.Error(err),
			zap.Any("query", qp.Query),
			zap.Duration("count_step_duration", time.Since(countStepStart)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}
	countStepDuration := time.Since(countStepStart)
	log.Debug(
		"查询mon节点列表：查询数据库模型总数成功",
		zap.Any("query", qp.Query),
		zap.Duration("count_step_duration", countStepDuration),
	)
	if count == 0 {
		return count, nil, nil
	}

	listStepStart := time.Now()
	log.Debug(
		"查询mon节点列表：开始查询数据库模型",
		zap.Any("query", qp.Query),
	)
	ms, err := s.nodeRepo.ListModel(ctx, qp)
	if err != nil {
		log.Error(
			"查询mon节点列表：查询数据库模型失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("list_step_duration", time.Since(listStepStart)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}
	listStepDuration := time.Since(listStepStart)
	log.Info(
		"查询mon节点列表：执行成功",
		zap.Int64("total_count", count),
		zap.Duration("count_step_duration", countStepDuration),
		zap.Duration("list_step_duration", listStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return count, ms, nil
}

func (s *MonNodeService) ExportMonNode(ctx context.Context, m monmodel.MonNodeModel) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"导出mon节点文件：开始执行",
		zap.Uint32("mon_node_id", m.ID),
	)

	log.Debug(
		"导出mon节点文件：入参详情",
		zap.Object("mon_node_model", &m),
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
	log.Debug(
		"导出mon节点文件：转换为导出格式",
		zap.Object("mon_node_vars", &monNode),
	)

	exportStepStart := time.Now()
	path := common.GetMonNodeExportPath(m.ID)
	log.Debug(
		"导出mon节点文件：开始写入文件",
		zap.String("path", path),
		zap.Object("mon_node_vars", &monNode),
	)
	if _, err := serializer.WriteYAML(path, monNode); err != nil {
		log.Error(
			"导出mon节点文件：写入文件失败",
			zap.Error(err),
			zap.String("path", path),
			zap.Object("mon_node_vars", &monNode),
			zap.Duration("export_step_duration", time.Since(exportStepStart)),
		)
		return errors.ErrExportCacheFileFailed.WithCause(err)
	}
	exportStepDuration := time.Since(exportStepStart)
	log.Debug(
		"导出mon节点文件：写入文件成功",
		zap.String("path", path),
		zap.Object("mon_node_vars", &monNode),
		zap.Duration("export_step_duration", exportStepDuration),
	)

	log.Info(
		"导出mon节点文件：执行成功",
		zap.String("path", path),
		zap.Uint32("mon_node_id", m.ID),
		zap.Duration("export_step_duration", exportStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}
