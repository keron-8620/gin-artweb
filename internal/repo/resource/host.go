package resource

import (
	"context"
	"time"

	"emperror.dev/errors"
	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"
	"gorm.io/gorm"

	resomodel "gin-artweb/internal/model/resource"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/shell"
)

// HostRepo 主机仓库实现
// 负责主机模型的CRUD操作和SSH连接管理
// 使用GORM进行数据库操作，使用SSH进行远程主机连接
type HostRepo struct {
	log           *zap.Logger             // 日志记录器
	gormDB        *gorm.DB                // GORM数据库连接
	timeouts      *config.DBTimeout       // 数据库操作超时配置
	slowThreshold *config.DBSlowThreshold // 数据库操作慢查询阈值配置
}

// NewHostRepo 创建主机仓库实例
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
//	*HostRepo: 主机仓库接口实现
func NewHostRepo(
	log *zap.Logger,
	gormDB *gorm.DB,
	timeouts *config.DBTimeout,
	slowThreshold *config.DBSlowThreshold,
) *HostRepo {
	return &HostRepo{
		log:           log,
		gormDB:        gormDB,
		timeouts:      timeouts,
		slowThreshold: slowThreshold,
	}
}

// CreateModel 创建主机模型
//
// 参数:
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	m: 主机模型，包含主机的详细信息
//
// 返回值:
//
//	error: 操作错误信息，成功则返回nil
//
// 功能:
//  1. 检查主机模型是否为空
//  2. 设置创建时间和更新时间
//  3. 执行数据库创建操作
//  4. 记录操作日志
func (r *HostRepo) CreateModel(
	ctx context.Context,
	m *resomodel.HostModel,
) error {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	// 检查参数
	if m == nil {
		err := errors.New("创建主机模型:模型不能为空")
		log.Error(
			"创建主机模型:模型不能为空",
			zap.Error(err),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return err
	}

	m.CreatedAt = startTime
	m.UpdatedAt = startTime

	log.Debug(
		"创建主机模型:开始执行",
		zap.Object("host_model", m),
	)

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()

	createStartTime := time.Now()
	err := database.DBCreate(dbCtx, r.gormDB, &resomodel.HostModel{}, m, nil)
	createDuration := time.Since(createStartTime)
	if err != nil {
		log.Error(
			"创建主机模型:数据库操作失败",
			zap.Error(err),
			zap.Object("host_model", m),
			zap.Duration("create_duration", createDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "创建主机模型:数据库操作失败")
	}

	if createDuration > r.slowThreshold.WriteSlow {
		log.Warn("创建主机模型:数据库创建耗时超过慢查询阈值，可能影响性能",
			zap.Uint32("host_id", m.ID),
			zap.Duration("create_duration", createDuration),
			zap.Duration("threshold", r.slowThreshold.WriteSlow),
		)
	}
	return nil
}

// UpdateModel 更新主机模型
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
func (r *HostRepo) UpdateModel(
	ctx context.Context,
	updateData map[string]any,
	conds ...any,
) error {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	if len(updateData) == 0 {
		err := errors.New("更新主机模型:更新数据为空")
		log.Error(
			"更新主机模型:更新数据为空",
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
	err := database.DBUpdate(dbCtx, r.gormDB, &resomodel.HostModel{}, updateData, nil, conds...)
	updateDuration := time.Since(updateStartTime)
	if err != nil {
		log.Error(
			"更新主机模型:数据库操作失败",
			zap.Error(err),
			zap.Any("update_data", updateData),
			zap.Any("conds", conds),
			zap.Duration("update_duration", updateDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "更新主机模型:数据库操作失败")
	}

	if updateDuration > r.slowThreshold.WriteSlow {
		log.Warn("更新主机模型:数据库更新耗时超过慢查询阈值，可能影响性能",
			zap.Duration("update_duration", updateDuration),
			zap.Duration("threshold", r.slowThreshold.WriteSlow),
		)
	}
	return nil
}

// DeleteModel 删除主机模型
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
func (r *HostRepo) DeleteModel(
	ctx context.Context,
	conds ...any,
) error {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"删除主机模型:删除条件",
		zap.Any("conds", conds),
	)

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()

	deleteStartTime := time.Now()
	err := database.DBDelete(dbCtx, r.gormDB, &resomodel.HostModel{}, conds...)
	deleteDuration := time.Since(deleteStartTime)
	if err != nil {
		log.Error(
			"删除主机模型:数据库操作失败",
			zap.Error(err),
			zap.Any("conds", conds),
			zap.Duration("delete_duration", deleteDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "删除主机模型:数据库操作失败")
	}

	if deleteDuration > r.slowThreshold.WriteSlow {
		log.Warn("删除主机模型:数据库删除耗时超过慢查询阈值，可能影响性能",
			zap.Duration("delete_duration", deleteDuration),
			zap.Duration("threshold", r.slowThreshold.WriteSlow),
		)
	}
	return nil
}

// GetModel 查询单个主机模型
//
// 参数:
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	preloads: 需要预加载的关联关系
//	conds: 查询条件，用于指定要查询的记录
//
// 返回值:
//
//	*resomodel.HostModel: 主机模型指针，包含主机的详细信息
//	error: 操作错误信息，成功则返回nil
//
// 功能:
//  1. 执行数据库查询操作
//  2. 预加载关联字段
//  3. 获取单个主机模型
//  4. 记录操作日志
func (r *HostRepo) GetModel(
	ctx context.Context,
	conds ...any,
) (*resomodel.HostModel, error) {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"查询主机模型:查询条件",
		zap.Any("conds", conds),
	)

	var m resomodel.HostModel

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()

	getStartTime := time.Now()
	err := database.DBGet(dbCtx, r.gormDB, nil, &m, conds...)
	getDuration := time.Since(getStartTime)
	if err != nil {
		log.Error(
			"查询主机模型:数据库操作失败",
			zap.Error(err),
			zap.Any("conds", conds),
			zap.Duration("get_duration", getDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.WrapIf(err, "查询主机模型:数据库操作失败")
	}

	log.Debug(
		"查询主机模型:查询到的模型详情",
		zap.Object("host_model", &m),
	)

	if getDuration > r.slowThreshold.ReadSlow {
		log.Warn("查询主机模型:数据库查询耗时超过慢查询阈值，可能影响性能",
			zap.Duration("get_duration", getDuration),
			zap.Duration("threshold", r.slowThreshold.ReadSlow),
		)
	}
	return &m, nil
}

// ListModel 查询主机模型列表
//
// 参数:
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	qp: 查询参数，包含分页、排序等查询条件
//
// 返回值:
//
//	int64: 总记录数
//	*[]resomodel.HostModel: 主机模型列表指针，包含符合条件的主机模型
//	error: 操作错误信息，成功则返回nil
//
// 功能:
//  1. 执行数据库查询操作
//  2. 获取主机模型列表
//  3. 返回总记录数和模型列表
//  4. 记录操作日志
func (r *HostRepo) ListModel(
	ctx context.Context,
	qp database.QueryParams,
) ([]resomodel.HostModel, error) {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"查询主机模型列表:入参详情",
		zap.Object("query_params", &qp),
	)

	var ms []resomodel.HostModel

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ListTimeout)
	defer cancel()

	listStartTime := time.Now()
	err := database.DBList(dbCtx, r.gormDB, &resomodel.HostModel{}, &ms, qp)
	listDuration := time.Since(listStartTime)
	if err != nil {
		log.Error(
			"查询主机模型列表:数据库操作失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("list_duration", listDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.WrapIf(err, "查询主机模型列表:数据库操作失败")
	}

	log.Debug(
		"查询主机模型列表:查询到的模型列表",
		zap.Uint32s("host_ids", resomodel.ListHostModelToUint32s(ms)),
	)

	if listDuration > r.slowThreshold.ListSlow {
		log.Warn("查询主机模型列表:数据库查询耗时超过慢查询阈值，可能影响性能",
			zap.Duration("list_duration", listDuration),
			zap.Duration("threshold", r.slowThreshold.ListSlow),
		)
	}
	return ms, nil
}

func (r *HostRepo) CountModel(
	ctx context.Context,
	query map[string]any,
) (int64, error) {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"查询主机模型总数:查询条件",
		zap.Any("query", query),
	)

	// 开启数据库事务
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()

	countStartTime := time.Now()
	count, err := database.DBCount(dbCtx, r.gormDB, &resomodel.HostModel{}, query)
	countDuration := time.Since(countStartTime)
	if err != nil {
		log.Error(
			"查询主机模型总数:数据库查询失败",
			zap.Error(err),
			zap.Any("query", query),
			zap.Duration("count_duration", countDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, errors.WrapIf(err, "查询主机模型总数:数据库查询失败")
	}

	log.Debug(
		"查询主机模型总数:查询到的记录数",
		zap.Int64("count", count),
	)

	if countDuration > r.slowThreshold.ReadSlow {
		log.Warn("查询主机模型总数:数据库查询耗时超过慢查询阈值，可能影响性能",
			zap.Duration("count_duration", countDuration),
			zap.Duration("threshold", r.slowThreshold.ReadSlow),
		)
	}
	return count, nil
}

// NewSSHClient 创建SSH客户端连接
//
// 参数:
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	sshIP: SSH服务器IP地址
//	sshPort: SSH服务器端口
//	sshUser: SSH用户名
//	sshAuths: SSH认证方法列表
//	timeout: 连接超时时间
//
// 返回值:
//
//	*ssh.Client: SSH客户端连接
//	error: 操作错误信息，成功则返回nil
//
// 功能:
//  1. 检查上下文是否有效
//  2. 创建SSH客户端连接
//  3. 记录操作日志
func (r *HostRepo) NewSSHClient(
	ctx context.Context,
	sshIP string,
	sshPort uint16,
	sshUser string,
	sshAuths []ssh.AuthMethod,
	timeout time.Duration,
) (*ssh.Client, error) {
	startTime := time.Now()
	if ctx.Err() != nil {
		return nil, errors.WrapIf(ctx.Err(), "上下文已取消")
	}

	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"创建ssh连接:入参详情",
		zap.String("ssh_ip", sshIP),
		zap.Uint16("ssh_port", sshPort),
		zap.String("ssh_user", sshUser),
	)

	connectsshStartTime := time.Now()
	client, err := shell.NewSSHClient(ctx, sshIP, sshPort, sshUser, sshAuths, false, timeout)
	connectsshDuration := time.Since(connectsshStartTime)
	if err != nil {
		log.Error(
			"创建ssh连接:ssh连接失败",
			zap.Error(err),
			zap.String("ssh_ip", sshIP),
			zap.Uint16("ssh_port", sshPort),
			zap.String("ssh_user", sshUser),
			zap.Duration("ssh_connect_duration", connectsshDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.WrapIf(err, "创建ssh连接:ssh连接失败")
	}

	return client, nil
}

// ExecuteCommand 执行SSH命令
//
// 参数:
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	session: SSH会话
//	command: 要执行的命令
//
// 返回值:
//
//	error: 操作错误信息，成功则返回nil
//
// 功能:
//  1. 检查上下文是否有效
//  2. 在SSH会话中执行命令
//  3. 记录操作日志
func (r *HostRepo) ExecuteCommand(
	ctx context.Context,
	session *ssh.Session,
	command string,
) error {
	startTime := time.Now()

	if ctx.Err() != nil {
		return errors.WrapIf(ctx.Err(), "上下文已取消")
	}

	if session == nil {
		return errors.New("session is nil")
	}

	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"执行命令:开始执行",
		zap.String("command", command),
	)

	executeCommandStartTime := time.Now()
	err := session.Run(command)
	executeCommandDuration := time.Since(executeCommandStartTime)
	if err != nil {
		log.Error(
			"执行命令:命令执行失败",
			zap.Error(err),
			zap.String("command", command),
			zap.Duration("execute_duration", executeCommandDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "执行命令:命令执行失败")
	}

	return nil
}
