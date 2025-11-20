package data

import (
	"context"
	goerrors "errors"
	"time"

	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"gin-artweb/internal/customer/biz"
	"gin-artweb/pkg/common"
	"gin-artweb/pkg/database"
	"gin-artweb/pkg/errors"
	"gin-artweb/pkg/log"
)

type recordRepo struct {
	log      *zap.Logger
	gormDB   *gorm.DB
	timeouts *database.DBTimeout
	cache    *cache.Cache
	maxNum   int
	ttl      time.Duration
}

func NewRecordRepo(
	log *zap.Logger,
	gormDB *gorm.DB,
	timeouts *database.DBTimeout,
	lockTime time.Duration,
	clearTime time.Duration,
	num int,
) biz.RecordRepo {
	return &recordRepo{
		log:      log,
		gormDB:   gormDB,
		timeouts: timeouts,
		cache:    cache.New(lockTime, clearTime),
		maxNum:   num,
		ttl:      lockTime,
	}
}

func (r *recordRepo) CreateModel(ctx context.Context, m *biz.LoginRecordModel) error {
	// 检查参数
	if m == nil {
		return goerrors.New("创建登录记录模型失败: 登录记录模型不能为空")
	}

	r.log.Debug(
		"开始创建登录记录模型",
		zap.Object(database.ModelKey, m),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)
	now := time.Now()
	m.LoginAt = now
	if err := database.DBCreate(ctx, r.gormDB, &biz.LoginRecordModel{}, m); err != nil {
		r.log.Error(
			"新增登陆记录模型失败",
			zap.Object(database.ModelKey, m),
			zap.Error(err),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return err
	}
	r.log.Debug(
		"创建登录记录模型成功",
		zap.Object(database.ModelKey, m),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(now)),
	)
	return nil
}

func (r *recordRepo) ListModel(
	ctx context.Context,
	qp database.QueryParams,
) (int64, *[]biz.LoginRecordModel, error) {
	r.log.Debug(
		"开始查询登录记录模型列表",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)
	now := time.Now()
	var ms []biz.LoginRecordModel
	count, err := database.DBList(ctx, r.gormDB, &biz.LoginRecordModel{}, &ms, qp)
	if err != nil {
		r.log.Error(
			"查询登陆记录列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
			zap.Duration(log.DurationKey, time.Since(now)),
		)
		return 0, nil, err
	}
	r.log.Debug(
		"查询登录记录模型列表成功",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		zap.Duration(log.DurationKey, time.Since(now)),
	)
	return count, &ms, nil
}

func (r *recordRepo) GetLoginFailNum(ctx context.Context, ip string) (int, error) {
	// 检查上下文
	if err := errors.CheckContext(ctx); err != nil {
		return 0, err
	}

	// 检查参数
	if ip == "" {
		return 0, goerrors.New("获取登录失败次数失败: IP地址不能为空")
	}

	r.log.Debug(
		"开始获取登录失败次数",
		zap.String("ip", ip),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	// 获取缓存的登陆失败次数，不存在返回允许失败次数的最大值
	num, exists := r.cache.Get(ip)
	if !exists {
		r.log.Debug(
			"未找到IP的登录失败记录, 返回最大允许失败次数",
			zap.String("ip", ip),
			zap.Int("max_fail_num", r.maxNum),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return r.maxNum, nil
	}

	n, _ := num.(int)
	r.log.Debug(
		"获取到IP的登录失败次数",
		zap.String("ip", ip),
		zap.Int("fail_num", n),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)
	return n, nil
}

func (r *recordRepo) SetLoginFailNum(ctx context.Context, ip string, num int) error {
	// 检查上下文
	if err := errors.CheckContext(ctx); err != nil {
		return err
	}

	// 检查参数
	if ip == "" {
		return goerrors.New("设置登录失败次数失败: IP地址不能为空")
	}

	r.log.Debug(
		"开始设置登录失败次数",
		zap.String("ip", ip),
		zap.Int("fail_num", num),
		zap.Duration("ttl", r.ttl),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	// 设置缓存的登陆失败次数
	r.cache.Set(ip, num, r.ttl)

	r.log.Debug(
		"设置登录失败次数成功",
		zap.String("ip", ip),
		zap.Int("fail_num", num),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)
	return nil
}
