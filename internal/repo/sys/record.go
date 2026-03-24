package sys

import (
	"context"
	"time"

	"emperror.dev/errors"
	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"
	"gorm.io/gorm"

	sysmodel "gin-artweb/internal/model/sys"
	"gin-artweb/internal/shared/config"
	"gin-artweb/internal/shared/ctxutil"
	"gin-artweb/internal/shared/database"
)

// LoginRecordRepo 登录记录仓库实现
// 负责登录记录的CRUD操作和登录失败次数的缓存管理
// 使用GORM进行数据库操作，使用cache进行登录失败次数的缓存

type LoginRecordRepo struct {
	log      *zap.Logger       // 日志记录器
	gormDB   *gorm.DB          // GORM数据库连接
	timeouts *config.DBTimeout // 数据库操作超时配置
	cache    *cache.Cache      // 缓存，用于存储登录失败次数
	maxNum   int               // 最大允许的登录失败次数
	ttl      time.Duration     // 缓存过期时间
}

// NewLoginRecordRepo 创建登录记录仓库实例
//
// 参数：
//
//	log: 日志记录器，用于记录操作日志
//	gormDB: GORM数据库连接，用于执行数据库操作
//	timeouts: 数据库操作超时配置，控制各类数据库操作的超时时间
//	lockTime: 缓存过期时间
//	clearTime: 缓存清理时间
//	num: 最大允许的登录失败次数
//
// 返回值：
//
//	sysmodel.LoginRecordRepo: 登录记录仓库接口实现
func NewLoginRecordRepo(
	log *zap.Logger,
	gormDB *gorm.DB,
	timeouts *config.DBTimeout,
	lockTime time.Duration,
	clearTime time.Duration,
	num int,
) *LoginRecordRepo {
	return &LoginRecordRepo{
		log:      log,
		gormDB:   gormDB,
		timeouts: timeouts,
		cache:    cache.New(lockTime, clearTime),
		maxNum:   num,
		ttl:      lockTime,
	}
}

// CreateModel 创建登录记录模型
//
// 参数：
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	m: 登录记录模型，包含登录记录的详细信息
//
// 返回值：
//
//	error: 操作错误信息，成功则返回nil
//
// 功能：
//  1. 检查登录记录模型是否为空
//  2. 设置登录时间
//  3. 执行数据库创建操作
//  4. 记录操作日志
func (r *LoginRecordRepo) CreateModel(ctx context.Context, m *sysmodel.LoginRecordModel) error {
	log := r.log.With(zap.String("trace_id", ctxutil.GetTraceID(ctx)))

	// 检查参数
	if m == nil {
		err := errors.New("创建登录记录模型：模型为空")
		log.Error(
			"创建登录记录模型：模型为空",
			zap.Error(err),
		)
		return err
	}

	log.Debug(
		"创建登录记录模型：开始执行",
		zap.Object("login_record_model", m),
	)
	now := time.Now()
	m.LoginAt = now
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.WriteTimeout)
	defer cancel()
	if err := database.DBCreate(dbCtx, r.gormDB, &sysmodel.LoginRecordModel{}, m, nil); err != nil {
		log.Error(
			"创建登录记录模型：数据库创建失败",
			zap.Object("login_record_model", m),
			zap.Error(err),
			zap.Duration("create_login_record_duration", time.Since(now)),
		)
		return errors.WrapIf(err, "创建登录记录模型：数据库创建失败")
	}
	log.Debug(
		"创建登录记录模型：执行成功",
		zap.Object("login_record_model", m),
		zap.Duration("create_login_record_duration", time.Since(now)),
	)
	return nil
}

// ListModel 查询登录记录模型列表
//
// 参数：
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	qp: 查询参数，包含分页、排序等查询条件
//
// 返回值：
//
//	int64: 总记录数
//	[]sysmodel.LoginRecordModel: 登录记录模型列表指针，包含符合条件的登录记录
//	error: 操作错误信息，成功则返回nil
//
// 功能：
//  1. 执行数据库查询操作
//  2. 获取登录记录模型列表
//  3. 返回总记录数和模型列表
//  4. 记录操作日志
func (r *LoginRecordRepo) ListModel(
	ctx context.Context,
	qp database.QueryParams,
) ([]sysmodel.LoginRecordModel, error) {
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"查询登录记录模型列表：开始执行",
		zap.Object("query_params", &qp),
	)
	now := time.Now()
	var ms []sysmodel.LoginRecordModel
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()
	if err := database.DBList(dbCtx, r.gormDB, &sysmodel.LoginRecordModel{}, &ms, qp); err != nil {
		log.Error(
			"查询登录记录模型列表：数据库查询失败",
			zap.Error(err),
			zap.Object("query_params", &qp),
			zap.Duration("list_login_record_duration", time.Since(now)),
		)
		return nil, errors.WrapIf(err, "查询登录记录模型列表：数据库查询失败")
	}
	log.Debug(
		"查询登录记录模型列表：执行成功",
		zap.Object("query_params", &qp),
		zap.Duration("list_login_record_duration", time.Since(now)),
	)
	return ms, nil
}

func (r *LoginRecordRepo) CountModel(
	ctx context.Context,
	query map[string]any,
) (int64, error) {
	log := ctxutil.NewLogger(r.log, ctx)

	log.Debug(
		"查询登录记录模型总数：开始执行",
		zap.Any("query", query),
	)
	now := time.Now()
	dbCtx, cancel := context.WithTimeout(ctx, r.timeouts.ReadTimeout)
	defer cancel()
	count, err := database.DBCount(dbCtx, r.gormDB, &sysmodel.LoginRecordModel{}, query)
	if err != nil {
		log.Error(
			"查询登录记录模型总数：数据库查询失败",
			zap.Error(err),
			zap.Any("query", query),
			zap.Duration("count_login_record_duration", time.Since(now)),
		)
		return 0, errors.WrapIf(err, "查询登录记录模型总数：数据库查询失败")
	}
	log.Debug(
		"查询登录记录模型总数：执行成功",
		zap.Any("query", query),
		zap.Int64("count", count),
		zap.Duration("count_login_record_duration", time.Since(now)),
	)
	return count, nil
}

// GetLoginFailNum 获取登录失败次数
//
// 参数：
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	ip: IP地址，用于标识登录失败的客户端
//
// 返回值：
//
//	int: 剩余的登录失败次数（未找到记录时返回最大允许失败次数）
//	error: 操作错误信息，成功则返回nil
//
// 功能：
//  1. 检查上下文是否有效
//  2. 检查IP地址是否为空
//  3. 从缓存中获取登录失败次数
//  4. 未找到记录时返回最大允许失败次数
//  5. 记录操作日志
func (r *LoginRecordRepo) GetLoginFailNum(ctx context.Context, ip string) (int, error) {
	// 检查上下文
	if ctx.Err() != nil {
		return 0, errors.WrapIf(ctx.Err(), "获取登录失败次数：上下文错误")
	}

	// 检查参数
	if ip == "" {
		return 0, errors.New("获取登录失败次数: IP地址不能为空")
	}

	log := r.log.With(zap.String("trace_id", ctxutil.GetTraceID(ctx)))

	log.Debug(
		"获取登录失败次数：开始执行",
		zap.String("ip", ip),
	)

	// 获取缓存的登录失败次数，不存在返回允许失败次数的最大值
	num, exists := r.cache.Get(ip)
	if !exists {
		log.Debug(
			"获取登录失败次数：未找到IP的登录失败记录, 返回最大允许失败次数",
			zap.String("ip", ip),
			zap.Int("max_fail_num", r.maxNum),
		)
		return r.maxNum, nil
	}

	n, _ := num.(int)
	log.Debug(
		"获取登录失败次数：获取到IP的登录失败次数",
		zap.String("ip", ip),
		zap.Int("fail_num", n),
	)
	return n, nil
}

// SetLoginFailNum 设置登录失败次数
//
// 参数：
//
//	ctx: 上下文，用于传递请求信息和控制超时
//	ip: IP地址，用于标识登录失败的客户端
//	num: 登录失败次数
//
// 返回值：
//
//	error: 操作错误信息，成功则返回nil
//
// 功能：
//  1. 检查上下文是否有效
//  2. 检查IP地址是否为空
//  3. 将登录失败次数设置到缓存中
//  4. 记录操作日志
func (r *LoginRecordRepo) SetLoginFailNum(ctx context.Context, ip string, num int) error {
	// 检查上下文
	if ctx.Err() != nil {
		return errors.WrapIf(ctx.Err(), "设置登录失败次数：上下文错误")
	}

	// 检查参数
	if ip == "" {
		return errors.New("设置登录失败次数: IP地址不能为空")
	}

	log := r.log.With(zap.String("trace_id", ctxutil.GetTraceID(ctx)))

	log.Debug(
		"设置登录失败次数：开始执行",
		zap.String("ip", ip),
		zap.Int("fail_num", num),
		zap.Duration("ttl", r.ttl),
	)

	// 设置缓存的登录失败次数
	r.cache.Set(ip, num, r.ttl)

	log.Debug(
		"设置登录失败次数：执行成功",
		zap.String("ip", ip),
		zap.Int("fail_num", num),
	)
	return nil
}
