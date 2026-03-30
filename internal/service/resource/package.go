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
	dto resomodel.UploadPackageBiz,
) (*resomodel.PackageModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info("创建程序包：开始执行")

	log.Debug(
		"创建程序包：输入参数",
		zap.Object("upload_package_biz", &dto),
	)

	newFileNameWithExt := uuid.NewString() + filepath.Ext(dto.Filename)
	m := resomodel.PackageModel{
		OriginFilename:  dto.Filename,
		StorageFilename: newFileNameWithExt,
		Label:           dto.Label,
		Version:         dto.Version,
	}

	createStepStart := time.Now()
	log.Debug(
		"创建程序包：开始创建数据库模型",
		zap.Object("package_model", &m),
	)
	if err := s.pkgRepo.CreateModel(ctx, &m); err != nil {
		log.Error(
			"创建程序包：创建数据库模型失败",
			zap.Error(err),
			zap.Object("package_model", &m),
			zap.Duration("create_step_duration", time.Since(createStepStart)),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, nil)
	}
	createStepDuration := time.Since(createStepStart)
	log.Debug(
		"创建程序包：创建数据库模型成功",
		zap.Uint32("package_id", m.ID),
		zap.Duration("create_step_duration", createStepDuration),
	)

	saveStepStart := time.Now()
	log.Debug("创建程序包：开始保存程序包文件")
	savePath := GetPackageStoragePath(newFileNameWithExt)
	if err := s.pkgRepo.SavePackageFile(ctx, dto.File, savePath, false); err != nil {
		log.Error(
			"创建程序包：程序包文件创建失败",
			zap.Error(err),
			zap.String("pkg_path", savePath),
			zap.Duration("save_step_duration", time.Since(saveStepStart)),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.ErrPackageSaveFailed.WithField("pkg_path", savePath)
	}
	saveStepDuration := time.Since(saveStepStart)
	log.Debug(
		"创建程序包：程序包文件创建成功",
		zap.String("pkg_path", savePath),
		zap.Duration("save_step_duration", saveStepDuration),
	)

	log.Info(
		"创建程序包：执行成功",
		zap.Uint32("package_id", m.ID),
		zap.Duration("create_step_duration", createStepDuration),
		zap.Duration("save_step_duration", saveStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return &m, nil
}

func (s *PackageService) DeletePackageByID(
	ctx context.Context,
	pkgId uint32,
) *errors.Error {
	if ctx.Err() != nil {
		return errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"删除程序包：开始执行",
		zap.Uint32("package_id", pkgId),
	)

	m, err := s.FindPackageByID(ctx, pkgId)
	if err != nil {
		log.Error(
			"删除程序包：查询程序包失败",
			zap.Error(err),
			zap.Uint32("package_id", pkgId),
		)
		return err
	}

	// 先从数据库删除
	deleteStepStart := time.Now()
	log.Debug(
		"删除程序包：开始删除数据库模型",
		zap.Uint32("package_id", pkgId),
	)
	if err := s.pkgRepo.DeleteModel(ctx, pkgId); err != nil {
		log.Error(
			"删除程序包：数据库删除失败",
			zap.Error(err),
			zap.Uint32("package_id", pkgId),
			zap.Duration("delete_step_duration", time.Since(deleteStepStart)),
		)
		return errors.NewGormError(err, map[string]any{"id": pkgId})
	}
	deleteStepDuration := time.Since(deleteStepStart)
	log.Debug(
		"删除程序包：数据库删除成功",
		zap.Uint32("package_id", pkgId),
		zap.Duration("delete_step_duration", deleteStepDuration),
	)

	// 再删除物理文件
	removeFileStepStart := time.Now()
	deletePath := GetPackageStoragePath(m.StorageFilename)
	log.Debug(
		"删除程序包：开始删除物理文件",
		zap.String("pkg_path", deletePath),
	)
	if rmErr := s.pkgRepo.RemovePackageFile(ctx, deletePath); rmErr != nil {
		log.Error(
			"删除程序包：删除物理文件失败",
			zap.Error(rmErr),
			zap.Uint32("package_id", pkgId),
			zap.String("pkg_path", deletePath),
			zap.Duration("remove_file_step_duration", time.Since(removeFileStepStart)),
		)
		return errors.ErrPackageRemoveFailed.WithField("pkg_path", deletePath)
	}
	removeFileStepDuration := time.Since(removeFileStepStart)
	log.Debug(
		"删除程序包：删除物理文件成功",
		zap.Uint32("package_id", pkgId),
		zap.String("pkg_path", deletePath),
		zap.Duration("remove_file_step_duration", removeFileStepDuration),
	)

	log.Info(
		"删除程序包：执行成功",
		zap.Uint32("package_id", pkgId),
		zap.Duration("delete_step_duration", deleteStepDuration),
		zap.Duration("remove_file_step_duration", removeFileStepDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

func (s *PackageService) FindPackageByID(
	ctx context.Context,
	pkgId uint32,
) (*resomodel.PackageModel, *errors.Error) {
	if ctx.Err() != nil {
		return nil, errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info(
		"查询程序包：开始执行",
		zap.Uint32("package_id", pkgId),
	)

	m, err := s.pkgRepo.GetModel(ctx, nil, pkgId)
	if err != nil {
		log.Error(
			"查询程序包：数据库查询失败",
			zap.Error(err),
			zap.Uint32("package_id", pkgId),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.NewGormError(err, map[string]any{"id": pkgId})
	}
	log.Debug(
		"查询程序包：查询到的数据库模型详情",
		zap.Object("package_model", m),
	)

	log.Info(
		"查询程序包：执行成功",
		zap.Uint32("package_id", pkgId),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return m, nil
}

func (s *PackageService) ListPackage(
	ctx context.Context,
	page, size int,
	dto resomodel.ListPackageDTO,
) (int64, []resomodel.PackageModel, *errors.Error) {
	if ctx.Err() != nil {
		return 0, nil, errors.FromError(ctx.Err())
	}

	startTime := time.Now()
	log := ctxutil.NewLogger(s.log, ctx)

	log.Info("查询程序包列表：开始执行")

	log.Debug(
		"查询程序包列表：参数详情",
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
	log.Debug(
		"查询程序包列表：数据库查询参数详情",
		zap.Object("query_params", &qp),
	)

	countStepStart := time.Now()
	log.Debug(
		"查询程序包列表：开始查询数据库模型总数",
		zap.Object("query_params", &qp),
	)
	count, err := s.pkgRepo.CountModel(ctx, qp.Query)
	countStepDuration := time.Since(countStepStart)
	if err != nil {
		log.Error(
			"查询程序包列表：查询数据库模型总数失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("count_step_duration", countStepDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}
	log.Debug(
		"查询程序包列表：查询数据库模型总数成功",
		zap.Int64("total_count", count),
		zap.Duration("count_step_duration", countStepDuration),
	)
	if count == 0 {
		log.Warn(
			"查询程序包列表：数据库模型总数为0",
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return count, nil, nil
	}

	listStepStart := time.Now()
	log.Debug(
		"查询程序包列表：开始查询数据库模型列表",
		zap.Object("query_params", &qp),
	)
	ms, err := s.pkgRepo.ListModel(ctx, qp)
	listStepDuration := time.Since(listStepStart)
	if err != nil {
		log.Error(
			"查询程序包列表：数据库查询失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("list_step_duration", listStepDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, nil, errors.NewGormError(err, nil)
	}
	log.Debug(
		"查询程序包列表：查询数据库模型列表成功",
		zap.Int("package_count", len(ms)),
		zap.Duration("list_step_duration", listStepDuration),
	)

	log.Info(
		"查询程序包列表：执行成功",
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return count, ms, nil
}

func GetPackageStoragePath(filename string) string {
	return filepath.Join(config.StorageDir, "packages", filename)
}
