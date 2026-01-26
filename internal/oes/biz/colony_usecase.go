package biz

import (
	"context"
	"fmt"
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
	OesColonyTableName = "oes_colony"
	OesColonyIDKey     = "oes_colony_id"
)

type OesColonyModel struct {
	database.StandardModel
	SystemType    string               `gorm:"column:system_type;type:varchar(20);comment:系统类型" json:"system_type"`
	ColonyNum     string               `gorm:"column:colony_num;type:varchar(2);uniqueIndex;comment:集群号" json:"colony_num"`
	ExtractedName string               `gorm:"column:extracted_name;type:varchar(50);comment:解压后名称" json:"extracted_name"`
	PackageID     uint32               `gorm:"column:package_id;comment:程序包ID" json:"package_id"`
	Package       bizReso.PackageModel `gorm:"foreignKey:PackageID;references:ID;constraint:OnDelete:CASCADE" json:"package"`
	XCounterID    uint32               `gorm:"column:xcounter_id;comment:xcounter包ID" json:"xcounter_id"`
	XCounter      bizReso.PackageModel `gorm:"foreignKey:XCounterID;references:ID;constraint:OnDelete:CASCADE" json:"xcounter"`
	MonNodeID     uint32               `gorm:"column:mon_node_id;not null;comment:mon节点ID" json:"mon_node_id"`
	MonNode       bizMon.MonNodeModel  `gorm:"foreignKey:MonNodeID;references:ID;constraint:OnDelete:CASCADE" json:"mon_node"`
}

func (m *OesColonyModel) TableName() string {
	return OesColonyTableName
}

func (m *OesColonyModel) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if m == nil {
		return errors.GormModelIsNil(OesColonyTableName)
	}
	if err := m.StandardModel.MarshalLogObject(enc); err != nil {
		return err
	}
	enc.AddString("system_type", m.SystemType)
	enc.AddString("colony_num", m.ColonyNum)
	enc.AddString("extracted_name", m.ExtractedName)
	enc.AddUint32("package_id", m.PackageID)
	enc.AddUint32("xcounter_id", m.XCounterID)
	enc.AddUint32("mon_node_id", m.MonNodeID)
	return nil
}

type OesColonyRepo interface {
	CreateModel(context.Context, *OesColonyModel) error
	UpdateModel(context.Context, map[string]any, ...any) error
	DeleteModel(context.Context, ...any) error
	FindModel(context.Context, []string, ...any) (*OesColonyModel, error)
	ListModel(context.Context, database.QueryParams) (int64, *[]OesColonyModel, error)
}

type OesColonyUsecase struct {
	log        *zap.Logger
	colonyRepo OesColonyRepo
}

func NewOesColonyUsecase(
	log *zap.Logger,
	colonyRepo OesColonyRepo,
) *OesColonyUsecase {
	return &OesColonyUsecase{
		log:        log,
		colonyRepo: colonyRepo,
	}
}

func (uc *OesColonyUsecase) CreateOesColony(
	ctx context.Context,
	m OesColonyModel,
) (*OesColonyModel, *errors.Error) {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return nil, errors.FromError(err)
	}

	uc.log.Info(
		"开始创建oes集群",
		zap.Object(database.ModelKey, &m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	if err := uc.colonyRepo.CreateModel(ctx, &m); err != nil {
		uc.log.Error(
			"创建oes集群失败",
			zap.Error(err),
			zap.Object(database.ModelKey, &m),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, nil)
	}

	nm, rErr := uc.FindOesColonyByID(ctx, []string{"Package", "XCounter", "MonNode"}, m.ID)
	if rErr != nil {
		return nil, rErr
	}

	if err := uc.OutportOesColonyData(ctx, nm); err != nil {
		return nil, err
	}

	uc.log.Info(
		"创建oes集群成功",
		zap.Object(database.ModelKey, nm),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return nm, nil
}

func (uc *OesColonyUsecase) UpdateOesColonyByID(
	ctx context.Context,
	oesColonyID uint32,
	data map[string]any,
) (*OesColonyModel, *errors.Error) {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return nil, errors.FromError(err)
	}

	uc.log.Info(
		"开始更新oes集群",
		zap.Uint32(OesColonyIDKey, oesColonyID),
		zap.Any(database.UpdateDataKey, data),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	data["id"] = oesColonyID
	if err := uc.colonyRepo.UpdateModel(ctx, data, "id = ?", oesColonyID); err != nil {
		uc.log.Error(
			"更新oes集群失败",
			zap.Error(err),
			zap.Uint32(OesColonyIDKey, oesColonyID),
			zap.Any(database.UpdateDataKey, data),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, data)
	}

	m, rErr := uc.FindOesColonyByID(ctx, []string{"Package", "XCounter", "MonNode"}, oesColonyID)
	if rErr != nil {
		return nil, rErr
	}

	if err := uc.OutportOesColonyData(ctx, m); err != nil {
		return nil, err
	}

	uc.log.Info(
		"更新oes集群成功",
		zap.Uint32(OesColonyIDKey, oesColonyID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return m, nil
}

func (uc *OesColonyUsecase) DeleteOesColonyByID(
	ctx context.Context,
	oesColonyID uint32,
) *errors.Error {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return errors.FromError(err)
	}

	uc.log.Info(
		"开始删除oes集群",
		zap.Uint32(OesColonyIDKey, oesColonyID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	if err := uc.colonyRepo.DeleteModel(ctx, oesColonyID); err != nil {
		uc.log.Error(
			"删除oes集群失败",
			zap.Error(err),
			zap.Uint32(OesColonyIDKey, oesColonyID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.NewGormError(err, map[string]any{"id": oesColonyID})
	}

	uc.log.Info(
		"删除oes集群成功",
		zap.Uint32(OesColonyIDKey, oesColonyID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return nil
}

func (uc *OesColonyUsecase) FindOesColonyByID(
	ctx context.Context,
	preloads []string,
	oesColonyID uint32,
) (*OesColonyModel, *errors.Error) {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return nil, errors.FromError(err)
	}

	uc.log.Info(
		"开始查询oes集群",
		zap.Strings(database.PreloadKey, preloads),
		zap.Uint32(OesColonyIDKey, oesColonyID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, err := uc.colonyRepo.FindModel(ctx, preloads, oesColonyID)
	if err != nil {
		uc.log.Error(
			"查询oes集群失败",
			zap.Error(err),
			zap.Uint32(OesColonyIDKey, oesColonyID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, map[string]any{"id": oesColonyID})
	}

	uc.log.Info(
		"查询oes集群成功",
		zap.Uint32(OesColonyIDKey, oesColonyID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return m, nil
}

func (uc *OesColonyUsecase) ListOesColony(
	ctx context.Context,
	qp database.QueryParams,
) (int64, *[]OesColonyModel, *errors.Error) {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return 0, nil, errors.FromError(err)
	}

	uc.log.Info(
		"开始查询角色列表",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	count, ms, err := uc.colonyRepo.ListModel(ctx, qp)
	if err != nil {
		uc.log.Error(
			"查询oes集群列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}

	uc.log.Info(
		"查询oes集群列表成功",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return count, ms, nil
}

func (uc *OesColonyUsecase) GetOesColonyBinDir(colonyNum string) string {
	return filepath.Join(config.StorageDir, "oes", "bin", colonyNum)
}

func (uc *OesColonyUsecase) GetOesColonyConfigDir(colonyNum string) string {
	return filepath.Join(config.StorageDir, "oes", "config", colonyNum)
}

func (uc *OesColonyUsecase) OutportOesColonyData(
	ctx context.Context,
	m *OesColonyModel,
) *errors.Error {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return errors.FromError(err)
	}

	uc.log.Info(
		"开始解压oes程序包并初始化集群配置文件",
		zap.Object(database.ModelKey, m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	colonyBinDir := uc.GetOesColonyBinDir(m.ColonyNum)
	colonyConfDir := uc.GetOesColonyConfigDir(m.ColonyNum)

	if _, err := os.Stat(colonyBinDir); !os.IsNotExist(err) {
		if err := os.RemoveAll(colonyBinDir); err != nil {
			uc.log.Error(
				"清理原oes集群配置文件失败",
				zap.Error(err),
				zap.String("path", colonyBinDir),
			)
			return ErrExportOesColonyFailed.WithCause(err)
		}
	}

	tmpDir, mErr := os.MkdirTemp("/tmp", "oes-")
	if mErr != nil {
		uc.log.Error(
			"创建oes程序包解压的tmp文件夹失败",
			zap.Error(mErr),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return ErrExportOesColonyFailed.WithCause(mErr)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			uc.log.Error(
				"删除oes程序包解压的tmp文件夹失败",
				zap.Error(err),
				zap.String("path", tmpDir),
			)
		}
	}()

	oesPkgPath := bizReso.PackageStoragePath(m.Package.StorageFilename)
	oesUnTarDirName, valiErr := archive.ValidateSingleDirTarGz(oesPkgPath)
	if valiErr != nil {
		uc.log.Error(
			"oes程序包校验失败",
			zap.Error(valiErr),
			zap.String("path", oesPkgPath),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return ErrExportOesColonyFailed.WithCause(valiErr)
	}

	if err := archive.UntarGz(oesPkgPath, tmpDir, archive.WithContext(ctx)); err != nil {
		uc.log.Error(
			"解压oes程序包失败",
			zap.Error(err),
			zap.String("src_path", oesPkgPath),
			zap.String("dst_path", colonyBinDir),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return ErrUntarGzOesPackage.WithCause(err)
	}

	oesTmpDir := filepath.Join(tmpDir, oesUnTarDirName)
	if err := fileutil.CopyDir(oesTmpDir, colonyBinDir, true); err != nil {
		uc.log.Error(
			"复制oes程序包解压目录失败",
			zap.Error(err),
			zap.String("src_path", oesTmpDir),
			zap.String("dst_path", colonyBinDir),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return ErrUntarGzOesPackage.WithCause(err)
	}

	xcterPkgPath := bizReso.PackageStoragePath(m.XCounter.StorageFilename)
	xcterUnTarDirName, valiErr := archive.ValidateSingleDirTarGz(xcterPkgPath)
	if valiErr != nil {
		uc.log.Error(
			"xcounter程序包校验失败",
			zap.Error(valiErr),
			zap.String("path", xcterPkgPath),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return ErrExportOesColonyFailed.WithCause(valiErr)
	}
	if err := archive.UntarGz(xcterPkgPath, tmpDir); err != nil {
		uc.log.Error(
			"压缩xcounter程序包失败",
			zap.Error(err),
			zap.String("src_path", xcterPkgPath),
			zap.String("dst_path", tmpDir),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return ErrUntarGzXCounterPackage.WithCause(err)
	}
	xcterTmpDir := filepath.Join(tmpDir, xcterUnTarDirName, "bin")
	oesBinDir := filepath.Join(colonyBinDir, "bin")
	if err := fileutil.CopyDir(xcterTmpDir, oesBinDir, true); err != nil {
		uc.log.Error(
			"复制xcounter程序包解压目录失败",
			zap.Error(err),
			zap.String("src_path", xcterTmpDir),
			zap.String("dst_path", oesBinDir),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return ErrUntarGzOesPackage.WithCause(err)
	}

	colonyConfAll := filepath.Join(colonyConfDir, "all")
	if _, err := os.Stat(colonyConfAll); os.IsNotExist(err) {
		colonyBinConf := filepath.Join(colonyBinDir, "conf")
		if err := fileutil.CopyDir(colonyBinConf, colonyConfAll, true); err != nil {
			uc.log.Error(
				"复制oes集群配置文件失败",
				zap.Error(err),
				zap.String("src_path", colonyBinConf),
				zap.String("dst_path", colonyConfAll),
				zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			)
			return errors.FromError(err)
		}
		srcPath := filepath.Join(config.ConfigDir, fmt.Sprintf("automatic_oes_%s.yaml", m.SystemType))
		dstPath := filepath.Join(colonyConfAll, "automatic.yaml")
		if err := fileutil.CopyFile(srcPath, dstPath); err != nil {
			uc.log.Error(
				"复制oes的automatic配置文件失败",
				zap.Error(err),
				zap.String("src_path", colonyConfDir),
				zap.String("dst_path", dstPath),
				zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			)
			return errors.FromError(err)
		}
	}
	oesVars := OesColonyVars{
		ID:        m.ID,
		ColonyNum: m.ColonyNum,
		PkgName:   m.ExtractedName,
		PackageID: m.PackageID,
		Version:   m.Package.Version,
		MonNodeID: m.MonNodeID,
	}
	oesColonyConf := filepath.Join(colonyConfAll, "colony.yaml")
	if _, err := serializer.WriteYAML(oesColonyConf, oesVars); err != nil {
		uc.log.Error(
			"导出oes集群配置变量文件失败",
			zap.Error(err),
			zap.String("path", oesColonyConf),
			zap.Object("oes_colony_vars", &oesVars),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return ErrExportOesColonyFailed.WithCause(err)
	}

	uc.log.Info(
		"解压oes程序包并初始化集群配置文件成功",
		zap.String("path", oesColonyConf),
		zap.Object("oes_colony_vars", &oesVars),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return nil
}

type OesColonyVars struct {
	ID         uint32 `json:"id" yaml:"id"`
	SystemType string `json:"system_type" yaml:"system_type"`
	ColonyNum  string `json:"colony_num" yaml:"colony_num"`
	PkgName    string `json:"pkg_name" yaml:"pkg_name"`
	PackageID  uint32 `json:"package_id" yaml:"package_id"`
	Version    string `json:"version" yaml:"version"`
	MonNodeID  uint32 `json:"mon_node_id" yaml:"mon_node_id"`
}

func (vs *OesColonyVars) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddUint32("id", vs.ID)
	enc.AddString("colony_num", vs.ColonyNum)
	enc.AddString("pkg_name", vs.PkgName)
	enc.AddUint32("package_id", vs.PackageID)
	enc.AddString("version", vs.Version)
	enc.AddUint32("mon_node_id", vs.MonNodeID)
	return nil
}
