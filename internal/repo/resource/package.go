package resource

import (
	"context"
	"io"
	"os"
	"time"

	"emperror.dev/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"

	resomodel "gin-artweb/internal/model/resource"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/database"
	"gin-artweb/pkg/fileutil"
)

// PackageRepo 程序包仓库实现
// 负责程序包模型的CRUD操作
// 使用GORM进行数据库操作
type PackageRepo struct {
	log      *zap.Logger       // 日志记录器
	gormDB   *gorm.DB          // GORM数据库连接
	timeouts *config.DBTimeout // 数据库操作超时配置
}

// NewPackageRepo 创建程序包仓库实例
//
// 参数:
//
//	log: 日志记录器，用于记录操作日志
//	gormDB: GORM数据库连接，用于执行数据库操作
//	timeouts: 数据库操作超时配置，控制各类数据库操作的超时时间
//
// 返回值:
//
//	*PackageRepo: 程序包仓库接口实现
func NewPackageRepo(
	log *zap.Logger,
	gormDB *gorm.DB,
	timeouts *config.DBTimeout,
) *PackageRepo {
	return &PackageRepo{
		log:      log,
		gormDB:   gormDB,
		timeouts: timeouts,
	}
}

// CreateModel 创建程序包模型
//
// 参数:
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	m: 程序包模型，包含程序包的详细信息
//
// 返回值:
//
//	error: 操作错误信息，成功则返回nil
//
// 功能:
//  1. 检查程序包模型是否为空
//  2. 设置上传时间
//  3. 执行数据库创建操作
//  4. 记录操作日志
func (r *PackageRepo) CreateModel(
	ctx context.Context,
	m *resomodel.PackageModel,
) error {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	// 检查参数
	if m == nil {
		err := errors.New("创建程序包模型:模型不能为空")
		log.Error(
			"创建程序包模型:模型不能为空",
			zap.Error(err),
		)
		return err
	}
	log.Debug(
		"创建程序包模型:开始执行",
		zap.Object("package_model", m),
	)

	m.UploadedAt = startTime
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	createPackageStartTime := time.Now()
	err := database.DBCreate(dbCtx, r.gormDB, &resomodel.PackageModel{}, m, nil)
	createPackageDuration := time.Since(createPackageStartTime)
	if err != nil {
		log.Error(
			"创建程序包模型:数据库操作失败",
			zap.Error(err),
			zap.Object("package_model", m),
			zap.Duration("create_duration", createPackageDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "创建程序包模型:数据库操作失败")
	}
	log.Debug(
		"创建程序包模型:执行成功",
		zap.Object("package_model", m),
		zap.Duration("create_duration", createPackageDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

// DeleteModel 删除程序包模型
//
// 参数:
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	conds: 查询条件，用于指定要删除的记录
//
// 返回值:
//
//	error: 操作错误信息，成功则返回nil
//
// 功能:
//  1. 执行数据库删除操作
//  2. 记录操作日志
func (r *PackageRepo) DeleteModel(
	ctx context.Context,
	conds ...any,
) error {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"删除程序包模型:开始执行",
		zap.Any("conds", conds),
	)
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	deletePackageStartTime := time.Now()
	err := database.DBDelete(dbCtx, r.gormDB, &resomodel.PackageModel{}, conds...)
	deletePackageDuration := time.Since(deletePackageStartTime)
	if err != nil {
		log.Error(
			"删除程序包模型:数据库操作失败",
			zap.Error(err),
			zap.Any("conds", conds),
			zap.Duration("delete_duration", deletePackageDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "删除程序包模型:数据库操作失败")
	}
	log.Debug(
		"删除程序包模型:执行成功",
		zap.Any("conds", conds),
		zap.Duration("delete_duration", deletePackageDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

// GetModel 查询单个程序包模型
//
// 参数:
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	preloads: 需要预加载的关联关系
//	conds: 查询条件，用于指定要查询的记录
//
// 返回值:
//
//	*resomodel.PackageModel: 程序包模型指针，包含程序包的详细信息
//	error: 操作错误信息，成功则返回nil
//
// 功能:
//  1. 执行数据库查询操作
//  2. 预加载关联字段
//  3. 获取单个程序包模型
//  4. 记录操作日志
func (r *PackageRepo) GetModel(
	ctx context.Context,
	preloads []string,
	conds ...any,
) (*resomodel.PackageModel, error) {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"查询程序包模型:开始执行",
		zap.Any("conds", conds),
	)

	var m resomodel.PackageModel
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()
	getPackageStartTime := time.Now()
	err := database.DBGet(dbCtx, r.gormDB, preloads, &m, conds...)
	getPackageDuration := time.Since(getPackageStartTime)
	if err != nil {
		log.Error(
			"查询程序包模型:数据库操作失败",
			zap.Error(err),
			zap.Any("conds", conds),
			zap.Duration("get_duration", getPackageDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.WrapIf(err, "查询程序包模型:数据库操作失败")
	}
	log.Debug(
		"查询程序包模型:执行成功",
		zap.Object("package_model", &m),
		zap.Any("conds", conds),
		zap.Duration("get_duration", getPackageDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return &m, nil
}

// ListModel 查询程序包模型列表
//
// 参数:
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	qp: 查询参数，包含分页、排序等查询条件
//
// 返回值:
//
//	int64: 总记录数
//	*[]resomodel.PackageModel: 程序包模型列表指针，包含符合条件的程序包模型
//	error: 操作错误信息，成功则返回nil
//
// 功能:
//  1. 执行数据库查询操作
//  2. 获取程序包模型列表
//  3. 返回总记录数和模型列表
//  4. 记录操作日志
func (r *PackageRepo) ListModel(
	ctx context.Context,
	qp database.QueryParams,
) ([]resomodel.PackageModel, error) {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"查询程序包模型列表:开始执行",
		zap.Object("query_params", &qp),
	)

	var ms []resomodel.PackageModel
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ListTimeout)
	defer cancel()
	listPackageStartTime := time.Now()
	err := database.DBList(dbCtx, r.gormDB, &resomodel.PackageModel{}, &ms, qp)
	listPackageDuration := time.Since(listPackageStartTime)
	if err != nil {
		log.Error(
			"查询程序包模型列表:数据库操作失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("list_package_duration", listPackageDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.WrapIf(err, "查询程序包模型列表:数据库操作失败")
	}
	log.Debug(
		"查询程序包模型列表:执行成功",
		zap.Object("query_params", &qp),
		zap.Duration("list_package_duration", listPackageDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return ms, nil
}

func (r *PackageRepo) CountModel(
	ctx context.Context,
	query map[string]any,
) (int64, error) {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"查询程序包模型总数:开始执行",
		zap.Any("query", query),
	)
	// 开启数据库事务
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()
	countPackageStartTime := time.Now()
	count, err := database.DBCount(dbCtx, r.gormDB, &resomodel.PackageModel{}, query)
	countPackageDuration := time.Since(countPackageStartTime)
	if err != nil {
		log.Error(
			"查询程序包模型总数:数据库查询失败",
			zap.Error(err),
			zap.Any("query", query),
			zap.Duration("count_package_duration", countPackageDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, errors.WrapIf(err, "查询程序包模型总数:数据库查询失败")
	}
	log.Debug(
		"查询程序包模型总数:执行成功",
		zap.Any("query", query),
		zap.Int64("total_count", count),
		zap.Duration("count_package_duration", countPackageDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return count, nil
}

func (r *PackageRepo) SavePackageFile(
	ctx context.Context,
	fileReader io.Reader,
	pkgPath string,
	overwrite bool,
) error {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"保存程序包:开始执行",
		zap.String("pkg_path", pkgPath),
		zap.Bool("overwrite", overwrite),
	)

	savePackageStartTime := time.Now()
	err := fileutil.WriteReaderToFile(ctx, fileReader, pkgPath, os.FileMode(0o644), overwrite)
	savePackageDuration := time.Since(savePackageStartTime)
	if err != nil {
		log.Error(
			"保存程序包:文件流写入失败",
			zap.Error(err),
			zap.String("pkg_path", pkgPath),
			zap.Bool("overwrite", overwrite),
			zap.Duration("save_package_duration", savePackageDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "保存程序包:文件流写入失败")
	}
	log.Debug(
		"保存程序包:执行成功",
		zap.String("pkg_path", pkgPath),
		zap.Bool("overwrite", overwrite),
		zap.Duration("save_package_duration", savePackageDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

func (r *PackageRepo) RemovePackageFile(
	ctx context.Context,
	pkgPath string,
) error {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"删除程序包文件:开始执行",
		zap.String("pkg_path", pkgPath),
	)

	removePackageStartTime := time.Now()
	err := fileutil.Remove(ctx, pkgPath)
	removePackageDuration := time.Since(removePackageStartTime)
	if err != nil {
		log.Error(
			"删除程序包文件:文件删除失败",
			zap.Error(err),
			zap.String("pkg_path", pkgPath),
			zap.Duration("remove_package_duration", removePackageDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "删除程序包文件:文件删除失败")
	}
	log.Debug(
		"删除程序包文件:执行成功",
		zap.String("pkg_path", pkgPath),
		zap.Duration("remove_package_duration", removePackageDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}
