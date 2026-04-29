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

// ScheduleRepo 计划任务仓库实现
// 负责计划任务模型的CRUD操作
// 使用GORM进行数据库操作
type ScheduleRepo struct {
	log           *zap.Logger             // 日志记录器
	gormDB        *gorm.DB                // GORM数据库连接
	timeouts      *config.DBTimeout       // 数据库操作超时配置
	slowThreshold *config.DBSlowThreshold // 数据库操作慢查询阈值配置
}

// NewScheduleRepo 创建计划任务仓库实例
//
// 参数:
//
//	log: 日志记录器，用于记录操作日志
//	gormDB: GORM数据库连接，用于执行数据库操作
//	timeouts: 数据库操作超时配置，控制各类数据库操作的超时时间
//	slowThreshold: 数据库操作慢查询阈值配置，用于判断是否需要记录慢查询日志
//
// 返回值:
//
//	jobmodel.ScheduleRepo: 计划任务仓库接口实现
func NewScheduleRepo(
	log *zap.Logger,
	gormDB *gorm.DB,
	timeouts *config.DBTimeout,
	slowThreshold *config.DBSlowThreshold,
) *ScheduleRepo {
	return &ScheduleRepo{
		log:           log,
		gormDB:        gormDB,
		timeouts:      timeouts,
		slowThreshold: slowThreshold,
	}
}

// CreateModel 创建计划任务模型
//
// 参数:
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	m: 计划任务模型，包含计划任务的详细信息
//
// 返回值:
//
//	error: 操作错误信息，成功则返回nil
//
// 功能:
//  1. 检查计划任务模型是否为空
//  2. 执行数据库创建操作
//  3. 记录操作日志
func (r *ScheduleRepo) CreateModel(
	ctx context.Context,
	m *jobmodel.ScheduleModel,
) error {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	// 检查参数
	if m == nil {
		err := errors.New("创建计划任务模型:模型不能为空")
		log.Error(
			"创建计划任务模型:模型不能为空",
			zap.Error(err),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return err
	}

	m.CreatedAt = startTime
	m.UpdatedAt = startTime

	log.Debug(
		"创建计划任务模型:开始执行",
		zap.Object("schedule_model", m),
	)

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()

	createStartTime := time.Now()
	err := database.DBCreate(dbCtx, r.gormDB, &jobmodel.ScheduleModel{}, m)
	createDuration := time.Since(createStartTime)
	if err != nil {
		log.Error(
			"创建计划任务模型:数据库操作失败",
			zap.Error(err),
			zap.Object("schedule_model", m),
			zap.Duration("create_duration", createDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "创建计划任务模型:数据库操作失败")
	}

	if createDuration > r.slowThreshold.WriteSlow {
		log.Warn("创建计划任务模型:数据库创建耗时超过慢查询阈值，可能影响性能",
			zap.Uint32("schedule_id", m.ID),
			zap.Duration("create_duration", createDuration),
			zap.Duration("threshold", r.slowThreshold.WriteSlow),
		)
	}
	return nil
}

// UpdateModel 更新计划任务模型
//
// 参数:
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	data: 更新数据，包含要更新的字段和值
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
func (r *ScheduleRepo) UpdateModel(
	ctx context.Context,
	updateData map[string]any,
	conds ...any,
) error {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	if len(updateData) == 0 {
		err := errors.New("更新计划任务模型:更新数据为空")
		log.Error(
			"更新计划任务模型:更新数据为空",
			zap.Error(err),
			zap.Any("update_data", updateData),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return err
	}

	updateData["updated_at"] = startTime

	log.Debug(
		"更新计划任务模型:更新数据",
		zap.Any("update_data", updateData),
		zap.Any("conds", conds),
	)

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()

	updateStartTime := time.Now()
	err := database.DBUpdateTx(dbCtx, r.gormDB, &jobmodel.ScheduleModel{}, updateData, conds...)
	updateDuration := time.Since(updateStartTime)
	if err != nil {
		log.Error(
			"更新计划任务模型:数据库操作失败",
			zap.Error(err),
			zap.Any("update_data", updateData),
			zap.Any("conds", conds),
			zap.Duration("update_duration", updateDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "更新计划任务模型:数据库操作失败")
	}

	if updateDuration > r.slowThreshold.WriteSlow {
		log.Warn("更新计划任务模型:数据库更新耗时超过慢查询阈值，可能影响性能",
			zap.Duration("update_duration", updateDuration),
			zap.Duration("threshold", r.slowThreshold.WriteSlow),
		)
	}
	return nil
}

// DeleteModel 删除计划任务模型
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
func (r *ScheduleRepo) DeleteModel(
	ctx context.Context,
	conds ...any,
) error {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"删除计划任务模型:删除条件",
		zap.Any("conds", conds),
	)

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()

	deleteStartTime := time.Now()
	err := database.DBDeleteTx(dbCtx, r.gormDB, &jobmodel.ScheduleModel{}, conds...)
	deleteDuration := time.Since(deleteStartTime)
	if err != nil {
		log.Error(
			"删除计划任务模型:数据库操作失败",
			zap.Error(err),
			zap.Any("conds", conds),
			zap.Duration("delete_duration", deleteDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "删除计划任务模型:数据库操作失败")
	}

	if deleteDuration > r.slowThreshold.WriteSlow {
		log.Warn("删除计划任务模型:数据库删除耗时超过慢查询阈值，可能影响性能",
			zap.Duration("delete_duration", deleteDuration),
			zap.Duration("threshold", r.slowThreshold.WriteSlow),
		)
	}
	return nil
}

// GetModel 查询单个计划任务模型
//
// 参数:
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	preloads: 需要预加载的关联关系
//	conds: 查询条件，用于指定要查询的记录
//
// 返回值:
//
//	*jobmodel.ScheduleModel: 计划任务模型指针，包含计划任务的详细信息
//	error: 操作错误信息，成功则返回nil
//
// 功能:
//  1. 执行数据库查询操作
//  2. 预加载关联字段
//  3. 获取单个计划任务模型
//  4. 记录操作日志
func (r *ScheduleRepo) GetModel(
	ctx context.Context,
	preloads []string,
	conds ...any,
) (*jobmodel.ScheduleModel, error) {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"查询计划任务模型:查询条件",
		zap.Strings("preloads", preloads),
		zap.Any("conds", conds),
	)

	var m jobmodel.ScheduleModel

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()

	getStartTime := time.Now()
	err := database.DBGet(dbCtx, r.gormDB, preloads, &m, conds...)
	getDuration := time.Since(getStartTime)
	if err != nil {
		log.Error(
			"查询计划任务模型:数据库操作失败",
			zap.Error(err),
			zap.Any("conds", conds),
			zap.Duration("get_duration", getDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.WrapIf(err, "查询计划任务模型:数据库操作失败")
	}

	log.Debug(
		"查询计划任务模型:查询到的模型详情",
		zap.Object("schedule_model", &m),
	)

	if getDuration > r.slowThreshold.ReadSlow {
		log.Warn("查询计划任务模型:数据库查询耗时超过慢查询阈值，可能影响性能",
			zap.Duration("get_duration", getDuration),
			zap.Duration("threshold", r.slowThreshold.ReadSlow),
		)
	}
	return &m, nil
}

// ListModel 查询计划任务模型列表
//
// 参数:
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	qp: 查询参数，包含分页、排序、过滤等条件
//
// 返回值:
//
//	int64: 查询结果总数
//	*[]jobmodel.ScheduleModel: 计划任务模型列表指针，包含查询到的计划任务详情
//	error: 操作错误信息，成功则返回nil
//
// 功能:
//  1. 执行数据库查询操作
//  2. 应用分页、排序、过滤等条件
//  3. 获取计划任务模型列表
//  4. 记录操作日志
func (r *ScheduleRepo) ListModel(
	ctx context.Context,
	qp database.QueryParams,
) ([]jobmodel.ScheduleModel, error) {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"查询计划任务模型列表:入参详情",
		zap.Object("query_params", &qp),
	)

	var ms []jobmodel.ScheduleModel

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ListTimeout)
	defer cancel()

	listStartTime := time.Now()
	err := database.DBList(dbCtx, r.gormDB, &jobmodel.ScheduleModel{}, &ms, qp)
	listDuration := time.Since(listStartTime)
	if err != nil {
		log.Error(
			"查询计划任务模型列表:数据库操作失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("list_duration", listDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.WrapIf(err, "查询计划任务模型列表:数据库操作失败")
	}

	log.Debug(
		"查询计划任务模型列表:查询到的模型列表",
		zap.Uint32s("schedule_ids", jobmodel.ListScheduleModelToUint32s(ms)),
	)

	if listDuration > r.slowThreshold.ListSlow {
		log.Warn("查询计划任务模型列表:数据库查询耗时超过慢查询阈值，可能影响性能",
			zap.Duration("list_duration", listDuration),
			zap.Duration("threshold", r.slowThreshold.ListSlow),
		)
	}
	return ms, nil
}

func (r *ScheduleRepo) CountModel(
	ctx context.Context,
	query map[string]any,
) (int64, error) {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"查询计划任务模型总数:入参详情",
		zap.Any("query", query),
	)

	// 开启数据库事务
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()

	countStartTime := time.Now()
	count, err := database.DBCount(dbCtx, r.gormDB, &jobmodel.ScheduleModel{}, query)
	countDuration := time.Since(countStartTime)
	if err != nil {
		log.Error(
			"查询计划任务模型总数:数据库查询失败",
			zap.Error(err),
			zap.Any("query", query),
			zap.Duration("count_duration", countDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, errors.WrapIf(err, "查询计划任务模型总数:数据库查询失败")
	}

	log.Debug(
		"查询计划任务模型总数:查询到的记录数",
		zap.Int64("count", count),
	)

	if countDuration > r.slowThreshold.ReadSlow {
		log.Warn("查询计划任务模型总数:数据库查询耗时超过慢查询阈值，可能影响性能",
			zap.Duration("count_duration", countDuration),
			zap.Duration("threshold", r.slowThreshold.ReadSlow),
		)
	}
	return count, nil
}

// CreateModels 创建计划任务模型列表
//
// 参数:
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	ms: 计划任务模型列表，包含计划任务的详细信息
//
// 返回值:
//
//	error: 操作错误信息，成功则返回nil
//
// 功能:
//  1. 检查计划任务模型列表是否为空
//  2. 执行数据库创建操作
//  3. 记录操作日志
func (r *ScheduleRepo) CreateModels(
	ctx context.Context,
	ms []jobmodel.ScheduleModel,
) error {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	// 检查参数
	if len(ms) == 0 {
		err := errors.New("创建计划任务模型列表:模型列表不能为空")
		log.Error(
			"创建计划任务模型列表:模型列表不能为空",
			zap.Error(err),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return err
	}

	for _, m := range ms {
		m.CreatedAt = startTime
		m.UpdatedAt = startTime
	}

	log.Debug(
		"创建计划任务模型列表:开始执行",
		zap.Any("schedule_models", ms),
	)

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout*time.Duration(len(ms)))
	defer cancel()

	createStartTime := time.Now()
	err := database.DBCreateTX(dbCtx, r.gormDB, &jobmodel.ScheduleModel{}, ms)
	createDuration := time.Since(createStartTime)
	if err != nil {
		log.Error(
			"创建计划任务模型列表:数据库操作失败",
			zap.Error(err),
			zap.Any("schedule_models", ms),
			zap.Duration("create_duration", createDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "创建计划任务模型列表:数据库操作失败")
	}

	if createDuration > r.slowThreshold.WriteSlow*time.Duration(len(ms)) {
		log.Warn("创建计划任务模型列表:数据库创建耗时超过慢查询阈值，可能影响性能",
			zap.Uint32s("schedule_ids", jobmodel.ListScheduleModelToUint32s(ms)),
			zap.Duration("threshold", r.slowThreshold.WriteSlow*time.Duration(len(ms))),
		)
	}
	return nil
}
