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
	log      *zap.Logger       // 日志记录器
	gormDB   *gorm.DB          // GORM数据库连接
	timeouts *config.DBTimeout // 数据库操作超时配置
}

// NewScriptRepo 创建脚本仓库实例
//
// 参数：
//
//	log: 日志记录器，用于记录操作日志
//	gormDB: GORM数据库连接，用于执行数据库操作
//	timeouts: 数据库操作超时配置，控制各类数据库操作的超时时间
//
// 返回值：
//
//	*ScriptRepo: 脚本仓库实例
func NewScriptRepo(
	log *zap.Logger,
	gormDB *gorm.DB,
	timeouts *config.DBTimeout,
) *ScriptRepo {
	return &ScriptRepo{
		log:      log,
		gormDB:   gormDB,
		timeouts: timeouts,
	}
}

// CreateModel 创建脚本模型
//
// 参数：
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	m: 脚本模型，包含脚本的详细信息
//
// 返回值：
//
//	error: 操作错误信息，成功则返回nil
//
// 功能：
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
		err := errors.New("创建脚本模型：模型不能为空")
		log.Error(
			"创建脚本模型: 模型不能为空",
			zap.Error(err),
		)
		return err
	}

	log.Debug(
		"创建脚本模型：开始执行",
		zap.Object("script_model", m),
	)

	m.CreatedAt = startTime
	m.UpdatedAt = startTime
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	createScriptStartTime := time.Now()
	err := database.DBCreate(dbCtx, r.gormDB, &jobmodel.ScriptModel{}, m, nil)
	createScriptDuration := time.Since(createScriptStartTime)
	if err != nil {
		log.Error(
			"创建脚本模型：数据库操作失败",
			zap.Error(err),
			zap.Object("script_model", m),
			zap.Duration("create_script_duration", createScriptDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "创建脚本模型：数据库操作失败")
	}
	log.Debug(
		"创建脚本模型：执行成功",
		zap.Object("script_model", m),
		zap.Duration("create_script_duration", createScriptDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

// UpdateModel 更新脚本模型
//
// 参数：
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	data: 更新数据，包含要更新的字段和值
//	conds: 查询条件，用于指定要更新的记录
//
// 返回值：
//
//	error: 操作错误信息，成功则返回nil
//
// 功能：
//  1. 检查更新数据是否为空
//  2. 执行数据库更新操作
//  3. 记录操作日志
func (r *ScriptRepo) UpdateModel(
	ctx context.Context,
	data map[string]any,
	conds ...any,
) error {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	// 检查参数
	if len(data) == 0 {
		err := errors.New("更新脚本模型：更新数据不能为空")
		log.Error(
			"更新脚本模型：更新数据不能为空",
			zap.Error(err),
			zap.Any("update_data", data),
			zap.Any("conds", conds),
		)
		return err
	}
	log.Debug(
		"更新脚本模型：开始执行",
		zap.Any("update_data", data),
		zap.Any("conds", conds),
	)

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	updateScriptStartTime := time.Now()
	err := database.DBUpdate(dbCtx, r.gormDB, &jobmodel.ScriptModel{}, data, nil, conds...)
	updateScriptDuration := time.Since(updateScriptStartTime)
	if err != nil {
		log.Error(
			"更新脚本模型：数据库操作失败",
			zap.Error(err),
			zap.Any("update_data", data),
			zap.Any("conds", conds),
			zap.Duration("update_script_duration", updateScriptDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "更新脚本模型：数据库操作失败")
	}
	log.Debug(
		"更新脚本模型：执行成功",
		zap.Any("update_data", data),
		zap.Any("conds", conds),
		zap.Duration("update_script_duration", updateScriptDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

// DeleteModel 删除脚本模型
//
// 参数：
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	conds: 查询条件，用于指定要删除的记录
//
// 返回值：
//
//	error: 操作错误信息，成功则返回nil
//
// 功能：
//  1. 执行数据库删除操作
//  2. 记录操作日志
func (r *ScriptRepo) DeleteModel(
	ctx context.Context,
	conds ...any,
) error {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"删除脚本模型：开始执行",
		zap.Any("conds", conds),
	)

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	deleteScriptStartTime := time.Now()
	err := database.DBDelete(dbCtx, r.gormDB, &jobmodel.ScriptModel{}, conds...)
	deleteScriptDuration := time.Since(deleteScriptStartTime)
	if err != nil {
		log.Error(
			"删除脚本模型：数据库操作失败",
			zap.Error(err),
			zap.Any("conds", conds),
			zap.Duration("delete_script_duration", deleteScriptDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "删除脚本模型：数据库操作失败")
	}
	log.Debug(
		"删除脚本模型：执行成功",
		zap.Any("conds", conds),
		zap.Duration("delete_script_duration", deleteScriptDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

// GetModel 查询单个脚本模型
//
// 参数：
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	conds: 查询条件，用于指定要查询的记录
//
// 返回值：
//
//	*jobmodel.ScriptModel: 脚本模型指针，包含脚本的详细信息
//	error: 操作错误信息，成功则返回nil
//
// 功能：
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
		"查询脚本模型：开始执行",
		zap.Any("conds", conds),
	)

	var m jobmodel.ScriptModel
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()
	getScriptStartTime := time.Now()
	err := database.DBGet(dbCtx, r.gormDB, nil, &m, conds...)
	getScriptDuration := time.Since(getScriptStartTime)
	if err != nil {
		log.Error(
			"查询脚本模型：数据库操作失败",
			zap.Error(err),
			zap.Any("conds", conds),
			zap.Duration("get_script_duration", getScriptDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.WrapIf(err, "查询脚本模型：数据库操作失败")
	}
	log.Debug(
		"查询脚本模型：执行成功",
		zap.Object("script_model", &m),
		zap.Any("conds", conds),
		zap.Duration("get_script_duration", getScriptDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return &m, nil
}

// ListModel 查询脚本模型列表
//
// 参数：
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	qp: 查询参数，包含分页、排序等查询条件
//
// 返回值：
//
//	int64: 总记录数
//	*[]jobmodel.ScriptModel: 脚本模型列表指针，包含符合条件的脚本模型
//	error: 操作错误信息，成功则返回nil
//
// 功能：
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
		"查询脚本模型列表：开始执行",
		zap.Object("query_params", &qp),
	)

	var ms []jobmodel.ScriptModel
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ListTimeout)
	defer cancel()
	listScriptStartTime := time.Now()
	err := database.DBList(dbCtx, r.gormDB, &jobmodel.ScriptModel{}, &ms, qp)
	listScriptDuration := time.Since(listScriptStartTime)
	if err != nil {
		log.Error(
			"查询脚本模型列表：数据库操作失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("list_script_duration", listScriptDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.WrapIf(err, "查询脚本模型列表：数据库操作失败")
	}
	log.Debug(
		"查询脚本模型列表：执行成功",
		zap.Object("query_params", &qp),
		zap.Duration("list_script_duration", listScriptDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return ms, nil
}

func (r *ScriptRepo) CountModel(
	ctx context.Context,
	query map[string]any,
) (int64, error) {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"查询脚本模型总数：开始执行",
		zap.Any("query", query),
	)

	var count int64
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()
	countScriptStartTime := time.Now()
	count, err := database.DBCount(dbCtx, r.gormDB, &jobmodel.ScriptModel{}, query)
	countScriptDuration := time.Since(countScriptStartTime)
	if err != nil {
		log.Error(
			"查询脚本模型总数：数据库操作失败",
			zap.Error(err),
			zap.Any("query", query),
			zap.Duration("count_script_duration", countScriptDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, errors.WrapIf(err, "查询脚本模型总数：数据库操作失败")
	}
	log.Debug(
		"查询脚本模型总数：执行成功",
		zap.Any("query", query),
		zap.Int64("count", count),
		zap.Duration("count_script_duration", countScriptDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
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
		"保存脚本文件：开始执行",
		zap.String("script_path", scriptPath),
		zap.Bool("overwrite", overwrite),
	)
	saveScriptStartTime := time.Now()
	err := fileutil.WriteReaderToFile(ctx, fileReader, scriptPath, os.FileMode(0o750), overwrite)
	saveScriptDuration := time.Since(saveScriptStartTime)
	if err != nil {
		log.Error(
			"保存脚本文件：文件写入失败",
			zap.Error(err),
			zap.String("script_path", scriptPath),
			zap.Duration("save_script_duration", saveScriptDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "保存脚本文件：文件写入失败")
	}
	log.Debug(
		"保存脚本文件：执行成功",
		zap.String("script_path", scriptPath),
		zap.Duration("save_script_duration", saveScriptDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

func (r *ScriptRepo) RemoveScriptFile(
	ctx context.Context,
	scriptPath string,
) error {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"删除脚本文件：开始执行",
		zap.String("script_path", scriptPath),
	)

	// 检查文件是否存在
	if _, err := os.Stat(scriptPath); err != nil {
		if os.IsNotExist(err) {
			log.Warn(
				"删除脚本文件：脚本文件不存在",
				zap.Error(err),
				zap.String("script_path", scriptPath),
				zap.Duration("total_duration", time.Since(startTime)),
			)
			return nil
		}
		log.Error(
			"删除脚本文件：检查脚本文件失败",
			zap.Error(err),
			zap.String("script_path", scriptPath),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "删除脚本文件：检查脚本文件失败")
	}
	// 删除文件
	removeScriptStartTime := time.Now()
	err := os.Remove(scriptPath)
	removeScriptDuration := time.Since(removeScriptStartTime)
	if err != nil {
		log.Error(
			"删除脚本文件：文件删除失败",
			zap.Error(err),
			zap.String("script_path", scriptPath),
			zap.Duration("remove_script_duration", removeScriptDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "删除脚本文件：文件删除失败")
	}
	return nil
}

// ListProjects 查询所有脚本的项目名称
//
// 参数：
//
//	ctx: 上下文，用于传递请求信息和控制超时
//
// 返回值：
//
//	[]string: 项目名称列表
//	error: 操作错误信息，成功则返回nil
//
// 功能：
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
		"查询脚本所有的项目名称：开始查询",
		zap.Any("query_params", query),
	)

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()

	var projects []string
	countScriptStartTime := time.Now()
	err := r.gormDB.WithContext(dbCtx).Model(&jobmodel.ScriptModel{}).Where(query).Distinct("project").Pluck("project", &projects).Error
	countScriptDuration := time.Since(countScriptStartTime)
	if err != nil {
		log.Error(
			"查询脚本所有的项目名称：数据库查询失败",
			zap.Error(err),
			zap.Any("query_params", query),
			zap.Duration("count_script_duration", countScriptDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.WrapIf(err, "查询脚本所有的项目名称：数据库查询失败")
	}

	log.Debug(
		"查询脚本所有的项目名称：查询项目名称成功",
		zap.Any("projects", projects),
		zap.Any("query_params", query),
		zap.Duration("count_script_duration", countScriptDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	return projects, nil
}

// ListLabels 查询所有脚本的标签名称
//
// 参数：
//
//	ctx: 上下文，用于传递请求信息和控制超时
//
// 返回值：
//
//	[]string: 标签名称列表
//	error: 操作错误信息，成功则返回nil
//
// 功能：
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
		"查询脚本所有的标签名称：开始查询",
		zap.Any("query_params", query),
	)

	var labels []string
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()

	// 查询所有唯一的标签名称
	countScriptStartTime := time.Now()
	err := r.gormDB.WithContext(dbCtx).Model(&jobmodel.ScriptModel{}).Where(query).Distinct("label").Pluck("label", &labels).Error
	countScriptDuration := time.Since(countScriptStartTime)
	if err != nil {
		log.Error(
			"查询脚本所有的标签名称：数据库查询失败",
			zap.Error(err),
			zap.Any("query_params", query),
			zap.Duration("count_script_duration", countScriptDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.WrapIf(err, "查询脚本所有的标签名称：数据库查询失败")
	}

	log.Debug(
		"查询脚本所有的标签名称：查询标签名称成功",
		zap.Any("labels", labels),
		zap.Any("query_params", query),
		zap.Duration("count_script_duration", countScriptDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	return labels, nil
}
