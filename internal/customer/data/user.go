package data

import (
	"context"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"gitee.com/keion8620/go-dango-gin/internal/customer/biz"
	"gitee.com/keion8620/go-dango-gin/pkg/database"
)

type userRepo struct {
	log    *zap.Logger
	gormDB *gorm.DB
}

func NewUserRepo(
	log *zap.Logger,
	gormDB *gorm.DB,
) biz.UserRepo {
	return &userRepo{
		log:    log,
		gormDB: gormDB,
	}
}

func (r *userRepo) CreateModel(ctx context.Context, m *biz.UserModel) error {
	now := time.Now()
	m.CreatedAt = now
	m.UpdatedAt = now
	if err := database.DBCreate(ctx, r.gormDB, &biz.UserModel{}, m); err != nil {
		r.log.Error(
			"failed to create user model",
			zap.Object("model", m),
			zap.Error(err),
		)
		return err
	}
	return nil
}

func (r *userRepo) UpdateModel(ctx context.Context, data map[string]any, conds ...any) error {
	if err := database.DBUpdate(ctx, r.gormDB, &biz.UserModel{}, data, nil, conds...); err != nil {
		r.log.Error(
			"failed to update user model",
			zap.Any("data", data),
			zap.Any("conditions", conds),
			zap.Error(err),
		)
		return err
	}
	return nil
}

func (r *userRepo) DeleteModel(ctx context.Context, conds ...any) error {
	if err := database.DBDelete(ctx, r.gormDB, &biz.UserModel{}, conds...); err != nil {
		r.log.Error(
			"failed to delete user model",
			zap.Any("conditions", conds),
			zap.Error(err),
		)
		return err
	}
	return nil
}

func (r *userRepo) FindModel(
	ctx context.Context,
	preloads []string,
	conds ...any,
) (*biz.UserModel, error) {
	var m biz.UserModel
	if err := database.DBFind(ctx, r.gormDB, preloads, &m, conds...); err != nil {
		r.log.Error(
			"failed to find user model",
			zap.Any("conditions", conds),
			zap.Error(err),
		)
		return nil, err
	}
	return &m, nil
}

func (r *userRepo) ListModel(
	ctx context.Context,
	qp database.QueryParams,
) (int64, []biz.UserModel, error) {
	var ms []biz.UserModel
	count, err := database.DBList(ctx, r.gormDB, &biz.UserModel{}, &ms, qp)
	if err != nil {
		r.log.Error(
			"failed to list user model",
			zap.Object("query_params", &qp),
			zap.Error(err),
		)
		return 0, nil, err
	}
	return count, ms, nil
}
