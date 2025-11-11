package data

import (
	"context"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"gin-artweb/internal/resource/biz"
	"gin-artweb/pkg/database"
)

type packageRepo struct {
	log    *zap.Logger
	gormDB *gorm.DB
}

func NewpackageRepo(
	log *zap.Logger,
	gormDB *gorm.DB,
) biz.PackageRepo {
	return &packageRepo{
		log:    log,
		gormDB: gormDB,
	}
}

func (r *packageRepo) CreateModel(ctx context.Context, m *biz.PackageModel) error {
	m.UploadedAt = time.Now()
	if err := database.DBCreate(ctx, r.gormDB, &biz.PackageModel{}, m); err != nil {
		r.log.Error(
			"新增程序包模型失败",
			zap.Object("model", m),
			zap.Error(err),
		)
		return err
	}
	return nil
}

func (r *packageRepo) DeleteModel(ctx context.Context, conds ...any) error {
	if err := database.DBDelete(ctx, r.gormDB, &biz.PackageModel{}, conds...); err != nil {
		r.log.Error(
			"删除程序包模型失败",
			zap.Any("conditions", conds),
			zap.Error(err),
		)
		return err
	}
	return nil
}

func (r *packageRepo) FindModel(
	ctx context.Context,
	preloads []string,
	conds ...any,
) (*biz.PackageModel, error) {
	var m biz.PackageModel
	if err := database.DBFind(ctx, r.gormDB, preloads, &m, conds...); err != nil {
		r.log.Error(
			"查询程序包模型失败",
			zap.Strings("preloads", preloads),
			zap.Any("conditions", conds),
			zap.Error(err),
		)
		return nil, err
	}
	return &m, nil
}

func (r *packageRepo) ListModel(
	ctx context.Context,
	qp database.QueryParams,
) (int64, []biz.PackageModel, error) {
	var ms []biz.PackageModel
	count, err := database.DBList(ctx, r.gormDB, &biz.PackageModel{}, &ms, qp)
	if err != nil {
		r.log.Error(
			"查询程序包列表失败",
			zap.Object("query_params", &qp),
			zap.Error(err),
		)
		return 0, nil, err
	}
	return count, ms, nil
}

