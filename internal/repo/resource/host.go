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
	log      *zap.Logger       // 日志记录器
	gormDB   *gorm.DB          // GORM数据库连接
	timeouts *config.DBTimeout // 数据库操作超时配置
}

// NewHostRepo 创建主机仓库实例
//
// 参数：
//
//	log: 日志记录器，用于记录操作日志
//	gormDB: GORM数据库连接，用于执行数据库操作
//	timeouts: 数据库操作超时配置，控制各类数据库操作的超时时间
//
// 返回值：
//
//	*HostRepo: 主机仓库接口实现
func NewHostRepo(
	log *zap.Logger,
	gormDB *gorm.DB,
	timeouts *config.DBTimeout,
) *HostRepo {
	return &HostRepo{
		log:      log,
		gormDB:   gormDB,
		timeouts: timeouts,
	}
}

// CreateModel 创建主机模型
//
// 参数：
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	m: 主机模型，包含主机的详细信息
//
// 返回值：
//
//	error: 操作错误信息，成功则返回nil
//
// 功能：
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
		err := errors.New("创建主机模型：模型不能为空")
		log.Error(
			"创建主机模型：模型不能为空",
			zap.Error(err),
		)
		return err
	}
	log.Debug(
		"创建主机模型：开始执行",
		zap.Object("host_model", m),
	)

	m.CreatedAt = startTime
	m.UpdatedAt = startTime
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	createHostStartTime := time.Now()
	err := database.DBCreate(dbCtx, r.gormDB, &resomodel.HostModel{}, m, nil)
	createHostDuration := time.Since(createHostStartTime)
	if err != nil {
		log.Error(
			"创建主机模型：数据库操作失败",
			zap.Error(err),
			zap.Object("host_model", m),
			zap.Duration("create_host_duration", createHostDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "创建主机模型：数据库操作失败")
	}
	log.Debug(
		"创建主机模型：执行成功",
		zap.Object("host_model", m),
		zap.Duration("create_host_duration", createHostDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

// UpdateModel 更新主机模型
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
func (r *HostRepo) UpdateModel(
	ctx context.Context,
	data map[string]any,
	conds ...any,
) error {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	// 检查参数
	if len(data) == 0 {
		err := errors.New("更新主机模型：更新数据不能为空")
		log.Error(
			"更新主机模型：更新数据不能为空",
			zap.Error(err),
			zap.Any("update_data", data),
			zap.Any("conds", conds),
		)
		return err
	}
	log.Debug(
		"更新主机模型：开始执行",
		zap.Any("update_data", data),
		zap.Any("conds", conds),
	)

	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	updateHostStartTime := time.Now()
	err := database.DBUpdate(dbCtx, r.gormDB, &resomodel.HostModel{}, data, nil, conds...)
	updateHostDuration := time.Since(updateHostStartTime)
	if err != nil {
		log.Error(
			"更新主机模型：数据库操作失败",
			zap.Error(err),
			zap.Any("update_data", data),
			zap.Any("conds", conds),
			zap.Duration("update_host_duration", updateHostDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "更新主机模型：数据库操作失败")
	}
	log.Debug(
		"更新主机模型：执行成功",
		zap.Any("update_data", data),
		zap.Any("conds", conds),
		zap.Duration("update_host_duration", updateHostDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

// DeleteModel 删除主机模型
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
func (r *HostRepo) DeleteModel(
	ctx context.Context,
	conds ...any,
) error {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"删除主机模型：开始执行",
		zap.Any("conds", conds),
	)
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	deleteHostStartTime := time.Now()
	err := database.DBDelete(dbCtx, r.gormDB, &resomodel.HostModel{}, conds...)
	deleteHostDuration := time.Since(deleteHostStartTime)
	if err != nil {
		log.Error(
			"删除主机模型：数据库操作失败",
			zap.Error(err),
			zap.Any("conds", conds),
			zap.Duration("delete_host_duration", deleteHostDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "删除主机模型：数据库操作失败")
	}
	log.Debug(
		"删除主机模型：执行成功",
		zap.Any("conds", conds),
		zap.Duration("delete_host_duration", deleteHostDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return nil
}

// GetModel 查询单个主机模型
//
// 参数：
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	preloads: 需要预加载的关联关系
//	conds: 查询条件，用于指定要查询的记录
//
// 返回值：
//
//	*resomodel.HostModel: 主机模型指针，包含主机的详细信息
//	error: 操作错误信息，成功则返回nil
//
// 功能：
//  1. 执行数据库查询操作
//  2. 预加载关联字段
//  3. 获取单个主机模型
//  4. 记录操作日志
func (r *HostRepo) GetModel(
	ctx context.Context,
	preloads []string,
	conds ...any,
) (*resomodel.HostModel, error) {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"查询主机模型：开始执行",
		zap.Any("conds", conds),
	)
	var m resomodel.HostModel
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()
	getHostStartTime := time.Now()
	err := database.DBGet(dbCtx, r.gormDB, preloads, &m, conds...)
	getHostDuration := time.Since(getHostStartTime)
	if err != nil {
		log.Error(
			"查询主机模型：数据库操作失败",
			zap.Error(err),
			zap.Any("conds", conds),
			zap.Duration("get_host_duration", getHostDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.WrapIf(err, "查询主机模型：数据库操作失败")
	}
	log.Debug(
		"查询主机模型：执行成功",
		zap.Object("host_model", &m),
		zap.Any("conds", conds),
		zap.Duration("get_host_duration", getHostDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return &m, nil
}

// ListModel 查询主机模型列表
//
// 参数：
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	qp: 查询参数，包含分页、排序等查询条件
//
// 返回值：
//
//	int64: 总记录数
//	*[]resomodel.HostModel: 主机模型列表指针，包含符合条件的主机模型
//	error: 操作错误信息，成功则返回nil
//
// 功能：
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
		"查询主机模型列表：开始执行",
		zap.Object("query_params", &qp),
	)
	var ms []resomodel.HostModel
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ListTimeout)
	defer cancel()
	listHostStartTime := time.Now()
	err := database.DBList(dbCtx, r.gormDB, &resomodel.HostModel{}, &ms, qp)
	listHostDuration := time.Since(listHostStartTime)
	if err != nil {
		log.Error(
			"查询主机模型列表：数据库操作失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("list_host_duration", listHostDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.WrapIf(err, "查询主机模型列表：数据库操作失败")
	}
	log.Debug(
		"查询主机模型列表：执行成功",
		zap.Object("query_params", &qp),
		zap.Duration("list_host_duration", listHostDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return ms, nil
}

func (r *HostRepo) CountModel(
	ctx context.Context,
	query map[string]any,
) (int64, error) {
	startTime := time.Now()
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"查询主机模型总数：开始执行",
		zap.Any("query", query),
	)
	// 开启数据库事务
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()
	countHostStartTime := time.Now()
	count, err := database.DBCount(dbCtx, r.gormDB, &resomodel.HostModel{}, query)
	countHostDuration := time.Since(countHostStartTime)
	if err != nil {
		log.Error(
			"查询主机模型总数：数据库查询失败",
			zap.Error(err),
			zap.Any("query", query),
			zap.Duration("count_host_duration", countHostDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return 0, errors.WrapIf(err, "查询主机模型总数：数据库查询失败")
	}
	log.Debug(
		"查询主机模型总数：执行成功",
		zap.Any("query", query),
		zap.Int64("count", count),
		zap.Duration("count_host_duration", countHostDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return count, nil
}

// NewSSHClient 创建SSH客户端连接
//
// 参数：
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	sshIP: SSH服务器IP地址
//	sshPort: SSH服务器端口
//	sshUser: SSH用户名
//	sshAuths: SSH认证方法列表
//	timeout: 连接超时时间
//
// 返回值：
//
//	*ssh.Client: SSH客户端连接
//	error: 操作错误信息，成功则返回nil
//
// 功能：
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
		"创建ssh连接：开始执行",
		zap.String("ssh_ip", sshIP),
		zap.Uint16("ssh_port", sshPort),
		zap.String("ssh_user", sshUser),
	)
	connectsshStartTime := time.Now()
	client, err := shell.NewSSHClient(ctx, sshIP, sshPort, sshUser, sshAuths, false, timeout)
	connectsshDuration := time.Since(connectsshStartTime)
	if err != nil {
		log.Error(
			"创建ssh连接：ssh连接失败",
			zap.Error(err),
			zap.String("ssh_ip", sshIP),
			zap.Uint16("ssh_port", sshPort),
			zap.String("ssh_user", sshUser),
			zap.Duration("ssh_connect_duration", connectsshDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return nil, errors.WrapIf(err, "创建ssh连接：ssh连接失败")
	}
	log.Debug(
		"创建ssh连接：执行成功",
		zap.String("ssh_ip", sshIP),
		zap.Uint16("ssh_port", sshPort),
		zap.String("ssh_user", sshUser),
		zap.Duration("ssh_connect_duration", connectsshDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)
	return client, nil
}

// ExecuteCommand 执行SSH命令
//
// 参数：
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	session: SSH会话
//	command: 要执行的命令
//
// 返回值：
//
//	error: 操作错误信息，成功则返回nil
//
// 功能：
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
		"执行命令：开始执行",
		zap.String("command", command),
	)

	executeCommandStartTime := time.Now()
	err := session.Run(command)
	executeCommandDuration := time.Since(executeCommandStartTime)
	if err != nil {
		log.Error(
			"执行命令：命令执行失败",
			zap.Error(err),
			zap.String("command", command),
			zap.Duration("execute_duration", executeCommandDuration),
			zap.Duration("total_duration", time.Since(startTime)),
		)
		return errors.WrapIf(err, "执行命令：命令执行失败")
	}

	log.Debug(
		"执行命令：执行成功",
		zap.String("command", command),
		zap.Duration("execute_duration", executeCommandDuration),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	return nil
}
