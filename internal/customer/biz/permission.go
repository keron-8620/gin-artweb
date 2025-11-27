package biz

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"gin-artweb/internal/shared/common"
	"gin-artweb/internal/shared/database"
	"gin-artweb/internal/shared/errors"
)

const (
	PermissionIDKey  = "permission_id"
	PermissionIDsKey = "permission_ids"
)

type PermissionModel struct {
	database.StandardModel
	URL    string `gorm:"column:url;type:varchar(150);not null;index:idx_permission_url_method_label;comment:HTTP的URL地址" json:"url"`
	Method string `gorm:"column:method;type:varchar(10);not null;index:idx_permission_url_method_label;comment:请求方法" json:"method"`
	Label  string `gorm:"column:label;type:varchar(50);not null;index:idx_permission_url_method_label;comment:标签" json:"label"`
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

func ListPermissionModelToUint32s(pms *[]PermissionModel) []uint32 {
	if pms == nil {
		return []uint32{}
	}
	ms := *pms
	if len(ms) == 0 {
		return []uint32{}
	}

	ids := make([]uint32, len(ms))
	for i, m := range ms {
		ids[i] = m.ID
	}
	return ids
}

type PermissionRepo interface {
	CreateModel(context.Context, *PermissionModel) error
	UpdateModel(context.Context, map[string]any, ...any) error
	DeleteModel(context.Context, ...any) error
	FindModel(context.Context, ...any) (*PermissionModel, error)
	ListModel(context.Context, database.QueryParams) (int64, *[]PermissionModel, error)
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
	if err := errors.CheckContext(ctx); err != nil {
		return nil, errors.FromError(err)
	}

	uc.log.Info(
		"开始创建权限",
		zap.Object(database.ModelKey, &m),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	if err := uc.permRepo.CreateModel(ctx, &m); err != nil {
		uc.log.Error(
			"创建权限失败",
			zap.Error(err),
			zap.Object(database.ModelKey, &m),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return nil, database.NewGormError(err, nil)
	}

	if err := uc.permRepo.AddPolicy(ctx, m); err != nil {
		uc.log.Error(
			"添加权限策略失败",
			zap.Error(err),
			zap.Object(database.ModelKey, &m),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return nil, ErrAddPolicy.WithCause(err)
	}

	uc.log.Info(
		"权限创建成功",
		zap.Object(database.ModelKey, &m),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)
	return &m, nil
}

func (uc *PermissionUsecase) UpdatePermissionByID(
	ctx context.Context,
	permID uint32,
	data map[string]any,
) *errors.Error {
	if err := errors.CheckContext(ctx); err != nil {
		return errors.FromError(err)
	}

	uc.log.Info(
		"开始更新权限",
		zap.Uint32(PermissionIDKey, permID),
		zap.Any(database.UpdateDataKey, data),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	if err := uc.permRepo.UpdateModel(ctx, data, "id = ?", permID); err != nil {
		uc.log.Error(
			"更新权限失败",
			zap.Error(err),
			zap.Uint32(PermissionIDKey, permID),
			zap.Any(database.UpdateDataKey, data),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return database.NewGormError(err, data)
	}

	m, rErr := uc.FindPermissionByID(ctx, permID)
	if rErr != nil {
		uc.log.Error(
			"查询更新后的权限失败",
			zap.Error(rErr),
			zap.Uint32(PermissionIDKey, permID),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return rErr
	}

	if err := uc.permRepo.RemovePolicy(ctx, *m, false); err != nil {
		uc.log.Error(
			"移除旧权限策略失败",
			zap.Error(err),
			zap.Uint32(PermissionIDKey, permID),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return ErrRemovePolicy.WithCause(err)
	}

	if err := uc.permRepo.AddPolicy(ctx, *m); err != nil {
		uc.log.Error(
			"添加新权限策略失败",
			zap.Error(err),
			zap.Uint32(PermissionIDKey, permID),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return ErrAddPolicy.WithCause(err)
	}

	uc.log.Info(
		"权限更新成功",
		zap.Uint32(PermissionIDKey, permID),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)
	return nil
}

func (uc *PermissionUsecase) DeletePermissionByID(
	ctx context.Context,
	permID uint32,
) *errors.Error {
	if err := errors.CheckContext(ctx); err != nil {
		return errors.FromError(err)
	}

	uc.log.Info(
		"开始删除权限",
		zap.Uint32(PermissionIDKey, permID),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	m, rErr := uc.FindPermissionByID(ctx, permID)
	if rErr != nil {
		uc.log.Error(
			"查询待删除权限失败",
			zap.Error(rErr),
			zap.Uint32(PermissionIDKey, permID),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return rErr
	}

	if err := uc.permRepo.DeleteModel(ctx, permID); err != nil {
		uc.log.Error(
			"删除权限失败",
			zap.Error(err),
			zap.Uint32(PermissionIDKey, permID),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return database.NewGormError(err, map[string]any{"id": permID})
	}

	if err := uc.permRepo.RemovePolicy(ctx, *m, true); err != nil {
		uc.log.Error(
			"移除权限策略失败",
			zap.Error(err),
			zap.Uint32(PermissionIDKey, permID),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return ErrRemovePolicy.WithCause(err)
	}

	uc.log.Info(
		"权限删除成功",
		zap.Uint32(PermissionIDKey, permID),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)
	return nil
}

func (uc *PermissionUsecase) FindPermissionByID(
	ctx context.Context,
	permID uint32,
) (*PermissionModel, *errors.Error) {
	if err := errors.CheckContext(ctx); err != nil {
		return nil, errors.FromError(err)
	}

	uc.log.Info(
		"开始查询权限",
		zap.Uint32(PermissionIDKey, permID),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	m, err := uc.permRepo.FindModel(ctx, permID)
	if err != nil {
		uc.log.Error(
			"查询权限失败",
			zap.Error(err),
			zap.Uint32(PermissionIDKey, permID),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return nil, database.NewGormError(err, map[string]any{"id": permID})
	}

	uc.log.Info(
		"查询权限成功",
		zap.Uint32(PermissionIDKey, permID),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)
	return m, nil
}

func (uc *PermissionUsecase) ListPermission(
	ctx context.Context,
	qp database.QueryParams,
) (int64, *[]PermissionModel, *errors.Error) {
	if err := errors.CheckContext(ctx); err != nil {
		return 0, nil, errors.FromError(err)
	}

	uc.log.Info(
		"开始查询权限列表",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	count, ms, err := uc.permRepo.ListModel(ctx, qp)
	if err != nil {
		uc.log.Error(
			"查询权限列表失败",
			zap.Error(err),
			zap.Object(database.QueryParamsKey, &qp),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return 0, nil, database.NewGormError(err, nil)
	}

	uc.log.Info(
		"查询权限列表成功",
		zap.Object(database.QueryParamsKey, &qp),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)
	return count, ms, nil
}

func (uc *PermissionUsecase) LoadPermissionPolicy(ctx context.Context) *errors.Error {
	if err := errors.CheckContext(ctx); err != nil {
		return errors.FromError(err)
	}

	uc.log.Info(
		"开始加载权限策略",
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)

	qp := database.QueryParams{
		Columns: []string{"id", "url", "method"},
	}

	_, pms, err := uc.ListPermission(ctx, qp)
	if err != nil {
		uc.log.Error(
			"加载权限策略时查询权限列表失败",
			zap.Error(err),
			zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
		)
		return err
	}

	var policyCount int
	if pms != nil {
		ms := *pms
		policyCount = len(ms)
		for i := range ms {
			if err := uc.permRepo.AddPolicy(ctx, ms[i]); err != nil {
				uc.log.Error(
					"加载权限策略失败",
					zap.Error(err),
					zap.Uint32(PermissionIDKey, ms[i].ID),
					zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
				)
				return ErrAddPolicy.WithCause(err)
			}
		}
	}

	uc.log.Info(
		"权限策略加载成功",
		zap.Int("policy_count", policyCount),
		zap.String(common.TraceIDKey, common.GetTraceID(ctx)),
	)
	return nil
}
