package resource

import (
	"context"
	"os"

	"go.uber.org/zap"

	resomodel "gin-artweb/internal/model/resource"
	resorepo "gin-artweb/internal/repository/resource"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
)

type PackageService struct {
	log     *zap.Logger
	pkgRepo *resorepo.PackageRepo
}

func NewPackageService(
	log *zap.Logger,
	pkgRepo *resorepo.PackageRepo,
) *PackageService {
	return &PackageService{
		log:     log,
		pkgRepo: pkgRepo,
	}
}

func (s *PackageService) CreatePackage(
	ctx context.Context,
	m resomodel.PackageModel,
) (*resomodel.PackageModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始创建程序包",
		zap.Object(database.ModelKey, &m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	if err := s.pkgRepo.CreateModel(ctx, &m); err != nil {
		s.log.Error(
			"创建程序包失败",
			zap.Error(err),
			zap.Object(database.ModelKey, &m),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, nil)
	}

	s.log.Info(
		"创建程序包成功",
		zap.Object(database.ModelKey, &m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return &m, nil
}

func (s *PackageService) DeletePackageById(
	ctx context.Context,
	pkgId uint32,
) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始删除程序包",
		zap.Uint32("package_id", pkgId),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, err := s.FindPackageById(ctx, pkgId)
	if err != nil {
		return err
	}

	// 先从数据库删除
	if err := s.pkgRepo.DeleteModel(ctx, pkgId); err != nil {
		s.log.Error(
			"删除程序包失败",
			zap.Error(err),
			zap.Uint32("package_id", pkgId),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.NewGormError(err, map[string]any{"id": pkgId})
	}

	// 再删除物理文件
	if rmErr := s.RemovePackage(ctx, *m); rmErr != nil {
		return rmErr
	}

	s.log.Info(
		"删除程序包成功",
		zap.Uint32("package_id", pkgId),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return nil
}

func (s *PackageService) FindPackageById(
	ctx context.Context,
	pkgId uint32,
) (*resomodel.PackageModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始查询程序包",
		zap.Uint32("package_id", pkgId),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	m, err := s.pkgRepo.GetModel(ctx, nil, pkgId)
	if err != nil {
		s.log.Error(
			"查询程序包失败",
			zap.Error(err),
			zap.Uint32("package_id", pkgId),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil, errors.NewGormError(err, map[string]any{"id": pkgId})
	}

	s.log.Info(
		"查询程序包成功",
		zap.Uint32("package_id", pkgId),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return m, nil
}

func (s *PackageService) ListPackage(
	ctx context.Context,
	qp database.QueryParams,
) (int64, *[]resomodel.PackageModel, *errors.Error) {
	if ctx.Err() != nil {
		return 0, nil, errors.FromError(ctx.Err())
	}

	s.log.Info(
		"开始查询程序包列表",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	count, ms, err := s.pkgRepo.ListModel(ctx, qp)
	if err != nil {
		s.log.Error(
			"查询程序包列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}

	s.log.Info(
		"查询程序包列表成功",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	return count, ms, nil
}

func (s *PackageService) RemovePackage(ctx context.Context, m resomodel.PackageModel) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	savePath := common.GetPackageStoragePath(m.StorageFilename)

	s.log.Info(
		"开始删除程序包文件",
		zap.String("path", savePath),
		zap.Uint32("package_id", m.ID),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	// 检查文件是否存在
	if _, statErr := os.Stat(savePath); os.IsNotExist(statErr) {
		// 文件不存在，视为删除成功
		s.log.Warn(
			"程序包文件不存在，无需删除",
			zap.String("path", savePath),
			zap.Uint32("package_id", m.ID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return nil
	} else if statErr != nil {
		// 其他 stat 错误
		s.log.Error(
			"检查程序包文件状态失败",
			zap.Error(statErr),
			zap.String("path", savePath),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.FromError(statErr)
	}

	// 执行删除操作
	if rmErr := os.Remove(savePath); rmErr != nil {
		s.log.Error(
			"删除程序包失败",
			zap.Error(rmErr),
			zap.String("path", savePath),
			zap.Uint32("package_id", m.ID),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return errors.FromError(rmErr)
	}

	s.log.Info(
		"删除程序包文件成功",
		zap.String("path", savePath),
		zap.Uint32("package_id", m.ID))
	return nil
}
