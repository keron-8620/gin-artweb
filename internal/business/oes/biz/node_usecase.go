package biz

import (
	"context"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	resoModel "gin-artweb/internal/infra/resource/model"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
	"gin-artweb/pkg/serializer"
)

const (
	OesNodeTableName = "oes_node"
	OesNodeIDKey     = "oes_node_id"
)

type OesNodeModel struct {
	database.StandardModel
	NodeRole    string              `gorm:"column:node_role;type:varchar(50);comment:节点角色" json:"role"`
	IsEnable    bool                `gorm:"column:is_enable;type:boolean;comment:是否启用" json:"is_enable"`
	OesColonyID uint32              `gorm:"column:oes_colony_id;not null;comment:oes集群ID" json:"oes_colony_id"`
	OesColony   OesColonyModel      `gorm:"foreignKey:OesColonyID;references:ID;constraint:OnDelete:CASCADE" json:"oes_colony"`
	HostID      uint32              `gorm:"column:host_id;not null;comment:主机ID" json:"host_id"`
	Host        resoModel.HostModel `gorm:"foreignKey:HostID;references:ID;constraint:OnDelete:CASCADE" json:"host"`
}

func (m *OesNodeModel) TableName() string {
	return OesNodeTableName
}

func (m *OesNodeModel) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if m == nil {
		return nil
	}
	if err := m.StandardModel.MarshalLogObject(enc); err != nil {
		return err
	}
	enc.AddString("role", m.NodeRole)
	enc.AddBool("is_enable", m.IsEnable)
	enc.AddUint32("oes_colony_id", m.OesColonyID)
	enc.AddUint32("host_id", m.HostID)
	return nil
}

type OesNodeRepo interface {
	CreateModel(context.Context, *OesNodeModel) error
	UpdateModel(context.Context, map[string]any, ...any) error
	DeleteModel(context.Context, ...any) error
	GetModel(context.Context, []string, ...any) (*OesNodeModel, error)
	ListModel(context.Context, database.QueryParams) (int64, *[]OesNodeModel, error)
}

type OesNodeUsecase struct {
	log      *zap.Logger
	nodeRepo OesNodeRepo
}

func NewOesNodeUsecase(
	log *zap.Logger,
	nodeRepo OesNodeRepo,
) *OesNodeUsecase {
	return &OesNodeUsecase{
		log:      log,
		nodeRepo: nodeRepo,
	}
}

func (uc *OesNodeUsecase) CreateOesNode(
	ctx context.Context,
	m OesNodeModel,
) (*OesNodeModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	uc.log.Info(
		"开始创建oes节点",
		zap.Object(database.ModelKey, &m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	if err := uc.nodeRepo.CreateModel(ctx, &m); err != nil {
		uc.log.Error(
			"创建oes节点失败",
			zap.Error(err),
			zap.Object(database.ModelKey, &m),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, nil)
	}

	// 查询oes节点关联数据
	nm, rErr := uc.FindOesNodeByID(ctx, []string{"OesColony", "Host"}, m.ID)
	if rErr != nil {
		return nil, rErr
	}

	// 导出oes节点缓存数据
	if err := uc.OutPortOesNodeData(ctx, nm); err != nil {
		return nil, err
	}

	uc.log.Info(
		"创建oes节点成功",
		zap.Object(database.ModelKey, nm),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return nm, nil
}

func (uc *OesNodeUsecase) UpdateOesNodeByID(
	ctx context.Context,
	oesNodeID uint32,
	data map[string]any,
) (*OesNodeModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	uc.log.Info(
		"开始更新oes节点",
		zap.Uint32(OesNodeIDKey, oesNodeID),
		zap.Any(database.UpdateDataKey, data),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	data["id"] = oesNodeID
	if err := uc.nodeRepo.UpdateModel(ctx, data, "id = ?", oesNodeID); err != nil {
		uc.log.Error(
			"更新oes节点失败",
			zap.Error(err),
			zap.Uint32(OesNodeIDKey, oesNodeID),
			zap.Any(database.UpdateDataKey, data),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, data)
	}

	// 查询oes节点关联数据
	m, rErr := uc.FindOesNodeByID(ctx, []string{"OesColony", "Host"}, oesNodeID)
	if rErr != nil {
		return nil, rErr
	}

	// 导出oes节点缓存数据
	if err := uc.OutPortOesNodeData(ctx, m); err != nil {
		return nil, err
	}

	uc.log.Info(
		"更新oes节点成功",
		zap.Uint32(OesNodeIDKey, oesNodeID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return m, nil
}

func (uc *OesNodeUsecase) DeleteOesNodeByID(
	ctx context.Context,
	oesNodeID uint32,
) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	uc.log.Info(
		"开始删除oes节点",
		zap.Uint32(OesNodeIDKey, oesNodeID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	if err := uc.nodeRepo.DeleteModel(ctx, oesNodeID); err != nil {
		uc.log.Error(
			"删除oes节点失败",
			zap.Error(err),
			zap.Uint32(OesNodeIDKey, oesNodeID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.NewGormError(err, map[string]any{"id": oesNodeID})
	}

	uc.log.Info(
		"删除oes节点成功",
		zap.Uint32(OesNodeIDKey, oesNodeID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return nil
}

func (uc *OesNodeUsecase) FindOesNodeByID(
	ctx context.Context,
	preloads []string,
	oesNodeID uint32,
) (*OesNodeModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	uc.log.Info(
		"开始查询oes节点",
		zap.Strings(database.PreloadKey, preloads),
		zap.Uint32(OesNodeIDKey, oesNodeID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, err := uc.nodeRepo.GetModel(ctx, preloads, oesNodeID)
	if err != nil {
		uc.log.Error(
			"查询oes节点失败",
			zap.Error(err),
			zap.Uint32(OesNodeIDKey, oesNodeID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, map[string]any{"id": oesNodeID})
	}

	uc.log.Info(
		"查询oes节点成功",
		zap.Uint32(OesNodeIDKey, oesNodeID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return m, nil
}

func (uc *OesNodeUsecase) ListOesNode(
	ctx context.Context,
	qp database.QueryParams,
) (int64, *[]OesNodeModel, *errors.Error) {
	if ctx.Err() != nil {
		return 0, nil, errors.FromError(ctx.Err())
	}

	uc.log.Info(
		"开始查询oes节点列表",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	count, ms, err := uc.nodeRepo.ListModel(ctx, qp)
	if err != nil {
		uc.log.Error(
			"查询oes节点列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}

	uc.log.Info(
		"查询oes节点列表成功",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return count, ms, nil
}

func (uc *OesNodeUsecase) OutPortOesNodeData(ctx context.Context, m *OesNodeModel) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	uc.log.Info(
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
	oesVars := OesNodeVars{
		ID:       m.ID,
		NodeRole: m.NodeRole,
		Specdir:  specdir,
		HostID:   m.HostID,
		IsEnable: m.IsEnable,
	}
	oesColonyConf := filepath.Join(config.StorageDir, "oes", "config", m.OesColony.ColonyNum, specdir, "node.yaml")
	if _, err := serializer.WriteYAML(oesColonyConf, oesVars); err != nil {
		uc.log.Error(
			"导出oes节点变量文件失败",
			zap.Error(err),
			zap.Uint32(OesNodeIDKey, m.ID),
			zap.String("colony_num", m.OesColony.ColonyNum),
			zap.String("path", oesColonyConf),
			zap.Object("oes_node_vars", &oesVars),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.ErrExportCacheFileFailed.WithCause(err)
	}

	uc.log.Info(
		"导出oes节点变量文件失败",
		zap.String("path", oesColonyConf),
		zap.Object("oes_colony_vars", &oesVars),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return nil
}

type OesNodeVars struct {
	ID       uint32 `json:"id" yaml:"id"`
	NodeRole string `json:"node_role" yaml:"node_role"`
	Specdir  string `json:"specdir" yaml:"specdir"`
	HostID   uint32 `json:"host_id" yaml:"host_id"`
	IsEnable bool   `json:"is_enable" yaml:"is_enable"`
}

func (vs *OesNodeVars) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddUint32("id", vs.ID)
	enc.AddString("node_role", vs.NodeRole)
	enc.AddString("specdir", vs.Specdir)
	enc.AddUint32("host_id", vs.HostID)
	enc.AddBool("is_enable", vs.IsEnable)
	return nil
}
