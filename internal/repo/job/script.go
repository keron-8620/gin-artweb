package job

import (
	"context"
	"io"
	"os"
	"time"

	"emperror.dev/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"

	jobmodel "gin-artweb/internal/model/job"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/database"
	"gin-artweb/pkg/fileutil"
)

// ScriptRepo 脚本仓库实现
// 负责脚本模型的CRUD操作
// 使用GORM进行数据库操作
type ScriptRepo struct {
	log           *zap.Logger             // 日志记录器
	gormDB        *gorm.DB                // GORM数据库连接
	timeouts      *config.DBTimeout       // 数据库操作超时配置
	slowThreshold *config.DBSlowThreshold // 数据库操作慢查询阈值配置
}

// NewScriptRepo 创建脚本仓库实例
//
// 参数:
//
//	log: 日志记录器，用于记录操作日志
//	gormDB: GORM数据库连接，用于执行数据库操作
//	timeouts: 数据库操作超时配置，控制各类数据库操作的超时时间
//	slowThreshold: 数据库操作慢查询阈值配置
//
// 返回值:
//
//	*ScriptRepo: 脚本仓库实例
func NewScriptRepo(
	log *zap.Logger,
	gormDB *gorm.DB,
	timeouts *config.DBTimeout,
	slowThreshold *config.DBSlowThreshold,
) *ScriptRepo {
	return &ScriptRepo{
		log:           log,
		gormDB:        gormDB,
		timeouts:      timeouts,
		slowThreshold: slowThreshold,
	}
}

// CreateModel 创建脚本模型
//
// 参数:
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	m: 脚本模型，包含脚本的详细信息
//
// 返回值:
//
//	error: 操作错误信息，成功则返回nil
//
// 功能:
//  1. 检查脚本模型是否为空
//  2. 设置创建时间和更新时间
//  3. 执行数据库创建操作
//  4. 记录操作日志
func (r *ScriptRepo) CreateModel(
	ctx context.Context,
	m *jobmodel.ScriptModel,
) error {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	// 检查参数
	if m == nil {
		err := errors.New("创建脚本模型:模型不能为空")
		log.Error(
			"创建脚本模型:模型不能为空",
			zap.Error(err),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return err
	}

	m.CreatedAt = startTime
	m.UpdatedAt = startTime

	log.Debug(
		"创建脚本模型:模型详情",
		zap.Object("script_model", m),
	)

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()

	createStartTime := time.Now()
	err := database.DBCreate(dbCtx, r.gormDB, &jobmodel.ScriptModel{}, m)
	createDuration := time.Since(createStartTime)
	if err != nil {
		log.Error(
			"创建脚本模型:数据库操作失败",
			zap.Error(err),
			zap.Object("script_model", m),
			zap.Duration("create_duration", createDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "创建脚本模型:数据库操作失败")
	}

	if createDuration > r.slowThreshold.WriteSlow {
		log.Warn("创建脚本模型:数据库创建耗时超过慢查询阈值，可能影响性能",
			zap.Uint32("script_id", m.ID),
			zap.Duration("create_duration", createDuration),
			zap.Duration("threshold", r.slowThreshold.WriteSlow),
		)
	}
	return nil
}

// UpdateModel 更新脚本模型
//
// 参数:
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	updateData: 更新数据，包含要更新的字段和值
//	conds: 查询条件，用于指定要更新的记录
//
// 返回值:
//
//	error: 操作错误信息，成功则返回nil
//
// 功能:
//  1. 检查更新数据是否为空
//  2. 执行数据库更新操作
//  3. 记录操作日志
func (r *ScriptRepo) UpdateModel(
	ctx context.Context,
	updateData map[string]any,
	conds ...any,
) error {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	if len(updateData) == 0 {
		err := errors.New("更新脚本模型:更新数据为空")
		log.Error(
			"更新脚本模型:更新数据为空",
			zap.Error(err),
			zap.Any("update_data", updateData),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return err
	}

	updateData["updated_at"] = startTime

	log.Debug(
		"更新脚本模型:更新数据",
		zap.Any("update_data", updateData),
		zap.Any("conds", conds),
	)

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()

	updateStartTime := time.Now()
	err := database.DBUpdateTx(dbCtx, r.gormDB, &jobmodel.ScriptModel{}, updateData, conds...)
	updateDuration := time.Since(updateStartTime)
	if err != nil {
		log.Error(
			"更新脚本模型:数据库操作失败",
			zap.Error(err),
			zap.Any("update_data", updateData),
			zap.Any("conds", conds),
			zap.Duration("update_duration", updateDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "更新脚本模型:数据库操作失败")
	}

	if updateDuration > r.slowThreshold.WriteSlow {
		log.Warn("更新脚本模型:数据库更新耗时超过慢查询阈值，可能影响性能",
			zap.Duration("update_duration", updateDuration),
			zap.Duration("threshold", r.slowThreshold.WriteSlow),
		)
	}
	return nil
}

// DeleteModel 删除脚本模型
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
func (r *ScriptRepo) DeleteModel(
	ctx context.Context,
	conds ...any,
) error {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"删除脚本模型:删除条件",
		zap.Any("conds", conds),
	)

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()

	deleteStartTime := time.Now()
	err := database.DBDeleteTx(dbCtx, r.gormDB, &jobmodel.ScriptModel{}, conds...)
	deleteDuration := time.Since(deleteStartTime)
	if err != nil {
		log.Error(
			"删除脚本模型:数据库操作失败",
			zap.Error(err),
			zap.Any("conds", conds),
			zap.Duration("delete_duration", deleteDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "删除脚本模型:数据库操作失败")
	}

	if deleteDuration > r.slowThreshold.WriteSlow {
		log.Warn("删除脚本模型:数据库删除耗时超过慢查询阈值，可能影响性能",
			zap.Duration("delete_duration", deleteDuration),
			zap.Duration("threshold", r.slowThreshold.WriteSlow),
		)
	}
	return nil
}

// GetModel 查询单个脚本模型
//
// 参数:
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	conds: 查询条件，用于指定要查询的记录
//
// 返回值:
//
//	*jobmodel.ScriptModel: 脚本模型指针，包含脚本的详细信息
//	error: 操作错误信息，成功则返回nil
//
// 功能:
//  1. 执行数据库查询操作
//  2. 获取单个脚本模型
//  3. 记录操作日志
func (r *ScriptRepo) GetModel(
	ctx context.Context,
	conds ...any,
) (*jobmodel.ScriptModel, error) {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"查询脚本模型:查询条件",
		zap.Any("conds", conds),
	)

	var m jobmodel.ScriptModel

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()

	getStartTime := time.Now()
	err := database.DBGet(dbCtx, r.gormDB, nil, &m, conds...)
	getDuration := time.Since(getStartTime)
	if err != nil {
		log.Error(
			"查询脚本模型:数据库操作失败",
			zap.Error(err),
			zap.Any("conds", conds),
			zap.Duration("get_duration", getDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.WrapIf(err, "查询脚本模型:数据库操作失败")
	}

	log.Debug(
		"查询脚本模型:查询到的模型详情",
		zap.Object("script_model", &m),
	)

	if getDuration > r.slowThreshold.ReadSlow {
		log.Warn("查询脚本模型:数据库查询耗时超过慢查询阈值，可能影响性能",
			zap.Duration("get_duration", getDuration),
			zap.Duration("threshold", r.slowThreshold.ReadSlow),
		)
	}
	return &m, nil
}

// ListModel 查询脚本模型列表
//
// 参数:
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	qp: 查询参数，包含分页、排序等查询条件
//
// 返回值:
//
//	int64: 总记录数
//	*[]jobmodel.ScriptModel: 脚本模型列表指针，包含符合条件的脚本模型
//	error: 操作错误信息，成功则返回nil
//
// 功能:
//  1. 执行数据库查询操作
//  2. 获取脚本模型列表
//  3. 返回总记录数和模型列表
//  4. 记录操作日志
func (r *ScriptRepo) ListModel(
	ctx context.Context,
	qp database.QueryParams,
) ([]jobmodel.ScriptModel, error) {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"查询脚本模型列表:入参详情",
		zap.Object("query_params", &qp),
	)

	var ms []jobmodel.ScriptModel

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ListTimeout)
	defer cancel()

	listStartTime := time.Now()
	err := database.DBList(dbCtx, r.gormDB, &jobmodel.ScriptModel{}, &ms, qp)
	listDuration := time.Since(listStartTime)
	if err != nil {
		log.Error(
			"查询脚本模型列表:数据库操作失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("list_duration", listDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.WrapIf(err, "查询脚本模型列表:数据库操作失败")
	}

	log.Debug(
		"查询脚本模型列表:查询到的模型列表",
		zap.Uint32s("script_ids", jobmodel.ListScriptModelToUint32s(ms)),
	)

	if listDuration > r.slowThreshold.ListSlow {
		log.Warn("查询脚本模型列表:数据库查询耗时超过慢查询阈值，可能影响性能",
			zap.Duration("list_duration", listDuration),
			zap.Duration("threshold", r.slowThreshold.ListSlow),
		)
	}
	return ms, nil
}

func (r *ScriptRepo) CountModel(
	ctx context.Context,
	query map[string]any,
) (int64, error) {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"查询API模型总数:查询条件",
		zap.Any("query", query),
	)

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()

	countStartTime := time.Now()
	count, err := database.DBCount(dbCtx, r.gormDB, &jobmodel.ScriptModel{}, query)
	countDuration := time.Since(countStartTime)
	if err != nil {
		log.Error(
			"查询脚本模型总数:数据库操作失败",
			zap.Error(err),
			zap.Any("query", query),
			zap.Duration("count_duration", countDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, errors.WrapIf(err, "查询脚本模型总数:数据库操作失败")
	}

	log.Debug(
		"查询脚本模型总数:查询到的记录数",
		zap.Int64("count", count),
	)

	if countDuration > r.slowThreshold.ReadSlow {
		log.Warn("查询脚本模型总数:数据库查询耗时超过慢查询阈值，可能影响性能",
			zap.Duration("count_duration", countDuration),
			zap.Duration("threshold", r.slowThreshold.ReadSlow),
		)
	}
	return count, nil
}

func (r *ScriptRepo) SaveScriptFile(
	ctx context.Context,
	fileReader io.Reader,
	scriptPath string,
	overwrite bool,
) error {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"保存脚本文件:入参详情",
		zap.String("script_path", scriptPath),
		zap.Bool("overwrite", overwrite),
	)

	saveStartTime := time.Now()
	err := fileutil.WriteReaderToFile(ctx, fileReader, scriptPath, os.FileMode(0o750), overwrite)
	saveDuration := time.Since(saveStartTime)
	if err != nil {
		log.Error(
			"保存脚本文件:文件写入失败",
			zap.Error(err),
			zap.String("script_path", scriptPath),
			zap.Duration("save_duration", saveDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "保存脚本文件:文件写入失败")
	}
	return nil
}

func (r *ScriptRepo) RemoveScriptFile(
	ctx context.Context,
	scriptPath string,
) error {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"删除脚本文件:入参详情",
		zap.String("script_path", scriptPath),
	)

	// 删除文件
	removeStartTime := time.Now()
	err := fileutil.Remove(ctx, scriptPath)
	removeDuration := time.Since(removeStartTime)
	if err != nil {
		log.Error(
			"删除脚本文件:文件删除失败",
			zap.Error(err),
			zap.String("script_path", scriptPath),
			zap.Duration("remove_duration", removeDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "删除脚本文件:文件删除失败")
	}
	return nil
}

func (r *ScriptRepo) MoveScriptFile(
	ctx context.Context,
	oldScriptPath string,
	newScriptPath string,
) error {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"移动脚本文件:入参详情",
		zap.String("old_script_path", oldScriptPath),
		zap.String("new_script_path", newScriptPath),
	)

	// 移动文件
	moveStartTime := time.Now()
	err := fileutil.Move(ctx, oldScriptPath, newScriptPath)
	moveDuration := time.Since(moveStartTime)
	if err != nil {
		log.Error(
			"移动脚本文件:文件移动失败",
			zap.Error(err),
			zap.String("old_script_path", oldScriptPath),
			zap.String("new_script_path", newScriptPath),
			zap.Duration("move_duration", moveDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "移动脚本文件:文件移动失败")
	}
	return nil
}

// ListProjects 查询所有脚本的项目名称
//
// 参数:
//
//	ctx: 上下文，用于传递请求信息和控制超时
//
// 返回值:
//
//	[]string: 项目名称列表
//	error: 操作错误信息，成功则返回nil
//
// 功能:
//  1. 执行数据库查询操作
//  2. 获取所有脚本的项目名称（去重）
//  3. 记录操作日志
func (r *ScriptRepo) ListProjects(
	ctx context.Context,
	query map[string]any,
) ([]string, error) {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"查询脚本所有的项目名称:查询条件",
		zap.Any("query_params", query),
	)

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()

	var projects []string

	listStartTime := time.Now()
	err := r.gormDB.WithContext(dbCtx).Model(&jobmodel.ScriptModel{}).Where(query).Distinct("project").Pluck("project", &projects).Error
	listDuration := time.Since(listStartTime)
	if err != nil {
		log.Error(
			"查询脚本所有的项目名称:数据库查询失败",
			zap.Error(err),
			zap.Any("query_params", query),
			zap.Duration("list_duration", listDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.WrapIf(err, "查询脚本所有的项目名称:数据库查询失败")
	}

	log.Debug(
		"查询脚本所有的项目名称:查询到的项目名称列表",
		zap.Strings("projects", projects),
	)

	if listDuration > r.slowThreshold.ListSlow {
		log.Warn("查询脚本所有的项目名称:数据库查询耗时超过慢查询阈值，可能影响性能",
			zap.Duration("list_duration", listDuration),
			zap.Duration("threshold", r.slowThreshold.ListSlow),
		)
	}
	return projects, nil
}

// ListLabels 查询所有脚本的标签名称
//
// 参数:
//
//	ctx: 上下文，用于传递请求信息和控制超时
//
// 返回值:
//
//	[]string: 标签名称列表
//	error: 操作错误信息，成功则返回nil
//
// 功能:
//  1. 执行数据库查询操作
//  2. 获取所有脚本的标签名称（去重）
//  3. 记录操作日志
func (r *ScriptRepo) ListLabels(
	ctx context.Context,
	query map[string]any,
) ([]string, error) {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"查询脚本所有的标签名称:查询条件",
		zap.Any("query_params", query),
	)

	var labels []string

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()

	// 查询所有唯一的标签名称
	listStartTime := time.Now()
	err := r.gormDB.WithContext(dbCtx).Model(&jobmodel.ScriptModel{}).Where(query).Distinct("label").Pluck("label", &labels).Error
	listDuration := time.Since(listStartTime)
	if err != nil {
		log.Error(
			"查询脚本所有的标签名称:数据库查询失败",
			zap.Error(err),
			zap.Any("query_params", query),
			zap.Duration("list_duration", listDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.WrapIf(err, "查询脚本所有的标签名称:数据库查询失败")
	}

	log.Debug(
		"查询脚本所有的标签名称:查询到的标签名称列表",
		zap.Strings("labels", labels),
	)

	if listDuration > r.slowThreshold.ListSlow {
		log.Warn("查询脚本所有的标签名称:数据库查询耗时超过慢查询阈值，可能影响性能",
			zap.Duration("list_duration", listDuration),
			zap.Duration("threshold", r.slowThreshold.ListSlow),
		)
	}

	return labels, nil
}
