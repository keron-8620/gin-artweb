package biz

import (
	"context"
	"os"
	"path/filepath"

	"go.uber.org/zap"

	"gin-artweb/internal/infra/resource/data"
	"gin-artweb/internal/infra/resource/model"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
)

type PackageUsecase struct {
	log     *zap.Logger
	pkgRepo *data.PackageRepo
}

func NewPackageUsecase(
	log *zap.Logger,
	pkgRepo *data.PackageRepo,
) *PackageUsecase {
	return &PackageUsecase{
		log:     log,
		pkgRepo: pkgRepo,
	}
}

func (uc *PackageUsecase) CreatePackage(
	ctx context.Context,
	m model.PackageModel,
) (*model.PackageModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	uc.log.Info(
		"开始创建程序包",
		zap.Object(database.ModelKey, &m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	if err := uc.pkgRepo.CreateModel(ctx, &m); err != nil {
		uc.log.Error(
			"创建程序包失败",
			zap.Error(err),
			zap.Object(database.ModelKey, &m),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, nil)
	}

	uc.log.Info(
		"创建程序包成功",
		zap.Object(database.ModelKey, &m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return &m, nil
}

func (uc *PackageUsecase) DeletePackageById(
	ctx context.Context,
	pkgId uint32,
) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	uc.log.Info(
		"开始删除程序包",
		zap.Uint32("package_id", pkgId),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
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
			zap.Uint32("package_id", pkgId),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.NewGormError(err, map[string]any{"id": pkgId})
	}

	// 再删除物理文件
	if rmErr := uc.RemovePackage(ctx, *m); rmErr != nil {
		return rmErr
	}

	uc.log.Info(
		"删除程序包成功",
		zap.Uint32("package_id", pkgId),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return nil
}

func (uc *PackageUsecase) FindPackageById(
	ctx context.Context,
	pkgId uint32,
) (*model.PackageModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	uc.log.Info(
		"开始查询程序包",
		zap.Uint32("package_id", pkgId),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, err := uc.pkgRepo.GetModel(ctx, nil, pkgId)
	if err != nil {
		uc.log.Error(
			"查询程序包失败",
			zap.Error(err),
			zap.Uint32("package_id", pkgId),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, map[string]any{"id": pkgId})
	}

	uc.log.Info(
		"查询程序包成功",
		zap.Uint32("package_id", pkgId),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return m, nil
}

func (uc *PackageUsecase) ListPackage(
	ctx context.Context,
	qp database.QueryParams,
) (int64, *[]model.PackageModel, *errors.Error) {
	if ctx.Err() != nil {
		return 0, nil, errors.FromError(ctx.Err())
	}

	uc.log.Info(
		"开始查询程序包列表",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	count, ms, err := uc.pkgRepo.ListModel(ctx, qp)
	if err != nil {
		uc.log.Error(
			"查询程序包列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}

	uc.log.Info(
		"查询程序包列表成功",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return count, ms, nil
}

func (uc *PackageUsecase) RemovePackage(ctx context.Context, m model.PackageModel) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	savePath := PackageStoragePath(m.StorageFilename)

	uc.log.Info(
		"开始删除程序包文件",
		zap.String("path", savePath),
		zap.Uint32("package_id", m.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	// 检查文件是否存在
	if _, statErr := os.Stat(savePath); os.IsNotExist(statErr) {
		// 文件不存在，视为删除成功
		uc.log.Warn(
			"程序包文件不存在，无需删除",
			zap.String("path", savePath),
			zap.Uint32("package_id", m.ID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil
	} else if statErr != nil {
		// 其他 stat 错误
		uc.log.Error(
			"检查程序包文件状态失败",
			zap.Error(statErr),
			zap.String("path", savePath),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.FromError(statErr)
	}

	// 执行删除操作
	if rmErr := os.Remove(savePath); rmErr != nil {
		uc.log.Error(
			"删除程序包失败",
			zap.Error(rmErr),
			zap.String("path", savePath),
			zap.Uint32("package_id", m.ID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.FromError(rmErr)
	}

	uc.log.Info(
		"删除程序包文件成功",
		zap.String("path", savePath),
		zap.Uint32("package_id", m.ID))
	return nil
}

func PackageStoragePath(filename string) string {
	return filepath.Join(config.StorageDir, "packages", filename)
}
