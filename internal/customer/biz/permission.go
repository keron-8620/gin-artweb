package biz

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"gin-artweb/pkg/database"
	"gin-artweb/pkg/errors"
)

type PermissionModel struct {
	database.StandardModel
	URL    string `gorm:"column:url;type:varchar(150);index:idx_permission_url_method_label;comment:HTTP的URL地址" json:"url"`
	Method string `gorm:"column:method;type:varchar(10);index:idx_permission_url_method_label;comment:请求方法" json:"method"`
	Label  string `gorm:"column:label;type:varchar(50);index:idx_permission_url_method_label;comment:标签" json:"label"`
	Descr  string `gorm:"column:descr;type:varchar(254);comment:描述" json:"descr"`
}

func (m *PermissionModel) TableName() string {
	return "customer_permission"
}

func (m *PermissionModel) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if err := m.StandardModel.MarshalLogObject(enc); err != nil {
		return err
	}
	enc.AddString("url", m.URL)
	enc.AddString("method", m.Method)
	enc.AddString("label", m.Label)
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
	RemovePolicy(context.Context, PermissionModel, bool) error
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
		return nil, database.NewGormError(err, nil)
	}
	if err := uc.permRepo.AddPolicy(ctx, m); err != nil {
		return nil, ErrAddPolicy.WithCause(err)
	}
	return &m, nil
}

func (uc *PermissionUsecase) UpdatePermissionByID(
	ctx context.Context,
	permID uint32,
	data map[string]any,
) *errors.Error {
	if err := uc.permRepo.UpdateModel(ctx, data, "id = ?", permID); err != nil {
		return database.NewGormError(err, data)
	}
	m, rErr := uc.FindPermissionByID(ctx, permID)
	if rErr != nil {
		return rErr
	}
	if err := uc.permRepo.RemovePolicy(ctx, *m, false); err != nil {
		return ErrRemovePolicy.WithCause(err)
	}
	if err := uc.permRepo.AddPolicy(ctx, *m); err != nil {
		return ErrAddPolicy.WithCause(err)
	}
	return nil
}

func (uc *PermissionUsecase) DeletePermissionByID(
	ctx context.Context,
	permID uint32,
) *errors.Error {
	m, rErr := uc.FindPermissionByID(ctx, permID)
	if rErr != nil {
		return rErr
	}
	if err := uc.permRepo.DeleteModel(ctx, permID); err != nil {
		return database.NewGormError(err, map[string]any{"id": permID})
	}
	if err := uc.permRepo.RemovePolicy(ctx, *m, true); err != nil {
		return ErrRemovePolicy.WithCause(err)
	}
	return nil
}

func (uc *PermissionUsecase) FindPermissionByID(
	ctx context.Context,
	permID uint32,
) (*PermissionModel, *errors.Error) {
	m, err := uc.permRepo.FindModel(ctx, nil, permID)
	if err != nil {
		return nil, database.NewGormError(err, map[string]any{"id": permID})
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
		return 0, nil, database.NewGormError(err, nil)
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
			return ErrAddPolicy.WithCause(err)
		}
	}
	return nil
}
