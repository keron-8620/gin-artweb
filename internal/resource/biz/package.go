package biz

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"gin-artweb/pkg/database"
	"gin-artweb/pkg/errors"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type PackageModel struct {
	database.BaseModel
	Label           string    `gorm:"column:label;type:varchar(50);index:idx_member;comment:标签" json:"label"`
	StorageFilename string    `gorm:"column:storage_filename;type:varchar(50);not null;uniqueIndex;comment:磁盘存储文件名" json:"storage_filename"`
	OriginFilename  string    `gorm:"column:origin_filename;type:varchar(255);comment:原始文件名" json:"origin_filename"`
	Version         string    `gorm:"column:version;type:varchar(50);comment:版本号" json:"version"`
	UploadedAt      time.Time `gorm:"column:uploaded_at;type:varchar(254);comment:上传时间" json:"uploaded_at"`
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
	ListModel(context.Context, database.QueryParams) (int64, []PackageModel, error)
}

type PackageUsecase struct {
	log     *zap.Logger
	pkgRepo PackageRepo
	dir     string
}

func NewPackageUsecase(
	log *zap.Logger,
	pkgRepo PackageRepo,
	dir string,
) *PackageUsecase {
	return &PackageUsecase{
		log:     log,
		pkgRepo: pkgRepo,
		dir:     dir,
	}
}

func (uc *PackageUsecase) CreatePackage(
	ctx context.Context,
	m PackageModel,
) (*PackageModel, *errors.Error) {
	if err := uc.pkgRepo.CreateModel(ctx, &m); err != nil {
		return nil, database.NewGormError(err, nil)
	}
	return &m, nil
}

func (uc *PackageUsecase) DeletePackageById(
	ctx context.Context,
	pkgId uint32,
) *errors.Error {
	m, err := uc.FindPackageById(ctx, pkgId)
	if err != nil {
		return err
	}

	// 先从数据库删除
	if err := uc.pkgRepo.DeleteModel(ctx, pkgId); err != nil {
		return database.NewGormError(err, map[string]any{"id": pkgId})
	}

	// 再删除物理文件
	if rmErr := uc.RemovePackage(ctx, *m); rmErr != nil {
		uc.log.Warn("程序包数据库记录已删除，但物理文件删除失败",
			zap.Uint32("id", pkgId),
			zap.String("storage_filename", m.StorageFilename))
	}
	return nil
}

func (uc *PackageUsecase) FindPackageById(
	ctx context.Context,
	pkgId uint32,
) (*PackageModel, *errors.Error) {
	m, err := uc.pkgRepo.FindModel(ctx, nil, pkgId)
	if err != nil {
		return nil, database.NewGormError(err, map[string]any{"id": pkgId})
	}
	return m, nil
}

func (uc *PackageUsecase) ListPackage(
	ctx context.Context,
	page, size int,
	query map[string]any,
	orderBy []string,
	isCount bool,
) (int64, []PackageModel, *errors.Error) {
	qp := database.QueryParams{
		Preloads: []string{},
		Query:    query,
		OrderBy:  orderBy,
		Limit:    max(size, 0),
		Offset:   max(page-1, 0),
		IsCount:  isCount,
	}
	count, ms, err := uc.pkgRepo.ListModel(ctx, qp)
	if err != nil {
		return 0, nil, database.NewGormError(err, nil)
	}
	return count, ms, nil
}

func (uc *PackageUsecase) PackagePath(filename string) string {
	return filepath.Join(uc.dir, filename)
}

func (uc *PackageUsecase) RemovePackage(ctx context.Context, m PackageModel) *errors.Error {
	savePath := uc.PackagePath(m.StorageFilename)

	// 检查文件是否存在
	if _, statErr := os.Stat(savePath); os.IsNotExist(statErr) {
		// 文件不存在，视为删除成功
		uc.log.Debug(
			"程序包文件不存在，无需删除",
			zap.String("path", savePath),
			zap.Uint32("package_id", m.ID))
		return nil
	} else if statErr != nil {
		// 其他 stat 错误
		uc.log.Error(
			"检查程序包文件状态失败",
			zap.String("path", savePath),
			zap.Error(statErr))
		return errors.FromError(statErr)
	}

	// 执行删除操作
	if rmErr := os.Remove(savePath); rmErr != nil {
		uc.log.Error("删除程序包失败",
			zap.String("path", savePath),
			zap.Uint32("package_id", m.ID),
			zap.Error(rmErr))
		return errors.FromError(rmErr)
	}

	uc.log.Info(
		"成功删除程序包文件",
		zap.String("path", savePath),
		zap.Uint32("package_id", m.ID))
	return nil
}
