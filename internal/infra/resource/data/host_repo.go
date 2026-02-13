package data

import (
	"context"
	"time"

	"emperror.dev/errors"
	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"
	"gorm.io/gorm"

	"gin-artweb/internal/infra/resource/model"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/log"
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
func (r *HostRepo) CreateModel(ctx context.Context, m *model.HostModel) error {
	// 检查参数
	if m == nil {
		err := errors.New("创建主机模型失败: 模型为空")
		r.log.Error(
			"创建主机模型失败: 模型为空",
			zap.Error(err),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return err
	}
	r.log.Debug(
		"开始创建主机模型",
		zap.Object(database.ModelKey, m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	now := time.Now()
	m.CreatedAt = now
	m.UpdatedAt = now
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	if err := database.DBCreate(dbCtx, r.gormDB, &model.HostModel{}, m, nil); err != nil {
		r.log.Error(
			"创建主机模型失败",
			zap.Error(err),
			zap.Object(database.ModelKey, m),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return errors.WrapIf(err, "创建主机模型失败")
	}
	r.log.Debug(
		"创建主机模型成功",
		zap.Object(database.ModelKey, m),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(now)),
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
func (r *HostRepo) UpdateModel(ctx context.Context, data map[string]any, conds ...any) error {
	if len(data) == 0 {
		err := errors.New("更新主机模型失败: 更新数据为空")
		r.log.Error(
			"更新主机模型失败: 更新数据为空",
			zap.Error(err),
			zap.Any(database.UpdateDataKey, data),
			zap.Any(database.ConditionsKey, conds),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		)
		return err
	}
	r.log.Debug(
		"开始更新主机模型",
		zap.Any(database.UpdateDataKey, data),
		zap.Any(database.ConditionsKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	now := time.Now()
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	if err := database.DBUpdate(dbCtx, r.gormDB, &model.HostModel{}, data, nil, conds...); err != nil {
		r.log.Error(
			"更新主机模型失败",
			zap.Error(err),
			zap.Any(database.UpdateDataKey, data),
			zap.Any(database.ConditionsKey, conds),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return errors.WrapIf(err, "更新主机模型失败")
	}
	r.log.Debug(
		"更新主机模型成功",
		zap.Any(database.UpdateDataKey, data),
		zap.Any(database.ConditionsKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(now)),
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
func (r *HostRepo) DeleteModel(ctx context.Context, conds ...any) error {
	r.log.Debug(
		"开始删除主机模型",
		zap.Any(database.ConditionsKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	now := time.Now()
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	if err := database.DBDelete(dbCtx, r.gormDB, &model.HostModel{}, conds...); err != nil {
		r.log.Error(
			"删除主机模型失败",
			zap.Error(err),
			zap.Any(database.ConditionsKey, conds),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return errors.WrapIf(err, "删除主机模型失败")
	}
	r.log.Debug(
		"删除主机模型成功",
		zap.Any(database.ConditionsKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(now)),
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
//	*model.HostModel: 主机模型指针，包含主机的详细信息
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
) (*model.HostModel, error) {
	r.log.Debug(
		"开始查询主机模型",
		zap.Any(database.ConditionsKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	now := time.Now()
	var m model.HostModel
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()
	if err := database.DBGet(dbCtx, r.gormDB, preloads, &m, conds...); err != nil {
		r.log.Error(
			"查询主机模型失败",
			zap.Error(err),
			zap.Any(database.ConditionsKey, conds),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return nil, errors.WrapIf(err, "查询主机模型失败")
	}
	r.log.Debug(
		"查询主机模型成功",
		zap.Object(database.ModelKey, &m),
		zap.Any(database.ConditionsKey, conds),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(now)),
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
//	*[]model.HostModel: 主机模型列表指针，包含符合条件的主机模型
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
) (int64, *[]model.HostModel, error) {
	r.log.Debug(
		"开始查询主机模型列表",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)
	now := time.Now()
	var ms []model.HostModel
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ListTimeout)
	defer cancel()
	count, err := database.DBList(dbCtx, r.gormDB, &model.HostModel{}, &ms, qp)
	if err != nil {
		r.log.Error(
			"查询主机模型列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return 0, nil, errors.WrapIf(err, "查询主机模型列表失败")
	}
	r.log.Debug(
		"查询主机模型列表成功",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(now)),
	)
	return count, &ms, nil
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
	if ctx.Err() != nil {
		return nil, errors.WrapIf(ctx.Err(), "上下文已取消")
	}
	r.log.Debug(
		"开始创建ssh连接",
		zap.String("ssh_ip", sshIP),
		zap.Uint16("ssh_port", sshPort),
		zap.String("ssh_user", sshUser),
	)
	now := time.Now()
	client, err := shell.NewSSHClient(ctx, sshIP, sshPort, sshUser, sshAuths, false, timeout)
	if err != nil {
		r.log.Error(
			"创建ssh连接失败",
			zap.Error(err),
			zap.String("ssh_ip", sshIP),
			zap.Uint16("ssh_port", sshPort),
			zap.String("ssh_user", sshUser),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return nil, errors.WrapIf(err, "创建ssh连接失败")
	}
	r.log.Debug(
		"创建ssh连接成功",
		zap.String("ssh_ip", sshIP),
		zap.Uint16("ssh_port", sshPort),
		zap.String("ssh_user", sshUser),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(now)),
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
	if ctx.Err() != nil {
		return errors.WrapIf(ctx.Err(), "上下文已取消")
	}

	r.log.Debug(
		"开始执行命令",
		zap.String("command", command),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
	)

	now := time.Now()
	if err := session.Run(command); err != nil {
		r.log.Error(
			"执行命令失败",
			zap.Error(err),
			zap.String("command", command),
			zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return errors.WrapIf(err, "执行命令失败")
	}

	r.log.Debug(
		"执行命令成功",
		zap.String("command", command),
		zap.String(ctxutil.TraceIDKey, ctxutil.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(now)),
	)

	return nil
}
