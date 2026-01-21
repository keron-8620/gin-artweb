package biz

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
	"gin-artweb/pkg/ctxutil"
)

const (
	PackageIDKey = "package_id"
)

type PackageModel struct {
	database.BaseModel
	Label           string    `gorm:"column:label;type:varchar(50);index:idx_package_label;comment:标签" json:"label"`
	StorageFilename string    `gorm:"column:storage_filename;type:varchar(50);not null;uniqueIndex;comment:磁盘存储文件名" json:"storage_filename"`
	OriginFilename  string    `gorm:"column:origin_filename;type:varchar(255);comment:原始文件名" json:"origin_filename"`
	Version         string    `gorm:"column:version;type:varchar(50);comment:版本号" json:"version"`
	UploadedAt      time.Time `gorm:"column:uploaded_at;autoCreateTime;comment:上传时间" json:"uploaded_at"`
}

func (m *PackageModel) TableName() string {
	return "resource_package"
}

func (m *PackageModel) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if err := m.BaseModel.MarshalLogObject(enc); err != nil {
		return err
	}
	enc.AddString("label", m.Label)
	enc.AddString("storage_filename", m.StorageFilename)
	enc.AddString("origin_filename", m.OriginFilename)
	enc.AddString("version", m.Version)
	enc.AddTime("uploaded_at", m.UploadedAt)
	return nil
}

type PackageRepo interface {
	CreateModel(context.Context, *PackageModel) error
	DeleteModel(context.Context, ...any) error
	FindModel(context.Context, []string, ...any) (*PackageModel, error)
	ListModel(context.Context, database.QueryParams) (int64, *[]PackageModel, error)
}

type PackageUsecase struct {
	log     *zap.Logger
	pkgRepo PackageRepo
}

func NewPackageUsecase(
	log *zap.Logger,
	pkgRepo PackageRepo,
) *PackageUsecase {
	return &PackageUsecase{
		log:     log,
		pkgRepo: pkgRepo,
	}
}

func (uc *PackageUsecase) CreatePackage(
	ctx context.Context,
	m PackageModel,
) (*PackageModel, *errors.Error) {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return nil, errors.FromError(err)
	}

	uc.log.Info(
		"开始创建程序包",
		zap.Object(database.ModelKey, &m),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)

	if err := uc.pkgRepo.CreateModel(ctx, &m); err != nil {
		uc.log.Error(
			"创建程序包失败",
			zap.Error(err),
			zap.Object(database.ModelKey, &m),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		return nil, database.NewGormError(err, nil)
	}

	uc.log.Info(
		"创建程序包成功",
		zap.Object(database.ModelKey, &m),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)
	return &m, nil
}

func (uc *PackageUsecase) DeletePackageById(
	ctx context.Context,
	pkgId uint32,
) *errors.Error {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return errors.FromError(err)
	}

	uc.log.Info(
		"开始删除程序包",
		zap.Uint32(PackageIDKey, pkgId),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)

	m, err := uc.FindPackageById(ctx, pkgId)
	if err != nil {
		return err
	}

	// 先从数据库删除
	if err := uc.pkgRepo.DeleteModel(ctx, pkgId); err != nil {
		uc.log.Error(
			"删除程序包失败",
			zap.Error(err),
			zap.Uint32(PackageIDKey, pkgId),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		return database.NewGormError(err, map[string]any{"id": pkgId})
	}

	// 再删除物理文件
	if rmErr := uc.RemovePackage(ctx, *m); rmErr != nil {
		return rmErr
	}

	uc.log.Info(
		"删除程序包成功",
		zap.Uint32(PackageIDKey, pkgId),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)
	return nil
}

func (uc *PackageUsecase) FindPackageById(
	ctx context.Context,
	pkgId uint32,
) (*PackageModel, *errors.Error) {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return nil, errors.FromError(err)
	}

	uc.log.Info(
		"开始查询程序包",
		zap.Uint32(PackageIDKey, pkgId),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)

	m, err := uc.pkgRepo.FindModel(ctx, nil, pkgId)
	if err != nil {
		uc.log.Error(
			"查询程序包失败",
			zap.Error(err),
			zap.Uint32(PackageIDKey, pkgId),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		return nil, database.NewGormError(err, map[string]any{"id": pkgId})
	}

	uc.log.Info(
		"查询程序包成功",
		zap.Uint32(PackageIDKey, pkgId),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)
	return m, nil
}

func (uc *PackageUsecase) ListPackage(
	ctx context.Context,
	qp database.QueryParams,
) (int64, *[]PackageModel, *errors.Error) {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return 0, nil, errors.FromError(err)
	}

	uc.log.Info(
		"开始查询程序包列表",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)

	count, ms, err := uc.pkgRepo.ListModel(ctx, qp)
	if err != nil {
		uc.log.Error(
			"查询程序包列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		return 0, nil, database.NewGormError(err, nil)
	}

	uc.log.Info(
		"查询程序包列表成功",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)
	return count, ms, nil
}

func (uc *PackageUsecase) RemovePackage(ctx context.Context, m PackageModel) *errors.Error {
	if err := ctxutil.CheckContext(ctx); err != nil {
		return errors.FromError(err)
	}

	savePath := PackageStoragePath(m.StorageFilename)

	uc.log.Info(
		"开始删除程序包文件",
		zap.String("path", savePath),
		zap.Uint32(PackageIDKey, m.ID),
		zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
	)

	// 检查文件是否存在
	if _, statErr := os.Stat(savePath); os.IsNotExist(statErr) {
		// 文件不存在，视为删除成功
		uc.log.Warn(
			"程序包文件不存在，无需删除",
			zap.String("path", savePath),
			zap.Uint32(PackageIDKey, m.ID),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		return nil
	} else if statErr != nil {
		// 其他 stat 错误
		uc.log.Error(
			"检查程序包文件状态失败",
			zap.Error(statErr),
			zap.String("path", savePath),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		return errors.FromError(statErr)
	}

	// 执行删除操作
	if rmErr := os.Remove(savePath); rmErr != nil {
		uc.log.Error(
			"删除程序包失败",
			zap.Error(rmErr),
			zap.String("path", savePath),
			zap.Uint32(PackageIDKey, m.ID),
			zap.String(string(ctxutil.TraceIDKey), ctxutil.GetTraceID(ctx)),
		)
		return errors.FromError(rmErr)
	}

	uc.log.Info(
		"删除程序包文件成功",
		zap.String("path", savePath),
		zap.Uint32(PackageIDKey, m.ID))
	return nil
}

func PackageStoragePath(filename string) string {
	return filepath.Join(config.StorageDir, "packages", filename)
}
