package resource

import (
	"context"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	resomodel "gin-artweb/internal/model/resource"
	resorepo "gin-artweb/internal/repo/resource"
	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
)

type PackageService struct {
	log        *zap.Logger
	pkgRepo    *resorepo.PackageRepo
	storageDir string
}

func NewPackageService(
	log *zap.Logger,
	pkgRepo *resorepo.PackageRepo,
	storageDir string,
) *PackageService {
	return &PackageService{
		log:        log,
		pkgRepo:    pkgRepo,
		storageDir: storageDir,
	}
}

func (s *PackageService) CreatePackage(
	ctx context.Context,
	dto resomodel.CreatePackageDTO,
) (*resomodel.PackageModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"创建程序包:开始执行",
		zap.Object("create_package_dto", &dto),
	)

	newFileNameWithExt := uuid.NewString() + filepath.Ext(dto.Filename)
	savePath := GetPackageStoragePath(newFileNameWithExt)
	if err := s.pkgRepo.SavePackageFile(ctx, dto.File, savePath, false); err != nil {
		log.Error(
			"创建程序包:程序包文件创建失败",
			zap.Error(err),
			zap.String("save_path", savePath),
			zap.Bool("overwrite", false),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.ErrPackageSaveFailed.WithField("pkg_path", savePath)
	}

	clearPackage := func() {
		if err := s.pkgRepo.RemovePackageFile(ctx, savePath); err != nil {
			log.Error(
				"创建程序包:删除缓存的程序包文件失败,请手动清理",
				zap.Error(err),
				zap.String("save_path", savePath),
				zap.Duration("total_duration", time.Since(startTime)),
			)
		}
	}

	m := resomodel.PackageModel{
		OriginFilename:  dto.Filename,
		StorageFilename: newFileNameWithExt,
		Label:           dto.Label,
		Version:         dto.Version,
	}

	if err := s.pkgRepo.CreateModel(ctx, &m); err != nil {
		log.Error(
			"创建程序包:创建数据库模型失败",
			zap.Error(err),
			zap.Object("package_model", &m),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		clearPackage()
		return nil, errors.NewGormError(err, nil)
	}

	log.Info(
		"创建程序包:执行成功",
		zap.Uint32("package_id", m.ID),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return &m, nil
}

func (s *PackageService) DeletePackageByID(
	ctx context.Context,
	pkgId uint32,
) *errors.Error {
	startTime := time.Now()
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"删除程序包:开始执行",
		zap.Uint32("package_id", pkgId),
	)

	m, err := s.FindPackageByID(ctx, pkgId)
	if err != nil {
		log.Error(
			"删除程序包:查询程序包失败",
			zap.Error(err),
			zap.Uint32("package_id", pkgId),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return err
	}

	if err := s.pkgRepo.DeleteModel(ctx, pkgId); err != nil {
		log.Error(
			"删除程序包:数据库删除失败",
			zap.Error(err),
			zap.Uint32("package_id", pkgId),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.NewGormError(err, map[string]any{"id": pkgId})
	}

	deletePath := GetPackageStoragePath(m.StorageFilename)
	if rmErr := s.pkgRepo.RemovePackageFile(ctx, deletePath); rmErr != nil {
		log.Error(
			"删除程序包:删除物理文件失败",
			zap.Error(rmErr),
			zap.Uint32("package_id", pkgId),
			zap.String("pkg_path", deletePath),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.ErrPackageRemoveFailed.WithField("pkg_path", deletePath)
	}

	log.Info(
		"删除程序包:执行成功",
		zap.Uint32("package_id", pkgId),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

func (s *PackageService) FindPackageByID(
	ctx context.Context,
	pkgId uint32,
) (*resomodel.PackageModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	m, err := s.pkgRepo.GetModel(ctx, pkgId)
	if err != nil {
		log.Error(
			"查询程序包:数据库查询失败",
			zap.Error(err),
			zap.Uint32("package_id", pkgId),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, map[string]any{"id": pkgId})
	}

	log.Debug(
		"查询程序包:查询到的数据库模型详情",
		zap.Object("package_model", m),
	)
	return m, nil
}

func (s *PackageService) ListPackage(
	ctx context.Context,
	page, size int,
	dto resomodel.ListPackageDTO,
) (int64, []resomodel.PackageModel, *errors.Error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return 0, nil, errors.FromError(ctx.Err())
	}
	log := ctxutil.NewLogger(s.log, ctx)

	log.Debug(
		"查询程序包列表:参数详情",
		zap.Int("page", page),
		zap.Int("size", size),
		zap.Object("list_package_dto", &dto),
	)

	limit, offset := common.Page2LimitOffset(page, size)
	qp := database.QueryParams{
		Limit:   limit,
		Offset:  offset,
		OrderBy: []string{"id ASC"},
		Query:   dto.ToQueryMap(),
	}

	count, err := s.pkgRepo.CountModel(ctx, qp.Query)
	if err != nil {
		log.Error(
			"查询程序包列表:查询数据库模型总数失败",
			zap.Error(err),
			zap.Any("query", qp.Query),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}

	if count == 0 {
		log.Warn(
			"查询程序包列表:数据库模型总数为0",
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, nil, nil
	}

	ms, err := s.pkgRepo.ListModel(ctx, qp)
	if err != nil {
		log.Error(
			"查询程序包列表:数据库查询失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}
	return count, ms, nil
}

func GetPackageStoragePath(filename string) string {
	return filepath.Join(config.StorageDir, "packages", filename)
}
