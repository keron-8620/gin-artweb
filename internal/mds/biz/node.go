package biz

import (
	"context"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	bizReso "gin-artweb/internal/resource/biz"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
	"gin-artweb/internal/shared/utils/serializer"
)

const MdsNodeIDKey = "mds_node_id"

type MdsNodeModel struct {
	database.StandardModel
	NodeRole    string            `gorm:"column:node_role;type:varchar(50);comment:节点角色" json:"role"`
	IsEnable    bool              `gorm:"column:is_enable;type:tinyint;comment:是否启用" json:"is_enable"`
	MdsColonyID uint32            `gorm:"column:mds_colony_id;not null;comment:mds集群ID" json:"mds_colony_id"`
	MdsColony   MdsColonyModel    `gorm:"foreignKey:MdsColonyID;references:ID;constraint:OnDelete:CASCADE" json:"mds_colony"`
	HostID      uint32            `gorm:"column:host_id;not null;comment:主机ID" json:"host_id"`
	Host        bizReso.HostModel `gorm:"foreignKey:HostID;references:ID;constraint:OnDelete:CASCADE" json:"host"`
}

func (m *MdsNodeModel) TableName() string {
	return "mds_node"
}

func (m *MdsNodeModel) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if err := m.StandardModel.MarshalLogObject(enc); err != nil {
		return err
	}
	enc.AddString("role", m.NodeRole)
	enc.AddBool("is_enable", m.IsEnable)
	return nil
}

type MdsNodeRepo interface {
	CreateModel(context.Context, *MdsNodeModel) error
	UpdateModel(context.Context, map[string]any, ...any) error
	DeleteModel(context.Context, ...any) error
	FindModel(context.Context, []string, ...any) (*MdsNodeModel, error)
	ListModel(context.Context, database.QueryParams) (int64, *[]MdsNodeModel, error)
}

type MdsNodeUsecase struct {
	log      *zap.Logger
	nodeRepo MdsNodeRepo
}

func NewMdsNodeUsecase(
	log *zap.Logger,
	nodeRepo MdsNodeRepo,
) *MdsNodeUsecase {
	return &MdsNodeUsecase{
		log:      log,
		nodeRepo: nodeRepo,
	}
}

func (uc *MdsNodeUsecase) CreateMdsNode(
	ctx context.Context,
	m MdsNodeModel,
) (*MdsNodeModel, *errors.Error) {
	if err := errors.CheckContext(ctx); err != nil {
		return nil, errors.FromError(err)
	}

	uc.log.Info(
		"开始创建mds节点",
		zap.Object(database.ModelKey, &m),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	if err := uc.nodeRepo.CreateModel(ctx, &m); err != nil {
		uc.log.Error(
			"创建mds节点失败",
			zap.Error(err),
			zap.Object(database.ModelKey, &m),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return nil, database.NewGormError(err, nil)
	}

	nm, rErr := uc.FindMdsNodeByID(ctx, []string{"MdsColony", "Host"}, m.ID)
	if rErr != nil {
		return nil, rErr
	}

	if err := uc.OutPortMdsNodeData(ctx, nm); err != nil {
		return nil, err
	}

	uc.log.Info(
		"创建mds节点成功",
		zap.Object(database.ModelKey, &m),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)
	return &m, nil
}

func (uc *MdsNodeUsecase) UpdateMdsNodeByID(
	ctx context.Context,
	mdsNodeID uint32,
	data map[string]any,
) *errors.Error {
	if err := errors.CheckContext(ctx); err != nil {
		return errors.FromError(err)
	}

	uc.log.Info(
		"开始更新mds节点",
		zap.Uint32(MdsNodeIDKey, mdsNodeID),
		zap.Any(database.UpdateDataKey, data),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	data["id"] = mdsNodeID
	if err := uc.nodeRepo.UpdateModel(ctx, data, "id = ?", mdsNodeID); err != nil {
		uc.log.Error(
			"更新mds节点失败",
			zap.Error(err),
			zap.Uint32(MdsNodeIDKey, mdsNodeID),
			zap.Any(database.UpdateDataKey, data),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return database.NewGormError(err, data)
	}

	nm, rErr := uc.FindMdsNodeByID(ctx, []string{"MdsColony", "Host"}, mdsNodeID)
	if rErr != nil {
		return rErr
	}

	if err := uc.OutPortMdsNodeData(ctx, nm); err != nil {
		return err
	}

	uc.log.Info(
		"更新mds节点成功",
		zap.Uint32(MdsNodeIDKey, mdsNodeID),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)
	return nil
}

func (uc *MdsNodeUsecase) DeleteMdsNodeByID(
	ctx context.Context,
	mdsNodeID uint32,
) *errors.Error {
	if err := errors.CheckContext(ctx); err != nil {
		return errors.FromError(err)
	}

	uc.log.Info(
		"开始删除mds节点",
		zap.Uint32(MdsNodeIDKey, mdsNodeID),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	if err := uc.nodeRepo.DeleteModel(ctx, mdsNodeID); err != nil {
		uc.log.Error(
			"删除mds节点失败",
			zap.Error(err),
			zap.Uint32(MdsNodeIDKey, mdsNodeID),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return database.NewGormError(err, map[string]any{"id": mdsNodeID})
	}

	uc.log.Info(
		"删除mds节点成功",
		zap.Uint32(MdsNodeIDKey, mdsNodeID),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)
	return nil
}

func (uc *MdsNodeUsecase) FindMdsNodeByID(
	ctx context.Context,
	preloads []string,
	mdsNodeID uint32,
) (*MdsNodeModel, *errors.Error) {
	if err := errors.CheckContext(ctx); err != nil {
		return nil, errors.FromError(err)
	}

	uc.log.Info(
		"开始查询mds节点",
		zap.Strings(database.PreloadKey, preloads),
		zap.Uint32(MdsNodeIDKey, mdsNodeID),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	m, err := uc.nodeRepo.FindModel(ctx, preloads, mdsNodeID)
	if err != nil {
		uc.log.Error(
			"查询mds节点失败",
			zap.Error(err),
			zap.Uint32(MdsNodeIDKey, mdsNodeID),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return nil, database.NewGormError(err, map[string]any{"id": mdsNodeID})
	}

	uc.log.Info(
		"查询mds节点成功",
		zap.Uint32(MdsNodeIDKey, mdsNodeID),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)
	return m, nil
}

func (uc *MdsNodeUsecase) ListMdsNode(
	ctx context.Context,
	qp database.QueryParams,
) (int64, *[]MdsNodeModel, *errors.Error) {
	if err := errors.CheckContext(ctx); err != nil {
		return 0, nil, errors.FromError(err)
	}

	uc.log.Info(
		"开始查询角色列表",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	count, ms, err := uc.nodeRepo.ListModel(ctx, qp)
	if err != nil {
		uc.log.Error(
			"查询mds节点列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return 0, nil, database.NewGormError(err, nil)
	}

	uc.log.Info(
		"查询mds节点列表成功",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)
	return count, ms, nil
}

func (uc *MdsNodeUsecase) OutPortMdsNodeData(ctx context.Context, m *MdsNodeModel) *errors.Error {
	if err := errors.CheckContext(ctx); err != nil {
		return errors.FromError(err)
	}

	uc.log.Info(
		"开始导出mds节点变量文件",
		zap.Object(database.ModelKey, m),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)
	var specdir string
	if m.NodeRole == "master" {
		specdir = "host_01"
	} else if m.NodeRole == "follow" {
		specdir = "host_02"
	} else {
		specdir = "host_03"
	}
	mdsVars := MdsNodeVars{
		ID:       m.ID,
		NodeRole: m.NodeRole,
		Specdir:  specdir,
		HostID:   m.HostID,
	}
	mdsColonyConf := filepath.Join(config.StorageDir, "export", "mds", m.MdsColony.ColonyNum, "colony.yml")
	if _, err := serializer.WriteYAML(mdsColonyConf, mdsVars); err != nil {
		uc.log.Error(
			"导出mds节点变量文件失败",
			zap.Error(err),
			zap.Uint32(MdsNodeIDKey, m.ID),
			zap.String("colony_num", m.MdsColony.ColonyNum),
			zap.String("path", mdsColonyConf),
			zap.Object("mds_node_vars", &mdsVars),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return ErrExportMdsColonyFailed.WithCause(err)
	}

	uc.log.Info(
		"导出mds节点变量文件失败",
		zap.String("path", mdsColonyConf),
		zap.Object("mds_colony_vars", &mdsVars),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)
	return nil
}

type MdsNodeVars struct {
	ID       uint32 `json:"id" yaml:"id"`
	NodeRole string `json:"node_role" yaml:"node_role"`
	Specdir  string `json:"specdir" yaml:"specdir"`
	HostID   uint32 `json:"host_id" yaml:"host_id"`
}

func (vs *MdsNodeVars) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddUint32("id", vs.ID)
	enc.AddString("node_role", vs.NodeRole)
	enc.AddString("specdir", vs.Specdir)
	enc.AddUint32("host_id", vs.HostID)
	return nil
}
