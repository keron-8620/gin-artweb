package biz

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	bizReso "gin-artweb/internal/resource/biz"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
	"gin-artweb/internal/shared/file"
)

const NodeIDKey = "node_id"

type NodeModel struct {
	database.StandardModel
	Name        string            `gorm:"column:name;type:varchar(50);not null;uniqueIndex;comment:名称" json:"name"`
	DeployPath  string            `gorm:"column:deploy_path;type:varchar(255);comment:部署路径" json:"deploy_path"`
	OutportPath string            `gorm:"column:outport_path;type:varchar(255);comment:导出路径" json:"outport_path"`
	JavaHome    string            `gorm:"column:java_home;type:varchar(255);comment:JAVA_HOME" json:"java_home"`
	URL         string            `gorm:"column:url;type:varchar(150);not null;uniqueIndex;comment:URL地址" json:"url"`
	HostID      uint32            `gorm:"column:host_id;not null;comment:主机ID" json:"host_id"`
	Host        bizReso.HostModel `gorm:"foreignKey:HostID;references:ID;constraint:OnDelete:CASCADE" json:"host"`
}

func (m *NodeModel) TableName() string {
	return "mon_node"
}

func (m *NodeModel) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if err := m.StandardModel.MarshalLogObject(enc); err != nil {
		return err
	}
	enc.AddString("name", m.Name)
	enc.AddString("deploy_path", m.DeployPath)
	enc.AddString("outport_path", m.OutportPath)
	enc.AddString("java_home", m.JavaHome)
	enc.AddString("url", m.URL)
	return nil
}

type NodeRepo interface {
	CreateModel(context.Context, *NodeModel) error
	UpdateModel(context.Context, map[string]any, ...any) error
	DeleteModel(context.Context, ...any) error
	FindModel(context.Context, []string, ...any) (*NodeModel, error)
	ListModel(context.Context, database.QueryParams) (int64, *[]NodeModel, error)
}

type NodeUsecase struct {
	log      *zap.Logger
	nodeRepo NodeRepo
	dir      string
}

func NewNodeUsecase(
	log *zap.Logger,
	nodeRepo NodeRepo,
	dir string,
) *NodeUsecase {
	return &NodeUsecase{
		log:      log,
		nodeRepo: nodeRepo,
		dir:      dir,
	}
}

func (uc *NodeUsecase) CreateMonNode(
	ctx context.Context,
	m NodeModel,
) (*NodeModel, *errors.Error) {
	if err := errors.CheckContext(ctx); err != nil {
		return nil, errors.FromError(err)
	}

	uc.log.Info(
		"开始创建mon节点",
		zap.Object(database.ModelKey, &m),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	if err := uc.nodeRepo.CreateModel(ctx, &m); err != nil {
		uc.log.Error(
			"创建mon节点失败",
			zap.Error(err),
			zap.Object(database.ModelKey, &m),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return nil, database.NewGormError(err, nil)
	}

	if err := uc.ExportMonNode(ctx, m); err != nil {
		return nil, err
	}

	uc.log.Info(
		"创建mon节点成功",
		zap.Object(database.ModelKey, &m),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)
	return uc.FindMonNodeByID(ctx, []string{"Host"}, m.ID)
}

func (uc *NodeUsecase) UpdateMonNodeByID(
	ctx context.Context,
	nodeID uint32,
	data map[string]any,
) *errors.Error {
	if err := errors.CheckContext(ctx); err != nil {
		return errors.FromError(err)
	}

	uc.log.Info(
		"开始更新mon节点",
		zap.Uint32(NodeIDKey, nodeID),
		zap.Any(database.UpdateDataKey, data),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	if err := uc.nodeRepo.UpdateModel(ctx, data, "id = ?", nodeID); err != nil {
		uc.log.Error(
			"更新mon节点失败",
			zap.Error(err),
			zap.Uint32(NodeIDKey, nodeID),
			zap.Any(database.UpdateDataKey, data),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return database.NewGormError(err, data)
	}

	m, rErr := uc.FindMonNodeByID(ctx, []string{"Host"}, nodeID)
	if rErr != nil {
		return rErr
	}

	if rErr := uc.ExportMonNode(ctx, *m); rErr != nil {
		return rErr
	}

	uc.log.Info(
		"更新mon节点成功",
		zap.Uint32(NodeIDKey, nodeID),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)
	return nil
}

func (uc *NodeUsecase) DeleteMonNodeByID(
	ctx context.Context,
	nodeID uint32,
) *errors.Error {
	if err := errors.CheckContext(ctx); err != nil {
		return errors.FromError(err)
	}

	uc.log.Info(
		"开始删除mon",
		zap.Uint32(NodeIDKey, nodeID),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	if err := uc.nodeRepo.DeleteModel(ctx, nodeID); err != nil {
		uc.log.Error(
			"删除mon失败",
			zap.Error(err),
			zap.Uint32(NodeIDKey, nodeID),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return database.NewGormError(err, map[string]any{"id": nodeID})
	}

	path := uc.ExportPath(nodeID)
	if err := os.RemoveAll(path); err != nil && !os.IsNotExist(err) {
		uc.log.Error(
			"删除mon节点文件失败",
			zap.Error(err),
			zap.String("path", path),
			zap.Uint32("mon_node_id", nodeID),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return ErrDeleteMonNodeFileFailed.WithCause(err)
	}

	uc.log.Info(
		"mon删除成功",
		zap.Uint32(NodeIDKey, nodeID),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)
	return nil
}

func (uc *NodeUsecase) FindMonNodeByID(
	ctx context.Context,
	preloads []string,
	nodeID uint32,
) (*NodeModel, *errors.Error) {
	if err := errors.CheckContext(ctx); err != nil {
		return nil, errors.FromError(err)
	}

	uc.log.Info(
		"开始查询mon",
		zap.Uint32(NodeIDKey, nodeID),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	m, err := uc.nodeRepo.FindModel(ctx, preloads, nodeID)
	if err != nil {
		uc.log.Error(
			"查询mon失败",
			zap.Error(err),
			zap.Uint32(NodeIDKey, nodeID),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return nil, database.NewGormError(err, map[string]any{"id": nodeID})
	}

	uc.log.Info(
		"查询mon成功",
		zap.Uint32(NodeIDKey, nodeID),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)
	return m, nil
}

func (uc *NodeUsecase) ListMonNode(
	ctx context.Context,
	qp database.QueryParams,
) (int64, *[]NodeModel, *errors.Error) {
	if err := errors.CheckContext(ctx); err != nil {
		return 0, nil, errors.FromError(err)
	}

	uc.log.Info(
		"开始查询mon列表",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	count, ms, err := uc.nodeRepo.ListModel(ctx, qp)
	if err != nil {
		uc.log.Error(
			"查询mon列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return 0, nil, database.NewGormError(err, nil)
	}

	uc.log.Info(
		"查询mon列表成功",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)
	return count, ms, nil
}

func (uc *NodeUsecase) ExportPath(pk uint32) string {
	filename := fmt.Sprintf("db_mon_%d.yaml", pk)
	return filepath.Join(uc.dir, filename)
}

func (uc *NodeUsecase) ExportMonNode(ctx context.Context, m NodeModel) *errors.Error {
	if err := errors.CheckContext(ctx); err != nil {
		return errors.FromError(err)
	}

	uc.log.Info(
		"开始导出mon节点文件",
		zap.Object(database.ModelKey, &m),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	monNode := MonNodeVars{
		ID:          m.ID,
		Name:        m.Name,
		DeployPath:  m.DeployPath,
		OutportPath: m.OutportPath,
		JavaHome:    m.JavaHome,
		URL:         m.URL,
		HostID:      m.HostID,
	}

	path := uc.ExportPath(m.ID)
	if err := file.WriteYAML(path, monNode); err != nil {
		uc.log.Error(
			"导出mon节点文件失败",
			zap.Error(err),
			zap.String("path", path),
			zap.Object("mon_node", &monNode),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return ErrExportMonNodeFailed.WithCause(err)
	}

	uc.log.Info(
		"导出mon节点文件成功",
		zap.String("path", path),
		zap.Object("mon_node", &monNode),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)
	return nil
}

type MonNodeVars struct {
	ID          uint32 `json:"id" yaml:"id"`
	Name        string `json:"name" yaml:"name"`
	DeployPath  string `json:"deploy_path" yaml:"deploy_path"`
	OutportPath string `json:"outport_path" yaml:"outport_path"`
	JavaHome    string `json:"java_home" yaml:"java_home"`
	URL         string `json:"url" yaml:"url"`
	HostID      uint32 `json:"host_id" yaml:"host_id"`
}

func (vs *MonNodeVars) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddUint32("id", vs.ID)
	enc.AddString("name", vs.Name)
	enc.AddString("deploy_path", vs.DeployPath)
	enc.AddString("outport_path", vs.OutportPath)
	enc.AddString("java_home", vs.JavaHome)
	enc.AddString("url", vs.URL)
	enc.AddUint32("host_id", vs.HostID)
	return nil
}
