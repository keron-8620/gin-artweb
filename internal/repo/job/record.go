package job

import (
	"context"
	"time"

	"emperror.dev/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"

	jobmodel "gin-artweb/internal/model/job"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/database"
)

// RecordRepo 脚本执行记录仓库实现
// 负责脚本执行记录模型的CRUD操作
// 使用GORM进行数据库操作
type RecordRepo struct {
	log      *zap.Logger       // 日志记录器
	gormDB   *gorm.DB          // GORM数据库连接
	timeouts *config.DBTimeout // 数据库操作超时配置
}

// NewRecordRepo 创建脚本执行记录仓库实例
//
// 参数：
//
//	log: 日志记录器，用于记录操作日志
//	gormDB: GORM数据库连接，用于执行数据库操作
//	timeouts: 数据库操作超时配置，控制各类数据库操作的超时时间
//
// 返回值：
//
//	jobmodel.ScriptRecordRepo: 脚本执行记录仓库接口实现
func NewRecordRepo(
	log *zap.Logger,
	gormDB *gorm.DB,
	timeouts *config.DBTimeout,
) *RecordRepo {
	return &RecordRepo{
		log:      log,
		gormDB:   gormDB,
		timeouts: timeouts,
	}
}

// CreateModel 创建脚本执行记录模型
//
// 参数：
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	m: 脚本执行记录模型，包含脚本执行记录的详细信息
//
// 返回值：
//
//	error: 操作错误信息，成功则返回nil
//
// 功能：
//  1. 检查脚本执行记录模型是否为空
//  2. 执行数据库创建操作
//  3. 记录操作日志
func (r *RecordRepo) CreateModel(
	ctx context.Context,
	m *jobmodel.ScriptRecordModel,
) error {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	// 检查参数
	if m == nil {
		err := errors.New("创建脚本执行记录模型：模型不能为空")
		log.Error(
			"创建脚本执行记录模型：模型不能为空",
			zap.Error(err),
		)
		return err
	}
	log.Debug(
		"创建脚本执行记录模型：开始执行",
		zap.Object("script_record_model", m),
	)

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	createScriptRecordStartTime := time.Now()
	err := database.DBCreate(dbCtx, r.gormDB, &jobmodel.ScriptRecordModel{}, m, nil)
	createScriptRecordDuration := time.Since(createScriptRecordStartTime)
	if err != nil {
		log.Error(
			"创建脚本执行记录模型：数据库操作失败",
			zap.Error(err),
			zap.Object("script_record_model", m),
			zap.Duration("create_script_record_duration", createScriptRecordDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "创建脚本执行记录模型：数据库操作失败")
	}
	log.Debug(
		"创建脚本执行记录模型：执行成功",
		zap.Object("script_record_model", m),
		zap.Duration("create_script_record_duration", createScriptRecordDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

// UpdateModel 更新脚本执行记录模型
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
func (r *RecordRepo) UpdateModel(
	ctx context.Context,
	data map[string]any,
	conds ...any,
) error {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	// 检查参数
	if len(data) == 0 {
		err := errors.New("更新脚本执行记录模型：更新数据不能为空")
		log.Error(
			"更新脚本执行记录模型：更新数据不能为空",
			zap.Error(err),
			zap.Any("update_data", data),
			zap.Any("conds", conds),
		)
		return err
	}
	log.Debug(
		"更新脚本执行记录模型：开始执行",
		zap.Any("update_data", data),
		zap.Any("conds", conds),
	)

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	updateScriptRecordStartTime := time.Now()
	err := database.DBUpdate(dbCtx, r.gormDB, &jobmodel.ScriptRecordModel{}, data, nil, conds...)
	updateScriptRecordDuration := time.Since(updateScriptRecordStartTime)
	if err != nil {
		log.Error(
			"更新脚本执行记录模型：数据库操作失败",
			zap.Error(err),
			zap.Any("update_data", data),
			zap.Any("conds", conds),
			zap.Duration("update_script_record_duration", updateScriptRecordDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "更新脚本执行记录模型：数据库操作失败")
	}
	log.Debug(
		"更新脚本执行记录模型：执行成功",
		zap.Any("update_data", data),
		zap.Any("conds", conds),
		zap.Duration("update_script_record_duration", updateScriptRecordDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

// DeleteModel 删除脚本执行记录模型
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
func (r *RecordRepo) DeleteModel(
	ctx context.Context,
	conds ...any,
) error {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"删除脚本执行记录模型：开始执行",
		zap.Any("conds", conds),
	)
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	deleteScriptRecordStartTime := time.Now()
	err := database.DBDelete(dbCtx, r.gormDB, &jobmodel.ScriptRecordModel{}, conds...)
	deleteScriptRecordDuration := time.Since(deleteScriptRecordStartTime)
	if err != nil {
		log.Error(
			"删除脚本执行记录模型：数据库操作失败",
			zap.Error(err),
			zap.Any("conds", conds),
			zap.Duration("delete_script_record_duration", deleteScriptRecordDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "删除脚本执行记录模型：数据库操作失败")
	}
	log.Debug(
		"删除脚本执行记录模型：执行成功",
		zap.Any("conds", conds),
		zap.Duration("delete_script_record_duration", deleteScriptRecordDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

// GetModel 查询单个脚本执行记录模型
//
// 参数：
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	preloads: 需要预加载的关联关系
//	conds: 查询条件，用于指定要查询的记录
//
// 返回值：
//
//	*jobmodel.ScriptRecordModel: 脚本执行记录模型指针，包含脚本执行记录的详细信息
//	error: 操作错误信息，成功则返回nil
//
// 功能：
//  1. 执行数据库查询操作
//  2. 预加载关联字段
//  3. 获取单个脚本执行记录模型
//  4. 记录操作日志
func (r *RecordRepo) GetModel(
	ctx context.Context,
	preloads []string,
	conds ...any,
) (*jobmodel.ScriptRecordModel, error) {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"查询脚本执行记录模型：开始执行",
		zap.Any("conds", conds),
	)
	var m jobmodel.ScriptRecordModel
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()
	getScriptRecordStartTime := time.Now()
	err := database.DBGet(dbCtx, r.gormDB, preloads, &m, conds...)
	getScriptRecordDuration := time.Since(getScriptRecordStartTime)
	if err != nil {
		log.Error(
			"查询脚本执行记录模型：数据库操作失败",
			zap.Error(err),
			zap.Any("conds", conds),
			zap.Duration("get_script_record_duration", getScriptRecordDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.WrapIf(err, "查询脚本执行记录模型：数据库操作失败")
	}
	log.Debug(
		"查询脚本执行记录模型：执行成功",
		zap.Object("script_record_model", &m),
		zap.Any("conds", conds),
		zap.Duration("get_script_record_duration", getScriptRecordDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return &m, nil
}

// ListModel 查询脚本执行记录模型列表
//
// 参数：
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	qp: 查询参数，包含分页、排序、过滤等条件
//
// 返回值：
//
//	int64: 查询结果总数
//	*[]jobmodel.ScriptRecordModel: 脚本执行记录模型列表指针，包含查询到的脚本执行记录详情
//	error: 操作错误信息，成功则返回nil
//
// 功能：
//  1. 执行数据库查询操作
//  2. 应用分页、排序、过滤等条件
//  3. 获取脚本执行记录模型列表
//  4. 记录操作日志
func (r *RecordRepo) ListModel(
	ctx context.Context,
	qp database.QueryParams,
) ([]jobmodel.ScriptRecordModel, error) {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"查询脚本执行记录模型列表：开始执行",
		zap.Object("query_params", &qp),
	)
	var ms []jobmodel.ScriptRecordModel
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ListTimeout)
	defer cancel()
	listScriptRecordStartTime := time.Now()
	err := database.DBList(dbCtx, r.gormDB, &jobmodel.ScriptRecordModel{}, &ms, qp)
	listScriptRecordDuration := time.Since(listScriptRecordStartTime)
	if err != nil {
		log.Error(
			"查询脚本执行记录模型列表：数据库操作失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("list_script_record_duration", listScriptRecordDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.WrapIf(err, "查询脚本执行记录模型列表：数据库操作失败")
	}
	log.Debug(
		"查询脚本执行记录模型列表：执行成功",
		zap.Object("query_params", &qp),
		zap.Duration("list_script_record_duration", listScriptRecordDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return ms, nil
}

func (r *RecordRepo) CountModel(
	ctx context.Context,
	query map[string]any,
) (int64, error) {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"查询脚本执行记录模型总数：开始执行",
		zap.Any("query", query),
	)
	var count int64
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()
	countScriptRecordStartTime := time.Now()
	count, err := database.DBCount(dbCtx, r.gormDB, &jobmodel.ScriptRecordModel{}, query)
	countScriptRecordDuration := time.Since(countScriptRecordStartTime)
	if err != nil {
		log.Error(
			"查询脚本执行记录模型总数：数据库操作失败",
			zap.Error(err),
			zap.Any("query", query),
			zap.Duration("count_script_record_duration", countScriptRecordDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, errors.WrapIf(err, "查询脚本执行记录模型总数：数据库操作失败")
	}
	log.Debug(
		"查询脚本执行记录模型总数：执行成功",
		zap.Any("query", query),
		zap.Int64("count", count),
		zap.Duration("count_script_record_duration", countScriptRecordDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return count, nil
}
