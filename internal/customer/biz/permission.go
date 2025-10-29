package biz

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"gitee.com/keion8620/go-dango-gin/pkg/database"
	"gitee.com/keion8620/go-dango-gin/pkg/errors"
)

type PermissionModel struct {
	database.StandardModel
	HttpUrl string `gorm:"column:http_url;type:varchar(150);index:idx_member;comment:HTTP的URL地址" json:"http_url"`
	Method  string `gorm:"column:method;type:varchar(10);index:idx_member;comment:请求方法" json:"method"`
	Descr   string `gorm:"column:descr;type:varchar(254);comment:描述" json:"descr"`
}

func (m *PermissionModel) TableName() string {
	return "customer_permission"
}

func (m *PermissionModel) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if err := m.StandardModel.MarshalLogObject(enc); err != nil {
		return err
	}
	enc.AddString("http_url", m.HttpUrl)
	enc.AddString("method", m.Method)
	enc.AddString("descr", m.Descr)
	return nil
}

type PermissionRepo interface {
	CreateModel(context.Context, *PermissionModel) error
	UpdateModel(context.Context, map[string]any, ...any) error
	DeleteModel(context.Context, ...any) error
	FindModel(context.Context, []string, ...any) (*PermissionModel, error)
	ListModel(context.Context, database.QueryParams) (int64, []PermissionModel, error)
	AddPolicy(context.Context, PermissionModel) error
	RemovePolicy(context.Context, PermissionModel) error
}

type PermissionUsecase struct {
	log      *zap.Logger
	permRepo PermissionRepo
}

func NewPermissionUsecase(
	log *zap.Logger,
	permRepo PermissionRepo,
) *PermissionUsecase {
	return &PermissionUsecase{
		log:      log,
		permRepo: permRepo,
	}
}

func (uc *PermissionUsecase) CreatePermission(
	ctx context.Context,
	m PermissionModel,
) (*PermissionModel, *errors.Error) {
	if err := uc.permRepo.CreateModel(ctx, &m); err != nil {
		rErr := database.NewGormError(err, nil)
		uc.log.Error(rErr.Error())
		return nil, rErr
	}
	if err := uc.permRepo.AddPolicy(ctx, m); err != nil {
		rErr := ErrAddPolicy.WithCause(err)
		uc.log.Error(rErr.Error())
		return nil, rErr
	}
	return &m, nil
}

func (uc *PermissionUsecase) UpdatePermissionById(
	ctx context.Context,
	permId uint,
	data map[string]any,
) *errors.Error {
	if err := uc.permRepo.UpdateModel(ctx, data, "id = ?", permId); err != nil {
		rErr := database.NewGormError(err, data)
		uc.log.Error(rErr.Error())
		return rErr
	}
	m, rErr := uc.FindPermissionById(ctx, permId)
	if rErr != nil {
		return rErr
	}
	if err := uc.permRepo.RemovePolicy(ctx, *m); err != nil {
		rErr := ErrRemovePolicy.WithCause(err)
		uc.log.Error(rErr.Error())
		return rErr
	}
	if err := uc.permRepo.AddPolicy(ctx, *m); err != nil {
		rErr := ErrAddPolicy.WithCause(err)
		uc.log.Error(rErr.Error())
		return rErr
	}
	return nil
}

func (uc *PermissionUsecase) DeletePermissionById(
	ctx context.Context,
	permId uint,
) *errors.Error {
	m, rErr := uc.FindPermissionById(ctx, permId)
	if rErr != nil {
		return rErr
	}
	if err := uc.permRepo.DeleteModel(ctx, permId); err != nil {
		rErr := database.NewGormError(err, map[string]any{"id": permId})
		uc.log.Error(rErr.Error())
		return rErr
	}
	if err := uc.permRepo.RemovePolicy(ctx, *m); err != nil {
		rErr := ErrRemovePolicy.WithCause(err)
		uc.log.Error(rErr.Error())
		return rErr
	}
	return nil
}

func (uc *PermissionUsecase) FindPermissionById(
	ctx context.Context,
	permId uint,
) (*PermissionModel, *errors.Error) {
	m, err := uc.permRepo.FindModel(ctx, nil, permId)
	if err != nil {
		rErr := database.NewGormError(err, map[string]any{"id": permId})
		uc.log.Error(rErr.Error())
		return nil, rErr
	}
	return m, nil
}

func (uc *PermissionUsecase) ListPermission(
	ctx context.Context,
	page, size int,
	query map[string]any,
	orderBy []string,
	isCount bool,
) (int64, []PermissionModel, *errors.Error) {
	qp := database.QueryParams{
		Preloads: []string{},
		Query:    query,
		OrderBy:  orderBy,
		Limit:    max(size, 0),
		Offset:   max(page-1, 0),
		IsCount:  isCount,
	}
	count, ms, err := uc.permRepo.ListModel(ctx, qp)
	if err != nil {
		rErr := database.NewGormError(err, nil)
		uc.log.Error(rErr.Error())
		return 0, nil, rErr
	}
	return count, ms, nil
}

func (uc *PermissionUsecase) LoadPermissionPolicy(ctx context.Context) error {
	_, pms, err := uc.ListPermission(ctx, 0, 0, nil, nil, false)
	if err != nil {
		return err
	}
	for _, pm := range pms {
		if err := uc.permRepo.AddPolicy(ctx, pm); err != nil {
			return err
		}
	}
	return nil
}
