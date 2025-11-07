package data

import (
	"context"
	"time"

	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"gin-artweb/internal/customer/biz"
	"gin-artweb/pkg/database"
)

type recordRepo struct {
	log    *zap.Logger
	gormDB *gorm.DB
	cache  *cache.Cache
	maxNum int
	ttl    time.Duration
}

func NewRecordRepo(
	log *zap.Logger,
	gormDB *gorm.DB,
	lockTime time.Duration,
	clearTime time.Duration,
	num int,
) biz.RecordRepo {
	return &recordRepo{
		log:    log,
		gormDB: gormDB,
		cache:  cache.New(lockTime, clearTime),
		maxNum: num,
		ttl:    lockTime,
	}
}

func (r *recordRepo) CreateModel(ctx context.Context, m *biz.LoginRecordModel) error {
	if err := database.DBCreate(ctx, r.gormDB, &biz.LoginRecordModel{}, m); err != nil {
		r.log.Error(
			"新增登陆记录模型失败",
			zap.Object("model", m),
			zap.Error(err),
		)
		return err
	}
	return nil
}

func (r *recordRepo) ListModel(
	ctx context.Context,
	qp database.QueryParams,
) (int64, []biz.LoginRecordModel, error) {
	var ms []biz.LoginRecordModel
	count, err := database.DBList(ctx, r.gormDB, &biz.LoginRecordModel{}, &ms, qp)
	if err != nil {
		r.log.Error(
			"查询登陆记录列表失败",
			zap.Object("query_params", &qp),
			zap.Error(err),
		)
		return 0, nil, err
	}
	return count, ms, nil
}

func (r *recordRepo) GetLoginFailNum(ctx context.Context, ip string) (int, error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
	}
	num, exists := r.cache.Get(ip)
	if !exists {
		return r.maxNum, nil
	}
	n, _ := num.(int)
	return n, nil
}

func (r *recordRepo) SetLoginFailNum(ctx context.Context, ip string, num int) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	r.cache.Set(ip, num, r.ttl)
	return nil
}
