package biz

import (
	"context"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	bizMon "gin-artweb/internal/mon/biz"
	bizReso "gin-artweb/internal/resource/biz"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
	"gin-artweb/pkg/archive"
	"gin-artweb/pkg/ctxutil"
	"gin-artweb/pkg/fileutil"
	"gin-artweb/pkg/serializer"
)

const (
	MdsColonyTableName = "mds_colony"
	MdsColonyIDKey     = "mds_colony_id"
)

type MdsColonyModel struct {
	database.StandardModel
	ColonyNum     string               `gorm:"column:colony_num;type:varchar(2);uniqueIndex;comment:集群号" json:"colony_num"`
	ExtractedName string               `gorm:"column:extracted_name;type:varchar(50);comment:解压后名称" json:"extracted_name"`
	PackageID     uint32               `gorm:"column:package_id;comment:程序包ID" json:"package_id"`
	Package       bizReso.PackageModel `gorm:"foreignKey:PackageID;references:ID;constraint:OnDelete:CASCADE" json:"package"`
	MonNodeID     uint32               `gorm:"column:mon_node_id;not null;comment:mon节点ID" json:"mon_node_id"`
	MonNode       bizMon.MonNodeModel  `gorm:"foreignKey:MonNodeID;references:ID;constraint:OnDelete:CASCADE" json:"mon_node"`
}

func (m *MdsColonyModel) TableName() string {
	return MdsColonyTableName
}

func (m *MdsColonyModel) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if m == nil {
		return errors.GormModelIsNil(MdsColonyTableName)
	}
	if err := m.StandardModel.MarshalLogObject(enc); err != nil {
		return err
	}
	enc.AddString("colony_num", m.ColonyNum)
	enc.AddString("extracted_name", m.ExtractedName)
	enc.AddUint32("package_id", m.PackageID)
	enc.AddUint32("mon_node_id", m.MonNodeID)
	return nil
}

type MdsColonyRepo interface {
	CreateModel(context.Context, *MdsColonyModel) error
	UpdateModel(context.Context, map[string]any, ...any) error
	DeleteModel(context.Context, ...any) error
	FindModel(context.Context, []string, ...any) (*MdsColonyModel, error)
	ListModel(context.Context, database.QueryParams) (int64, *[]MdsColonyModel, error)
}

type MdsColonyUsecase struct {
	log        *zap.Logger
	colonyRepo MdsColonyRepo
}

func NewMdsColonyUsecase(
	log *zap.Logger,
	colonyRepo MdsColonyRepo,
) *MdsColonyUsecase {
	return &MdsColonyUsecase{
		log:        log,
		colonyRepo: colonyRepo,
	}
}

func (uc *MdsColonyUsecase) CreateMdsColony(
	ctx context.Context,
	m MdsColonyModel,
) (*MdsColonyModel, *errors.Error) {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return nil, errors.FromError(err)
	}

	uc.log.Info(
		"开始创建mds集群",
		zap.Object(database.ModelKey, &m),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)

	if err := uc.colonyRepo.CreateModel(ctx, &m); err != nil {
		uc.log.Error(
			"创建mds集群失败",
			zap.Error(err),
			zap.Object(database.ModelKey, &m),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, nil)
	}

	nm, rErr := uc.FindMdsColonyByID(ctx, []string{"Package", "MonNode"}, m.ID)
	if rErr != nil {
		return nil, rErr
	}

	if err := uc.OutportMdsColonyData(ctx, nm); err != nil {
		return nil, err
	}

	uc.log.Info(
		"创建mds集群成功",
		zap.Object(database.ModelKey, nm),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)
	return nm, nil
}

func (uc *MdsColonyUsecase) UpdateMdsColonyByID(
	ctx context.Context,
	mdsColonyID uint32,
	data map[string]any,
) (*MdsColonyModel, *errors.Error) {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return nil, errors.FromError(err)
	}

	uc.log.Info(
		"开始更新mds集群",
		zap.Uint32(MdsColonyIDKey, mdsColonyID),
		zap.Any(database.UpdateDataKey, data),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)

	data["id"] = mdsColonyID
	if err := uc.colonyRepo.UpdateModel(ctx, data, "id = ?", mdsColonyID); err != nil {
		uc.log.Error(
			"更新mds集群失败",
			zap.Error(err),
			zap.Uint32(MdsColonyIDKey, mdsColonyID),
			zap.Any(database.UpdateDataKey, data),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, data)
	}

	m, rErr := uc.FindMdsColonyByID(ctx, []string{"Package", "MonNode"}, mdsColonyID)
	if rErr != nil {
		return nil, rErr
	}

	if err := uc.OutportMdsColonyData(ctx, m); err != nil {
		return nil, err
	}

	uc.log.Info(
		"更新mds集群成功",
		zap.Uint32(MdsColonyIDKey, mdsColonyID),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)
	return m, nil
}

func (uc *MdsColonyUsecase) DeleteMdsColonyByID(
	ctx context.Context,
	mdsColonyID uint32,
) *errors.Error {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return errors.FromError(err)
	}

	uc.log.Info(
		"开始删除mds集群",
		zap.Uint32(MdsColonyIDKey, mdsColonyID),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)

	if err := uc.colonyRepo.DeleteModel(ctx, mdsColonyID); err != nil {
		uc.log.Error(
			"删除mds集群失败",
			zap.Error(err),
			zap.Uint32(MdsColonyIDKey, mdsColonyID),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		return errors.NewGormError(err, map[string]any{"id": mdsColonyID})
	}

	uc.log.Info(
		"删除mds集群成功",
		zap.Uint32(MdsColonyIDKey, mdsColonyID),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)
	return nil
}

func (uc *MdsColonyUsecase) FindMdsColonyByID(
	ctx context.Context,
	preloads []string,
	mdsColonyID uint32,
) (*MdsColonyModel, *errors.Error) {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return nil, errors.FromError(err)
	}

	uc.log.Info(
		"开始查询mds集群",
		zap.Strings(database.PreloadKey, preloads),
		zap.Uint32(MdsColonyIDKey, mdsColonyID),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)

	m, err := uc.colonyRepo.FindModel(ctx, preloads, mdsColonyID)
	if err != nil {
		uc.log.Error(
			"查询mds集群失败",
			zap.Error(err),
			zap.Uint32(MdsColonyIDKey, mdsColonyID),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, map[string]any{"id": mdsColonyID})
	}

	uc.log.Info(
		"查询mds集群成功",
		zap.Uint32(MdsColonyIDKey, mdsColonyID),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)
	return m, nil
}

func (uc *MdsColonyUsecase) ListMdsColony(
	ctx context.Context,
	qp database.QueryParams,
) (int64, *[]MdsColonyModel, *errors.Error) {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return 0, nil, errors.FromError(err)
	}

	uc.log.Info(
		"开始查询角色列表",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)

	count, ms, err := uc.colonyRepo.ListModel(ctx, qp)
	if err != nil {
		uc.log.Error(
			"查询mds集群列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}

	uc.log.Info(
		"查询mds集群列表成功",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)
	return count, ms, nil
}

func (uc *MdsColonyUsecase) GetMdsColonyBinDir(colonyNum string) string {
	return filepath.Join(config.StorageDir, "mds", "bin", colonyNum)
}

func (uc *MdsColonyUsecase) GetMdsColonyConfigDir(colonyNum string) string {
	return filepath.Join(config.StorageDir, "mds", "config", colonyNum)
}

func (uc *MdsColonyUsecase) OutportMdsColonyData(
	ctx context.Context,
	m *MdsColonyModel,
) *errors.Error {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return errors.FromError(err)
	}

	uc.log.Info(
		"开始解压mds程序包并初始化集群配置文件",
		zap.Object(database.ModelKey, m),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)

	colonyBinDir := uc.GetMdsColonyBinDir(m.ColonyNum)
	colonyConfDir := uc.GetMdsColonyConfigDir(m.ColonyNum)

	if _, err := os.Stat(colonyBinDir); !os.IsNotExist(err) {
		if err := os.RemoveAll(colonyBinDir); err != nil {
			uc.log.Error(
				"清理原mds集群配置文件失败",
				zap.Error(err),
				zap.String("path", colonyBinDir),
			)
			return ErrExportMdsColonyFailed.WithCause(err)
		}
	}

	tmpDir, mErr := os.MkdirTemp("/tmp", "mds-")
	if mErr != nil {
		uc.log.Error(
			"创建mds程序包解压的tmp文件夹失败",
			zap.Error(mErr),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		return ErrExportMdsColonyFailed.WithCause(mErr)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			uc.log.Error(
				"删除mds程序包解压的tmp文件夹失败",
				zap.Error(err),
				zap.String("path", tmpDir),
			)
		}
	}()

	mdsPkgPath := bizReso.PackageStoragePath(m.Package.StorageFilename)
	mdsUnTarDirName, valiErr := archive.ValidateSingleDirTarGz(mdsPkgPath)
	if valiErr != nil {
		uc.log.Error(
			"mds程序包校验失败",
			zap.Error(valiErr),
			zap.String("path", mdsPkgPath),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		return ErrExportMdsColonyFailed.WithCause(valiErr)
	}
	if err := archive.UntarGz(mdsPkgPath, tmpDir, archive.WithContext(ctx)); err != nil {
		uc.log.Error(
			"解压mds程序包失败",
			zap.Error(err),
			zap.String("path", m.Package.StorageFilename),
			zap.String("dest", colonyBinDir),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		return ErrUntarGzMdsPackage.WithCause(err)
	}

	mdsTmpDir := filepath.Join(tmpDir, mdsUnTarDirName)
	if err := fileutil.CopyDir(mdsTmpDir, colonyBinDir, true); err != nil {
		uc.log.Error(
			"复制mds程序包解压目录失败",
			zap.Error(err),
			zap.String("src_path", mdsTmpDir),
			zap.String("dst_path", colonyBinDir),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		return ErrUntarGzMdsPackage.WithCause(err)
	}

	colonyConfAll := filepath.Join(colonyConfDir, "all")
	if _, err := os.Stat(colonyConfAll); os.IsNotExist(err) {
		colonyBinConf := filepath.Join(colonyBinDir, "conf")
		if err := fileutil.CopyDir(colonyBinConf, colonyConfAll, true); err != nil {
			uc.log.Error(
				"复制mds集群配置文件失败",
				zap.Error(err),
				zap.String("src_path", colonyBinConf),
				zap.String("dst_path", colonyConfAll),
				zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
			)
			return errors.FromError(err)
		}
		srcPath := filepath.Join(config.ConfigDir, "automatic_mds.yaml")
		dstPath := filepath.Join(colonyConfAll, "automatic.yaml")
		if err := fileutil.CopyFile(srcPath, dstPath); err != nil {
			uc.log.Error(
				"复制mds的automatic配置文件失败",
				zap.Error(err),
				zap.String("src_path", colonyConfDir),
				zap.String("dst_path", dstPath),
				zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
			)
			return errors.FromError(err)
		}
	}
	mdsVars := MdsColonyVars{
		ID:        m.ID,
		ColonyNum: m.ColonyNum,
		PkgName:   m.ExtractedName,
		PackageID: m.PackageID,
		Version:   m.Package.Version,
		MonNodeID: m.MonNodeID,
	}
	mdsColonyConf := filepath.Join(colonyConfAll, "colony.yaml")
	if _, err := serializer.WriteYAML(mdsColonyConf, mdsVars); err != nil {
		uc.log.Error(
			"导出mds集群配置变量文件失败",
			zap.Error(err),
			zap.String("path", mdsColonyConf),
			zap.Object("mds_colony_vars", &mdsVars),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		return ErrExportMdsColonyFailed.WithCause(err)
	}

	uc.log.Info(
		"解压mds程序包并初始化集群配置文件成功",
		zap.String("path", mdsColonyConf),
		zap.Object("mds_colony_vars", &mdsVars),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)
	return nil
}

type MdsColonyVars struct {
	ID        uint32 `json:"id" yaml:"id"`
	ColonyNum string `json:"colony_num" yaml:"colony_num"`
	PkgName   string `json:"pkg_name" yaml:"pkg_name"`
	PackageID uint32 `json:"package_id" yaml:"package_id"`
	Version   string `json:"version" yaml:"version"`
	MonNodeID uint32 `json:"mon_node_id" yaml:"mon_node_id"`
}

func (vs *MdsColonyVars) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddUint32("mds_colony_id", vs.ID)
	enc.AddString("colony_num", vs.ColonyNum)
	enc.AddString("pkg_name", vs.PkgName)
	enc.AddUint32("package_id", vs.PackageID)
	enc.AddString("version", vs.Version)
	enc.AddUint32("mon_node_id", vs.MonNodeID)
	return nil
}
